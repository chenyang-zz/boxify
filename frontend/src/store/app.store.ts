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

import { create } from "zustand";
import { persist } from "zustand/middleware";

interface AppState {
  isPropertyOpen: boolean;
  setIsPropertyOpen: (open: boolean) => void;
}

export const useAppStore = create<AppState>()(
  persist(
    (set) => ({
      isPropertyOpen: true,
      setIsPropertyOpen: (open: boolean) =>
        set(() => ({ isPropertyOpen: open })),
    }),
    {
      name: "boxify-app-store",
      version: 1,
      migrate: (persistedState, version) => {
        return persistedState as AppState;
      },
      partialize: (state) => ({ isPropertyOpen: state.isPropertyOpen }),
    },
  ),
);
