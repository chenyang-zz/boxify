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
import StandardForm from "./components/StandardForm";
import { useRef } from "react";

function ConnectionEdit() {
  const standardFormRef = useRef<HTMLFormElement>(null);

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
              <StandardForm ref={standardFormRef} />
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
            onClick={() => {
              standardFormRef.current?.requestSubmit();
            }}
          >
            保存
          </Button>
        </footer>
      </div>
    </div>
  );
}

export default ConnectionEdit;
