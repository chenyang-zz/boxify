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

import { Button } from "@/components/ui/button";
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
import { EyeOffIcon } from "lucide-react";
import { FC } from "react";

enum Environment {
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

enum AuthMethod {
  Password = "password",
  InQuiry = "inquiry",
}

const authMethodOptions = [
  { label: "密码", value: AuthMethod.Password },
  { label: "每次询问", value: AuthMethod.InQuiry },
];

interface StandardFormProps {
  ref: React.Ref<HTMLFormElement>;
}

const StandardForm: FC<StandardFormProps> = ({ ref }) => {
  return (
    <Card className="w-full h-full p-5">
      <form
        ref={ref}
        onSubmit={(e) => {
          e.preventDefault();
          const formData = new FormData(e.currentTarget);
          const data = Object.fromEntries(formData.entries());

          console.log(data);
        }}
        className="h-full"
      >
        <FieldSet className="w-full ">
          <FieldGroup className="w-full gap-5">
            <div className="grid grid-cols-2 gap-5">
              <Field className="gap-2">
                <FieldLabel htmlFor="color-tag" className="text-xs">
                  颜色标签
                </FieldLabel>
                <Input
                  id="color-tag"
                  type="text"
                  className="text-xs"
                  placeholder="Max Leiter"
                />
              </Field>
              <Field className="gap-2">
                <FieldLabel htmlFor="environment" className="text-xs">
                  环境
                </FieldLabel>
                <Select defaultValue={Environment.None}>
                  <SelectTrigger id="environment" className="w-45">
                    <SelectValue />
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
              <Field className="gap-2">
                <FieldLabel htmlFor="name" className="text-xs">
                  名称
                </FieldLabel>
                <Input id="name" className="text-xs" />
              </Field>
              <Field className="gap-2">
                <FieldLabel htmlFor="host" className="text-xs">
                  Host
                </FieldLabel>
                <Input id="host" className="text-xs" />
              </Field>
              <Field className="gap-2">
                <FieldLabel htmlFor="user" className="text-xs">
                  User
                </FieldLabel>
                <Input id="user" className="text-xs" />
              </Field>
              <Field className="gap-2">
                <FieldLabel htmlFor="port" className="text-xs">
                  端口
                </FieldLabel>
                <Input id="port" type="number" className="text-xs" />
              </Field>
              <Field className="gap-2">
                <FieldLabel htmlFor="auth-method" className="text-xs">
                  验证方式
                </FieldLabel>
                <Select defaultValue={AuthMethod.Password}>
                  <SelectTrigger id="auth-method" className="w-45">
                    <SelectValue />
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
              <Field className="gap-2">
                <FieldLabel htmlFor="password" className="text-xs">
                  密码
                </FieldLabel>
                <InputGroup>
                  <InputGroupInput
                    id="password"
                    type="password"
                    className="text-xs"
                  />
                  <InputGroupAddon align="inline-end">
                    <EyeOffIcon />
                  </InputGroupAddon>
                </InputGroup>
              </Field>
            </div>

            <Field className="gap-2">
              <FieldLabel htmlFor="remark" className="text-xs">
                备注
              </FieldLabel>
              <Textarea id="remark" className="text-xs resize-none" />
            </Field>
          </FieldGroup>
        </FieldSet>
      </form>
    </Card>
  );
};

export default StandardForm;
