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

export interface LLMConfig {
  base_url?: string;
  api_key?: string;
  model_name?: string;
  temperature?: number;
  max_tokens?: number;
}

interface LLMSettingProps {
  config?: LLMConfig;
  onChange?: (config: LLMConfig) => void;
}

export const LLMSetting: FC<LLMSettingProps> = ({
  config: initialConfig,
  onChange,
}) => {
  const [config, setConfig] = useState<LLMConfig>({
    base_url: "",
    api_key: "",
    model_name: "",
    temperature: 0.7,
    max_tokens: 8192,
    ...initialConfig,
  });

  const handleChange = (field: keyof LLMConfig, value: string) => {
    const next = { ...config, [field]: value };
    setConfig(next);
    onChange?.(next);
  };

  const handleNumberChange = (field: keyof LLMConfig, value: string) => {
    const numValue = value === "" ? undefined : Number(value);
    const next = { ...config, [field]: numValue };
    setConfig(next);
    onChange?.(next);
  };

  return (
    <form className="w-full" onSubmit={(e) => e.preventDefault()}>
      {/* 主标题 */}
      <h2 className="text-xl font-semibold text-foreground mb-1 text-left">
        模型提供商
      </h2>
      <p className="text-sm text-muted-foreground mb-6 text-left">
        配置 LLM 连接信息和模型参数。
      </p>

      {/* 配置卡片 */}
      <div className="bg-card rounded-xl p-5 space-y-6 border border-border">
        <SettingField
          id="base_url"
          label="提供商基础地址"
          description="请填写模型提供商的基础 URL 地址，需兼容 OpenAI 格式"
        >
          <Input
            id="base_url"
            type="url"
            placeholder="https://api.openai.com/v1"
            value={config.base_url ?? ""}
            onChange={(e) => handleChange("base_url", e.target.value)}
            className="w-[280px]"
          />
        </SettingField>

        <SettingField
          id="api_key"
          label="提供商密钥"
          description="请填写模型提供商 API 密钥"
        >
          <Input
            id="api_key"
            type="password"
            placeholder="sk-..."
            value={config.api_key ?? ""}
            onChange={(e) => handleChange("api_key", e.target.value)}
            className="w-[280px]"
          />
        </SettingField>

        <SettingField
          id="model_name"
          label="模型名"
          description="模型必须支持工具调用、图像识别等功能"
        >
          <Input
            id="model_name"
            type="text"
            placeholder="gpt-4o"
            value={config.model_name ?? ""}
            onChange={(e) => handleChange("model_name", e.target.value)}
            className="w-[280px]"
          />
        </SettingField>

        <SettingField
          id="temperature"
          label="温度 (temperature)"
          description="温度越低输出越确定稳定，越高越具创造性，默认为 0.7"
        >
          <Input
            id="temperature"
            type="number"
            placeholder="0.7"
            value={config.temperature ?? 0.7}
            onChange={(e) =>
              handleNumberChange("temperature", e.target.value)
            }
            min={0}
            max={2}
            step={0.1}
            className="w-[200px]"
          />
        </SettingField>

        <SettingField
          id="max_tokens"
          label="最大输出 Token 数"
          description="单次回复允许生成的最大 Token 数量，默认为 8192"
        >
          <Input
            id="max_tokens"
            type="number"
            placeholder="8192"
            value={config.max_tokens ?? 8192}
            onChange={(e) =>
              handleNumberChange("max_tokens", e.target.value)
            }
            min={1}
            max={128000}
            className="w-[200px]"
          />
        </SettingField>
      </div>
    </form>
  );
};

export default LLMSetting;
