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

import { FC } from "react";
import FileTreeItem from "./FileTreeItem";
import { PropertyItemType, usePropertyStore } from "@/store/property.store";

interface FileTreeProps {
  data: PropertyItemType[];
}

export const FIleTreeList: FC<FileTreeProps> = ({ data }) => {
  return (
    <div className="text-foreground text-sm flex-1 overflow-auto">
      {data.map((item, index) => (
        <FileTreeItem key={index} item={item} />
      ))}
    </div>
  );
};

const FileTree: FC = () => {
  const propertyList = usePropertyStore((state) => state.propertyList);

  return <FIleTreeList data={propertyList}></FIleTreeList>;
};

export default FileTree;
