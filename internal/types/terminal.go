// Copyright 2026 chenyang
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

// TerminalCreateResult 终端创建结果
type TerminalCreateResult struct {
	BaseResult
	Data *TerminalEnvironmentInfo `json:"data,omitempty"` // 环境信息
}

// TerminalEnvironmentInfo 终端环境信息
type TerminalEnvironmentInfo struct {
	WorkPath  string     `json:"workPath,omitempty"`  // 当前工作路径
	PythonEnv *PythonEnv `json:"pythonEnv,omitempty"` // Python 环境信息
}

// PythonEnv Python 环境信息
type PythonEnv struct {
	HasPython bool   `json:"hasPython"`         // 是否安装了 Python
	Version   string `json:"version,omitempty"` // Python 版本
	EnvActive bool   `json:"envActive"`         // 是否激活了虚拟环境
	EnvType   string `json:"envType,omitempty"` // 环境类型: venv, virtualenv, conda, pipenv, poetry
	EnvName   string `json:"envName,omitempty"` // 虚拟环境名称
	EnvPath   string `json:"envPath,omitempty"` // 虚拟环境路径
}

// GitInfo Git 信息
type GitInfo struct {
	IsRepo        bool   `json:"isRepo"`           // 是否是 Git 仓库
	Branch        string `json:"branch,omitempty"` // 当前分支
	ModifiedFiles int    `json:"modifiedFiles"`    // 修改的文件数
	AddedLines    int    `json:"addedLines"`       // 新增代码行数
	DeletedLines  int    `json:"deletedLines"`     // 删除代码行数
}

// TerminalTestConfigResult 终端配置测试结果
type TerminalTestConfigResult struct {
	BaseResult
	Data *TerminalTestConfigData `json:"data,omitempty"` // 测试数据
}

// TerminalTestConfigData 终端配置测试数据
type TerminalTestConfigData struct {
	Rows                   uint16 `json:"rows"`                             // 终端行数
	Cols                   uint16 `json:"cols"`                             // 终端列数
	WorkPath               string `json:"workPath,omitempty"`               // 工作路径
	WorkPathValid          bool   `json:"workPathValid,omitempty"`          // 工作路径是否有效
	RequestedShell         string `json:"requestedShell"`                   // 请求的 shell 类型
	DetectedShell          string `json:"detectedShell"`                    // 检测到的 shell 类型
	ShellPath              string `json:"shellPath"`                        // shell 路径
	Available              bool   `json:"available"`                        // shell 是否可用
	InitialCommand         string `json:"initialCommand,omitempty"`         // 初始命令
	InitialCommandExecuted bool   `json:"initialCommandExecuted,omitempty"` // 初始命令是否执行
	Output                 string `json:"output,omitempty"`                 // 初始命令输出
}

// TerminalListExecutableCommandsResult 获取可执行命令结果
type TerminalListExecutableCommandsResult struct {
	BaseResult
	Data *TerminalListExecutableCommandsData `json:"data,omitempty"` // 可执行命令数据
}

// TerminalListExecutableCommandsData 可执行命令数据
type TerminalListExecutableCommandsData struct {
	ResolvedShell   string                       `json:"resolvedShell"`   // 实际解析后的终端类型
	Commands        []*TerminalExecutableCommand `json:"commands"`        // 可执行命令列表
	DefaultCommands []string                     `json:"defaultCommands"` // shell 默认命令（通常为内建命令）
	Count           int                          `json:"count"`           // 命令总数
}

// TerminalExecutableCommand 可执行命令信息
type TerminalExecutableCommand struct {
	Name string `json:"name"` // 命令名称
	Path string `json:"path"` // 命令绝对路径
}

// TerminalFullscreenChangedEvent 终端全屏交互模式切换事件。
type TerminalFullscreenChangedEvent struct {
	SessionID     string `json:"sessionId"`     // 会话 ID
	InFullscreen  bool   `json:"inFullscreen"`  // 是否处于全屏交互模式
	ChangedAtUnix int64  `json:"changedAtUnix"` // 事件时间（Unix 毫秒）
}
