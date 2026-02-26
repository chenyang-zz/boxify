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

import { ConnectionEnum } from "@/common/constrains";
import { AuthMethod, Environment } from "@/common/enums/connection";
import { ConnectionStandard, TerminalStandard } from "@/types/initial-data";
import { ShellType } from "@wails/service";

export function generateDefaultFormData(
  type?: ConnectionEnum,
  tab = "standard",
): ConnectionStandard {
  if (tab === "standard") {
    switch (type) {
      case ConnectionEnum.MYSQL:
        return {
          tagColor: "",
          environment: Environment.None,
          name: "",
          host: "localhost",
          user: "root",
          port: 3306,
          authMethod: AuthMethod.Password,
          password: "",
          remark: "",
        };
      case ConnectionEnum.SSH:
        return {
          tagColor: "",
          environment: Environment.None,
          name: "",
          host: "",
          user: "root",
          port: 22,
          authMethod: AuthMethod.Password,
          password: "",
          remark: "",
        };
      case ConnectionEnum.TERMINAL:
        return {
          name: "",
          shell: ShellType.ShellTypeAuto,
        };
    }
  }
  return {
    tagColor: "",
    environment: Environment.None,
    name: "",
    host: "localhost",
    user: "root",
    port: 3306,
    authMethod: AuthMethod.Password,
    password: "",
    remark: "",
  };
}
