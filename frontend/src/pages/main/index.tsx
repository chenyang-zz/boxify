import UtilBar from "@/components/UtilBar";
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from "@/components/ui/resizable";
import TitleBar from "@/components/TitleBar";
import PropertyTree from "@/components/PropertyTree";
import { TooltipProvider } from "@/components/ui/tooltip";
import { useShallow } from "zustand/react/shallow";
import { useAppStore } from "@/store/app.store";
import { Toaster } from "@/components/ui/sonner";
import Tabs from "@/components/Tabs";
import { WindowService } from "@wails/service";
import { callWails } from "@/lib/utils";
import { useDataSync } from "@/hooks/useDataSync";
import { DataChannel } from "@/store/data-sync.store";
import { useOpenWindowWithData } from "@/hooks/useOpenWindowWithData";

function MainApp() {
  const isOpen = useAppStore(useShallow((state) => state.isPropertyOpen));

  useDataSync(DataChannel.Settings, (event) => {
    if (event.dataType === "settings:update") {
      console.log("[设置窗口] 收到配置更新", event.data);
    }
  });

  const { openWindowWithData } = useOpenWindowWithData();

  return (
    <TooltipProvider>
      <div
        id="App"
        className="h-screen w-screen bg-background flex flex-col overflow-hidden text-foreground"
      >
        <TitleBar />
        <div className="flex-1 flex overflow-hidden">
          <UtilBar />
          <main className="h-full flex flex-1 pb-2 overflow-hidden">
            <ResizablePanelGroup orientation="horizontal">
              {isOpen && (
                <>
                  <ResizablePanel defaultSize="250px" maxSize="400px">
                    <PropertyTree />
                  </ResizablePanel>
                  <ResizableHandle className=" opacity-0" />
                </>
              )}
              <ResizablePanel className="pl-1 pr-2">
                <div
                  onClick={() =>
                    openWindowWithData("settings", { title: "设置3" })
                  }
                >
                  打开窗口
                </div>
                <Tabs />
              </ResizablePanel>
            </ResizablePanelGroup>
          </main>
        </div>
        <Toaster />
      </div>
    </TooltipProvider>
  );
}

export default MainApp;
