package plugin

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// 插件管理器所需的最小配置。
type Config struct {
	OpenClawDir string // OpenClaw 运行目录（含 extensions/openclaw.json）。
	DataDir     string // Boxify 持久化插件状态与缓存的目录。
}

// 读取 OpenClaw 配置文件，不存在时返回空配置。
func (c *Config) ReadOpenClawJSON() (map[string]interface{}, error) {
	data, err := os.ReadFile(c.openClawJSONPath())
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]interface{}{}, nil
		}
		return nil, err
	}
	cfg := map[string]interface{}{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// 写回 OpenClaw 配置文件。
func (c *Config) WriteOpenClawJSON(cfg map[string]interface{}) error {
	if cfg == nil {
		cfg = map[string]interface{}{}
	}
	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(c.openClawJSONPath()), 0o755); err != nil {
		return err
	}
	return os.WriteFile(c.openClawJSONPath(), out, 0o644)
}

// 返回 OpenClaw 主配置文件路径。
func (c *Config) openClawJSONPath() string {
	return filepath.Join(c.OpenClawDir, "openclaw.json")
}
