import { useEffect, useRef, useState } from "react";
import { Chrome, Github } from "lucide-react";
import { Events, System, Window } from "@wailsio/runtime";
import {
  AuthOAuthCompletedEvent,
  AuthService,
  WindowService,
} from "@wails/service";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { useInitialData } from "@/hooks/useInitialData";
import { callWails, callWailsWithOptions, currentPageId } from "@/lib/utils";
import boxifyLogo from "../../../../boxify-logo-transparent.png";

type LoginProvider = "github" | "google" | "other";

interface LoginInitialData {
  reason?: string;
}

const loginProviderLabels: Record<LoginProvider, string> = {
  github: "GitHub",
  google: "Google",
  other: "其他登录",
};

const macControlButtonClass =
  "size-3 rounded-full transition-opacity hover:opacity-85";

// OAuth 登录成功后打开主窗口，并关闭当前登录窗口。
async function openMainAfterOAuthLogin() {
  await callWails(WindowService.OpenPage, "index");
  await callWails(WindowService.ClosePage, currentPageId());
}

// 兼容 Wails 事件 payload 包装格式，提取后端发送的数据。
function toAuthOAuthCompletedEvent(
  event: unknown,
): AuthOAuthCompletedEvent | null {
  const maybeWrapped = event as { data?: AuthOAuthCompletedEvent };
  if (maybeWrapped?.data && typeof maybeWrapped.data.success === "boolean") {
    return maybeWrapped.data;
  }
  const maybePayload = event as AuthOAuthCompletedEvent;
  if (typeof maybePayload?.success === "boolean") {
    return maybePayload;
  }
  return null;
}

// 渲染登录窗口左上角原生风格操作按钮。
function LoginWindowControls() {
  const isMac = System.IsMac();

  if (!isMac) {
    return null;
  }

  // 关闭登录窗口。
  const handleClose = () => {
    Window.Close().catch((err) => {
      console.error("关闭登录窗口失败:", err);
    });
  };

  // 最小化登录窗口。
  const handleMinimise = () => {
    Window.Minimise().catch((err) => {
      console.error("最小化登录窗口失败:", err);
    });
  };

  // 切换登录窗口最大化状态。
  const handleToggleMaximise = () => {
    Window.ToggleMaximise().catch((err) => {
      console.error("切换登录窗口缩放失败:", err);
    });
  };

  return (
    <div
      className="absolute left-5 top-5 z-10 flex items-center gap-2"
      style={{ "--wails-draggable": "no-drag" } as React.CSSProperties}
    >
      <button
        type="button"
        aria-label="关闭窗口"
        title="关闭"
        className={`${macControlButtonClass} bg-[#FF5F57]`}
        onClick={handleClose}
      />
      <button
        type="button"
        aria-label="最小化窗口"
        title="最小化"
        className={`${macControlButtonClass} bg-[#FEBC2E]`}
        onClick={handleMinimise}
      />
      <button
        type="button"
        aria-label="缩放窗口"
        title="缩放"
        className={`${macControlButtonClass} bg-[#28C840]`}
        onClick={handleToggleMaximise}
      />
    </div>
  );
}

// 渲染 Boxify 独立登录页，提供 GitHub、Google 和其他登录入口。
function LoginPage() {
  const { initialData } = useInitialData<LoginInitialData>();
  const [pendingProvider, setPendingProvider] = useState<LoginProvider | null>(
    null,
  );
  const pendingProviderRef = useRef<LoginProvider | null>(null);
  const reasonToastRef = useRef("");
  const loginReason =
    typeof initialData?.data?.reason === "string"
      ? initialData.data.reason.trim()
      : "";

  // 同步最新登录等待态，避免取消后处理迟到的 OAuth 完成事件。
  const updatePendingProvider = (provider: LoginProvider | null) => {
    pendingProviderRef.current = provider;
    setPendingProvider(provider);
  };

  useEffect(() => {
    // 监听后端 Deep Link 登录完成事件，登录窗口只负责收尾跳转。
    const unbind = Events.On("auth:oauth-completed", (event: unknown) => {
      const payload = toAuthOAuthCompletedEvent(event);
      if (!payload || (payload.provider && payload.provider !== "github")) {
        return;
      }
      if (pendingProviderRef.current !== "github") {
        return;
      }

      if (!payload.success) {
        toast.error("GitHub 登录失败", {
          description: payload.message || "请重新尝试",
          style: { textAlign: "left" },
        });
        updatePendingProvider(null);
        return;
      }

      void openMainAfterOAuthLogin().finally(() => {
        updatePendingProvider(null);
      });
    });

    return () => {
      unbind();
    };
  }, []);

  useEffect(() => {
    if (!loginReason || reasonToastRef.current === loginReason) {
      return;
    }
    reasonToastRef.current = loginReason;
    toast.warning("需要重新登录", {
      description: loginReason,
      style: { textAlign: "left" },
    });
  }, [loginReason]);

  // 处理登录按钮点击，GitHub 走真实 OAuth，其他入口保留占位提示。
  const handleLogin = async (provider: LoginProvider) => {
    if (pendingProvider) {
      return;
    }

    if (provider !== "github") {
      toast.info("该登录方式暂未开放");
      return;
    }

    updatePendingProvider(provider);
    try {
      await callWailsWithOptions(AuthService.StartOAuthLogin, ["github"], {
        timeoutMs: 30000,
        timeoutMessage: "打开 GitHub 授权页超时，请重新尝试",
      });
    } catch {
      updatePendingProvider(null);
    }
  };

  // 取消当前浏览器授权等待，并回到初始登录页。
  const handleCancelLogin = () => {
    updatePendingProvider(null);
  };

  if (pendingProvider === "github") {
    return (
      <main
        className="relative flex h-screen w-screen items-center justify-center overflow-hidden bg-background px-8 py-10 text-foreground"
        style={{ "--wails-draggable": "drag" } as React.CSSProperties}
      >
        <LoginWindowControls />
        <section
          className="flex w-full max-w-[340px] flex-col items-center"
          style={{ "--wails-draggable": "no-drag" } as React.CSSProperties}
        >
          <div className="mb-9 flex size-16 items-center justify-center rounded-[18px] bg-card shadow-xl ring-1 ring-border">
            <img
              src={boxifyLogo}
              alt="Boxify"
              className="size-14 object-contain"
            />
          </div>

          <p className="mb-8 text-center text-[15px] font-medium leading-none text-muted-foreground">
            请继续在浏览器中登录
          </p>

          <Button
            type="button"
            variant="outline"
            className="h-11 w-full rounded-full text-[15px] font-semibold"
            onClick={handleCancelLogin}
          >
            取消登录
          </Button>
        </section>
      </main>
    );
  }

  return (
    <main
      className="relative flex h-screen w-screen items-center justify-center overflow-hidden bg-background px-8 py-10 text-foreground"
      style={{ "--wails-draggable": "drag" } as React.CSSProperties}
    >
      <LoginWindowControls />
      <section
        className="flex w-full max-w-[340px] flex-col items-center"
        style={{ "--wails-draggable": "no-drag" } as React.CSSProperties}
      >
        <div className="mb-9 flex size-16 items-center justify-center rounded-[18px] bg-card shadow-xl ring-1 ring-border">
          <img
            src={boxifyLogo}
            alt="Boxify"
            className="size-14 object-contain"
          />
        </div>

        <h1 className="mb-4 text-center text-[28px] font-semibold leading-tight tracking-normal text-foreground">
          欢迎使用 Boxify
        </h1>

        <div className="mb-8 inline-flex h-7 items-center rounded-full bg-primary/10 px-4 text-sm font-medium text-primary ring-1 ring-primary/20">
          <span className="mr-2">✓</span>
          支持 GitHub / Google 登录
        </div>

        {loginReason ? (
          <div className="mb-5 w-full rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-center text-sm font-medium text-destructive">
            {loginReason}
          </div>
        ) : null}

        <div className="flex w-full flex-col gap-3">
          <Button
            type="button"
            disabled={pendingProvider !== null}
            className="h-12 rounded-full bg-primary text-[15px] font-semibold text-primary-foreground shadow-lg hover:bg-primary/90"
            onClick={() => handleLogin("github")}
          >
            <Github data-icon="inline-start" />
            使用 GitHub 登录
          </Button>

          <Button
            type="button"
            variant="outline"
            disabled={pendingProvider !== null}
            className="h-12 rounded-full text-[15px] font-semibold"
            onClick={() => handleLogin("google")}
          >
            <Chrome data-icon="inline-start" />
            {pendingProvider === "google" ? "正在登录..." : "使用 Google 登录"}
          </Button>

          <Button
            type="button"
            variant="link"
            disabled={pendingProvider !== null}
            className="h-11 rounded-full text-[15px] font-medium"
            onClick={() => handleLogin("other")}
          >
            {pendingProvider === "other"
              ? "正在登录..."
              : loginProviderLabels.other}
          </Button>
        </div>
      </section>
    </main>
  );
}

export default LoginPage;
