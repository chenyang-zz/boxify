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

import { FC, ReactNode, useState } from "react";
import { Database, FolderOpen, List, SquareTerminal } from "lucide-react";
import { cn } from "@/lib/utils";
import { Separator } from "../ui/separator";
import { useAppStore } from "@/store/app.store";
import { useShallow } from "zustand/react/shallow";

enum UtilType {
  ALL = "all",
  TERMINAL = "terminal",
  DATABASE = "database",
}

interface UtilItemType {
  type: UtilType;
  icon: React.ReactNode;
}

const UtilItems: UtilItemType[] = [
  {
    type: UtilType.ALL,
    icon: <List className="size-5 " />,
  },
  {
    type: UtilType.TERMINAL,
    icon: <SquareTerminal className="size-5 " />,
  },
  {
    type: UtilType.DATABASE,
    icon: <Database className="size-5" />,
  },
];

interface UtilItemProps {
  icon: ReactNode;
  isActive: boolean;
  onClick: () => void;
}

const UtilItem: FC<UtilItemProps> = ({ icon, isActive, onClick }) => {
  return (
    <div
      className={cn(
        "inline-flex items-center justify-center hover:bg-accent hover:text-accent-foreground h-8 w-8 rounded-xl cursor-default",
        isActive &&
          "bg-primary text-primary-foreground hover:bg-primary/90 hover:text-primary-foreground",
      )}
      onClick={() => onClick()}
    >
      {icon}
    </div>
  );
};

const UtilBar: FC = () => {
  const isOpen = useAppStore(useShallow((state) => state.isPropertyOpen));
  const [activeTab, setActiveTab] = useState<UtilType>(UtilType.TERMINAL);

  return (
    <div className="p-1 h-full  gap-2 flex flex-col text-foreground">
      <UtilItem
        isActive={isOpen}
        icon={<FolderOpen className="size-5" />}
        onClick={() => {
          useAppStore.getState().setIsPropertyOpen(!isOpen);
        }}
      />
      <Separator />
      {UtilItems.map((item, index) => (
        <UtilItem
          key={index}
          icon={item.icon}
          isActive={activeTab === item.type}
          onClick={() => setActiveTab(item.type)}
        />
      ))}
    </div>
  );
};

export default UtilBar;
