# AhaVault - 前端项目

## 技术栈
- React 18+ (TypeScript)
- Vite (构建工具)
- TailwindCSS (样式)
- Axios (HTTP 客户端)
- Tus-JS-Client (断点续传)
- Web Worker (SHA-256 计算)

## 目录结构

```
web/
├── src/
│   ├── components/       # React 组件
│   │   ├── common/       # 通用组件 (Button, Input, Modal)
│   │   ├── upload/       # 上传相关组件
│   │   ├── share/        # 分享相关组件
│   │   └── admin/        # 管理员组件
│   ├── pages/            # 页面组件
│   │   ├── Home.tsx      # 首页 (取件码输入)
│   │   ├── Cabinet.tsx   # 我的文件柜
│   │   ├── Share.tsx     # 分享管理
│   │   └── Admin.tsx     # 管理后台
│   ├── services/         # API 服务层
│   │   ├── api.ts        # Axios 实例配置
│   │   ├── fileService.ts
│   │   └── shareService.ts
│   ├── hooks/            # 自定义 Hooks
│   ├── utils/            # 工具函数
│   │   ├── crypto.ts     # 加密相关
│   │   └── format.ts     # 格式化工具
│   ├── workers/          # Web Workers
│   │   └── sha256.worker.ts
│   ├── types/            # TypeScript 类型定义
│   ├── assets/           # 静态资源
│   └── styles/           # 全局样式
├── public/               # 公共资源
└── package.json
```

## 开发指南

### 安装依赖
```bash
npm install
```

### 启动开发服务器
```bash
npm run dev
```

### 构建生产版本
```bash
npm run build
```

## 核心功能模块

### 1. 文件上传
- 使用 Web Worker 计算 SHA-256 避免 UI 阻塞
- Tus 协议支持断点续传
- 前端 Magic Bytes 校验

### 2. 取件码系统
- 8 位取件码输入 (带遮罩)
- IP 限流 + Captcha 验证
- 访问密码二次验证

### 3. 主题切换
- 支持深色/浅色模式
- 使用 TailwindCSS dark: 前缀

## 注意事项

- 所有 API 调用必须通过 HTTPS
- 敏感数据不得存储在 localStorage
- 使用 React.StrictMode 开发
