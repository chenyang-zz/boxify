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

import { useMemo } from "react";
import { useEventStore } from "@/store/event.store";
import { EventType } from "@wails/events/models";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  ChevronDownIcon,
  ExpandIcon,
  FileCode2Icon,
  GitBranchIcon,
  RotateCcwIcon,
  XIcon,
} from "lucide-react";
import { useTerminalStore } from "../store/terminal.store";

interface GitReviewPanelProps {
  sessionId: string;
}

function formatFileStatus(indexStatus: string, workTreeStatus: string): string {
  const toReadable = (code: string): string => {
    switch (code) {
      case "M":
        return "Modified";
      case "A":
        return "Added";
      case "D":
        return "Deleted";
      case "R":
        return "Renamed";
      case "C":
        return "Copied";
      case "U":
        return "Unmerged";
      case "?":
        return "Untracked";
      default:
        return "";
    }
  };

  const idx = toReadable(indexStatus);
  const wt = toReadable(workTreeStatus);

  if (idx && wt && idx !== wt) {
    return `${idx}/${wt}`;
  }

  return idx || wt || "Changed";
}

export function GitReviewPanel({ sessionId }: GitReviewPanelProps) {
  const closeReviewPanel = useTerminalStore((state) => state.closeReviewPanel);
  const gitStatusEvent = useEventStore(
    (state) => state.latestEvents[EventType.EventTypeGitStatusChanged],
  );

  const files = useMemo(
    () => gitStatusEvent?.data.status.files ?? [],
    [gitStatusEvent],
  );

  const handleClose = () => {
    closeReviewPanel(sessionId);
  };

  return (
    <div className="h-full w-full flex flex-col bg-card text-card-foreground border-l">
      <div className="flex items-center justify-between px-3 py-2 border-b">
        <h3 className="text-sm font-medium">Code review</h3>
        <div className="flex items-center gap-1">
          <Button variant="ghost" size="icon" className="h-7 w-7" disabled>
            <ExpandIcon className="size-4" />
          </Button>
          <Button variant="ghost" size="icon" className="h-7 w-7" onClick={handleClose}>
            <XIcon className="size-4" />
          </Button>
        </div>
      </div>

      <div className="px-3 py-2 border-b flex items-center gap-1.5 text-xs text-muted-foreground">
        <GitBranchIcon className="size-3.5" />
        <span>{gitStatusEvent?.data.status.head || "No branch"}</span>
        <span className="mx-1">|</span>
        <span>{files.length} files</span>
        <span className="text-green-500">+{gitStatusEvent?.data.status.addedLines ?? 0}</span>
        <span className="text-red-500">-{gitStatusEvent?.data.status.deletedLines ?? 0}</span>
      </div>

      <div className="px-3 py-2 border-b flex items-center justify-between">
        <Button
          variant="ghost"
          className="h-7 px-2 text-xs text-muted-foreground"
        >
          Uncommitted changes
          <ChevronDownIcon className="ml-1 size-3.5" />
        </Button>
        <Button variant="ghost" className="h-7 px-2 text-xs text-muted-foreground" disabled>
          <RotateCcwIcon className="size-3.5 mr-1" />
          Discard all
        </Button>
      </div>

      <div className="flex-1 overflow-auto p-2 space-y-2">
        {files.length === 0 && (
          <div className="h-full flex items-center justify-center text-xs text-muted-foreground">
            Working tree clean
          </div>
        )}

        {files.map((file) => (
          <div
            key={`${file.path}-${file.indexStatus}-${file.workTreeStatus}`}
            className="rounded-md border bg-background/60"
          >
            <div className="px-2 py-1.5 border-b flex items-center justify-between gap-2">
              <div className="min-w-0 flex items-center gap-1.5 text-xs font-medium">
                <FileCode2Icon className="size-3.5 shrink-0" />
                <span className="truncate">{file.path}</span>
              </div>
              <Badge variant="secondary" className="text-[10px]">
                {formatFileStatus(file.indexStatus, file.workTreeStatus)}
              </Badge>
            </div>
            <div className="px-2 py-1.5 text-[11px] text-muted-foreground">
              XY: {file.indexStatus || "-"}{file.workTreeStatus || "-"}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
