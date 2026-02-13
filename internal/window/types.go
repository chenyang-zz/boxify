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

package window

// WindowType 窗口类型
type WindowType int

const (
	// WindowTypeMain 主窗口 - 应用程序主窗口
	WindowTypeMain WindowType = iota

	// WindowTypeSingleton 单例窗口 - 如设置窗口，只能打开一个实例
	WindowTypeSingleton

	// WindowTypeModal 模态窗口 - 阻塞父窗口的对话框
	WindowTypeModal
)

// ParseWindowType 解析窗口类型字符串
func ParseWindowType(typeStr string) WindowType {
	switch typeStr {
	case "main":
		return WindowTypeMain
	case "singleton":
		return WindowTypeSingleton
	case "modal":
		return WindowTypeModal
	default:
		return WindowTypeSingleton
	}
}
