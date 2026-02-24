import * as React from "react";
import { Loader2Icon, TriangleAlertIcon } from "lucide-react";
import { useState } from "react";

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
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";

export interface DeleteConfirmDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title?: string;
  checkboxLabel?: string;
  onConfirm: () => void | Promise<void>;
  cancelText?: string;
  confirmText?: string;
}

const DEFAULT_PROPS = {
  title: "确认删除",
  checkboxLabel: "我了解此操作是永久且无法撤销",
  cancelText: "取消",
  confirmText: "确认删除",
} as const;

export function DeleteConfirmDialog({
  open,
  onOpenChange,
  title = DEFAULT_PROPS.title,
  checkboxLabel = DEFAULT_PROPS.checkboxLabel,
  onConfirm,
  cancelText = DEFAULT_PROPS.cancelText,
  confirmText = DEFAULT_PROPS.confirmText,
}: DeleteConfirmDialogProps) {
  const [checked, setChecked] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  // 当弹窗关闭时，重置复选框状态
  React.useEffect(() => {
    if (!open) {
      setChecked(false);
    }
  }, [open]);

  const handleConfirm = async () => {
    setIsLoading(true);
    try {
      await onConfirm();
      onOpenChange(false);
    } catch (error) {
      // 保持弹窗打开，让用户看到错误
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent className=" text-foreground">
        <AlertDialogHeader>
          <AlertDialogTitle>{title}</AlertDialogTitle>
        </AlertDialogHeader>

        <div className="flex items-top gap-2 py-2 text-muted-foreground">
          <Checkbox
            id="delete-confirm-checkbox"
            checked={checked}
            onCheckedChange={(checked) => setChecked(checked === true)}
          />
          <Label
            htmlFor="delete-confirm-checkbox"
            className="text-sm leading-tight cursor-pointer peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
          >
            {checkboxLabel}
          </Label>
        </div>

        <AlertDialogFooter>
          <AlertDialogCancel disabled={isLoading}>
            {cancelText}
          </AlertDialogCancel>
          <AlertDialogAction
            variant="destructive"
            disabled={!checked || isLoading}
            onClick={(e) => {
              e.preventDefault();
              handleConfirm();
            }}
          >
            {isLoading ? (
              <>
                <Loader2Icon className="mr-2 size-4 animate-spin" />
                处理中...
              </>
            ) : (
              confirmText
            )}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
