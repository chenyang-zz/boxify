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
import {
  InputGroup,
  InputGroupAddon,
  InputGroupInput,
} from "@/components/ui/input-group";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { EyeIcon, EyeOffIcon } from "lucide-react";
import type { ConnectionStandard } from "@/types/initial-data";
import { FC, useEffect, useState } from "react";
import ColorTagSelector from "@/components/ColorTagSelector";
import ColorTagBadge from "@/components/ColorTagBadge";

export enum Environment {
  None = "none",
  Development = "development",
  Testing = "testing",
  Production = "production",
}

const environmentOptions = [
  { label: "无", value: Environment.None },
  { label: "开发", value: Environment.Development },
  { label: "测试", value: Environment.Testing },
  { label: "生产", value: Environment.Production },
];

export enum AuthMethod {
  Password = "password",
  InQuiry = "inquiry",
}

const authMethodOptions = [
  { label: "密码", value: AuthMethod.Password },
  { label: "每次询问", value: AuthMethod.InQuiry },
];

interface StandardFormProps {
  initialData?: ConnectionStandard; // 初始数据（编辑场景）
  onSubmit: (data: ConnectionStandard) => void; // 提交回调
}

interface FormErrors {
  name?: string;
  host?: string;
  user?: string;
  port?: string;
  password?: string;
}

const defaultFormData: ConnectionStandard = {
  tagColor: "",
  environment: Environment.None,
  name: "",
  host: "localhost",
  user: "root",
  port: 3306,
  validationWay: AuthMethod.Password,
  password: "",
  remark: "",
};

const StandardForm: FC<StandardFormProps> = ({ initialData, onSubmit }) => {
  // 表单状态管理
  const [formData, setFormData] = useState<ConnectionStandard>({
    ...defaultFormData,
  });

  const [errors, setErrors] = useState<FormErrors>({});
  const [showPassword, setShowPassword] = useState(false);

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
      newErrors.name = "连接名称不能为空";
    }
    if (!formData.host.trim()) {
      newErrors.host = "主机地址不能为空";
    }
    if (!formData.user.trim()) {
      newErrors.user = "用户名不能为空";
    }
    if (!formData.port || formData.port <= 0) {
      newErrors.port = "端口必须大于0";
    }
    if (formData.validationWay === "password" && !formData.password) {
      newErrors.password = "密码不能为空";
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
    field: keyof ConnectionStandard,
    value: string | number,
  ) => {
    setFormData((prev: ConnectionStandard) => ({ ...prev, [field]: value }));
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
                  value={formData.tagColor}
                  onChange={(color) => handleChange("tagColor", color)}
                />
              </Field>
              {/* 环境 */}
              <Field className="gap-2">
                <FieldLabel htmlFor="environment" className="text-xs">
                  环境
                </FieldLabel>
                <Select
                  value={formData.environment}
                  onValueChange={(value) => handleChange("environment", value)}
                >
                  <SelectTrigger id="environment" className="w-full">
                    <SelectValue placeholder="选择环境" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectGroup>
                      {environmentOptions.map((option) => (
                        <SelectItem key={option.value} value={option.value}>
                          {option.label}
                        </SelectItem>
                      ))}
                    </SelectGroup>
                  </SelectContent>
                </Select>
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

              {/* Host */}
              <Field className="gap-2">
                <FieldLabel htmlFor="host" className="text-xs">
                  Host
                </FieldLabel>
                <Input
                  id="host"
                  className={`text-xs ${errors.host ? "border-red-500" : ""}`}
                  value={formData.host}
                  onChange={(e) => handleChange("host", e.target.value)}
                />
              </Field>

              {/* User */}
              <Field className="gap-2">
                <FieldLabel htmlFor="user" className="text-xs">
                  User
                </FieldLabel>
                <Input
                  id="user"
                  className={`text-xs ${errors.user ? "border-red-500" : ""}`}
                  value={formData.user}
                  onChange={(e) => handleChange("user", e.target.value)}
                />
              </Field>

              {/* 端口 */}
              <Field className="gap-2">
                <FieldLabel htmlFor="port" className="text-xs">
                  端口
                </FieldLabel>
                <Input
                  id="port"
                  type="number"
                  className={`text-xs ${errors.port ? "border-red-500" : ""}`}
                  value={formData.port}
                  onChange={(e) =>
                    handleChange("port", parseInt(e.target.value) || 0)
                  }
                />
              </Field>

              {/* 验证方式 */}
              <Field className="gap-2">
                <FieldLabel htmlFor="auth-method" className="text-xs">
                  验证方式
                </FieldLabel>
                <Select
                  value={formData.validationWay}
                  onValueChange={(value) =>
                    handleChange("validationWay", value)
                  }
                >
                  <SelectTrigger id="auth-method" className="w-full">
                    <SelectValue placeholder="选择验证方式" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectGroup>
                      {authMethodOptions.map((option) => (
                        <SelectItem key={option.value} value={option.value}>
                          {option.label}
                        </SelectItem>
                      ))}
                    </SelectGroup>
                  </SelectContent>
                </Select>
              </Field>

              {/* 密码 - 仅在选择密码验证时显示 */}
              <Field className="gap-2">
                <FieldLabel htmlFor="password" className="text-xs">
                  密码
                </FieldLabel>
                <InputGroup
                  className={
                    errors.password && formData.validationWay === "password"
                      ? "border-red-500"
                      : ""
                  }
                >
                  <InputGroupInput
                    id="password"
                    type={showPassword ? "text" : "password"}
                    className={`text-xs`}
                    value={
                      formData.validationWay === "password"
                        ? formData.password
                        : ""
                    }
                    disabled={formData.validationWay !== "password"}
                    onChange={(e) => handleChange("password", e.target.value)}
                  />
                  <InputGroupAddon
                    align="inline-end"
                    className="cursor-pointer"
                    onClick={() => setShowPassword(!showPassword)}
                  >
                    {showPassword ? <EyeIcon /> : <EyeOffIcon />}
                  </InputGroupAddon>
                </InputGroup>
              </Field>
            </div>

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

export default StandardForm;
