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

import { StoreApi, UseBoundStore } from "zustand";

export type StoreMethods<T> = {
  [K in keyof T as T[K] extends Function ? K : never]: T[K];
};

export type StoreState<T> = Omit<T, keyof StoreMethods<T>>;

export function StoreMethods<T>(
  store: UseBoundStore<StoreApi<T>>,
): StoreMethods<T> {
  return store.getState() as StoreMethods<T>;
}
