package process

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ensureOpenClawConfig 启动前修复 openclaw.json 的关键兼容项与插件配置。
func (m *Manager) ensureOpenClawConfig() {
	ocDir := m.resolveOpenClawDir()
	cfgPath := filepath.Join(ocDir, "openclaw.json")

	var cfg map[string]interface{}
	created := false
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		_ = os.MkdirAll(filepath.Dir(cfgPath), 0o755)
		cfg = map[string]interface{}{}
		created = true
	} else if err := json.Unmarshal(data, &cfg); err != nil {
		cfg = map[string]interface{}{}
		created = true
	}

	changed := created
	if NormalizeOpenClawConfig(cfg) {
		changed = true
	}

	gw, _ := cfg["gateway"].(map[string]interface{})
	if gw == nil {
		gw = map[string]interface{}{}
		cfg["gateway"] = gw
	}
	if gw["mode"] != "local" {
		gw["mode"] = "local"
		changed = true
	}
	if _, ok := cfg["meta"]; ok {
		delete(cfg, "meta")
		changed = true
	}
	if _, ok := cfg["workspace"]; ok {
		delete(cfg, "workspace")
		changed = true
	}

	qqExtDir := filepath.Join(ocDir, "extensions", "qq")
	qqInstalled := false
	if _, statErr := os.Stat(qqExtDir); statErr == nil {
		qqInstalled = true
	}
	if qqInstalled {
		changed = m.ensureQQPluginConfig(cfg, qqExtDir) || changed
	}

	if changed {
		out, marshalErr := json.MarshalIndent(cfg, "", "  ")
		if marshalErr != nil {
			m.logger.Warn("openclaw.json 序列化失败", "error", marshalErr)
		} else if writeErr := os.WriteFile(cfgPath, out, 0o644); writeErr != nil {
			m.logger.Warn("openclaw.json 写入失败", "error", writeErr)
		} else {
			m.logger.Info("openclaw.json 配置已自动修复")
		}
	}

	qqInstallPath := ""
	if pl, ok := cfg["plugins"].(map[string]interface{}); ok {
		if ins, ok := pl["installs"].(map[string]interface{}); ok {
			if qqIns, ok := ins["qq"].(map[string]interface{}); ok {
				if p, ok := qqIns["installPath"].(string); ok {
					qqInstallPath = p
				}
			}
		}
	}
	m.patchQQPluginChannel(ocDir, qqInstallPath)
}

// ensureQQPluginConfig 确保 QQ channel/plugins 配置项存在且有效。
func (m *Manager) ensureQQPluginConfig(cfg map[string]interface{}, qqExtDir string) bool {
	changed := false

	ch, _ := cfg["channels"].(map[string]interface{})
	if ch == nil {
		ch = map[string]interface{}{}
		cfg["channels"] = ch
	}
	qq, _ := ch["qq"].(map[string]interface{})
	if qq == nil {
		qq = map[string]interface{}{}
		ch["qq"] = qq
	}
	if qq["wsUrl"] == nil || qq["wsUrl"] == "" {
		qq["wsUrl"] = "ws://127.0.0.1:3001"
		changed = true
	}
	if qq["enabled"] == nil {
		qq["enabled"] = true
		changed = true
	}

	pl, _ := cfg["plugins"].(map[string]interface{})
	if pl == nil {
		pl = map[string]interface{}{}
		cfg["plugins"] = pl
	}
	ent, _ := pl["entries"].(map[string]interface{})
	if ent == nil {
		ent = map[string]interface{}{}
		pl["entries"] = ent
	}
	if ent["qq"] == nil {
		ent["qq"] = map[string]interface{}{"enabled": true}
		changed = true
	}
	ins, _ := pl["installs"].(map[string]interface{})
	if ins == nil {
		ins = map[string]interface{}{}
		pl["installs"] = ins
	}
	qqIns, _ := ins["qq"].(map[string]interface{})
	if qqIns == nil {
		qqIns = map[string]interface{}{}
		ins["qq"] = qqIns
	}
	if p, _ := qqIns["installPath"].(string); p == "" {
		qqIns["installPath"] = qqExtDir
		changed = true
	} else if _, statErr := os.Stat(p); statErr != nil {
		qqIns["installPath"] = qqExtDir
		changed = true
	}
	source, _ := qqIns["source"].(string)
	if source != "npm" && source != "archive" && source != "path" {
		qqIns["source"] = "path"
		changed = true
	}
	if qqIns["version"] == nil || qqIns["version"] == "" {
		qqIns["version"] = "latest"
		changed = true
	}

	return changed
}

// patchQQPluginChannel 修复 QQ channel 插件的 Promise 生命周期与日志回传实现。
func (m *Manager) patchQQPluginChannel(ocDir, installPath string) {
	paths := []string{}
	if ocDir != "" {
		paths = append(paths,
			filepath.Join(ocDir, "extensions", "qq", "src", "channel.ts"),
			filepath.Join(ocDir, "extensions", "qq", "dist", "channel.js"),
		)
	}
	if installPath != "" {
		paths = append(paths,
			filepath.Join(installPath, "src", "channel.ts"),
			filepath.Join(installPath, "dist", "channel.js"),
		)
	}
	if m.cfg.OpenClawApp != "" {
		paths = append(paths,
			filepath.Join(m.cfg.OpenClawApp, "extensions", "qq", "src", "channel.ts"),
			filepath.Join(m.cfg.OpenClawApp, "extensions", "qq", "dist", "channel.js"),
		)
	}

	oldPattern := regexp.MustCompile(`(?s)return\s*\(\)\s*=>\s*\{\s*client\.disconnect\(\);\s*clients\.delete\(account\.accountId\);\s*stopFileServer\(\);\s*\};`)
	newReturn := `return new Promise((resolve) => {
        const cleanup = () => {
          client.disconnect();
          clients.delete(account.accountId);
          stopFileServer();
          resolve();
        };
        const abortSignal = (ctx && ctx.abortSignal) ? ctx.abortSignal : undefined;
        if (abortSignal) {
          if (abortSignal.aborted) { cleanup(); return; }
          abortSignal.addEventListener("abort", cleanup, { once: true });
        }
        client.on("close", () => {
          cleanup();
        });
      });`

	loggerPattern := regexp.MustCompile(`(?s)function\s+postLogEntry\s*\([^)]*\)\s*\{.*?\n\}`)
	managerURLPattern := regexp.MustCompile(`const\s+MANAGER_LOG_URL\s*=\s*"[^"]*";`)
	managerPort := 19527
	if m.cfg.ManagerPort > 0 {
		managerPort = m.cfg.ManagerPort
	}
	managerURLLine := fmt.Sprintf(`const MANAGER_LOG_URL = "http://127.0.0.1:%d/api/events/log";`, managerPort)
	loggerReplacement := `function postLogEntry(summary, detail, source) {
  try {
    const payload = {
      source: source || "openclaw",
      type: "openclaw.reply",
      summary,
      detail,
    };
    const f = (globalThis && globalThis.fetch) ? globalThis.fetch.bind(globalThis) : null;
    if (f) {
      f(MANAGER_LOG_URL, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      }).catch(() => {});
    }
  } catch {}
}`

	patchedAny := false
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		content := string(data)
		patched := content
		fileChanged := false

		if !strings.Contains(patched, "return new Promise") && oldPattern.MatchString(patched) {
			patched = oldPattern.ReplaceAllString(patched, newReturn)
			fileChanged = true
		}
		if managerURLPattern.MatchString(patched) {
			next := managerURLPattern.ReplaceAllString(patched, managerURLLine)
			if next != patched {
				patched = next
				fileChanged = true
			}
		}
		if strings.Contains(patched, `const http = require("http")`) && loggerPattern.MatchString(patched) {
			patched = loggerPattern.ReplaceAllString(patched, loggerReplacement)
			fileChanged = true
		}
		if !fileChanged {
			continue
		}
		if err := os.WriteFile(path, []byte(patched), 0o644); err != nil {
			m.logger.Warn("QQ channel 补丁写入失败", "path", path, "error", err)
			continue
		}
		patchedAny = true
		m.logger.Info("QQ channel 兼容补丁已应用", "path", path)
	}

	if !patchedAny {
		m.logger.Debug("QQ channel 兼容补丁未命中")
	}
}
