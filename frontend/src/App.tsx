import UtilBar from "@/components/UtilBar";
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from "./components/ui/resizable";
import TitleBar from "./components/TitleBar";
import PropertyTree from "./components/PropertyTree";
import { TooltipProvider } from "./components/ui/tooltip";
import { useShallow } from "zustand/react/shallow";
import { useAppStore } from "./store/app.store";
import { Toaster } from "@/components/ui/sonner";
import Tabs from "./components/Tabs";

function App() {
  const isOpen = useAppStore(useShallow((state) => state.isPropertyOpen));
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

export default App;
