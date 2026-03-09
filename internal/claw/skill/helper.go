package skill

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/chenyang-zz/boxify/internal/types"
)

// buildBlockSet 从 openclaw.json 构建技能 blocklist 索引。
func buildBlockSet(ocConfig map[string]interface{}) map[string]bool {
	blockSet := make(map[string]bool)
	skillsCfg, _ := ocConfig["skills"].(map[string]interface{})
	blocklist, _ := skillsCfg["blocklist"].([]interface{})
	for _, raw := range blocklist {
		id, _ := raw.(string)
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		blockSet[id] = true
	}
	return blockSet
}

// scanDir 扫描技能目录并构建技能列表。
func (m *Manager) scanDir(dir string, blockSet map[string]bool, result *[]types.ClawSkill, seen map[string]bool) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		m.logger.Debug("跳过不可读技能目录", "dir", dir, "error", err)
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillID := entry.Name()
		if seen[skillID] {
			continue
		}

		skillDir := filepath.Join(dir, skillID)
		skillMDPath := filepath.Join(skillDir, "SKILL.md")
		if _, err := os.Stat(skillMDPath); err != nil {
			continue
		}
		seen[skillID] = true

		pkgInfo := loadJSONFile(filepath.Join(skillDir, "package.json"))
		skill := types.ClawSkill{
			ID:          skillID,
			Name:        firstNonEmptyString(stringFromMap(pkgInfo, "name"), skillID),
			Description: stringFromMap(pkgInfo, "description"),
			Version:     stringFromMap(pkgInfo, "version"),
			Enabled:     !blockSet[skillID],
			Source:      detectSource(dir),
			Path:        skillDir,
		}

		if mdData, err := os.ReadFile(skillMDPath); err == nil {
			applyMarkdownMetadata(&skill, string(mdData))
		}
		*result = append(*result, skill)
	}
}

// detectSource 根据技能目录推断来源标签。
func detectSource(dir string) string {
	switch {
	case strings.Contains(dir, string(filepath.Separator)+"work"+string(filepath.Separator)):
		return "workspace"
	case strings.Contains(dir, string(filepath.Separator)+"app"+string(filepath.Separator)):
		return "app-skill"
	default:
		return "skill"
	}
}

// applyMarkdownMetadata 从 SKILL.md 提取名称与描述。
func applyMarkdownMetadata(skill *types.ClawSkill, mdContent string) {
	if skill == nil {
		return
	}
	re := regexp.MustCompile(`(?s)^---\n(.*?)\n---`)
	if match := re.FindStringSubmatch(mdContent); len(match) > 1 {
		frontMatter := match[1]
		if nameMatch := regexp.MustCompile(`(?m)^name:\s*(.+)$`).FindStringSubmatch(frontMatter); len(nameMatch) > 1 {
			skill.Name = strings.Trim(strings.TrimSpace(nameMatch[1]), `"'`)
		}
		if descMatch := regexp.MustCompile(`(?m)^description:\s*["']?(.+?)["']?$`).FindStringSubmatch(frontMatter); len(descMatch) > 1 {
			skill.Description = strings.Trim(strings.TrimSpace(descMatch[1]), `"'`)
		}
	}
	if strings.TrimSpace(skill.Description) != "" {
		return
	}
	for _, line := range strings.Split(mdContent, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "---") {
			continue
		}
		if len(line) > 200 {
			line = line[:200]
		}
		skill.Description = line
		return
	}
}

// loadJSONFile 加载 JSON 文件为 map。
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

// stringFromMap 从 map 中读取字符串字段。
func stringFromMap(data map[string]interface{}, key string) string {
	if data == nil {
		return ""
	}
	value, _ := data[key].(string)
	return strings.TrimSpace(value)
}

// firstNonEmptyString 返回首个非空字符串。
func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

// uniqueExistingDirs 返回存在的唯一目录列表。
func uniqueExistingDirs(dirs []string) []string {
	result := make([]string, 0, len(dirs))
	seen := make(map[string]struct{})
	for _, dir := range dirs {
		dir = strings.TrimSpace(dir)
		if dir == "" {
			continue
		}
		dir = filepath.Clean(dir)
		if _, err := os.Stat(dir); err != nil {
			continue
		}
		if _, ok := seen[dir]; ok {
			continue
		}
		seen[dir] = struct{}{}
		result = append(result, dir)
	}
	return result
}

// detectGlobalOpenClawDir 检测 npm 全局 openclaw 包目录。
func detectGlobalOpenClawDir() string {
	out, err := exec.Command("npm", "root", "-g").Output()
	if err != nil {
		return ""
	}
	npmRoot := strings.TrimSpace(string(out))
	if npmRoot == "" {
		return ""
	}
	candidate := filepath.Join(npmRoot, "openclaw")
	if _, err := os.Stat(candidate); err != nil {
		return ""
	}
	return candidate
}

// existingDir 返回存在的目录，否则返回空字符串。
func existingDir(dir string) string {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return ""
	}
	if _, err := os.Stat(dir); err != nil {
		return ""
	}
	return dir
}
