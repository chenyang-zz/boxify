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

package service

import (
	"context"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// ServiceLifecycle 服务生命周期接口
type ServiceLifecycle interface {
	ServiceStartup(ctx context.Context, options application.ServiceOptions) error
	ServiceShutdown() error
}

// ServiceCleanup 可选的清理接口
type ServiceCleanup interface {
	Cleanup() error
}

// RegisterLifecycle 注册生命周期方法到服务（辅助函数）
func RegisterLifecycle(service any) {
	if lifecycle, ok := service.(ServiceLifecycle); ok {
		// Wails v3 会自动调用这些方法
		_ = lifecycle
	}
}
