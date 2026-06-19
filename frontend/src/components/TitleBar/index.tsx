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

import { FC, useEffect, useState } from "react";
import { System, Window } from "@wailsio/runtime";
import { AuthService, WindowService } from "@wails/service";
import { Button } from "../ui/button";
import { Bell, LogOut, Search } from "lucide-react";
import { Input } from "../ui/input";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "../ui/dropdown-menu";
import {
  getCurrentUserProfile,
  handleApiAuthError,
  type AuthMeUser,
} from "@/api/auth";
import { callWails } from "@/lib/utils";
import boxifyLogo from "../../../../boxify-logo-transparent.png";

const macControlButtonClass =
  "size-3 rounded-full transition-opacity hover:opacity-85";

// getUserStatusLabel 根据登录用户状态生成标题栏副标题。
function getUserStatusLabel(user: AuthMeUser | null): string {
  if (!user) {
    return "未登录";
  }
  if (user.is_active === false) {
    return "已停用";
  }
  return user.is_admin ? "管理员" : "普通用户";
}

// getAvatarInitial 返回用户名首字符，缺失时使用稳定兜底。
function getAvatarInitial(username?: string): string {
  const first = Array.from(username?.trim() ?? "")[0];
  return first ? first.toUpperCase() : "U";
}

const TitleBar: FC = () => {
  // 仅在 macOS 渲染窗口控制按钮，保持各平台原生习惯一致。
  const isMac = System.IsMac();
  const [currentUser, setCurrentUser] = useState<AuthMeUser | null>(null);
  const [avatarFailed, setAvatarFailed] = useState(false);
  const username = currentUser?.username?.trim() || "当前用户";
  const avatarUrl = avatarFailed ? "" : currentUser?.avatar_url?.trim() || "";
  const avatarInitial = getAvatarInitial(currentUser?.username);
  const userStatusLabel = getUserStatusLabel(currentUser);
  const providerLabel = currentUser?.oauth_provider?.trim();
  const userEmail = currentUser?.email?.trim();

  useEffect(() => {
    let disposed = false;

    // 标题栏读取远端用户信息失败时静默降级，认证过期则复用统一跳转。
    getCurrentUserProfile()
      .then((user) => {
        if (disposed) {
          return;
        }
        setCurrentUser(user);
        setAvatarFailed(false);
      })
      .catch(async (error) => {
        if (await handleApiAuthError(error)) {
          return;
        }
        console.warn("[Auth] 登录用户读取异常:", error);
      });

    return () => {
      disposed = true;
    };
  }, []);

  // 关闭当前窗口。
  const handleWindowClose = () => {
    Window.Close().catch((err) => {
      console.error("关闭窗口失败:", err);
    });
  };

  // 最小化当前窗口。
  const handleWindowMinimise = () => {
    Window.Minimise().catch((err) => {
      console.error("最小化窗口失败:", err);
    });
  };

  // 切换窗口最大化状态。
  const handleWindowToggleMaximise = () => {
    Window.ToggleMaximise().catch((err) => {
      console.error("切换窗口最大化失败:", err);
    });
  };

  // 退出当前登录状态，并切回登录窗口。
  const handleLogout = async () => {
    try {
      await callWails(AuthService.Logout);
      await callWails(WindowService.OpenPage, "login");
      await callWails(WindowService.ClosePage, "index");
    } catch {
      // callWails 已展示中文错误提示，这里只阻止未处理的异步异常。
    }
  };

  return (
    <header
      className="grid h-14 w-full shrink-0 cursor-default grid-cols-[minmax(210px,1fr)_minmax(260px,520px)_minmax(210px,1fr)] items-center border-b bg-background px-5 text-foreground"
      style={{ "--wails-draggable": "drag" } as React.CSSProperties}
    >
      <div
        className="flex min-w-0 items-center gap-4"
        style={{ "--wails-draggable": "no-drag" } as React.CSSProperties}
      >
        {isMac && (
          <div className="flex items-center gap-2">
            <button
              type="button"
              aria-label="关闭窗口"
              title="关闭"
              className={`${macControlButtonClass} bg-[#FF5F57]`}
              onClick={handleWindowClose}
            />
            <button
              type="button"
              aria-label="最小化窗口"
              title="最小化"
              className={`${macControlButtonClass} bg-[#FEBC2E]`}
              onClick={handleWindowMinimise}
            />
            <button
              type="button"
              aria-label="缩放窗口"
              title="缩放"
              className={`${macControlButtonClass} bg-[#28C840]`}
              onClick={handleWindowToggleMaximise}
            />
          </div>
        )}
        <div className="flex min-w-0 items-center gap-3" aria-label="Boxify">
          <div className="flex size-8 shrink-0 items-center justify-center rounded-lg bg-[#2f6df6] shadow-sm">
            <img
              src={boxifyLogo}
              alt=""
              aria-hidden="true"
              className="size-7 object-contain"
            />
          </div>
          <span className="truncate text-[18px] font-semibold leading-none tracking-normal text-foreground">
            Boxify
          </span>
        </div>
      </div>

      <div
        className="relative"
        style={{ "--wails-draggable": "no-drag" } as React.CSSProperties}
      >
        <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          aria-label="搜索"
          placeholder="搜索会话、任务..."
          className="h-9 rounded-md bg-background pl-9 text-sm shadow-xs"
        />
      </div>

      <div
        className="flex min-w-0 items-center justify-end gap-3"
        style={{ "--wails-draggable": "no-drag" } as React.CSSProperties}
      >
        <Button
          size="icon-sm"
          variant="ghost"
          className="relative"
          aria-label="通知"
        >
          <Bell />
          <span className="absolute right-2 top-2 size-1.5 rounded-full bg-destructive" />
        </Button>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <button
              type="button"
              className="flex min-w-0 items-center gap-3  border-l  pl-3 pr-1 outline-none transition-colors "
              aria-label="当前用户菜单"
              title={providerLabel ? `登录来源：${providerLabel}` : undefined}
            >
              <div className="flex size-8 shrink-0 items-center justify-center rounded-full bg-secondary text-xs font-semibold text-secondary-foreground">
                {avatarUrl ? (
                  <img
                    src={avatarUrl}
                    alt=""
                    aria-hidden="true"
                    className="size-8 rounded-full object-cover"
                    onError={() => setAvatarFailed(true)}
                  />
                ) : (
                  avatarInitial
                )}
              </div>
              <div className="hidden min-w-0 text-left sm:block">
                <div className="truncate text-xs font-semibold leading-none">
                  {username}
                </div>
                <div className="mt-1 truncate text-[11px] leading-none text-muted-foreground">
                  {userStatusLabel}
                </div>
              </div>
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-56">
            <DropdownMenuLabel className="min-w-0">
              <div className="truncate text-sm font-semibold">{username}</div>
              {userEmail ? (
                <div className="mt-1 truncate text-xs font-normal text-muted-foreground">
                  {userEmail}
                </div>
              ) : null}
              {providerLabel ? (
                <div className="mt-1 truncate text-xs font-normal text-muted-foreground">
                  {providerLabel}
                </div>
              ) : null}
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              variant="destructive"
              onClick={() => {
                void handleLogout();
              }}
            >
              <LogOut />
              <span>退出登录</span>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  );
};

export default TitleBar;
