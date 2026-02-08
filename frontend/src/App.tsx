import { useState } from "react";
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

function App() {
  const isOpen = useAppStore(useShallow((state) => state.isPropertyOpen));
  return (
    <TooltipProvider>
      <div id="App" className="h-screen w-screen bg-background flex flex-col">
        <TitleBar />
        <div className="flex-1 flex">
          <UtilBar />
          <main className="w-full h-full flex flex-1 pb-2">
            <ResizablePanelGroup orientation="horizontal">
              {isOpen && (
                <>
                  <ResizablePanel defaultSize="200px">
                    <PropertyTree />
                  </ResizablePanel>
                  <ResizableHandle />
                </>
              )}

              <ResizablePanel></ResizablePanel>
            </ResizablePanelGroup>
          </main>
        </div>
      </div>
    </TooltipProvider>
  );
}

export default App;
