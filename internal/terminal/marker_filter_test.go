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
	"testing"
	"time"
)

func TestNewMarkerFilter(t *testing.T) {
	filter := NewMarkerFilter()
	if filter == nil {
		t.Fatal("NewMarkerFilter returned nil")
	}
	if filter.startMarkerRegex == nil {
		t.Error("startMarkerRegex should not be nil")
	}
	if filter.endMarkerRegex == nil {
		t.Error("endMarkerRegex should not be nil")
	}
	if filter.fallbackDelay != 3*time.Second {
		t.Errorf("expected fallbackDelay 3s, got %v", filter.fallbackDelay)
	}
}

func TestMarkerFilter_Process_StartEndMarker(t *testing.T) {
	filter := NewMarkerFilter()

	// 模拟命令开始标记
	startMarker := "\x1b]133;A\x1b\\"
	result := filter.Process([]byte(startMarker))

	if result.CommandEnded {
		t.Error("command should not be ended after start marker")
	}
	if len(result.Output) != 0 {
		t.Errorf("expected no output after start marker, got %q", result.Output)
	}
	if !filter.InCommandOutput() {
		t.Error("should be in command output after start marker")
	}
}

func TestMarkerFilter_Process_CommandOutput(t *testing.T) {
	filter := NewMarkerFilter()

	// 发送开始标记
	startMarker := "\x1b]133;A\x1b\\"
	filter.Process([]byte(startMarker))

	// 发送命令输出
	output := "Hello, World!\n"
	result := filter.Process([]byte(output))

	if result.CommandEnded {
		t.Error("command should not be ended during output")
	}
	if string(result.Output) != output {
		t.Errorf("expected output %q, got %q", output, string(result.Output))
	}
}

func TestMarkerFilter_Process_EndMarker(t *testing.T) {
	filter := NewMarkerFilter()

	// 完整流程：开始标记 -> 输出 -> 结束标记
	startMarker := "\x1b]133;A\x1b\\"
	output := "test output\n"
	endMarker := "\x1b]133;D;0\x1b\\"

	// 处理开始标记
	filter.Process([]byte(startMarker))

	// 处理输出
	filter.Process([]byte(output))

	// 处理结束标记
	result := filter.Process([]byte(endMarker))

	if !result.CommandEnded {
		t.Error("command should be ended after end marker")
	}
	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}
}

func TestMarkerFilter_Process_EndMarkerWithExitCode(t *testing.T) {
	tests := []struct {
		name           string
		endMarker      string
		expectedExit   int
	}{
		{"exit code 0", "\x1b]133;D;0\x1b\\", 0},
		{"exit code 1", "\x1b]133;D;1\x1b\\", 1},
		{"exit code 127", "\x1b]133;D;127\x1b\\", 127},
		{"exit code 255", "\x1b]133;D;255\x1b\\", 255},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewMarkerFilter()

			// 处理开始和结束标记
			filter.Process([]byte("\x1b]133;A\x1b\\"))
			result := filter.Process([]byte(tt.endMarker))

			if !result.CommandEnded {
				t.Error("command should be ended")
			}
			if result.ExitCode != tt.expectedExit {
				t.Errorf("expected exit code %d, got %d", tt.expectedExit, result.ExitCode)
			}
		})
	}
}

func TestMarkerFilter_Process_AlternativeEndMarker(t *testing.T) {
	filter := NewMarkerFilter()

	// 使用 BEL (\x07) 作为终止符的标记
	startMarker := "\x1b]133;A\x07"
	endMarker := "\x1b]133;D;0\x07"

	filter.Process([]byte(startMarker))
	result := filter.Process([]byte(endMarker))

	if !result.CommandEnded {
		t.Error("command should be ended with BEL terminator")
	}
}

func TestMarkerFilter_Process_OutputBeforeStartMarker(t *testing.T) {
	filter := NewMarkerFilter()

	// 在开始标记之前的输出应该被丢弃（命令回显、提示符等）
	beforeOutput := "prompt$ "
	startMarker := "\x1b]133;A\x1b\\"
	commandOutput := "actual output\n"
	endMarker := "\x1b]133;D;0\x1b\\"

	// 发送开始标记之前的输出
	result1 := filter.Process([]byte(beforeOutput))
	if len(result1.Output) != 0 {
		t.Errorf("output before start marker should be discarded, got %q", result1.Output)
	}

	// 发送开始标记和命令输出
	filter.Process([]byte(startMarker))
	result2 := filter.Process([]byte(commandOutput))

	if string(result2.Output) != commandOutput {
		t.Errorf("expected command output %q, got %q", commandOutput, string(result2.Output))
	}

	// 发送结束标记
	filter.Process([]byte(endMarker))
}

func TestMarkerFilter_Process_MultipleCommands(t *testing.T) {
	filter := NewMarkerFilter()

	// 模拟多个命令的执行
	commands := []struct {
		output   string
		exitCode int
	}{
		{"first command output\n", 0},
		{"second command output\n", 1},
		{"third command output\n", 0},
	}

	for i, cmd := range commands {
		// 重置过滤器
		if i > 0 {
			filter.Reset()
		}

		// 开始标记
		filter.Process([]byte("\x1b]133;A\x1b\\"))

		// 命令输出
		result := filter.Process([]byte(cmd.output))
		if string(result.Output) != cmd.output {
			t.Errorf("command %d: expected output %q, got %q", i+1, cmd.output, string(result.Output))
		}

		// 结束标记
		endMarker := "\x1b]133;D;" + string(rune('0'+cmd.exitCode)) + "\x1b\\"
		endResult := filter.Process([]byte(endMarker))
		if !endResult.CommandEnded {
			t.Errorf("command %d: should be ended", i+1)
		}
	}
}

func TestMarkerFilter_Reset(t *testing.T) {
	filter := NewMarkerFilter()

	// 进入命令输出状态
	filter.Process([]byte("\x1b]133;A\x1b\\"))
	if !filter.InCommandOutput() {
		t.Error("should be in command output")
	}

	// 重置
	filter.Reset()
	if filter.InCommandOutput() {
		t.Error("should not be in command output after reset")
	}
}

func TestMarkerFilter_Process_SplitMarker(t *testing.T) {
	filter := NewMarkerFilter()

	// 模拟标记被分割成多个数据块
	part1 := "\x1b]133"
	part2 := ";A\x1b\\"

	filter.Process([]byte(part1))
	result := filter.Process([]byte(part2))

	// 标记应该被正确处理
	if result.CommandEnded {
		t.Error("command should not be ended after start marker")
	}
	if !filter.InCommandOutput() {
		t.Error("should be in command output after complete start marker")
	}
}

func TestMarkerFilter_Process_FallbackMode(t *testing.T) {
	filter := NewMarkerFilter()
	// 修改 fallbackDelay 以便测试
	filter.fallbackDelay = 10 * time.Millisecond

	// 等待超过 fallbackDelay
	time.Sleep(20 * time.Millisecond)

	// 发送不带标记的数据，应该进入降级模式
	data := "output without markers\n"
	result := filter.Process([]byte(data))

	if !filter.inFallbackMode {
		t.Error("should be in fallback mode after timeout")
	}
	if string(result.Output) != data {
		t.Errorf("fallback mode should pass through data, got %q", string(result.Output))
	}
}

func TestMarkerFilter_Process_CompleteFlow(t *testing.T) {
	filter := NewMarkerFilter()

	// 模拟完整的终端输出流
	prompt := "user@host:~$ "
	echo := "echo hello\n"
	startMarker := "\x1b]133;A\x1b\\"
	output := "hello\n"
	endMarker := "\x1b]133;D;0\x1b\\"
	nextPrompt := "user@host:~$ "

	// 提示符（应该被丢弃）
	r1 := filter.Process([]byte(prompt))
	if len(r1.Output) != 0 {
		t.Errorf("prompt should be discarded, got %q", r1.Output)
	}

	// 命令回显（应该被丢弃）
	r2 := filter.Process([]byte(echo))
	if len(r2.Output) != 0 {
		t.Errorf("command echo should be discarded, got %q", r2.Output)
	}

	// 开始标记
	filter.Process([]byte(startMarker))

	// 实际输出（应该被保留）
	r3 := filter.Process([]byte(output))
	if string(r3.Output) != output {
		t.Errorf("command output should be kept, got %q", string(r3.Output))
	}

	// 结束标记
	r4 := filter.Process([]byte(endMarker))
	if !r4.CommandEnded {
		t.Error("command should be ended")
	}

	// 下一个提示符（应该被丢弃）
	r5 := filter.Process([]byte(nextPrompt))
	if len(r5.Output) != 0 {
		t.Errorf("next prompt should be discarded, got %q", r5.Output)
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"0", 0},
		{"1", 1},
		{"10", 10},
		{"127", 127},
		{"255", 255},
		{"abc", 0},   // 无效字符
		{"12a3", 123}, // 部分有效
		{"", 0},       // 空字符串
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseInt(tt.input)
			if result != tt.expected {
				t.Errorf("parseInt(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMarkerFilter_Process_EmptyData(t *testing.T) {
	filter := NewMarkerFilter()

	result := filter.Process([]byte{})
	if len(result.Output) != 0 {
		t.Errorf("empty input should produce empty output, got %q", result.Output)
	}
	if result.CommandEnded {
		t.Error("empty input should not end command")
	}
}

func TestMarkerFilter_Process_ConcurrentSafety(t *testing.T) {
	filter := NewMarkerFilter()
	done := make(chan bool)

	// 启动多个 goroutine 并发处理
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				filter.Process([]byte("\x1b]133;A\x1b\\"))
				filter.Process([]byte("test\n"))
				filter.Process([]byte("\x1b]133;D;0\x1b\\"))
			}
			done <- true
		}()
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestMarkerFilter_Process_OSC1337(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedOutput string
	}{
		{
			name:           "OSC 1337 SetMark with ST terminator",
			input:          "\x1b]1337;SetMark\x1b\\",
			expectedOutput: "",
		},
		{
			name:           "OSC 1337 SetMark with BEL terminator",
			input:          "\x1b]1337;SetMark\x07",
			expectedOutput: "",
		},
		{
			name:           "OSC 1337 CurrentDir with ST terminator",
			input:          "\x1b]1337;CurrentDir=file://localhost/Users/test\x1b\\",
			expectedOutput: "",
		},
		{
			name:           "OSC 1337 CurrentDir with BEL terminator",
			input:          "\x1b]1337;CurrentDir=file://localhost/Users/test\x07",
			expectedOutput: "",
		},
		{
			name:           "OSC 1337 UserVars",
			input:          "\x1b]1337;User=user@host\x1b\\",
			expectedOutput: "",
		},
		{
			name:           "OSC 1337 CursorShape",
			input:          "\x1b]1337;CursorShape=1\x07",
			expectedOutput: "",
		},
		{
			name:           "OSC 1337 followed by normal text",
			input:          "\x1b]1337;SetMark\x1b\\hello world",
			expectedOutput: "hello world",
		},
		{
			name:           "normal text followed by OSC 1337",
			input:          "hello\x1b]1337;SetMark\x1b\\",
			expectedOutput: "hello",
		},
		{
			name:           "text with OSC 1337 in middle",
			input:          "hello\x1b]1337;SetMark\x1b\\world",
			expectedOutput: "helloworld",
		},
		{
			name:           "multiple OSC 1337 sequences",
			input:          "\x1b]1337;SetMark\x1b\\text\x1b]1337;CurrentDir=/tmp\x07more",
			expectedOutput: "textmore",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewMarkerFilter()
			// 先发送开始标记进入命令输出模式
			filter.Process([]byte("\x1b]133;A\x1b\\"))

			result := filter.Process([]byte(tt.input))
			if string(result.Output) != tt.expectedOutput {
				t.Errorf("expected output %q, got %q", tt.expectedOutput, string(result.Output))
			}
		})
	}
}

func TestMarkerFilter_Process_OSC1337WithOSC133(t *testing.T) {
	filter := NewMarkerFilter()

	// 第一部分：提示符和 OSC 1337（开始标记之前，应该被丢弃）
	r1 := filter.Process([]byte("prompt$ \x1b]1337;CurrentDir=/Users/test\x07"))
	if len(r1.Output) != 0 {
		t.Errorf("content before start marker should be discarded, got %q", r1.Output)
	}

	// 第二部分：开始标记
	filter.Process([]byte("\x1b]133;A\x1b\\"))

	// 第三部分：命令输出和 OSC 1337 混合（OSC 1337 应该被过滤，输出保留）
	r2 := filter.Process([]byte("command output\n\x1b]1337;SetMark\x1b\\"))
	if string(r2.Output) != "command output\n" {
		t.Errorf("expected 'command output\\n', got %q", string(r2.Output))
	}

	// 第四部分：结束标记
	r3 := filter.Process([]byte("\x1b]133;D;0\x1b\\"))
	if !r3.CommandEnded {
		t.Error("command should be ended")
	}

	// 第五部分：结束后的提示符和 OSC 1337（应该被丢弃）
	r4 := filter.Process([]byte("\x1b]1337;SetMark\x1b\\next prompt$ "))
	if len(r4.Output) != 0 {
		t.Errorf("content after end marker should be discarded, got %q", r4.Output)
	}
}

func TestMarkerFilter_Process_OSC1337EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedOutput string
	}{
		{
			name:           "OSC 1337 with special characters in path",
			input:          "\x1b]1337;CurrentDir=file://host/path with spaces\x1b\\",
			expectedOutput: "",
		},
		{
			name:           "OSC 1337 with unicode",
			input:          "\x1b]1337;Title=测试标题\x07",
			expectedOutput: "",
		},
		{
			name:           "OSC 1337 with RGB color",
			input:          "\x1b]1337;SetColors=fg=ffffff bg=000000\x1b\\",
			expectedOutput: "",
		},
		{
			name:           "empty OSC 1337",
			input:          "\x1b]1337;\x1b\\",
			expectedOutput: "",
		},
		{
			name:           "OSC 1337 split across chunks - part1",
			input:          "\x1b]1337;CurrentDir=/path",
			expectedOutput: "", // 不完整的序列，应该被缓冲
		},
		{
			name:           "OSC 1337 CurrentDir with file:// and BEL terminator",
			input:          "\x1b]1337;CurrentDir=file://SheepdeMacBook-Pro-7.local/Users/sheepzhao\x07",
			expectedOutput: "",
		},
		{
			name:           "OSC 1337 CurrentDir with file:// and ST terminator",
			input:          "\x1b]1337;CurrentDir=file://localhost/Users/test\x1b\\",
			expectedOutput: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewMarkerFilter()
			// 进入命令输出模式
			filter.Process([]byte("\x1b]133;A\x1b\\"))

			result := filter.Process([]byte(tt.input))
			if string(result.Output) != tt.expectedOutput {
				t.Errorf("expected output %q, got %q", tt.expectedOutput, string(result.Output))
			}
		})
	}
}

func TestMarkerFilter_Process_OSC1337Debug(t *testing.T) {
	filter := NewMarkerFilter()
	// 进入命令输出模式
	filter.Process([]byte("\x1b]133;A\x1b\\"))

	// 实际的 OSC 1337 CurrentDir 序列（来自 iTerm2 shell 集成）
	input := "\x1b]1337;CurrentDir=file://SheepdeMacBook-Pro-7.local/Users/sheepzhao\x07"
	result := filter.Process([]byte(input))

	t.Logf("输入: %q", input)
	t.Logf("输出: %q", string(result.Output))

	if string(result.Output) != "" {
		t.Errorf("OSC 1337 CurrentDir should be filtered, got %q", string(result.Output))
	}
}

func TestMarkerFilter_Process_OSC1337SplitAcrossChunks(t *testing.T) {
	filter := NewMarkerFilter()
	// 进入命令输出模式
	filter.Process([]byte("\x1b]133;A\x1b\\"))

	// OSC 1337 序列被分割成多个数据块
	part1 := "\x1b]1337;CurrentDir=file://SheepdeMac"
	part2 := "Book-Pro-7.local/Users/sheepzhao\x07"

	// 发送第一部分
	r1 := filter.Process([]byte(part1))
	t.Logf("part1 输出: %q", string(r1.Output))

	// 发送第二部分
	r2 := filter.Process([]byte(part2))
	t.Logf("part2 输出: %q", string(r2.Output))

	// 合并输出应该为空
	combined := string(r1.Output) + string(r2.Output)
	if combined != "" {
		t.Errorf("split OSC 1337 should be filtered, got %q", combined)
	}
}

func TestMarkerFilter_Process_OSC7(t *testing.T) {
	filter := NewMarkerFilter()
	// 进入命令输出模式
	filter.Process([]byte("\x1b]133;A\x1b\\"))

	// OSC 7 也是 file:// 格式，用于设置当前目录
	// 格式: ESC ] 7 ; file://host/path BEL
	input := "\x1b]7;file://SheepdeMacBook-Pro-7.local/Users/sheepzhao\x07"
	result := filter.Process([]byte(input))

	t.Logf("输入: %q", input)
	t.Logf("输出: %q", string(result.Output))

	// OSC 7 目前是否被过滤？
	if string(result.Output) != "" {
		t.Logf("OSC 7 not filtered, output: %q", string(result.Output))
	}
}

func TestMarkerFilter_Process_RealWorldOutput(t *testing.T) {
	filter := NewMarkerFilter()
	// 进入命令输出模式
	filter.Process([]byte("\x1b]133;A\x1b\\"))

	// 模拟真实的终端输出，包含 OSC 1337 和回车
	input := "\x1b]1337;CurrentDir=file://SheepdeMacBook-Pro-7.local/Users/sheepzhao\x07\r\x1b]1337;SetMark\x07"
	result := filter.Process([]byte(input))

	t.Logf("输入: %q", input)
	t.Logf("输出: %q", string(result.Output))

	// 应该只有 \r 被保留（因为它不是 OSC 序列的一部分）
	if string(result.Output) != "\r" {
		t.Errorf("expected only \\r, got %q", string(result.Output))
	}
}
