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

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import WindowHeader from "@/components/WindowHeader";
import StandardForm, {
  AuthMethod,
  Environment,
} from "./components/StandardForm";
import type {
  ConnectionEditInitialData,
  ConnectionStandard,
} from "@/types/initial-data";
import { useInitialData } from "@/hooks/useInitialData";
import { useEffect, useRef, useState } from "react";
import { DataSyncAPI } from "@/lib/data-sync";
import { DataChannel } from "@/store/data-sync.store";
import { useWindowListener } from "@/hooks/useWindowListener";
import { currentPageId, currentWindowName } from "@/lib/utils";
import { WindowService } from "@wails/service";

const defaultStandardFormData: ConnectionStandard = {
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

function ConnectionEdit() {
  const isEdit = useRef(false);

  // 接收初始数据
  const { initialData, isLoading, clearInitialData } =
    useInitialData<ConnectionEditInitialData>();

  useWindowListener(({ isClose, data }) => {
    if (isClose && data.name === currentWindowName()) {
      console.log("当前窗口已关闭, 清理相关资源");
      clearInitialData();
    }
  });

  // 检查初始数据是否加载
  useEffect(() => {
    if (initialData && !isLoading) {
      console.log("=== 接收到的初始数据 ===");
      console.log("数据:", initialData);
      if (initialData.data.uuid) {
        isEdit.current = true;
      }
      console.log("====================");
    }
  }, [initialData, isLoading]);

  // 处理标准配置表单提交
  const handleSaveStandard = (data: ConnectionStandard) => {
    console.log("=== 保存连接配置 ===");
    console.log("标准配置:", data);
    console.log("==================");

    // 发送给主窗口
    DataSyncAPI.sendToWindow(
      "main",
      DataChannel.Connection,
      isEdit.current ? "connection:update" : "connection:save",
      { ...data, uuid: initialData?.data.uuid },
    );

    // 关闭当前窗口
    WindowService.ClosePage(currentPageId()).catch((err) => {
      console.error("关闭窗口失败:", err);
    });
  };

  return (
    <div className="flex flex-col h-screen bg-background w-screen">
      <WindowHeader title="MYSQL 配置编辑" />
      <div className="px-8 pt-2 pb-4 flex-1 flex flex-col">
        <div className="flex-1 w-full">
          <Tabs
            defaultValue="overview"
            className="w-full h-full flex flex-col items-center"
          >
            <TabsList className="shrink-0">
              <TabsTrigger value="overview">标准</TabsTrigger>
              <TabsTrigger value="analytics">高级</TabsTrigger>
              <TabsTrigger value="reports">参数</TabsTrigger>
              <TabsTrigger value="settings">SSL</TabsTrigger>
            </TabsList>
            <TabsContent value="overview" className="w-full flex-1">
              <StandardForm
                initialData={{
                  ...defaultStandardFormData,
                  ...initialData?.data.standard,
                }}
                onSubmit={handleSaveStandard}
              />
            </TabsContent>
            <TabsContent value="analytics" className="w-full">
              <Card>
                <CardHeader>
                  <CardTitle>Analytics</CardTitle>
                  <CardDescription>
                    Track performance and user engagement metrics. Monitor
                    trends and identify growth opportunities.
                  </CardDescription>
                </CardHeader>
                <CardContent className="text-muted-foreground text-sm">
                  Page views are up 25% compared to last month.
                </CardContent>
              </Card>
            </TabsContent>
            <TabsContent value="reports">
              <Card>
                <CardHeader>
                  <CardTitle>Reports</CardTitle>
                  <CardDescription>
                    Generate and download your detailed reports. Export data in
                    multiple formats for analysis.
                  </CardDescription>
                </CardHeader>
                <CardContent className="text-muted-foreground text-sm">
                  You have 5 reports ready and available to export.
                </CardContent>
              </Card>
            </TabsContent>
            <TabsContent value="settings">
              <Card>
                <CardHeader>
                  <CardTitle>Settings</CardTitle>
                  <CardDescription>
                    Manage your account preferences and options. Customize your
                    experience to fit your needs.
                  </CardDescription>
                </CardHeader>
                <CardContent className="text-muted-foreground text-sm">
                  Configure notifications, security, and themes.
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>
        </div>
        <footer className=" mt-3 flex shrink-0 items-center justify-between">
          <Button
            variant="ghost"
            size="sm"
            className=" text-green-600 hover:text-green-600"
          >
            测试连接
          </Button>
          <Button
            variant="ghost"
            size="sm"
            className=" text-blue-600 hover:text-blue-600"
            type="submit"
            form="standard-form"
          >
            保存
          </Button>
        </footer>
      </div>
    </div>
  );
}

export default ConnectionEdit;
