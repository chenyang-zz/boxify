import Editor, { loader, type OnMount } from "@monaco-editor/react";
import * as monaco from "monaco-editor";
import { FC, useCallback, useEffect, useMemo, useRef } from "react";
import { ConnectionEnum } from "@/common/constrains";
import { resolveCssVar } from "@/lib/theme";

loader.config({ monaco });

interface FilterExpressionInputProps {
  value: string;
  columns: string[];
  databaseType?: ConnectionEnum.MYSQL | ConnectionEnum.POSTGRESQL | null;
  pending: boolean;
  error?: string;
  onChange: (nextValue: string) => void;
  onApply: (expression: string) => void;
}

type SuggestionKind = "column" | "keyword" | "operator" | "function";

interface FilterSuggestion {
  label: string;
  insertText: string;
  kind: SuggestionKind;
  detail?: string;
}

const FILTER_LANGUAGE_ID = "boxify-filter-sql";
const FILTER_THEME_LIGHT = "boxify-filter-light";
const FILTER_THEME_DARK = "boxify-filter-dark";
let filterThemesRegistered = false;

const BASE_FILTER_KEYWORDS: FilterSuggestion[] = [
  { label: "AND", insertText: "AND", kind: "keyword" },
  { label: "OR", insertText: "OR", kind: "keyword" },
  { label: "LIKE", insertText: "LIKE", kind: "keyword" },
  { label: "NOT LIKE", insertText: "NOT LIKE", kind: "keyword" },
  { label: "IS NULL", insertText: "IS NULL", kind: "keyword" },
  { label: "IS NOT NULL", insertText: "IS NOT NULL", kind: "keyword" },
];

const BASE_FILTER_OPERATORS: FilterSuggestion[] = [
  { label: "=", insertText: "=", kind: "operator" },
  { label: "!=", insertText: "!=", kind: "operator" },
  { label: "<>", insertText: "<>", kind: "operator" },
  { label: ">", insertText: ">", kind: "operator" },
  { label: ">=", insertText: ">=", kind: "operator" },
  { label: "<", insertText: "<", kind: "operator" },
  { label: "<=", insertText: "<=", kind: "operator" },
];

const FILTER_CONSTANTS: FilterSuggestion[] = [
  { label: "NULL", insertText: "NULL", kind: "keyword", detail: "Constant" },
  { label: "TRUE", insertText: "TRUE", kind: "keyword", detail: "Constant" },
  { label: "FALSE", insertText: "FALSE", kind: "keyword", detail: "Constant" },
];

const MYSQL_FILTER_FUNCTIONS: FilterSuggestion[] = [
  {
    label: "LOWER",
    insertText: "LOWER(${1:str})",
    kind: "function",
    detail: "LOWER(str)",
  },
  {
    label: "UPPER",
    insertText: "UPPER(${1:str})",
    kind: "function",
    detail: "UPPER(str)",
  },
  {
    label: "TRIM",
    insertText: "TRIM(${1:str})",
    kind: "function",
    detail: "TRIM(str)",
  },
  {
    label: "LENGTH",
    insertText: "LENGTH(${1:str})",
    kind: "function",
    detail: "LENGTH(str)",
  },
  {
    label: "ABS",
    insertText: "ABS(${1:x})",
    kind: "function",
    detail: "ABS(x)",
  },
  {
    label: "ROUND",
    insertText: "ROUND(${1:x}${2:, ${3:d}})",
    kind: "function",
    detail: "ROUND(x [, d])",
  },
  {
    label: "CONCAT",
    insertText: "CONCAT(${1:str1}, ${2:str2})",
    kind: "function",
    detail: "CONCAT(str1, str2, ...)",
  },
  {
    label: "SUBSTRING",
    insertText: "SUBSTRING(${1:str}, ${2:pos}${3:, ${4:len}})",
    kind: "function",
    detail: "SUBSTRING(str, pos [, len])",
  },
  {
    label: "COALESCE",
    insertText: "COALESCE(${1:expr1}, ${2:expr2})",
    kind: "function",
    detail: "COALESCE(expr1, expr2, ...)",
  },
  { label: "NOW", insertText: "NOW()", kind: "function", detail: "NOW()" },
];

const POSTGRES_FILTER_FUNCTIONS: FilterSuggestion[] = [
  {
    label: "LOWER",
    insertText: "LOWER(${1:str})",
    kind: "function",
    detail: "LOWER(str)",
  },
  {
    label: "UPPER",
    insertText: "UPPER(${1:str})",
    kind: "function",
    detail: "UPPER(str)",
  },
  {
    label: "BTRIM",
    insertText: "BTRIM(${1:str})",
    kind: "function",
    detail: "BTRIM(str)",
  },
  {
    label: "LENGTH",
    insertText: "LENGTH(${1:str})",
    kind: "function",
    detail: "LENGTH(str)",
  },
  {
    label: "ABS",
    insertText: "ABS(${1:x})",
    kind: "function",
    detail: "ABS(x)",
  },
  {
    label: "ROUND",
    insertText: "ROUND(${1:x}${2:, ${3:d}})",
    kind: "function",
    detail: "ROUND(x [, d])",
  },
  {
    label: "CONCAT",
    insertText: "CONCAT(${1:str1}, ${2:str2})",
    kind: "function",
    detail: "CONCAT(str1, str2, ...)",
  },
  {
    label: "SUBSTRING",
    insertText: "SUBSTRING(${1:str}, ${2:pos}${3: FOR ${4:len}})",
    kind: "function",
    detail: "SUBSTRING(str, pos [FOR len])",
  },
  {
    label: "COALESCE",
    insertText: "COALESCE(${1:expr1}, ${2:expr2})",
    kind: "function",
    detail: "COALESCE(expr1, expr2, ...)",
  },
  { label: "NOW", insertText: "NOW()", kind: "function", detail: "NOW()" },
];

const POSTGRES_FILTER_KEYWORDS: FilterSuggestion[] = [
  { label: "ILIKE", insertText: "ILIKE", kind: "keyword" },
  { label: "NOT ILIKE", insertText: "NOT ILIKE", kind: "keyword" },
];

// 根据建议类型映射为 Monaco 的补全项类型。
function toCompletionItemKind(
  kind: SuggestionKind,
): monaco.languages.CompletionItemKind {
  switch (kind) {
    case "column":
      return monaco.languages.CompletionItemKind.Field;
    case "function":
      return monaco.languages.CompletionItemKind.Function;
    case "operator":
      return monaco.languages.CompletionItemKind.Operator;
    default:
      return monaco.languages.CompletionItemKind.Keyword;
  }
}

// 确保过滤表达式语言只注册一次。
function ensureFilterLanguageRegistered() {
  const exists = monaco.languages
    .getLanguages()
    .some((lang) => lang.id === FILTER_LANGUAGE_ID);
  if (exists) {
    return;
  }

  monaco.languages.register({ id: FILTER_LANGUAGE_ID });
  monaco.languages.setMonarchTokensProvider(FILTER_LANGUAGE_ID, {
    tokenizer: {
      root: [
        [/\b(AND|OR|LIKE|NOT|IS|NULL|TRUE|FALSE|ILIKE)\b/i, "keyword"],
        [/[A-Za-z_][A-Za-z0-9_]*(?=\s*\()/, "function"],
        [/[A-Za-z_][A-Za-z0-9_]*/, "variable"],
        [/<=|>=|<>|!=|=|<|>/, "operator"],
        [/'[^']*'/, "string"],
        [/`[^`]*`|"[^"]*"/, "variable"],
        [/\d+(\.\d+)?/, "number"],
        [/[()]/, "delimiter.parenthesis"],
      ],
    },
  });
}

// 注册筛选输入器主题，避免打开输入器时重复切换 Monaco 全局主题导致闪烁。
function ensureFilterThemesRegistered() {
  if (filterThemesRegistered) {
    return;
  }
  monaco.editor.defineTheme(FILTER_THEME_LIGHT, {
    base: "vs",
    inherit: true,
    rules: [
      { token: "keyword", foreground: "0284c7", fontStyle: "bold" },
      { token: "function", foreground: "c2410c" },
      { token: "variable", foreground: "15803d" },
      { token: "operator", foreground: "475569" },
    ],
    colors: {
      "editor.background": resolveCssVar("--background", "#ffffff"),
      "editorCursor.foreground": resolveCssVar("--primary", "#0f172a"),
    },
  });
  monaco.editor.defineTheme(FILTER_THEME_DARK, {
    base: "vs-dark",
    inherit: true,
    rules: [
      { token: "keyword", foreground: "7dd3fc", fontStyle: "bold" },
      { token: "function", foreground: "fdba74" },
      { token: "variable", foreground: "4ade80" },
      { token: "operator", foreground: "cbd5e1" },
    ],
    colors: {
      "editor.background": resolveCssVar("--background", "#1f2937"),
      "editorCursor.foreground": resolveCssVar("--primary", "#e2e8f0"),
    },
  });
  filterThemesRegistered = true;
}

// 筛选表达式输入：使用 Monaco 提供补全与光标跟随候选列表。
const FilterExpressionInput: FC<FilterExpressionInputProps> = ({
  value,
  columns,
  databaseType,
  pending,
  error,
  onChange,
  onApply,
}) => {
  const editorRef = useRef<monaco.editor.IStandaloneCodeEditor | null>(null);
  const completionProviderRef = useRef<monaco.IDisposable | null>(null);
  const keydownDisposableRef = useRef<monaco.IDisposable | null>(null);
  const editorTheme = document.documentElement.classList.contains("dark")
    ? FILTER_THEME_DARK
    : FILTER_THEME_LIGHT;

  const suggestions = useMemo<FilterSuggestion[]>(() => {
    const columnSuggestions = columns.map((column) => ({
      label: column,
      insertText: column,
      kind: "column" as const,
      detail: "Column",
    }));

    const dbKeywords =
      databaseType === ConnectionEnum.POSTGRESQL
        ? [...BASE_FILTER_KEYWORDS, ...POSTGRES_FILTER_KEYWORDS]
        : BASE_FILTER_KEYWORDS;
    const dbFunctions =
      databaseType === ConnectionEnum.POSTGRESQL
        ? POSTGRES_FILTER_FUNCTIONS
        : MYSQL_FILTER_FUNCTIONS;

    return [
      ...columnSuggestions,
      ...dbKeywords,
      ...BASE_FILTER_OPERATORS,
      ...FILTER_CONSTANTS,
      ...dbFunctions,
    ];
  }, [columns, databaseType]);

  const registerCompletionProvider = useCallback(() => {
    completionProviderRef.current?.dispose();
    completionProviderRef.current =
      monaco.languages.registerCompletionItemProvider(FILTER_LANGUAGE_ID, {
        triggerCharacters: [
          ...Array.from(
            "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_",
          ),
        ],
        provideCompletionItems: (model, position) => {
          const word = model.getWordUntilPosition(position);
          const range: monaco.IRange = {
            startLineNumber: position.lineNumber,
            endLineNumber: position.lineNumber,
            startColumn: word.startColumn,
            endColumn: word.endColumn,
          };

          return {
            suggestions: suggestions.map((item) => ({
              label: item.label,
              kind: toCompletionItemKind(item.kind),
              insertText: item.insertText,
              insertTextRules:
                item.kind === "function"
                  ? monaco.languages.CompletionItemInsertTextRule
                      .InsertAsSnippet
                  : monaco.languages.CompletionItemInsertTextRule.None,
              detail: item.detail ?? item.kind.toUpperCase(),
              range,
              sortText:
                item.kind === "column"
                  ? "0"
                  : item.kind === "keyword"
                    ? "1"
                    : "2",
            })),
          };
        },
      });
  }, [suggestions]);

  const handleEditorMount: OnMount = useCallback(
    (editor) => {
      ensureFilterLanguageRegistered();
      editorRef.current = editor;
      registerCompletionProvider();

      // Enter 执行筛选，Shift+Enter 换行。
      keydownDisposableRef.current?.dispose();
      keydownDisposableRef.current = editor.onKeyDown((event) => {
        if (event.keyCode === monaco.KeyCode.Enter && !event.shiftKey) {
          event.preventDefault();
          event.stopPropagation();
          onApply(editor.getValue());
        }
      });
    },
    [onApply, registerCompletionProvider],
  );

  useEffect(() => {
    if (!editorRef.current) {
      return;
    }
    registerCompletionProvider();
  }, [registerCompletionProvider]);

  useEffect(
    () => () => {
      completionProviderRef.current?.dispose();
      keydownDisposableRef.current?.dispose();
    },
    [],
  );

  return (
    <div>
      <div className="relative bg-background">
        <Editor
          beforeMount={() => {
            ensureFilterLanguageRegistered();
            ensureFilterThemesRegistered();
          }}
          value={value}
          onMount={handleEditorMount}
          onChange={(nextValue) => onChange(nextValue ?? "")}
          language={FILTER_LANGUAGE_ID}
          theme={editorTheme}
          options={{
            readOnly: pending,
            minimap: { enabled: false },
            lineNumbers: "off",
            glyphMargin: false,
            folding: false,
            lineDecorationsWidth: 0,
            lineNumbersMinChars: 0,
            overviewRulerBorder: false,
            hideCursorInOverviewRuler: true,
            scrollBeyondLastLine: false,
            wordWrap: "on",
            quickSuggestions: { other: true, comments: false, strings: true },
            suggestOnTriggerCharacters: true,
            acceptSuggestionOnEnter: "on",
            tabCompletion: "on",
            fontSize: 12,
            padding: { top: 8, bottom: 8 },
            renderLineHighlight: "none",
            automaticLayout: true,
          }}
          height="96px"
        />
      </div>
      <div className="flex px-2 h-6  items-center justify-between text-[10px] text-muted-foreground">
        <span>回车执行筛选，Shift+回车换行</span>
        <span>选中补全后按回车键确认</span>
      </div>
    </div>
  );
};

export default FilterExpressionInput;
