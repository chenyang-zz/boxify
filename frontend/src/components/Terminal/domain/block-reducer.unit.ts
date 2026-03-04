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
  appendBatchToBlockLastLine,
  appendLineToBlockOutput,
  createRunningBlock,
  finalizeBlock,
  updateBlockStatus,
} from "./block-reducer";

function assert(condition: boolean, message: string): void {
  if (!condition) {
    throw new Error(message);
  }
}

function assertEqual<T>(actual: T, expected: T, message: string): void {
  if (actual !== expected) {
    throw new Error(`${message}: expected "${expected}", got "${actual}"`);
  }
}

// 运行 block reducer 领域单测，覆盖 block 生命周期核心路径。
export function runBlockReducerUnitTests(): void {
  const nowFactory = { now: () => 1000 };
  const lineFactory = {
    now: () => 2000,
    createLineId: () => "line-1",
  };

  const block = createRunningBlock(
    "block-1",
    "echo hello",
    { workPath: "/tmp", gitBranch: "main" },
    nowFactory,
  );
  assertEqual(block.status, "running", "block initial status");
  assertEqual(block.startTime, 1000, "block start time");
  assertEqual(block.workPath, "/tmp", "block workPath");

  const withLine = appendLineToBlockOutput(
    [block],
    "block-1",
    "hello",
    [],
    lineFactory,
  );
  assert(Boolean(withLine), "append line result");
  assertEqual(withLine?.[0].output.length, 1, "append line length");
  assertEqual(withLine?.[0].output[0].id, "line-1", "append line id");

  const withBatch = appendBatchToBlockLastLine(
    withLine || [],
    "block-1",
    [
      { content: " world", formattedContent: [] },
      { content: "!", formattedContent: [] },
    ],
    lineFactory,
  );
  assert(Boolean(withBatch), "append batch result");
  assertEqual(withBatch?.[0].output.length, 1, "append batch should merge into last line");
  assertEqual(withBatch?.[0].output[0].content, "hello world!", "append batch content");

  const finalized = finalizeBlock(withBatch || [], "block-1", 0, {
    now: () => 3000,
  });
  assert(Boolean(finalized), "finalize result");
  assertEqual(finalized?.[0].status, "success", "finalize status");
  assertEqual(finalized?.[0].endTime, 3000, "finalize end time");

  const updated = updateBlockStatus(finalized || [], "block-1", "pending");
  assert(Boolean(updated), "update status result");
  assertEqual(updated?.[0].status, "pending", "manual status update");
}
