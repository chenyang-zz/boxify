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
	"encoding/base64"
	"io"

	"log/slog"
)

// EventEmitter 事件发射接口（解耦 Wails 依赖）
type EventEmitter interface {
	Emit(event string, data map[string]interface{})
}

// OutputHandler 输出处理器
type OutputHandler struct {
	emitter EventEmitter
	logger  *slog.Logger
}

// NewOutputHandler 创建输出处理器
func NewOutputHandler(emitter EventEmitter, logger *slog.Logger) *OutputHandler {
	return &OutputHandler{
		emitter: emitter,
		logger:  logger,
	}
}

// StartOutputLoop 启动输出读取循环
func (h *OutputHandler) StartOutputLoop(session *Session) {
	buf := make([]byte, 1024)

	for {
		select {
		case <-session.Context().Done():
			// 收到退出信号
			return
		default:
			n, err := session.Pty.Read(buf)
			if err != nil {
				if err != io.EOF && session.Context().Err() == nil {
					// 只有在 context 未取消时才报告错误
					if h.logger != nil {
						h.logger.Error("读取 PTY 输出失败", "sessionId", session.ID, "error", err)
					}
					h.emitError(session.ID, err.Error())
				}
				return
			}

			// 使用过滤器处理输出
			result := session.Filter().Process(buf[:n])

			// 获取当前 block ID
			blockID := session.CurrentBlock()

			// 只有有过滤后输出时才发送
			if len(result.Output) > 0 {
				if h.logger != nil {
					h.logger.Info("提取过滤后终端输出", "text", string(result.Output))
				}
				h.emitOutput(session.ID, blockID, result.Output)
			}

			// 工作路径变化时发送事件
			if result.PwdChanged {
				h.emitPwdUpdate(session.ID, result.Pwd)
			}

			// 命令结束时发送事件
			if result.CommandEnded {
				h.emitCommandEnd(session.ID, blockID, result.ExitCode)
			}
		}
	}
}

// emitOutput 发送输出事件
func (h *OutputHandler) emitOutput(sessionID, blockID string, output []byte) {
	if h.emitter == nil {
		return
	}

	encoded := base64.StdEncoding.EncodeToString(output)
	h.emitter.Emit("terminal:output", map[string]interface{}{
		"sessionId": sessionID,
		"blockId":   blockID,
		"data":      encoded,
	})
}

// emitError 发送错误事件
func (h *OutputHandler) emitError(sessionID, message string) {
	if h.emitter == nil {
		return
	}

	h.emitter.Emit("terminal:error", map[string]interface{}{
		"sessionId": sessionID,
		"message":   message,
	})
}

// emitCommandEnd 发送命令结束事件
func (h *OutputHandler) emitCommandEnd(sessionID, blockID string, exitCode int) {
	if h.emitter == nil {
		return
	}

	h.emitter.Emit("terminal:command_end", map[string]interface{}{
		"sessionId": sessionID,
		"blockId":   blockID,
		"exitCode":  exitCode,
	})
}

// emitPwdUpdate 发送工作路径更新事件
func (h *OutputHandler) emitPwdUpdate(sessionID, pwd string) {
	if h.emitter == nil {
		return
	}

	h.emitter.Emit("terminal:pwd_update", map[string]interface{}{
		"sessionId": sessionID,
		"pwd":       pwd,
	})
}
