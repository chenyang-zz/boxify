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

import type { BlockStatus, OutputLine, TerminalBlock } from "../types/block";

interface BlockContext {
  workPath?: string;
  gitBranch?: string;
}

interface OutputChunk {
  content: string;
  formattedContent: OutputLine["formattedContent"];
}

interface LineFactory {
  now: () => number;
  createLineId: () => string;
}

interface BlockFactory {
  now: () => number;
}

function appendOutputToLastLine(
  output: OutputLine[],
  content: string,
  formattedContent: OutputLine["formattedContent"],
  factory: LineFactory,
): OutputLine[] {
  const lastLine = output[output.length - 1];
  if (!lastLine) {
    return [
      {
        id: factory.createLineId(),
        content,
        formattedContent,
        timestamp: factory.now(),
      },
    ];
  }

  return [
    ...output.slice(0, -1),
    {
      ...lastLine,
      content: lastLine.content + content,
      formattedContent: [...lastLine.formattedContent, ...formattedContent],
    },
  ];
}

function updateBlocksById(
  blocks: TerminalBlock[],
  blockId: string,
  updater: (block: TerminalBlock) => TerminalBlock,
): TerminalBlock[] | null {
  const idx = blocks.findIndex((block) => block.id === blockId);
  if (idx === -1) return null;

  return blocks.map((block, i) => (i === idx ? updater(block) : block));
}

export function createRunningBlock(
  id: string,
  command: string,
  context: BlockContext | undefined,
  factory: BlockFactory,
): TerminalBlock {
  return {
    id,
    command,
    output: [],
    status: "running",
    startTime: factory.now(),
    workPath: context?.workPath,
    gitBranch: context?.gitBranch,
  };
}

export function appendLineToBlockOutput(
  blocks: TerminalBlock[],
  blockId: string,
  content: string,
  formattedContent: OutputLine["formattedContent"],
  factory: LineFactory,
): TerminalBlock[] | null {
  return updateBlocksById(blocks, blockId, (block) => ({
    ...block,
    output: [
      ...block.output,
      {
        id: factory.createLineId(),
        content,
        formattedContent,
        timestamp: factory.now(),
      },
    ],
  }));
}

export function appendChunkToBlockLastLine(
  blocks: TerminalBlock[],
  blockId: string,
  content: string,
  formattedContent: OutputLine["formattedContent"],
  factory: LineFactory,
): TerminalBlock[] | null {
  return updateBlocksById(blocks, blockId, (block) => ({
    ...block,
    output: appendOutputToLastLine(block.output, content, formattedContent, factory),
  }));
}

export function appendBatchToBlockLastLine(
  blocks: TerminalBlock[],
  blockId: string,
  chunks: OutputChunk[],
  factory: LineFactory,
): TerminalBlock[] | null {
  if (chunks.length === 0) return blocks;

  const merged = chunks.reduce(
    (acc, chunk) => ({
      content: acc.content + chunk.content,
      formattedContent: [...acc.formattedContent, ...chunk.formattedContent],
    }),
    { content: "", formattedContent: [] as OutputLine["formattedContent"] },
  );

  return appendChunkToBlockLastLine(
    blocks,
    blockId,
    merged.content,
    merged.formattedContent,
    factory,
  );
}

export function finalizeBlock(
  blocks: TerminalBlock[],
  blockId: string,
  exitCode: number,
  factory: BlockFactory,
): TerminalBlock[] | null {
  return updateBlocksById(blocks, blockId, (block) => ({
    ...block,
    status: exitCode === 0 ? "success" : "error",
    endTime: factory.now(),
    exitCode,
  }));
}

export function updateBlockStatus(
  blocks: TerminalBlock[],
  blockId: string,
  status: BlockStatus,
): TerminalBlock[] | null {
  return updateBlocksById(blocks, blockId, (block) => ({
    ...block,
    status,
  }));
}
