# Git 提交规范

使用中文提交信息，格式：

```
<图标> <类型>(<范围>): <简短描述>

<详细描述（可选）>

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
```

## 类型

| 类型 | 图标 | 说明 |
|------|------|------|
| `feat` | ✨ | 新功能 |
| `fix` | 🐛 | 修复 bug |
| `refactor` | ♻️ | 重构代码（不改变功能） |
| `docs` | 📝 | 文档更新 |
| `style` | 💄 | 代码格式调整（不影响逻辑） |
| `test` | ✅ | 测试相关 |
| `chore` | 🔧 | 构建/工具链相关 |
| `perf` | ⚡ | 性能优化 |
| `ci` | 👷 | CI/CD 相关 |
| `revert` | ⏪ | 回滚提交 |

## 范围（可选）

范围用于说明提交影响的模块，例如：
- `terminal`: 终端模块
- `git`: Git 相关功能
- `ui`: 界面相关
- `api`: API 接口

## 示例

```
✨ feat(terminal): 添加目录选择器搜索功能

- 支持模糊搜索过滤目录
- 高亮匹配的搜索关键词
- 添加键盘快捷键支持

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
```

```
🐛 fix(git): 修复路径解析错误

修复了包含空格的路径无法正确解析的问题

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
```
