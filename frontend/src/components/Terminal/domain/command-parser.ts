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

export type TokenType =
  | "whitespace"
  | "command"
  | "option"
  | "operator"
  | "variable"
  | "string"
  | "path"
  | "argument";

export interface CommandToken {
  value: string;
  type: TokenType;
  valid?: boolean;
}

const OPERATOR_SET = new Set([
  "|",
  "||",
  "&&",
  ";",
  ">",
  ">>",
  "<",
  "<<",
  "2>",
  "2>>",
  "&",
]);

const ASSIGNMENT_RE = /^[A-Za-z_][A-Za-z0-9_]*=.*/;

// 将输入拆分为保留空白和操作符的 token 列表，供后续高亮使用。
export function splitCommandTokens(input: string): string[] {
  const tokens: string[] = [];
  let buffer = "";
  let quote: "'" | '"' | null = null;

  for (let i = 0; i < input.length; i += 1) {
    const ch = input[i];

    if (quote) {
      buffer += ch;
      if (ch === quote) {
        quote = null;
      }
      continue;
    }

    if (ch === "'" || ch === '"') {
      buffer += ch;
      quote = ch;
      continue;
    }

    if (/\s/.test(ch)) {
      if (buffer) {
        tokens.push(buffer);
        buffer = "";
      }
      let spaceRun = ch;
      while (i + 1 < input.length && /\s/.test(input[i + 1])) {
        i += 1;
        spaceRun += input[i];
      }
      tokens.push(spaceRun);
      continue;
    }

    if (ch === "&" || ch === "|" || ch === ";" || ch === "<" || ch === ">") {
      if (buffer) {
        tokens.push(buffer);
        buffer = "";
      }
      let operator = ch;
      const next = input[i + 1];
      if (
        (ch === "&" || ch === "|" || ch === ">" || ch === "<") &&
        next === ch
      ) {
        i += 1;
        operator += next;
      }
      tokens.push(operator);
      continue;
    }

    if (ch === "2" && (input[i + 1] === ">" || input[i + 1] === "<")) {
      if (buffer) {
        tokens.push(buffer);
        buffer = "";
      }
      const operator = `${ch}${input[i + 1]}`;
      i += 1;
      if (input[i + 1] === ">") {
        i += 1;
        tokens.push(`${operator}>`);
      } else {
        tokens.push(operator);
      }
      continue;
    }

    buffer += ch;
  }

  if (buffer) {
    tokens.push(buffer);
  }

  return tokens;
}

// 规范化命令 token（去掉首尾引号），用于命令有效性匹配。
export function normalizeCommandToken(token: string): string {
  return token.replace(/^["']|["']$/g, "");
}

// 对 token 进行语义分类，并对主命令做有效性标记。
export function classifyCommandTokens(
  input: string,
  commandSet: Set<string>,
): CommandToken[] {
  const rawTokens = splitCommandTokens(input);
  const commandIndex = rawTokens.findIndex((token) => {
    if (!token || /^\s+$/.test(token)) return false;
    if (ASSIGNMENT_RE.test(token)) return false;
    return !OPERATOR_SET.has(token);
  });

  return rawTokens.map((token, index) => {
    if (!token) return { value: token, type: "argument" };
    if (/^\s+$/.test(token)) return { value: token, type: "whitespace" };
    if (OPERATOR_SET.has(token)) return { value: token, type: "operator" };

    if (index === commandIndex) {
      const normalized = normalizeCommandToken(token);
      const valid = normalized.length > 0 && commandSet.has(normalized);
      return { value: token, type: "command", valid };
    }

    if (token.startsWith("-")) return { value: token, type: "option" };
    if (token.startsWith("$")) return { value: token, type: "variable" };
    if (
      (token.startsWith('"') && token.endsWith('"')) ||
      (token.startsWith("'") && token.endsWith("'"))
    ) {
      return { value: token, type: "string" };
    }
    if (
      token.includes("/") ||
      token.startsWith("./") ||
      token.startsWith("../") ||
      token.startsWith("~/")
    ) {
      return { value: token, type: "path" };
    }

    return { value: token, type: "argument" };
  });
}

// 返回不同 token 对应的展示样式类名。
export function commandTokenClassName(token: CommandToken): string {
  switch (token.type) {
    case "command":
      if (token.valid) return "text-emerald-400";
      return "text-foreground underline decoration-red-400 decoration-dashed underline-offset-4";
    case "option":
      return "text-amber-300";
    case "operator":
      return "text-pink-300";
    case "variable":
      return "text-sky-300";
    case "string":
      return "text-teal-300";
    case "path":
      return "text-violet-300";
    default:
      return "text-foreground";
  }
}
