import { Chrome, Ellipsis, Github } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import boxifyLogo from "../../../../boxify-logo-transparent.png";

type LoginProvider = "github" | "google" | "other";

const loginProviderLabels: Record<LoginProvider, string> = {
  github: "GitHub",
  google: "Google",
  other: "其他登录",
};

// 处理登录入口点击，后续接入 OAuth 时替换为真实授权流程。
function handleLogin(provider: LoginProvider) {
  toast.info(`${loginProviderLabels[provider]} 登录即将开放`);
}

// 渲染 Boxify 独立登录页，提供 GitHub、Google 和其他登录入口。
function LoginPage() {
  return (
    <main className="flex h-screen w-screen items-center justify-center overflow-hidden bg-[#f6f8fc] px-8 py-10 text-[#111827]">
      <section className="flex w-full max-w-[340px] flex-col items-center">
        <div className="mb-9 flex size-16 items-center justify-center rounded-[18px] bg-white shadow-[0_18px_45px_rgba(31,41,55,0.12)] ring-1 ring-[#e6eaf2]">
          <img src={boxifyLogo} alt="Boxify" className="size-14 object-contain" />
        </div>

        <h1 className="mb-4 text-center text-[28px] font-semibold leading-tight tracking-normal text-[#111827]">
          欢迎使用 Boxify
        </h1>

        <div className="mb-8 inline-flex h-7 items-center rounded-full bg-[#e8efff] px-4 text-sm font-medium text-[#3f5fb5] ring-1 ring-[#d7e2ff]">
          <span className="mr-2 text-[#5574d8]">✓</span>
          支持 GitHub / Google 登录
        </div>

        <div className="flex w-full flex-col gap-3">
          <Button
            type="button"
            className="h-12 rounded-full bg-[#111827] text-[15px] font-semibold text-white shadow-[0_12px_28px_rgba(17,24,39,0.16)] hover:bg-[#1f2937]"
            onClick={() => handleLogin("github")}
          >
            <Github data-icon="inline-start" />
            使用 GitHub 登录
          </Button>

          <Button
            type="button"
            variant="outline"
            className="h-12 rounded-full border-[#d7dce6] bg-white text-[15px] font-semibold text-[#111827] shadow-none hover:bg-[#eef3fb] hover:text-[#111827]"
            onClick={() => handleLogin("google")}
          >
            <Chrome data-icon="inline-start" />
            使用 Google 登录
          </Button>

          <Button
            type="button"
            variant="ghost"
            className="h-11 rounded-full text-[15px] font-medium text-[#64748b] hover:bg-[#eef3fb] hover:text-[#111827]"
            onClick={() => handleLogin("other")}
          >
            <Ellipsis data-icon="inline-start" />
            其他登录
          </Button>
        </div>
      </section>
    </main>
  );
}

export default LoginPage;
