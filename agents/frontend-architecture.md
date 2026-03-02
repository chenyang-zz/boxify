# Boxify 前端架构（当前仓库）

更新时间：2026-03-02

```text
Boxify/frontend/
├── package.json
├── pnpm-lock.yaml
├── vite.config.ts
├── tsconfig.json
├── index.html
├── src/
│   ├── main.tsx                    # 前端入口
│   ├── App.tsx                     # 主应用组件
│   ├── style.css
│   ├── assets/                     # 字体、图片等资源
│   ├── common/                     # 通用常量/枚举
│   ├── constants/                  # 业务常量
│   ├── components/                 # 功能组件与 UI 组件
│   │   ├── ui/
│   │   ├── TitleBar/
│   │   ├── PropertyTree/
│   │   ├── UtilBar/
│   │   ├── DBTable/
│   │   ├── FileTree/
│   │   ├── Terminal/
│   │   └── WindowHeader/
│   ├── hooks/                      # React Hooks
│   ├── lib/                        # 前端业务库（连接、SQL、属性、同步）
│   ├── pages/                      # 页面级模块（main/settings/connectionEdit）
│   ├── providers/                  # Provider（如 ThemeProvider）
│   ├── store/                      # 状态管理
│   └── types/                      # 前端类型定义
├── bindings/                       # Wails 生成绑定
├── public/                         # 静态资源
├── dist/                           # 构建产物
└── plugins/                        # 前端插件
```

## 分层摘要

1. 启动层：`src/main.tsx`
2. 页面层：`src/pages`
3. 组件层：`src/components`
4. 状态与逻辑层：`src/store`、`src/hooks`、`src/lib`
5. 桥接层：`bindings`（Wails 前后端调用绑定）
