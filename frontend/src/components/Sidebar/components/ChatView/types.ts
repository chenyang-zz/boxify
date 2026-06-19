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

import type { ElementType, ReactNode } from "react";
import type {
  ListSessionItem,
  SidebarProjectItem,
} from "@/types/api/session";

export type PinnedItem =
  | { kind: "session"; item: ListSessionItem }
  | { kind: "project"; item: SidebarProjectItem };

export type DeleteTarget =
  | { kind: "session"; item: ListSessionItem }
  | { kind: "project"; item: SidebarProjectItem };

export interface ActionRowProps {
  icon: ElementType;
  label: string;
  disabled?: boolean;
  onClick?: () => void;
}

export interface ChildrenProps {
  children: ReactNode;
}

export interface ProjectBlockProps {
  project: SidebarProjectItem;
  expanded: boolean;
  selectedSessionId: string;
  onToggle: () => void;
  onSelectSession: (sessionId: string) => void;
  onRequestDeleteProject: (project: SidebarProjectItem) => void;
  onRequestDeleteSession: (session: ListSessionItem) => void;
}

export interface SidebarItemContextMenuProps {
  children: ReactNode;
  onDelete: () => void;
  type?: "session" | "project";
  session?: ListSessionItem;
}

export interface ChatMoveContextValue {
  projects: SidebarProjectItem[];
  onMoveSessionToProject: (
    session: ListSessionItem,
    targetProjectId: string,
  ) => void;
}
