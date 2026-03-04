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

import {
  classifyCommandTokens,
  normalizeCommandToken,
  splitCommandTokens,
} from "./command-parser";

function assertEqual<T>(actual: T, expected: T, message: string): void {
  if (actual !== expected) {
    throw new Error(`${message}: expected "${expected}", got "${actual}"`);
  }
}

function assert(condition: boolean, message: string): void {
  if (!condition) {
    throw new Error(message);
  }
}

// 运行命令解析领域单测，便于后续接入统一测试框架时直接复用。
export function runCommandParserUnitTests(): void {
  const splitTokens = splitCommandTokens(`FOO=1 ls -la | grep "test file"`);
  assertEqual(splitTokens[0], "FOO=1", "assignment token split");
  assertEqual(splitTokens[2], "ls", "main command token split");
  assertEqual(splitTokens[6], "|", "pipe operator split");

  const normalized = normalizeCommandToken(`"git"`);
  assertEqual(normalized, "git", "normalize quoted command");

  const commandSet = new Set(["ls", "grep", "git"]);
  const classified = classifyCommandTokens(
    `FOO=1 ls -la ./src | grep "hello world"`,
    commandSet,
  );

  const commandToken = classified.find((token) => token.type === "command");
  assert(Boolean(commandToken), "command token should exist");
  assertEqual(commandToken?.value, "ls", "main command classification");
  assertEqual(commandToken?.valid, true, "main command validity");

  const hasOption = classified.some((token) => token.type === "option");
  const hasPath = classified.some((token) => token.type === "path");
  const hasOperator = classified.some((token) => token.type === "operator");
  const hasString = classified.some((token) => token.type === "string");

  assert(hasOption, "should classify option token");
  assert(hasPath, "should classify path token");
  assert(hasOperator, "should classify operator token");
  assert(hasString, "should classify string token");
}
