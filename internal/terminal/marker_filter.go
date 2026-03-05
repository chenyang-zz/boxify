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
	"encoding/base64"
	"log/slog"
	"regexp"
	"strings"
	"sync"
	"time"
)

// MarkerFilter 基于 OSC 133 标记的输出过滤器
// 使用 Shell Hooks 或命令包装注入的标记来识别命令边界
type MarkerFilter struct {
	mu               sync.Mutex
	buffer           bytes.Buffer
	inCommandOutput  bool   // 是否在命令输出区域内
	inInteractive    bool   // 是否处于 alternate screen 交互模式
	currentPwd       string // 当前工作路径
	startMarkerRegex *regexp.Regexp
	endMarkerRegex   *regexp.Regexp
	pwdMarkerRegex   *regexp.Regexp // OSC 1337;Pwd 序列
	oscGenericFilter *regexp.Regexp // 通用 OSC 序列过滤 (0, 1, 2, 7)
	osc1337Filter    *regexp.Regexp // OSC 1337 序列过滤（在代码中排除 Pwd）
	privateModeRegex *regexp.Regexp // CSI ? Ps h/l 序列（DEC 私有模式）
	createdAt        time.Time      // 过滤器创建时间
	markerDetected   bool           // 是否已检测到标记
	fallbackDelay    time.Duration  // 降级延迟时间
	inFallbackMode   bool           // 是否处于降级模式（透传所有输出）
	privateModeTail  string         // 处理跨 chunk 的不完整私有模式序列
	activeModes      map[string]struct{}
	logger           *slog.Logger
}

// NewMarkerFilter 创建标记过滤器
func NewMarkerFilter(logger *slog.Logger) *MarkerFilter {
	return &MarkerFilter{
		logger: logger,
		// 匹配 \x1b]133;A\x1b\ 或 \x1b]133;A\x07 (OSC 133;A - 命令开始)
		startMarkerRegex: regexp.MustCompile(`\x1b\]133;A(?:\x1b\\|\x07)`),
		// 匹配 \x1b]133;D;{exit_code}\x1b\ 或 \x1b]133;D;{exit_code}\x07 (OSC 133;D - 命令结束)
		endMarkerRegex: regexp.MustCompile(`\x1b\]133;D;(\d+)(?:\x1b\\|\x07)`),
		// 匹配 \x1b]1337;Pwd;{base64}\x1b\ 或 \x1b]1337;Pwd;{base64}\x07 (OSC 1337;Pwd - 工作路径更新)
		pwdMarkerRegex: regexp.MustCompile(`\x1b\]1337;Pwd;([A-Za-z0-9+/=]+)(?:\x1b\\|\x07)`),
		// 匹配需要过滤的通用 OSC 序列：OSC 0, 1, 2, 7（窗口标题等）
		// 使用 ; 确保只匹配以数字开头后跟 ; 的序列，避免误匹配 OSC 133, 1337 等
		oscGenericFilter: regexp.MustCompile(`\x1b\](?:0|1|2|7);[^\x07\x1b]*(?:\x1b\\|\x07)`),
		// 匹配 OSC 1337 序列（在代码中排除 Pwd）
		osc1337Filter: regexp.MustCompile(`\x1b\]1337;[^\x07\x1b]*(?:\x1b\\|\x07)`),
		// 匹配 DEC 私有模式切换：ESC[?Ps h/l
		privateModeRegex: regexp.MustCompile(`\x1b\[\?(\d+)([hl])`),
		createdAt:        time.Now(),
		fallbackDelay:    3 * time.Second, // 3秒后如果没检测到标记，进入降级模式
		activeModes:      make(map[string]struct{}),
	}
}

// ProcessResult 处理结果
type ProcessResult struct {
	Output                 []byte // 过滤后的输出
	CommandEnded           bool   // 命令是否结束
	ExitCode               int    // 命令退出码（仅在 CommandEnded 为 true 时有效）
	PwdChanged             bool   // 工作路径是否变化
	Pwd                    string // 新的工作路径（仅在 PwdChanged 为 true 时有效）
	InteractionModeChanged bool   // 交互模式是否切换
	InInteractive          bool   // 当前是否处于交互模式
}

// Process 处理输出数据
func (f *MarkerFilter) Process(data []byte) ProcessResult {
	f.mu.Lock()
	defer f.mu.Unlock()
	wasInCommandOutput := f.inCommandOutput

	// 检查是否需要进入降级模式
	if !f.markerDetected && !f.inFallbackMode {
		if time.Since(f.createdAt) > f.fallbackDelay {
			interactionModeChanged, inInteractive := f.detectInteractionModeTransition(
				data,
				wasInCommandOutput,
				false,
			)
			f.inFallbackMode = true
			// 输出之前缓冲的内容
			f.buffer.Write(data)
			output := f.buffer.Bytes()
			f.buffer.Reset()
			return ProcessResult{
				Output:                 output,
				CommandEnded:           false,
				ExitCode:               0,
				InteractionModeChanged: interactionModeChanged,
				InInteractive:          inInteractive,
			}
		}
	}

	// 降级模式：透传所有输出
	if f.inFallbackMode {
		interactionModeChanged, inInteractive := f.detectInteractionModeTransition(
			data,
			wasInCommandOutput,
			false,
		)
		return ProcessResult{
			Output:                 data,
			CommandEnded:           false,
			ExitCode:               0,
			InteractionModeChanged: interactionModeChanged,
			InInteractive:          inInteractive,
		}
	}

	f.buffer.Write(data)
	content := f.buffer.String()

	// 先移除通用 OSC 序列（OSC 0, 1, 2, 7 - 窗口标题等）
	content = string(f.oscGenericFilter.ReplaceAll([]byte(content), []byte("")))

	// 移除 OSC 1337 序列（排除 Pwd，Pwd 由 pwdMarkerRegex 处理）
	// 由于 Go RE2 不支持负向先行断言，需要手动检查
	content = f.filterOSC1337(content)

	var result bytes.Buffer
	var commandEnded bool
	var exitCode int
	var pwdChanged bool
	var pwd string

	// 循环处理所有标记
	for {
		startMatch := f.startMarkerRegex.FindStringIndex(content)
		endMatch := f.endMarkerRegex.FindStringSubmatchIndex(content)
		pwdMatch := f.pwdMarkerRegex.FindStringSubmatchIndex(content)

		// 确定最先匹配的标记
		nextMarkerIdx := -1
		nextMarkerLen := 0
		markerType := 0 // 0: none, 1: start, 2: end, 3: pwd

		// 找出最先出现的标记
		if startMatch != nil {
			nextMarkerIdx = startMatch[0]
			nextMarkerLen = startMatch[1] - startMatch[0]
			markerType = 1
		}
		if endMatch != nil && (nextMarkerIdx == -1 || endMatch[0] < nextMarkerIdx) {
			nextMarkerIdx = endMatch[0]
			nextMarkerLen = endMatch[1] - endMatch[0]
			markerType = 2
		}
		if pwdMatch != nil && (nextMarkerIdx == -1 || pwdMatch[0] < nextMarkerIdx) {
			nextMarkerIdx = pwdMatch[0]
			nextMarkerLen = pwdMatch[1] - pwdMatch[0]
			markerType = 3
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

		switch markerType {
		case 2:
			// 结束标记
			// 提取退出码
			if endMatch[2] != -1 && endMatch[3] != -1 {
				exitCodeStr := content[endMatch[2]:endMatch[3]]
				exitCode = parseInt(exitCodeStr)
			}
			f.inCommandOutput = false
			commandEnded = true
		case 3:
			// Pwd 标记
			// 提取并解码 Base64 编码的路径
			if pwdMatch[2] != -1 && pwdMatch[3] != -1 {
				encoded := content[pwdMatch[2]:pwdMatch[3]]
				if decoded, err := base64.StdEncoding.DecodeString(encoded); err == nil {
					newPwd := string(decoded)
					// 只有路径真正改变时才设置 PwdChanged
					if newPwd != f.currentPwd {
						f.currentPwd = newPwd
						pwd = newPwd
						pwdChanged = true
					}
				}
			}
		case 1:
			// 开始标记
			f.inCommandOutput = true
		}

		// 移除已处理的内容
		content = content[nextMarkerIdx+nextMarkerLen:]
	}

	// 处理剩余内容
	if len(content) > 0 {
		if f.inCommandOutput {
			// 仅缓存“未完成的 OSC 标记尾部”，其余内容立即透传，避免命令输出延迟到结束才释放。
			prefix, tail, hasIncompleteTail := f.splitIncompleteMarkerTail(content)
			if hasIncompleteTail {
				if len(prefix) > 0 {
					result.WriteString(prefix)
				}
				f.buffer.Reset()
				f.buffer.WriteString(tail)
			} else {
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

	allowHintEnter := wasInCommandOutput || f.inCommandOutput || f.startMarkerRegex.Match(data)
	interactionModeChanged, inInteractive := f.detectInteractionModeTransition(
		data,
		allowHintEnter,
		commandEnded,
	)

	return ProcessResult{
		Output:                 result.Bytes(),
		CommandEnded:           commandEnded,
		ExitCode:               exitCode,
		PwdChanged:             pwdChanged,
		Pwd:                    pwd,
		InteractionModeChanged: interactionModeChanged,
		InInteractive:          inInteractive,
	}
}

// detectInteractionModeTransition 检测交互模式切换。
// 规则参考 Warp/FinalTerm 类终端常见做法：
// - 优先依赖 DEC 私有模式切换（alt-screen、鼠标追踪等）而非宽泛控制序列启发式；
// - 命令结束时强制退出交互模式，防止缺少 leave 序列导致状态悬挂。
func (f *MarkerFilter) detectInteractionModeTransition(
	data []byte,
	allowHintEnter bool,
	commandEnded bool,
) (changed bool, inInteractive bool) {
	_ = allowHintEnter // 仅保留参数兼容，当前策略不使用弱启发式。
	prev := f.inInteractive

	combined := f.privateModeTail + string(data)
	matches := f.privateModeRegex.FindAllStringSubmatchIndex(combined, -1)
	for _, match := range matches {
		if len(match) < 6 || match[2] == -1 || match[3] == -1 || match[4] == -1 || match[5] == -1 {
			continue
		}

		mode := combined[match[2]:match[3]]
		action := combined[match[4]:match[5]]
		if !isInteractiveMode(mode) {
			continue
		}

		// 统一约束：进入交互态仅允许发生在命令执行窗口内。
		// 退出交互态允许随时生效，用于清理可能的悬挂状态。
		entersInteractive := action == "h"

		// DECTCEM: ?25l 隐藏光标通常用于交互界面，?25h 为恢复。
		if mode == "25" {
			if action == "l" {
				entersInteractive = true
				f.activeModes[mode] = struct{}{}
			} else {
				entersInteractive = false
				delete(f.activeModes, mode)
			}
		} else if action == "h" {
			f.activeModes[mode] = struct{}{}
		} else {
			delete(f.activeModes, mode)
		}

		if entersInteractive && !allowHintEnter && !isAltScreenMode(mode) {
			// 不允许在非命令执行窗口进入交互态，回滚刚刚写入的 mode。
			delete(f.activeModes, mode)
		}
	}

	// 保留末尾用于处理跨 chunk 的 ESC[?Ps h/l 序列。
	if len(combined) > 32 {
		f.privateModeTail = combined[len(combined)-32:]
	} else {
		f.privateModeTail = combined
	}

	if commandEnded {
		f.activeModes = make(map[string]struct{})
	}

	f.inInteractive = len(f.activeModes) > 0
	changed = prev != f.inInteractive

	return changed, f.inInteractive
}

// splitIncompleteMarkerTail 将内容拆分为“可立即输出前缀”和“需继续等待的未完成 OSC 尾部”。
// 仅对不完整的 OSC 序列（ESC ] ...，且尚未遇到 BEL / ST）进行缓存。
func (f *MarkerFilter) splitIncompleteMarkerTail(s string) (prefix string, tail string, hasIncompleteTail bool) {
	if s == "" {
		return "", "", false
	}

	lastEsc := strings.LastIndexByte(s, 0x1b)
	if lastEsc == -1 {
		return s, "", false
	}

	tail = s[lastEsc:]
	if tail == "\x1b" {
		return s[:lastEsc], tail, true
	}

	// 仅处理 OSC 前缀，其他 ANSI 序列（例如 CSI 颜色）不做缓存，直接透传。
	if !strings.HasPrefix(tail, "\x1b]") {
		return s, "", false
	}

	// 已经包含终止符，说明是完整 OSC，无需缓存。
	if strings.Contains(tail, "\x07") || strings.Contains(tail, "\x1b\\") {
		return s, "", false
	}

	return s[:lastEsc], tail, true
}

// Reset 重置过滤器状态
func (f *MarkerFilter) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.buffer.Reset()
	f.inCommandOutput = false
	f.inInteractive = false
	f.privateModeTail = ""
	f.activeModes = make(map[string]struct{})
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
			result = result*10 + int(c-'0')
		}
	}
	return result
}

// isInteractiveMode 判断 DEC 私有模式是否属于交互态信号。
func isInteractiveMode(mode string) bool {
	switch mode {
	case
		"1",    // DECCKM 应用光标键
		"9",    // X10 鼠标
		"25",   // DECTCEM 光标显示/隐藏（特殊处理）
		"47",   // alt-screen
		"1000", // VT200 鼠标
		"1001", // 高亮跟踪
		"1002", // Button-event 鼠标
		"1003", // Any-event 鼠标
		"1004", // FocusIn/Out
		"1005", // UTF-8 鼠标编码
		"1006", // SGR 鼠标编码
		"1015", // urxvt 鼠标编码
		"1016", // SGR 像素鼠标模式
		"1047", // alt-screen
		"1049": // alt-screen + cursor save/restore
		return true
	default:
		return false
	}
}

// isAltScreenMode 判断是否为 alternate screen 相关模式。
func isAltScreenMode(mode string) bool {
	return mode == "47" || mode == "1047" || mode == "1049"
}

// filterOSC1337 过滤 OSC 1337 序列，但保留 Pwd 序列
func (f *MarkerFilter) filterOSC1337(content string) string {
	// 使用 ReplaceAllFunc 来检查每个匹配是否是 Pwd 序列
	return string(f.osc1337Filter.ReplaceAllFunc([]byte(content), func(match []byte) []byte {
		// 如果是 OSC 1337;Pwd 序列，保留它
		if bytes.HasPrefix(match, []byte("\x1b]1337;Pwd")) {
			return match
		}
		// 其他 OSC 1337 序列，移除
		return []byte("")
	}))
}
