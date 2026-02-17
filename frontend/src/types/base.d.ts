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

/** null to undefined type mapping */
type NullToUndefined<T> = T extends null
  ? undefined
  : T extends object
    ? { [K in keyof T]: NullToUndefined<T[K]> }
    : T;

/** null to optional type mapping */
type NullToOptional<T> = {
  [K in keyof T as null extends T[K] ? K : never]?: Exclude<T[K], null>;
} & {
  [K in keyof T as null extends T[K] ? never : K]: T[K];
};
