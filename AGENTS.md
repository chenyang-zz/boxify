<!--
 Copyright 2026 chenyang
 
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at
 
     https://www.apache.org/licenses/LICENSE-2.0
 
 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
-->

## Language
Use Chinese

## Commenting Requirements (Mandatory)

All code changes MUST include sufficient and meaningful comments.

### General Rules

- Every exported function must have a descriptive comment explaining:
  - What it does
  - Its parameters
  - Return values
  - Possible side effects
- Complex logic must include inline comments explaining WHY, not WHAT.
- Do not add trivial comments that restate obvious code.
- If modifying existing code, improve outdated or unclear comments.
- Do not remove existing useful comments.

### Go Specific

- Follow GoDoc style for exported functions.
- Struct fields must include comments if their meaning is not obvious.
- Concurrency logic MUST include explanation comments.

### Frontend Specific

- Explain non-obvious state logic.
- Explain side-effects in useEffect.
- Explain custom hooks behavior.

### Hard Rule

Pull requests or patches without proper comments are considered incomplete.