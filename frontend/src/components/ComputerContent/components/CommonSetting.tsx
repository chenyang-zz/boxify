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

import { FC, useState } from "react";
import { Input } from "@/components/ui/input";
import { SettingField } from "./SettingField";

export interface CommonConfig {
  max_iterations?: number;
  max_retries?: number;
  max_search_results?: number;
}

interface CommonSettingProps {
  config?: CommonConfig;
  onChange?: (config: CommonConfig) => void;
}

export const CommonSetting: FC<CommonSettingProps> = ({
  config: initialConfig,
  onChange,
}) => {
  const [config, setConfig] = useState<CommonConfig>({
    max_iterations: 100,
    max_retries: 3,
    max_search_results: 10,
    ...initialConfig,
  });

  const handleChange = (field: keyof CommonConfig, value: string) => {
    const numValue = value === "" ? undefined : Number(value);
    const next = { ...config, [field]: numValue };
    setConfig(next);
    onChange?.(next);
  };

  return (
    <form className="w-full" onSubmit={(e) => e.preventDefault()}>
      {/* 主标题 */}
      <h2 className="text-xl font-semibold text-foreground mb-1 text-left">
        通用配置
      </h2>
      <p className="text-sm text-muted-foreground mb-6 text-left">
        配置 Agent 执行过程中的通用参数。
      </p>

      {/* 配置卡片 */}
      <div className="bg-card rounded-xl p-5 space-y-6 border border-border">
        <SettingField
          id="max_iterations"
          label="最大计划迭代次数"
          description="执行 Agent 最大能迭代循环调用工具的次数，默认为 100"
        >
          <Input
            id="max_iterations"
            type="number"
            placeholder="100"
            value={config.max_iterations ?? 100}
            onChange={(e) => handleChange("max_iterations", e.target.value)}
            min={0}
            max={200}
            className="w-[200px]"
          />
        </SettingField>

        <SettingField
          id="max_retries"
          label="最大重试次数"
          description="LLM / Tool 调用失败时的最大重试次数，默认为 3"
        >
          <Input
            id="max_retries"
            type="number"
            placeholder="3"
            value={config.max_retries ?? 3}
            onChange={(e) => handleChange("max_retries", e.target.value)}
            min={0}
            max={10}
            className="w-[200px]"
          />
        </SettingField>

        <SettingField
          id="max_search_results"
          label="最大搜索结果"
          description="搜索工具返回的最大结果数量，默认为 10"
        >
          <Input
            id="max_search_results"
            type="number"
            placeholder="10"
            value={config.max_search_results ?? 10}
            onChange={(e) =>
              handleChange("max_search_results", e.target.value)
            }
            min={0}
            max={30}
            className="w-[200px]"
          />
        </SettingField>
      </div>
    </form>
  );
};

export default CommonSetting;
