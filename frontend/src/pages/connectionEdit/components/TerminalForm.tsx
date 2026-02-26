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

import { Card } from "@/components/ui/card";
import { Field, FieldGroup, FieldLabel, FieldSet } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import type { TerminalStandard } from "@/types/initial-data";
import { FC, useEffect, useState } from "react";
import ColorTagSelector from "@/components/ColorTagSelector";
import { CommonFormProps } from "../types";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { ShellOptions } from "@/common/constrains";

interface FormErrors {
  name?: string;
}

const TerminalForm: FC<CommonFormProps<TerminalStandard>> = ({
  initialData,
  onSubmit,
}) => {
  // 表单状态管理
  const [formData, setFormData] = useState<TerminalStandard>({
    ...initialData,
  });

  const [errors, setErrors] = useState<FormErrors>({});

  // 初始化数据填充（编辑场景）
  useEffect(() => {
    if (initialData) {
      setFormData({ ...initialData });
      setErrors({});
    }
  }, [initialData]);

  // 字段验证函数
  const validateForm = (): boolean => {
    const newErrors: FormErrors = {};

    if (!formData.name.trim()) {
      newErrors.name = "终端名称不能为空";
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  // 提交处理
  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!validateForm()) {
      return; // 验证失败，不提交
    }

    onSubmit(formData);
  };

  // 通用字段变更处理
  const handleChange = (
    field: keyof TerminalStandard,
    value: string | number,
  ) => {
    setFormData((prev: TerminalStandard) => ({ ...prev, [field]: value }));
    // 清除该字段的错误提示
    const errorKey = field as keyof FormErrors;
    if (errors[errorKey]) {
      setErrors((prev: FormErrors) => {
        const newErrors = { ...prev };
        (newErrors as any)[errorKey] = undefined;
        return newErrors;
      });
    }
  };

  return (
    <Card className="w-full h-full p-5">
      <form id="standard-form" onSubmit={handleSubmit} className="h-full">
        <FieldSet className="w-full">
          <FieldGroup className="w-full gap-5">
            <div className="grid grid-cols-2 gap-5">
              {/* 颜色标签 */}
              <Field className="gap-2">
                <FieldLabel htmlFor="color-tag" className="text-xs">
                  颜色标签
                </FieldLabel>
                <ColorTagSelector
                  className="h-9"
                  value={formData.tagColor || ""}
                  onChange={(color) => handleChange("tagColor", color)}
                />
              </Field>

              {/* 名称 */}
              <Field className="gap-2">
                <FieldLabel htmlFor="name" className="text-xs">
                  名称
                </FieldLabel>
                <Input
                  id="name"
                  className={`text-xs ${errors.name ? "border-red-500" : ""}`}
                  value={formData.name}
                  onChange={(e) => handleChange("name", e.target.value)}
                />
              </Field>
            </div>

            {/* Shell */}
            <Field className="gap-2">
              <FieldLabel htmlFor="shell" className="text-xs">
                Shell
              </FieldLabel>
              <Select
                value={formData.shell}
                onValueChange={(value) => handleChange("shell", value)}
              >
                <SelectTrigger id="shell" className="w-full">
                  <SelectValue placeholder="选择Shell" />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    {ShellOptions.map((option) => (
                      <SelectItem key={option.value} value={option.value}>
                        {option.label}
                      </SelectItem>
                    ))}
                  </SelectGroup>
                </SelectContent>
              </Select>
            </Field>

            {/* 工作路径 */}
            <Field className="gap-2">
              <FieldLabel htmlFor="workpath" className="text-xs">
                工作路径
              </FieldLabel>
              <Input
                id="workpath"
                className="text-xs"
                value={formData.workpath}
                onChange={(e) => handleChange("workpath", e.target.value)}
              />
            </Field>

            {/* 初始执行命令 */}
            <Field className="gap-2">
              <FieldLabel htmlFor="initialCommand" className="text-xs">
                初始执行命令
              </FieldLabel>
              <Textarea
                id="initialCommand"
                className="text-xs resize-none"
                value={formData.initialCommand}
                onChange={(e) => handleChange("initialCommand", e.target.value)}
              />
            </Field>

            {/* 备注 */}
            <Field className="gap-2">
              <FieldLabel htmlFor="remark" className="text-xs">
                备注
              </FieldLabel>
              <Textarea
                id="remark"
                className="text-xs resize-none"
                value={formData.remark}
                onChange={(e) => handleChange("remark", e.target.value)}
              />
            </Field>
          </FieldGroup>
        </FieldSet>
      </form>
    </Card>
  );
};

export default TerminalForm;
