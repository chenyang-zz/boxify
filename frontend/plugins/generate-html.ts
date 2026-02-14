import fs from "fs";
import path from "path";
import { resolve } from "path";

// 页面配置接口
export interface PageConfig {
  /** 页面唯一标识 */
  id: string;
  /** 页面标题 */
  title: string;
  /** React 入口文件路径（相对于 src） */
  entry: string;
  /** 挂载容器的 DOM ID */
  containerId: string;
  /** 路由路径（可选，默认使用 id） */
  route?: string;
  /** 是否为主应用（默认：false） */
  isMain?: boolean;
}

// 从 JSON 文件加载页面配置
function loadPagesConfig(): PageConfig[] {
  const configPath = resolve(__dirname, "../../page.config.json");
  const configContent = fs.readFileSync(configPath, "utf-8");
  const config = JSON.parse(configContent);
  return config.pages || [];
}

export const pagesConfig: PageConfig[] = loadPagesConfig();

// 获取所有页面入口点（用于 Vite 配置）
export function getPageEntries() {
  return pagesConfig.reduce(
    (acc, page) => {
      const htmlPath = resolve(__dirname, "..", `${page.id}.html`);
      acc[page.id] = htmlPath;
      return acc;
    },
    {} as Record<string, string>,
  );
}

// HTML 生成插件
function generateHTML() {
  return {
    name: "generate-html",
    configResolved() {
      const pagesConfig = loadPagesConfig();

      const HTML_TEMPLATE = (page: {
        id: string;
        title: string;
        entry: string;
        containerId: string;
        window: {
          name: string;
        };
      }) =>
        `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
  <meta name="page-id" content="${page.id}"/>
  <meta name="window-name" content="${page.window.name}"/>
  <title>${page.title}</title>
</head>
<body>
  <div id="${page.containerId}"></div>
  <script src="./src/${page.entry}" type="module"></script>
</body>
</html>
`;

      pagesConfig.forEach(
        (page: {
          id: string;
          title: string;
          entry: string;
          containerId: string;
          window: {
            name: string;
          };
        }) => {
          console.log(page.title);

          const htmlContent = HTML_TEMPLATE(page);
          const htmlPath = path.resolve(__dirname, "..", `${page.id}.html`);
          fs.writeFileSync(htmlPath, htmlContent, "utf-8");
          console.log(`✓ Generated: ${page.id}.html`);
        },
      );
      console.log(`✓ Generated ${pagesConfig.length} page(s)\n`);
    },
  };
}

export default generateHTML;
