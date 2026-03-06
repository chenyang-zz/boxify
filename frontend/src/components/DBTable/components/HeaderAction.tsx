import { FC } from "react";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import {
  ArrowUpNarrowWideIcon,
  DownloadIcon,
  FileUpIcon,
  FunnelPlusIcon,
  MinusIcon,
  PlusIcon,
  Redo2Icon,
  RefreshCcwIcon,
  SaveIcon,
  TimerResetIcon,
  Undo2Icon,
} from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import type { DBTableActionState } from "../types";

interface HeaderActionProps {
  state: DBTableActionState;
  onToggleTransaction: () => void;
  onRefresh: () => void;
  onAddRow: () => void;
  onDeleteRows: () => void;
  onSave: () => void;
  onUndo: () => void;
  onRedo: () => void;
  onToggleFilter: () => void;
  onSort: () => void;
  onImport: () => void;
  onExport: (format: "csv" | "json" | "md") => void;
}

// 表头操作区：只负责按钮展示与事件透传。
const HeaderAction: FC<HeaderActionProps> = ({
  state,
  onToggleTransaction,
  onRefresh,
  onAddRow,
  onDeleteRows,
  onSave,
  onUndo,
  onRedo,
  onToggleFilter,
  onSort,
  onImport,
  onExport,
}) => {
  const sortTip =
    state.sortState.direction === "none"
      ? "未排序"
      : `按列排序：${state.sortState.direction === "asc" ? "升序" : "降序"}`;

  return (
    <header className="shrink-0 px-2 py-1 flex items-center gap-0.5 shadow shadow-background">
      <Button
        variant={state.inTransaction ? "secondary" : "ghost"}
        size="sm"
        onClick={onToggleTransaction}
        disabled={state.pending}
      >
        <TimerResetIcon className="size-4" />
        {state.inTransaction ? "结束事务" : "开始事务"}
      </Button>
      <Separator orientation="vertical" />
      <Button
        className="rounded-full"
        size="icon-sm"
        variant="ghost"
        onClick={onRefresh}
        disabled={state.pending}
        title="重新加载数据"
      >
        <RefreshCcwIcon />
      </Button>
      <Button
        className="rounded-full"
        size="icon-sm"
        variant="ghost"
        onClick={onAddRow}
        disabled={state.pending}
        title="新增行"
      >
        <PlusIcon />
      </Button>
      <Button
        className="rounded-full"
        size="icon-sm"
        variant="ghost"
        onClick={onDeleteRows}
        disabled={state.pending || !state.hasSelection}
        title="删除/恢复选中行"
      >
        <MinusIcon />
      </Button>
      <Button
        className="rounded-full"
        size="icon-sm"
        variant={state.dirty ? "secondary" : "ghost"}
        onClick={onSave}
        disabled={state.pending || !state.dirty}
        title="保存变更"
      >
        <SaveIcon />
      </Button>
      <Button
        className="rounded-full"
        size="icon-sm"
        variant="ghost"
        onClick={onUndo}
        disabled={state.pending || !state.canUndo}
        title="撤销"
      >
        <Undo2Icon />
      </Button>
      <Separator orientation="vertical" />
      <Button
        className="rounded-full"
        size="icon-sm"
        variant="ghost"
        onClick={onRedo}
        disabled={state.pending || !state.canRedo}
        title="重做"
      >
        <Redo2Icon />
      </Button>
      <Button
        className="rounded-full"
        size="icon-sm"
        variant={state.showFilterInput ? "secondary" : "ghost"}
        onClick={onToggleFilter}
        disabled={state.pending}
        title="筛选"
      >
        <FunnelPlusIcon />
      </Button>
      <Button
        className="rounded-full"
        size="icon-sm"
        variant={state.sortState.direction === "none" ? "ghost" : "secondary"}
        onClick={onSort}
        disabled={state.pending}
        title={sortTip}
      >
        <ArrowUpNarrowWideIcon />
      </Button>
      <Separator orientation="vertical" />
      <Button
        className="rounded-full"
        size="icon-sm"
        variant="ghost"
        onClick={onImport}
        disabled={state.pending}
        title="导入 CSV/JSON"
      >
        <FileUpIcon />
      </Button>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            className="rounded-full"
            size="icon-sm"
            variant="ghost"
            disabled={state.pending}
            title="导出"
          >
            <DownloadIcon />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent>
          <DropdownMenuItem onClick={() => onExport("csv")}>导出 CSV</DropdownMenuItem>
          <DropdownMenuItem onClick={() => onExport("json")}>导出 JSON</DropdownMenuItem>
          <DropdownMenuItem onClick={() => onExport("md")}>导出 Markdown</DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </header>
  );
};

export default HeaderAction;
