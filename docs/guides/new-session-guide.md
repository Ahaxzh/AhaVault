# 🚀 Claude 新会话快速启动指南

**目的**: 确保新的 Claude 会话能快速进入工作状态
**适用场景**: 当你需要开启新的 Claude Code 会话时
**预计耗时**: 1 分钟

---

## 📋 新会话启动步骤

### 方式一：标准启动（推荐）⭐

**你只需复制粘贴以下消息给新 Claude 会话**：

```
请先阅读以下文档了解项目状态：
1. docs/knowledge/INDEX.md
2. docs/memory-bank/progress.md

然后告诉我：
- 当前项目进度
- 下一步优先级任务
- 是否有阻塞问题
```

**Claude 会自动**：
1. ✅ 读取 `docs/knowledge/INDEX.md`（50 行，了解文档结构）
2. ✅ 读取 `docs/memory-bank/progress.md`（100 行，了解当前进度）
3. ✅ 向你汇报："项目当前进度 X%，下一步任务是..."
4. ✅ 准备好开始工作

**预计耗时**: 30 秒

---

### 方式二：极速启动（适合紧急情况）

如果你只想让 Claude 快速开始某个具体任务：

```
请阅读 docs/memory-bank/progress.md，
然后开始执行优先级 P0 任务：补充后端测试。
```

---

### 方式三：从特定任务开始

如果你明确知道要做什么：

```
请阅读 docs/tasks/backend/05-api.md，
然后实现 handlers/upload.go 的 Tus 上传功能。
```

---

## 🎯 新会话的预期行为

### ✅ 好的 Claude 会话应该

1. **主动汇报进度**
   ```
   我已阅读进度文档，当前项目：
   - 整体进度: 40%
   - 后端核心: 80% (进行中)
   - 前端应用: 0% (未开始)
   - 下一步优先级任务: 补充后端测试 (P0)

   根据 docs/tasks/testing/backend-tests.md，
   我将开始补充 file_service_test.go。
   是否开始？
   ```

2. **遵循协作规范**
   - 按照 `Claude.md` 中的代码规范编写代码
   - 完成任务后立即更新 `docs/memory-bank/progress.md`
   - 提交代码时遵循 Conventional Commits

3. **提出合理问题**
   - 如果文档有歧义，会主动询问澄清
   - 不会自己猜测需求

### ❌ 不好的会话表现

1. **不读文档就开始**
   - 直接问："你想让我做什么？"
   - 需要你重复解释项目背景

2. **不更新文档**
   - 完成任务后忘记更新 `progress.md`
   - 不遵循文档更新规范

---

## 🔧 常见问题

### Q1: 新会话不知道项目进度怎么办？

**A**: 明确告诉它读哪些文档：
```
请依次阅读：
1. docs/knowledge/INDEX.md
2. docs/memory-bank/progress.md
3. docs/tasks/README.md
```

### Q2: 新会话不遵循代码规范怎么办？

**A**: 提醒它：
```
请遵循 Claude.md 中的代码规范：
- 文件必须有注释头
- 每个函数必须有注释
- 必须编写对应的测试文件
```

### Q3: 如何让新会话继续上一个会话的工作？

**A**:
```
请阅读 docs/memory-bank/progress.md 的"最近更新"部分，
继续上一个会话的工作：完成 handlers/upload.go。
```

### Q4: 会话中断后如何恢复？

**A**: Claude Code 会自动保存会话历史，你可以：
1. 使用 `/resume` 命令恢复会话
2. 或开启新会话并告诉它读取 `progress.md`

---

## 📝 给 Claude 的标准开场白模板

**复制粘贴即可使用**：

```
你好 Claude！欢迎加入 AhaVault 项目开发。

请按照以下步骤进入工作状态：

1. 阅读 docs/knowledge/INDEX.md - 了解文档结构
2. 阅读 docs/memory-bank/progress.md - 了解当前进度
3. 阅读 docs/tasks/README.md - 查看优先级任务

然后向我汇报：
- 当前项目进度
- 下一步优先级任务（P0/P1）
- 是否有阻塞问题

最后，遵循 Claude.md 中的协作规范开始工作。
```

---

## 🎯 高级用法

### 多任务并行

如果你想同时开启多个会话处理不同任务：

**会话 A（后端测试）**:
```
请专注于补充后端测试，阅读 docs/tasks/testing/backend-tests.md
```

**会话 B（API 开发）**:
```
请专注于实现 API Handler，阅读 docs/tasks/backend/05-api.md
```

### 指定技术栈

如果是前端任务，提醒 Claude 使用正确的技术栈：

```
请阅读 docs/memory-bank/tech-stack.md 的"前端技术栈"部分，
然后使用 React + TypeScript + TailwindCSS 开发前端组件。
```

---

## 🔄 会话工作流程（完整示例）

```
你:
  请阅读 docs/knowledge/INDEX.md 和 docs/memory-bank/progress.md，
  然后告诉我当前进度和下一步任务。

Claude:
  我已阅读文档，当前状态：
  - 整体进度: 40%
  - 后端核心: 80% (缺少测试)
  - 前端应用: 0%
  - P0 任务: 补充后端测试 (file_service_test.go)

  我将开始补充测试。是否开始？

你:
  可以，开始吧。记得遵循 Claude.md 的规范。

Claude:
  好的！我会：
  1. 编写符合规范的测试代码
  2. 确保覆盖率达到 70%
  3. 完成后更新 docs/memory-bank/progress.md
  4. 提交代码时使用 Conventional Commits

  现在开始编写 file_service_test.go...
  [开始工作]

Claude (完成后):
  ✅ 已完成 file_service_test.go
  ✅ 测试覆盖率: 75%
  ✅ 已更新 docs/memory-bank/progress.md
  ✅ 已提交代码: test(services): add file_service tests

  下一步任务: 补充 share_service_test.go
  是否继续？
```

---

## 📚 相关文档

- **协作规范**: [Claude.md](../../Claude.md)
- **知识索引**: [knowledge/INDEX.md](../knowledge/INDEX.md)
- **当前进度**: [memory-bank/progress.md](../memory-bank/progress.md)
- **任务列表**: [tasks/README.md](../tasks/README.md)

---

## 💡 最佳实践

1. ✅ **每次新会话都让 Claude 先读文档** - 确保它了解最新进度
2. ✅ **明确指出优先级** - 告诉 Claude 哪些是 P0 紧急任务
3. ✅ **提醒遵循规范** - 开场时提一句"遵循 Claude.md 规范"
4. ✅ **及时反馈** - 如果 Claude 做得好，夸奖它；如果有问题，及时纠正
5. ✅ **保持文档更新** - 完成任务后确保 `progress.md` 是最新的

---

## 🚫 常见错误

1. ❌ **不告诉 Claude 读文档** - 导致它不知道项目状态
2. ❌ **忘记提醒规范** - 导致代码质量不符合要求
3. ❌ **不更新 progress.md** - 导致下一个会话无法了解进度
4. ❌ **同时开多个会话但不协调** - 可能导致冲突

---

<div align="center">

**🎯 现在你知道如何高效开启新会话了！**

*保持文档更新，新会话无缝衔接*

**维护者**: Claude AI + 开发团队
**最后更新**: 2026-02-04

</div>
