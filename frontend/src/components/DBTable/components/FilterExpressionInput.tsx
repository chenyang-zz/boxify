import { FC, KeyboardEvent, useMemo, useRef, useState } from "react";
import { Textarea } from "@/components/ui/textarea";
import { cn } from "@/lib/utils";

interface FilterExpressionInputProps {
  value: string;
  columns: string[];
  pending: boolean;
  error?: string;
  onChange: (nextValue: string) => void;
  onApply: () => void;
}

interface FilterSuggestion {
  label: string;
  insertText: string;
  kind: "column" | "keyword" | "operator";
}

const KEYWORD_SUGGESTIONS: FilterSuggestion[] = [
  { label: "AND", insertText: "AND", kind: "keyword" },
  { label: "OR", insertText: "OR", kind: "keyword" },
  { label: "LIKE", insertText: "LIKE", kind: "keyword" },
  { label: "NOT LIKE", insertText: "NOT LIKE", kind: "keyword" },
  { label: "IS NULL", insertText: "IS NULL", kind: "keyword" },
  { label: "IS NOT NULL", insertText: "IS NOT NULL", kind: "keyword" },
  { label: "( )", insertText: "()", kind: "operator" },
  { label: "=", insertText: "=", kind: "operator" },
  { label: "!=", insertText: "!=", kind: "operator" },
  { label: ">", insertText: ">", kind: "operator" },
  { label: ">=", insertText: ">=", kind: "operator" },
  { label: "<", insertText: "<", kind: "operator" },
  { label: "<=", insertText: "<=", kind: "operator" },
];

function getCaretWordRange(value: string, caret: number): { start: number; token: string } {
  let start = caret;
  while (start > 0 && /[A-Za-z0-9_]/.test(value[start - 1])) {
    start -= 1;
  }
  return {
    start,
    token: value.slice(start, caret),
  };
}

function kindBadgeClassName(kind: FilterSuggestion["kind"]): string {
  switch (kind) {
    case "column":
      return "text-emerald-500";
    case "keyword":
      return "text-blue-500";
    default:
      return "text-orange-500";
  }
}

// 筛选表达式输入：提供变量/语法补全和快捷执行。
const FilterExpressionInput: FC<FilterExpressionInputProps> = ({
  value,
  columns,
  pending,
  error,
  onChange,
  onApply,
}) => {
  const textareaRef = useRef<HTMLTextAreaElement | null>(null);
  const [activeIndex, setActiveIndex] = useState(0);
  const [focus, setFocus] = useState(false);

  const suggestions = useMemo(() => {
    const el = textareaRef.current;
    const caret = el?.selectionStart ?? value.length;
    const { token } = getCaretWordRange(value, caret);
    const keyword = token.trim().toLowerCase();

    const columnSuggestions: FilterSuggestion[] = columns.map((column) => ({
      label: column,
      insertText: column,
      kind: "column",
    }));

    const all = [...columnSuggestions, ...KEYWORD_SUGGESTIONS];
    if (!keyword) {
      return all.slice(0, 10);
    }

    return all
      .filter((item) => item.label.toLowerCase().includes(keyword))
      .slice(0, 10);
  }, [columns, value]);

  // 用选中的建议替换光标前的 token。
  const applySuggestion = (suggestion: FilterSuggestion) => {
    const el = textareaRef.current;
    if (!el) {
      return;
    }

    const caret = el.selectionStart ?? value.length;
    const { start } = getCaretWordRange(value, caret);
    const nextValue = `${value.slice(0, start)}${suggestion.insertText}${value.slice(caret)}`;
    onChange(nextValue);

    requestAnimationFrame(() => {
      const nextCaret = start + suggestion.insertText.length;
      el.focus();
      el.setSelectionRange(nextCaret, nextCaret);
    });
  };

  // 处理键盘交互：回车执行、方向键切换候选、Tab 选中候选。
  const handleKeyDown = (event: KeyboardEvent<HTMLTextAreaElement>) => {
    if (event.key === "ArrowDown" && suggestions.length > 0) {
      event.preventDefault();
      setActiveIndex((prev) => (prev + 1) % suggestions.length);
      return;
    }

    if (event.key === "ArrowUp" && suggestions.length > 0) {
      event.preventDefault();
      setActiveIndex((prev) => (prev - 1 + suggestions.length) % suggestions.length);
      return;
    }

    if (event.key === "Tab" && suggestions.length > 0) {
      event.preventDefault();
      const suggestion = suggestions[activeIndex] ?? suggestions[0];
      if (suggestion) {
        applySuggestion(suggestion);
      }
      return;
    }

    if (event.key === "Enter" && !event.shiftKey) {
      event.preventDefault();
      onApply();
    }
  };

  return (
    <div className="space-y-1.5">
      <Textarea
        ref={textareaRef}
        value={value}
        onChange={(event) => {
          onChange(event.target.value);
          setActiveIndex(0);
        }}
        onKeyDown={handleKeyDown}
        onFocus={() => setFocus(true)}
        onBlur={() => {
          setTimeout(() => {
            setFocus(false);
          }, 120);
        }}
        placeholder="id = 1 AND name LIKE '%box%'"
        className="min-h-20 resize-y text-xs"
        disabled={pending}
      />
      <div className="flex items-center justify-between text-[11px] text-muted-foreground">
        <span>回车执行筛选，Shift+回车换行</span>
        <span>Tab 选中建议</span>
      </div>
      {focus && suggestions.length > 0 && (
        <div className="rounded-md border border-border bg-popover p-1 shadow-sm max-h-40 overflow-auto">
          {suggestions.map((suggestion, index) => (
            <button
              key={`${suggestion.kind}-${suggestion.label}-${index}`}
              type="button"
              className={cn(
                "w-full text-left px-2 py-1 text-xs rounded-sm flex items-center justify-between hover:bg-accent",
                index === activeIndex && "bg-accent",
              )}
              onMouseDown={(event) => {
                event.preventDefault();
                applySuggestion(suggestion);
                setActiveIndex(index);
              }}
            >
              <span className="truncate">{suggestion.label}</span>
              <span className={cn("text-[10px] uppercase", kindBadgeClassName(suggestion.kind))}>
                {suggestion.kind}
              </span>
            </button>
          ))}
        </div>
      )}
      {error && <p className="text-[11px] text-destructive">{error}</p>}
    </div>
  );
};

export default FilterExpressionInput;
