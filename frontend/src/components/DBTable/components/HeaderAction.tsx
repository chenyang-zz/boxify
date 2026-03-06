import { FC, useCallback, useState } from "react";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import {
  ArrowUpNarrowWideIcon,
  CheckIcon,
  DownloadIcon,
  FileUpIcon,
  FunnelPlusIcon,
  MinusIcon,
  PlusIcon,
  RefreshCcwIcon,
  RotateCcwIcon,
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
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import type { DBTableActionState } from "../types";

interface HeaderActionProps {
  state: DBTableActionState;
  onStartTransaction: () => void;
  onSaveTransaction: () => void;
  onCommitTransaction: () => void;
  onRollbackTransaction: () => void;
  onRefresh: () => void | Promise<void>;
  onAddRow: () => void;
  onDeleteRows: () => void;
  onUndo: () => void;
  onToggleFilter: () => void;
  onSort: () => void;
  onImport: () => void;
  onExport: (format: "csv" | "json" | "md") => void;
}

// 表头操作区：只负责按钮展示与事件透传。
const HeaderAction: FC<HeaderActionProps> = ({
  state,
  onStartTransaction,
  onSaveTransaction,
  onCommitTransaction,
  onRollbackTransaction,
  onRefresh,
  onAddRow,
  onDeleteRows,
  onUndo,
  onToggleFilter,
  onSort,
  onImport,
  onExport,
}) => {
  const [refreshConfirmOpen, setRefreshConfirmOpen] = useState(false);

  // 处理刷新点击：有未保存变更时要求用户二次确认。
  const handleRefreshClick = useCallback(() => {
    if (state.dirty) {
      setRefreshConfirmOpen(true);
      return;
    }
    void onRefresh();
  }, [onRefresh, state.dirty]);

  // 执行最终刷新动作并关闭确认弹窗。
  const handleConfirmRefresh = useCallback(() => {
    setRefreshConfirmOpen(false);
    void onRefresh();
  }, [onRefresh]);

  const sortTip =
    state.sortState.direction === "none"
      ? "未排序"
      : `按列排序：${state.sortState.direction === "asc" ? "升序" : "降序"}`;

  return (
    <>
      <header className="shrink-0 px-2 py-1 flex items-center gap-0.5 shadow shadow-background">
        {state.inTransaction ? (
          <>
            <Button
              variant={state.dirty ? "secondary" : "ghost"}
              size="sm"
              onClick={onCommitTransaction}
              disabled={state.pending}
            >
              <CheckIcon className="size-4" />
              提交
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={onRollbackTransaction}
              disabled={state.pending}
            >
              <RotateCcwIcon className="size-4" />
              回退
            </Button>
          </>
        ) : (
          <Button
            variant="ghost"
            size="sm"
            onClick={onStartTransaction}
            disabled={state.pending}
          >
            <TimerResetIcon className="size-4" />
            开始事务
          </Button>
        )}
        <Separator orientation="vertical" />
        <Button
          className="rounded-full"
          size="icon-sm"
          variant="ghost"
          onClick={handleRefreshClick}
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
          title="删除行"
        >
          <MinusIcon />
        </Button>
        <Button
          className="rounded-full"
          size="icon-sm"
          variant="ghost"
          onClick={onSaveTransaction}
          disabled={state.pending || !state.dirty}
          title="保存"
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
            <DropdownMenuItem onClick={() => onExport("csv")}>
              导出 CSV
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => onExport("json")}>
              导出 JSON
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => onExport("md")}>
              导出 Markdown
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </header>

      <AlertDialog
        open={refreshConfirmOpen}
        onOpenChange={setRefreshConfirmOpen}
      >
        <AlertDialogContent className="text-foreground">
          <AlertDialogHeader className="place-items-start text-left">
            <AlertDialogTitle>确认要刷新吗</AlertDialogTitle>
            <AlertDialogDescription>
              刷新将导致为保存数据丢失！
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter className="justify-end">
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              onClick={handleConfirmRefresh}
            >
              确认刷新
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
};

export default HeaderAction;
