package plugin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// 从 openclaw.json 提取插件 entries 与 installs。
func extractOpenClawPluginState(ocConfig map[string]interface{}) (map[string]interface{}, map[string]interface{}) {
	pluginsCfg, _ := ocConfig["plugins"].(map[string]interface{})
	pluginEntries, _ := pluginsCfg["entries"].(map[string]interface{})
	pluginInstalls, _ := pluginsCfg["installs"].(map[string]interface{})
	if pluginEntries == nil {
		pluginEntries = map[string]interface{}{}
	}
	if pluginInstalls == nil {
		pluginInstalls = map[string]interface{}{}
	}
	return pluginEntries, pluginInstalls
}

// 过滤技能中心展示的插件状态，优先回退到最近一次未污染备份中的插件集合。
func filterSkillPluginState(
	openClawDir string,
	entries map[string]interface{},
	installs map[string]interface{},
) (map[string]interface{}, map[string]interface{}) {
	if len(entries) <= 10 {
		return entries, installs
	}

	allowedIDs := latestBackupPluginIDs(openClawDir)
	for pluginID := range localOpenClawPluginIDs(openClawDir) {
		allowedIDs[pluginID] = struct{}{}
	}
	if len(allowedIDs) == 0 {
		return entries, installs
	}

	filteredEntries := make(map[string]interface{})
	filteredInstalls := make(map[string]interface{})
	for pluginID := range allowedIDs {
		if entry, ok := entries[pluginID]; ok {
			filteredEntries[pluginID] = entry
		}
		if install, ok := installs[pluginID]; ok {
			filteredInstalls[pluginID] = install
		}
	}
	if len(filteredEntries) == 0 && len(filteredInstalls) == 0 {
		return entries, installs
	}
	return filteredEntries, filteredInstalls
}

// 读取当前 OpenClaw 实例本地 extensions 目录中的插件 ID。
func localOpenClawPluginIDs(openClawDir string) map[string]struct{} {
	result := make(map[string]struct{})
	entries, err := os.ReadDir(filepath.Join(openClawDir, "extensions"))
	if err != nil {
		return result
	}
	for _, entry := range entries {
		if entry.IsDir() {
			result[entry.Name()] = struct{}{}
		}
	}
	return result
}

// 返回最近一次 openclaw 备份中的插件 entry 集合。
func latestBackupPluginIDs(openClawDir string) map[string]struct{} {
	result := make(map[string]struct{})
	backupDir := filepath.Join(openClawDir, "backups")
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return result
	}

	latestName := ""
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "pre-edit-") || !strings.HasSuffix(name, ".json") {
			continue
		}
		if name > latestName {
			latestName = name
		}
	}
	if latestName == "" {
		return result
	}

	payload := loadJSONFile(filepath.Join(backupDir, latestName))
	pluginsCfg, _ := payload["plugins"].(map[string]interface{})
	backupEntries, _ := pluginsCfg["entries"].(map[string]interface{})
	for pluginID := range backupEntries {
		result[pluginID] = struct{}{}
	}
	return result
}

// 扫描插件目录并转换为技能中心插件项。
func scanSkillCenterPluginDir(dir string, entries, installs map[string]interface{}, result *[]SkillCenterPlugin, seen map[string]bool, source string) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, entry := range dirEntries {
		if !entry.IsDir() {
			continue
		}
		pluginID := entry.Name()
		if seen[pluginID] {
			continue
		}
		seen[pluginID] = true

		pluginDir := filepath.Join(dir, pluginID)
		pkgInfo := loadJSONFile(filepath.Join(pluginDir, "package.json"))
		pluginInfo := loadJSONFile(filepath.Join(pluginDir, "openclaw.plugin.json"))

		item := SkillCenterPlugin{
			ID:          pluginID,
			Name:        firstNonEmptyString(stringFromMap(pluginInfo, "name"), stringFromMap(pkgInfo, "name"), pluginID),
			Description: firstNonEmptyString(stringFromMap(pluginInfo, "description"), stringFromMap(pkgInfo, "description")),
			Version:     stringFromMap(pkgInfo, "version"),
			Enabled:     true,
			Source:      source,
			Path:        pluginDir,
		}
		if entryCfg, ok := entries[pluginID].(map[string]interface{}); ok {
			if enabled, ok := entryCfg["enabled"].(bool); ok {
				item.Enabled = enabled
			}
		}
		if installCfg, ok := installs[pluginID].(map[string]interface{}); ok {
			if version, ok := installCfg["version"].(string); ok && strings.TrimSpace(version) != "" {
				item.Version = version
			}
			if installedAt, ok := installCfg["installedAt"].(string); ok {
				item.InstalledAt = installedAt
			}
			if installPath, ok := installCfg["installPath"].(string); ok && strings.TrimSpace(installPath) != "" {
				item.Path = installPath
			}
		}
		*result = append(*result, item)
	}
}

// 追加仅存在于 openclaw.json 中的插件项。
func appendConfigOnlyPlugins(entries, installs map[string]interface{}, result *[]SkillCenterPlugin, seen map[string]bool) {
	for pluginID, rawEntry := range entries {
		if seen[pluginID] {
			continue
		}
		seen[pluginID] = true

		item := SkillCenterPlugin{
			ID:      pluginID,
			Name:    pluginID,
			Enabled: true,
			Source:  "config",
		}
		if entryCfg, ok := rawEntry.(map[string]interface{}); ok {
			if enabled, ok := entryCfg["enabled"].(bool); ok {
				item.Enabled = enabled
			}
		}
		if installCfg, ok := installs[pluginID].(map[string]interface{}); ok {
			item.Version = stringFromMap(installCfg, "version")
			item.InstalledAt = stringFromMap(installCfg, "installedAt")
			item.Path = stringFromMap(installCfg, "installPath")
		}
		*result = append(*result, item)
	}
}

// 按 openclaw.json 中声明的 installPath 补扫当前实例插件，避免误扫全局插件目录。
func scanConfiguredPluginInstalls(entries, installs map[string]interface{}, result *[]SkillCenterPlugin, seen map[string]bool) {
	for pluginID, rawInstall := range installs {
		if seen[pluginID] {
			continue
		}

		installCfg, ok := rawInstall.(map[string]interface{})
		if !ok {
			continue
		}
		installPath := strings.TrimSpace(stringFromMap(installCfg, "installPath"))
		if installPath == "" {
			continue
		}
		info, err := os.Stat(installPath)
		if err != nil || !info.IsDir() {
			continue
		}

		seen[pluginID] = true
		pkgInfo := loadJSONFile(filepath.Join(installPath, "package.json"))
		pluginInfo := loadJSONFile(filepath.Join(installPath, "openclaw.plugin.json"))
		item := SkillCenterPlugin{
			ID:          pluginID,
			Name:        firstNonEmptyString(stringFromMap(pluginInfo, "name"), stringFromMap(pkgInfo, "name"), pluginID),
			Description: firstNonEmptyString(stringFromMap(pluginInfo, "description"), stringFromMap(pkgInfo, "description")),
			Version:     firstNonEmptyString(stringFromMap(installCfg, "version"), stringFromMap(pkgInfo, "version")),
			Enabled:     true,
			Source:      "installed",
			InstalledAt: stringFromMap(installCfg, "installedAt"),
			Path:        installPath,
		}
		if entryCfg, ok := entries[pluginID].(map[string]interface{}); ok {
			if enabled, ok := entryCfg["enabled"].(bool); ok {
				item.Enabled = enabled
			}
		}
		*result = append(*result, item)
	}
}

// 加载 JSON 文件为 map。
func loadJSONFile(path string) map[string]interface{} {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil
	}
	return payload
}

// 从 map 中读取字符串字段。
func stringFromMap(data map[string]interface{}, key string) string {
	if data == nil {
		return ""
	}
	value, _ := data[key].(string)
	return strings.TrimSpace(value)
}

// 返回首个非空字符串。
func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
