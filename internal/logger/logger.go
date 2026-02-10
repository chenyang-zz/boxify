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

package logger

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/chenyang-zz/boxify/internal/utils"
)

const (
	envLogDir  = "Boxify_LOG_DIR"
	appDirName = "boxify"

	logFileName         = "boxify.log"
	logRotateMaxBytes   = 10 * 1024 * 1024 // 10 MB
	logRotateMaxBackups = 10               // 保留最近的10个日志文件，超过这个数量的旧日志将被删除
)

var (
	once    sync.Once
	logMu   sync.Mutex
	logInst *log.Logger
	logFile *os.File
	logPath string
)

// Init 初始化日志系统，确保只执行一次。根据环境变量或系统默认路径创建日志文件，并设置日志输出。
// 日志文件会自动轮转和清理旧日志。
func Init(ctx context.Context) {
	once.Do(func() {
		path, out := initOutput(ctx)
		logMu.Lock()
		defer logMu.Unlock()
		logPath = path
		logInst = log.New(out, "", log.Ldate|log.Ltime|log.Lmicroseconds)
		logInst.Printf("[INFO] 日志初始化完成，日志文件：%s", logPath)
	})
}

// Path 返回当前日志文件的路径。如果日志系统尚未初始化，则会先进行初始化。
// 返回的路径可能是环境变量指定的目录、系统默认配置目录，或者临时目录中的日志文件路径。
func Path() string {
	Init(context.Background())
	logMu.Lock()
	defer logMu.Unlock()
	return logPath
}

// Close 关闭日志系统，释放资源。会将日志输出重置为标准错误，并关闭日志文件。
// 如果日志系统尚未初始化，则会先进行初始化。调用此函数后，日志文件将不再被写入，并且相关资源将被清理。
func Close() {
	Init(context.Background())
	logMu.Lock()
	defer logMu.Unlock()
	if logInst != nil {
		logInst.SetOutput(os.Stderr)
	}
	if logFile != nil {
		_ = logFile.Close()
		logFile = nil
	}
}

// Infof 输出信息级别的日志消息，格式化字符串和参数列表。
// 调用此函数会将日志消息以“信息”级别输出到日志文件中。
// 如果日志系统尚未初始化，则会先进行初始化。
func Infof(format string, args ...any) {
	printf("INFO", format, args...)
}

// Warnf 输出警告级别的日志消息，格式化字符串和参数列表。
// 调用此函数会将日志消息以“警告”级别输出到日志文件中。
// 如果日志系统尚未初始化，则会先进行初始化。
func Warnf(format string, args ...any) {
	printf("WARN", format, args...)
}

// Errorf 输出错误级别的日志消息，格式化字符串和参数列表。
// 调用此函数会将日志消息以“错误”级别输出到日志文件中。
// 如果日志系统尚未初始化，则会先进行初始化。
func Errorf(format string, args ...any) {
	printf("ERROR", format, args...)
}

// ErrorfWithTrace 输出错误级别的日志消息，包含错误链信息。
// 接受一个错误对象、格式字符串和参数列表。
// 函数会将格式化后的消息与错误链信息一起输出到日志文件中。
// 如果错误对象为nil，则只输出格式化后的消息；
// 否则，会调用ErrorChain函数获取错误链的字符串表示，并将其附加到日志消息中。
// 调用此函数会将日志消息以“错误”级别输出到日志文件中。
// 如果日志系统尚未初始化，则会先进行初始化。
func ErrorfWithTrace(err error, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if err == nil {
		Errorf("%s", msg)
		return
	}
	Errorf("%s；错误链：%s", msg, ErrorChain(err))
}

// ErrorChain 返回错误链的字符串表示，包含所有独特的错误消息，按顺序连接。
// 对于每个错误，都会调用Error()方法获取其消息，并将它们连接成一个字符串。
// 函数会避免重复的错误消息，并且如果错误链过长（超过20层），会在末尾添加提示信息表示已截断。
func ErrorChain(err error) string {
	if err == nil {
		return "<nil>"
	}

	var parts []string
	seen := make(map[string]struct{})
	cur := err
	truncated := false
	for i := 0; cur != nil && i < 20; i++ {
		s := cur.Error()
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			parts = append(parts, s)
		}
		cur = errors.Unwrap(cur)
	}
	if cur != nil {
		truncated = true
	}

	if len(parts) == 0 {
		return err.Error()
	}

	if truncated {
		parts = append(parts, "（错误链过长，已截断）")
	}
	return strings.Join(parts, " -> ")
}

// printf 是一个内部函数，用于格式化日志消息并输出到日志文件。
// 它接受日志级别、格式字符串和可变参数列表。
// 函数首先确保日志系统已初始化，然后获取当前的日志实例，并使用指定的格式输出日志消息。
// 如果日志实例不可用，则函数会直接返回，不进行任何操作。
func printf(level string, format string, args ...any) {
	Init(context.Background())
	logMu.Lock()
	inst := logInst
	logMu.Unlock()
	if inst == nil {
		return
	}
	inst.Printf("[%s] %s", level, fmt.Sprintf(format, args...))
}

// initOutput 初始化日志输出，优先使用环境变量指定的目录，如果不可用则使用系统默认配置目录，最后退回到临时目录。
// 返回日志文件路径和对应的io.Writer。
func initOutput(ctx context.Context) (string, io.Writer) {
	dir := strings.TrimSpace(os.Getenv(envLogDir))

	if dir == "" {
		base, err := os.UserConfigDir()
		if err != nil || strings.TrimSpace(base) == "" {
			base = os.TempDir()
		}
		dir = filepath.Join(base, appDirName, "logs")
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return filepath.Join(dir, logFileName), os.Stderr
	}

	path := filepath.Join(dir, logFileName)
	rotateIfNeeded(path, dir)

	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return path, os.Stderr
	}

	logFile = f
	writer := io.Writer(f)

	if utils.IsDev(ctx) {
		writer = io.MultiWriter(os.Stdout, f)
	}
	return path, writer
}

// rotateIfNeeded 检查日志文件大小，如果超过限制则进行轮转，并清理旧日志。
func rotateIfNeeded(path, dir string) {
	fi, err := os.Stat(path)
	if err != nil || fi.IsDir() {
		return
	}

	if fi.Size() < logRotateMaxBytes {
		return
	}

	ts := time.Now().Format("20060102-150405")
	rotated := filepath.Join(dir, "boxify-"+ts+".log")
	if err = os.Rename(path, rotated); err != nil {
		return
	}
	cleanupOldLogs(dir)
}

// cleanupOldLogs 删除超过保留数量的旧日志文件，保留最新的logRotateMaxBackups个日志。
func cleanupOldLogs(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	type item struct {
		name string
		path string
	}
	var logs []item
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasPrefix(name, "boxify-") || !strings.HasSuffix(name, ".log") {
			continue
		}

		logs = append(logs, item{name: name, path: filepath.Join(dir, name)})
	}

	sort.Slice(logs, func(i, j int) bool {
		return logs[i].name > logs[j].name
	})
	if len(logs) <= logRotateMaxBackups {
		return
	}
	for _, it := range logs[logRotateMaxBackups:] {
		_ = os.Remove(it.path)
	}
}
