package process

import "strings"

// NormalizeOpenClawConfig 对 openclaw.json 做兼容性清洗，返回 true 表示发生修改。
func NormalizeOpenClawConfig(cfg map[string]interface{}) bool {
	if cfg == nil {
		return false
	}
	changed := false

	if gateway, ok := cfg["gateway"].(map[string]interface{}); ok && gateway != nil {
		if mode, ok := gateway["mode"].(string); ok && strings.TrimSpace(mode) == "hosted" {
			gateway["mode"] = "remote"
			changed = true
		}
		if custom, ok := gateway["customBindHost"].(string); !ok || strings.TrimSpace(custom) == "" {
			if bindAddr, ok := gateway["bindAddress"].(string); ok && strings.TrimSpace(bindAddr) != "" {
				gateway["customBindHost"] = strings.TrimSpace(bindAddr)
				changed = true
			}
		}
		if _, ok := gateway["bindAddress"]; ok {
			delete(gateway, "bindAddress")
			changed = true
		}
	}

	if hooks, ok := cfg["hooks"].(map[string]interface{}); ok && hooks != nil {
		if p, ok := hooks["path"].(string); !ok || strings.TrimSpace(p) == "" {
			if legacyPath, ok := hooks["basePath"].(string); ok && strings.TrimSpace(legacyPath) != "" {
				hooks["path"] = strings.TrimSpace(legacyPath)
				changed = true
			}
		}
		if token, ok := hooks["token"].(string); !ok || strings.TrimSpace(token) == "" {
			if legacyToken, ok := hooks["secret"].(string); ok && strings.TrimSpace(legacyToken) != "" {
				hooks["token"] = legacyToken
				changed = true
			}
		}
		if _, ok := hooks["basePath"]; ok {
			delete(hooks, "basePath")
			changed = true
		}
		if _, ok := hooks["secret"]; ok {
			delete(hooks, "secret")
			changed = true
		}
	}

	return changed
}
