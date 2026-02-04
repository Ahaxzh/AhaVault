# 前端页面开发任务清单

**模块名称**: 前端页面开发
**负责人**: Claude AI
**最后更新**: 2026-02-05
**当前进度**: 20%
**状态**: 🟡 进行中

---

## 📋 任务清单

### 核心页面

- [x] `pages/Home.tsx` - 首页（取件码输入 - App.tsx）
- [ ] `pages/Login.tsx` - 登录页
- [ ] `pages/Register.tsx` - 注册页
- [ ] `pages/Cabinet.tsx` - 我的文件柜
- [ ] `pages/Share.tsx` - 分享管理页
- [ ] `pages/Admin.tsx` - 管理后台
- [ ] `pages/NotFound.tsx` - 404 页面

### API 服务层

- [x] `services/api.ts` - Axios 实例配置
- [ ] `services/authService.ts` - 认证 API
- [ ] `services/fileService.ts` - 文件 API
- [x] `services/shareService.ts` - 分享 API
- [ ] `services/uploadService.ts` - Tus 上传封装

### 自定义 Hooks

- [ ] `hooks/useAuth.ts` - 认证状态管理
- [ ] `hooks/useFileUpload.ts` - 文件上传
- [ ] `hooks/useFileList.ts` - 文件列表
- [ ] `hooks/useShare.ts` - 分享管理
- [ ] `hooks/useToast.ts` - 消息提示

### 工具函数

- [ ] `utils/format.ts` - 格式化工具
- [ ] `utils/validation.ts` - 表单验证
- [ ] `utils/storage.ts` - LocalStorage 封装

### Web Workers

- [ ] `workers/sha256.worker.ts` - SHA-256 计算

---

**测试要求**:
- 每个页面需要 E2E 测试
- 每个服务/Hook/工具需要单元测试

**目标覆盖率**: ≥70%

---

**维护者**: Claude AI
**预计完成**: 2026-02-21
