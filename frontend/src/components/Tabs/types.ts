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

import { TabState } from "@/store/tabs.store";

export interface TabProps {
  tab: TabState;
  isActive: boolean;
  onClose: (tabId: string) => void;
  onPin: (tabId: string) => void;
  onUnpin: (tabId: string) => void;
  onSelect: (tabId: string) => void;
}

export interface TabBarProps {
  tabs: TabState[];
  activeTabId: string | null;
  onTabSelect: (tabId: string) => void;
  onTabClose: (tabId: string) => void;
  onTabPin: (tabId: string) => void;
  onTabUnpin: (tabId: string) => void;
  onTabMove?: (fromIndex: number, toIndex: number) => void;
}

export interface TabContentProps {
  tabs: TabState[];
  activeTabId: string | null;
}

export interface TabContextMenuProps {
  tab: TabState;
  children?: React.ReactNode;
}
