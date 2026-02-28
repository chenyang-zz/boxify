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

package terminal

import (
	"bytes"
	"regexp"
	"sync"
	"time"
)

// MarkerFilter 基于 OSC 133 标记的输出过滤器
// 使用 Shell Hooks 或命令包装注入的标记来识别命令边界
type MarkerFilter struct {
	mu                sync.Mutex
	buffer            bytes.Buffer
	inCommandOutput   bool          // 是否在命令输出区域内
	startMarkerRegex  *regexp.Regexp
	endMarkerRegex    *regexp.Regexp
	oscFilterRegex    *regexp.Regexp // 通用 OSC 序列过滤（排除 OSC 133）
	createdAt         time.Time     // 过滤器创建时间
	markerDetected    bool          // 是否已检测到标记
	fallbackDelay     time.Duration // 降级延迟时间
	inFallbackMode    bool          // 是否处于降级模式（透传所有输出）
}

// NewMarkerFilter 创建标记过滤器
func NewMarkerFilter() *MarkerFilter {
	return &MarkerFilter{
		// 匹配 \x1b]133;A\x1b\ 或 \x1b]133;A\x07 (OSC 133;A - 命令开始)
		startMarkerRegex: regexp.MustCompile(`\x1b\]133;A(?:\x1b\\|\x07)`),
		// 匹配 \x1b]133;D;{exit_code}\x1b\ 或 \x1b]133;D;{exit_code}\x07 (OSC 133;D - 命令结束)
		endMarkerRegex: regexp.MustCompile(`\x1b\]133;D;(\d+)(?:\x1b\\|\x07)`),
		// 匹配需要过滤的 OSC 序列：OSC 0, 1, 2, 7, 1337 等
		// OSC 133 用于命令边界标记，不需要过滤
		oscFilterRegex: regexp.MustCompile(`\x1b\](?:0|1|2|7|1337);[^\x07\x1b]*(?:\x1b\\|\x07)`),
		createdAt:      time.Now(),
		fallbackDelay:  3 * time.Second, // 3秒后如果没检测到标记，进入降级模式
	}
}

// ProcessResult 处理结果
type ProcessResult struct {
	Output       []byte // 过滤后的输出
	CommandEnded bool   // 命令是否结束
	ExitCode     int    // 命令退出码（仅在 CommandEnded 为 true 时有效）
}

// Process 处理输出数据
func (f *MarkerFilter) Process(data []byte) ProcessResult {
	f.mu.Lock()
	defer f.mu.Unlock()

	// 检查是否需要进入降级模式
	if !f.markerDetected && !f.inFallbackMode {
		if time.Since(f.createdAt) > f.fallbackDelay {
			f.inFallbackMode = true
			// 输出之前缓冲的内容
			f.buffer.Write(data)
			output := f.buffer.Bytes()
			f.buffer.Reset()
			return ProcessResult{
				Output:       output,
				CommandEnded: false,
				ExitCode:     0,
			}
		}
	}

	// 降级模式：透传所有输出
	if f.inFallbackMode {
		return ProcessResult{
			Output:       data,
			CommandEnded: false,
			ExitCode:     0,
		}
	}

	f.buffer.Write(data)
	content := f.buffer.String()

	// 先移除非 OSC 133 的 OSC 序列（如 OSC 0, 1, 2, 7, 1337 等）
	content = string(f.oscFilterRegex.ReplaceAll([]byte(content), []byte("")))

	var result bytes.Buffer
	var commandEnded bool
	var exitCode int

	// 循环处理所有标记
	for {
		startMatch := f.startMarkerRegex.FindStringIndex(content)
		endMatch := f.endMarkerRegex.FindStringSubmatchIndex(content)

		// 确定最先匹配的标记
		nextMarkerIdx := -1
		nextMarkerLen := 0
		isEndMarker := false

		if startMatch != nil && (endMatch == nil || startMatch[0] < endMatch[0]) {
			nextMarkerIdx = startMatch[0]
			nextMarkerLen = startMatch[1] - startMatch[0]
			isEndMarker = false
		} else if endMatch != nil {
			nextMarkerIdx = endMatch[0]
			nextMarkerLen = endMatch[1] - endMatch[0]
			isEndMarker = true
		}

		// 没有找到更多标记
		if nextMarkerIdx == -1 {
			break
		}

		// 标记已检测到
		f.markerDetected = true

		// 处理标记之前的内容
		if nextMarkerIdx > 0 {
			before := content[:nextMarkerIdx]
			if f.inCommandOutput {
				result.WriteString(before)
			}
			// 不在命令输出区域时，丢弃内容（命令回显、提示符等）
		}

		if isEndMarker {
			// 结束标记
			// 提取退出码
			if endMatch[2] != -1 && endMatch[3] != -1 {
				exitCodeStr := content[endMatch[2]:endMatch[3]]
				exitCode = parseInt(exitCodeStr)
			}
			f.inCommandOutput = false
			commandEnded = true
		} else {
			// 开始标记
			f.inCommandOutput = true
		}

		// 移除已处理的内容
		content = content[nextMarkerIdx+nextMarkerLen:]
	}

	// 处理剩余内容
	if len(content) > 0 {
		if f.inCommandOutput {
			// 检查是否可能是未完成的标记
			if f.isPossibleMarkerStart(content) {
				// 保留到缓冲区等待更多数据
				f.buffer.Reset()
				f.buffer.WriteString(content)
			} else {
				// 确定是命令输出
				result.WriteString(content)
				f.buffer.Reset()
			}
		} else {
			// 不在命令输出区域，保留到缓冲区（可能是提示符的一部分）
			f.buffer.Reset()
			f.buffer.WriteString(content)
		}
	} else {
		f.buffer.Reset()
	}

	return ProcessResult{
		Output:       result.Bytes(),
		CommandEnded: commandEnded,
		ExitCode:     exitCode,
	}
}

// isPossibleMarkerStart 检查是否可能是标记的开始
func (f *MarkerFilter) isPossibleMarkerStart(s string) bool {
	// 检查是否以 ESC 开头（可能是未完成的 OSC 序列）
	for i := 0; i < len(s) && i < 10; i++ {
		if s[i] == 0x1b {
			return true
		}
	}
	return false
}

// Reset 重置过滤器状态
func (f *MarkerFilter) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.buffer.Reset()
	f.inCommandOutput = false
}

// InCommandOutput 返回当前是否在命令输出区域
func (f *MarkerFilter) InCommandOutput() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.inCommandOutput
}

func parseInt(s string) int {
	var result int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int(c - '0')
		}
	}
	return result
}
