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

import { usePropertyStore } from "@/store/property.store";
import { createContext, FC, ReactNode, useContext, useState } from "react";

export const fileTreeContext = {
  selectedUUID: "", // 当前选中的文件或文件夹的UUID
  setSelectedUUID: (uuid: string) => {}, // 更新选中的UUID
};

export type FileTreeContextType = typeof fileTreeContext;

const FileTreeContext = createContext<FileTreeContextType>(fileTreeContext);

const FileTreeProvider: FC<{
  children: ReactNode;
}> = ({ children }) => {
  const [selectedUUID, setSelectedUUID] = useState("");

  return (
    <FileTreeContext.Provider
      value={{
        selectedUUID,
        setSelectedUUID,
      }}
    >
      {children}
    </FileTreeContext.Provider>
  );
};

export const useFileTreeContext = () => {
  const context = useContext(FileTreeContext);
  if (!context) {
    throw new Error(
      "useFileTreeContext must be used within a FileTreeProvider",
    );
  }
  return context;
};

export default FileTreeProvider;
