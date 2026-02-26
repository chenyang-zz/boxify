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

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import WindowHeader from "@/components/WindowHeader";
import StandardForm from "./components/StandardForm";
import type {
  ConnectionEditInitialData,
  CommonStandard,
  TerminalStandard,
  ConnectionStandard,
} from "@/types/initial-data";
import { useInitialData } from "@/hooks/useInitialData";
import { FC, useEffect, useMemo, useRef, useState } from "react";
import { DataSyncAPI } from "@/lib/data-sync";
import { DataChannel } from "@/store/data-sync.store";
import { useWindowListener } from "@/hooks/useWindowListener";
import { callWails, currentPageId, currentWindowName } from "@/lib/utils";
import { TerminalConfig, TerminalService, WindowService } from "@wails/service";
import { generateDefaultFormData } from "./lib";
import { ConnectionEnum } from "@/common/constrains";
import TerminalForm from "./components/TerminalForm";
import { toast } from "sonner";
import { Spinner } from "@/components/ui/spinner";

let isTest = false;

function ConnectionEdit() {
  const isEdit = useRef(false);
  const [loading, setLoading] = useState(false);

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
  const handleSave = async (data: ConnectionStandard) => {
    console.log("=== 配置 ===");
    console.log("标准配置:", data);
    console.log("==================");

    if (isTest) {
      try {
        setLoading(true);
        switch (initialData?.data.type) {
          case ConnectionEnum.TERMINAL:
            const terminalData = data as TerminalStandard;
            await callWails(
              TerminalService.TestConfig,
              TerminalConfig.createFrom({
                shell: terminalData.shell,
                workPath: terminalData.workpath,
                initialCommand: terminalData.initialCommand,
              }),
            );
            break;
        }
        toast.success("Success", {
          description: "连接成功",
          style: {
            textAlign: "left",
          },
        });
      } catch {
      } finally {
        setLoading(false);
      }
      return;
    }

    // 发送给主窗口
    DataSyncAPI.sendToWindow(
      "main",
      DataChannel.Connection,
      isEdit.current ? "connection:update" : "connection:save",
      { ...data, uuid: initialData?.data.uuid, type: initialData?.data.type },
    );

    // 关闭当前窗口
    WindowService.ClosePage(currentPageId()).catch((err) => {
      console.error("关闭窗口失败:", err);
    });
  };

  const formItem = useMemo(() => {
    if (initialData?.data.type === ConnectionEnum.TERMINAL) {
      return (
        <TerminalForm
          initialData={{
            ...(generateDefaultFormData(
              initialData?.data.type,
              "standard",
            ) as TerminalStandard),
            ...(initialData?.data.standard as TerminalStandard),
          }}
          onSubmit={handleSave}
        />
      );
    }

    return (
      <Tabs
        defaultValue="standard"
        className="w-full h-full flex flex-col items-center"
      >
        <TabsList className="shrink-0">
          <TabsTrigger value="overview">标准</TabsTrigger>
          {/* <TabsTrigger value="analytics">高级</TabsTrigger>
              <TabsTrigger value="reports">参数</TabsTrigger>
              <TabsTrigger value="settings">SSL</TabsTrigger> */}
        </TabsList>
        <TabsContent value="overview" className="w-full flex-1">
          <StandardForm
            initialData={{
              ...(generateDefaultFormData(
                initialData?.data.type,
                "standard",
              ) as CommonStandard),
              ...(initialData?.data.standard as CommonStandard),
            }}
            onSubmit={handleSave}
          />
        </TabsContent>
        {/* <TabsContent value="analytics" className="w-full">
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
            </TabsContent> */}
      </Tabs>
    );
  }, [initialData]);

  return (
    <div className="flex flex-col h-screen bg-background w-screen">
      <WindowHeader title={initialData?.data?.title || ""} />
      <div className="px-8 pt-2 pb-4 flex-1 flex flex-col">
        <div className="flex-1 w-full">{formItem}</div>
        <footer className=" mt-3 flex shrink-0 items-center justify-between">
          <Button
            variant="ghost"
            size="sm"
            disabled={loading}
            className=" text-green-600 hover:text-green-600"
            type="submit"
            form="standard-form"
            onClick={() => (isTest = true)}
          >
            {loading && <Spinner data-icon="inline-start" />}
            测试连接
          </Button>
          <Button
            variant="ghost"
            size="sm"
            className=" text-blue-600 hover:text-blue-600"
            type="submit"
            form="standard-form"
            onClick={() => (isTest = false)}
          >
            保存
          </Button>
        </footer>
      </div>
    </div>
  );
}

export default ConnectionEdit;
