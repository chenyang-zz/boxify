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
import Tabs from "@/components/Tabs";
import Sidebar, { useActiveView } from "@/components/Sidebar";
import ClawContent from "@/components/ClawContent";

function MainApp() {
  const isOpen = useAppStore(useShallow((state) => state.isPropertyOpen));
  const activeView = useActiveView();

  return (
    <TooltipProvider>
      <div
        id="App"
        className="h-screen w-screen bg-background flex flex-col overflow-hidden text-foreground"
      >
        <TitleBar />
        <main className="h-full flex flex-1 overflow-hidden">
          <ResizablePanelGroup orientation="horizontal">
            {isOpen && (
              <>
                <ResizablePanel
                  defaultSize="250px"
                  minSize="250px"
                  maxSize="70%"
                  className="border-r"
                >
                  <Sidebar />
                </ResizablePanel>
                <ResizableHandle className=" opacity-0" />
              </>
            )}
            <ResizablePanel>
              {activeView === "control" ? <ClawContent /> : <Tabs />}
            </ResizablePanel>
          </ResizablePanelGroup>
        </main>
      </div>
    </TooltipProvider>
  );
}

export default MainApp;
