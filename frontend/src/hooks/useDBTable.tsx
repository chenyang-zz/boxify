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

import { DBFileType } from "@/common/constrains";
import { getDBTableValuesByUUID } from "@/lib/db-table";
import { getPropertyItemByUUID } from "@/lib/property";
import { useEffect, useState } from "react";

export const useDBTable = (uuid: string) => {
  const [columns, setColumns] = useState<string[]>([]);
  const [values, setValues] = useState<Record<string, any>[]>([]);

  useEffect(() => {
    if (!uuid) return;
    const item = getPropertyItemByUUID(uuid);
    if (!item || item.type !== DBFileType.TABLE) return;
    async function getColumns() {
      const res = await getDBTableValuesByUUID(uuid);
      if (!res) return;
      setColumns(res.fields);
      setValues(res.data as Record<string, any>[]);
    }
    getColumns();
  }, [uuid]);

  return {
    columns,
    values,
  };
};
