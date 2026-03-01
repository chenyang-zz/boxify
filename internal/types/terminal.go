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
