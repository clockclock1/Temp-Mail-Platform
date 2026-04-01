# Temp Mail Platform (Go + Vue)

一个临时邮箱系统，Docker 部署时前后端打包为同一镜像（单容器），满足以下目标：

- 只需给子域名配置 MX 记录即可接收邮件
- 必须登录后才能创建邮箱
- 完整用户权限管理（用户、角色、权限）
- 后端 Go，前端 Vue，数据库使用文件型 SQLite
- 支持 GitHub Actions 的 Go 全平台编译和 Docker 多架构构建

## 架构

- 后端: `backend` (Go + Gin + GORM + SQLite)
- 前端: `frontend` (Vue3 + Vite)
- Docker 生产部署: 单镜像（Go API + SMTP + Vue 静态站点）
- 数据库文件: `backend/data/tempmail.db`（可改）
- 邮件原文: `backend/data/messages/*.eml`
- SMTP 收信服务: 默认 `:2525`
- HTTP API: 默认 `:8080`

## 核心能力

- 登录鉴权: JWT
- 权限管理: RBAC（用户/角色/权限）
- 域名管理: 启用/禁用接收域名
- 临时邮箱: 支持 TTL 过期时间
- 邮件收信: SMTP 收件后存储到 SQLite + 原始 EML 文件
- 邮件查看: 列表、详情、删除、下载原始邮件
- 统计接口: 用户/域名/邮箱/邮件数量

## 子域名 MX 配置

假设你要用 `mail.example.com` 作为收件域名：

1. 给 `mail.example.com` 添加 `A` 记录，指向你的服务器公网 IP。
2. 给 `mail.example.com` 添加 `MX` 记录，值指向 `mail.example.com`（或你的邮件入口主机名）。
3. 确保公网可访问 SMTP 25 端口，并转发到本服务监听端口（默认容器内 `2525`）。
4. 登录系统后在域名管理里新增 `mail.example.com`。
5. 创建邮箱如 `test@mail.example.com` 即可收信。

## 后端本地运行

```bash
cd backend
cp .env.example .env
go mod tidy
go run ./cmd/server
```

默认管理员账号来自环境变量：

- `DEFAULT_ADMIN_USER=admin`
- `DEFAULT_ADMIN_PASS=admin123456`

## 前端本地运行

```bash
cd frontend
npm install
npm run dev
```

默认访问：`http://localhost:5173`

## Docker 运行（单容器）

```bash
cp backend/.env.example backend/.env
docker compose up -d --build
```

- Web 控制台: `http://localhost:8080`
- 后端 API: `http://localhost:8080/api/v1`
- SMTP: `localhost:2525`（生产通常映射公网 25）

## 主要 API

基地址: `/api/v1`

- 认证
  - `POST /auth/login`
  - `GET /auth/me`
- 域名
  - `GET /domains/available`（登录后可见）
  - `GET /domains`（需 `domain:manage`）
  - `POST /domains`（需 `domain:manage`）
  - `PUT /domains/:id`（需 `domain:manage`）
  - `DELETE /domains/:id`（需 `domain:manage`）
- 邮箱
  - `GET /mailboxes`（需 `mailbox:read`）
  - `POST /mailboxes`（需 `mailbox:create`）
  - `DELETE /mailboxes/:id`（需 `mailbox:delete`）
  - `GET /mailboxes/:id/messages`（需 `message:read`）
- 邮件
  - `GET /messages/:id`（需 `message:read`）
  - `GET /messages/:id/raw`（需 `message:read`）
  - `DELETE /messages/:id`（需 `message:delete`）
- 权限管理
  - `GET /users`、`POST /users`、`PATCH /users/:id`、`DELETE /users/:id`（需 `user:manage`）
  - `GET /roles`、`POST /roles`、`PUT /roles/:id`、`DELETE /roles/:id`（需 `role:manage`）
  - `GET /permissions`（需 `role:manage`）
- 统计
  - `GET /stats`（需 `stats:read`）

## 兼容旧格式 API（与你的文档一致）

已兼容 `查看邮件.md` 和 `新建邮箱地址.md` 里的调用路径与头部格式：

- `POST /admin/new_address`（`x-admin-auth`）
- `POST /api/new_address`（`Authorization: Bearer <用户JWT>` 或 `x-user-token`，也支持管理员头）
- `GET /api/mails`（`Authorization: Bearer <address jwt>`）
- `GET /admin/mails`（`x-admin-auth`，支持 `address` 过滤）
- `DELETE /admin/mails/:id`（`x-admin-auth`）
- `DELETE /admin/delete_address/:id`（`x-admin-auth`）
- `DELETE /admin/clear_inbox/:id`（`x-admin-auth`）
- `DELETE /admin/clear_sent_items/:id`（`x-admin-auth`，当前实现为兼容 no-op）
- `POST /user_api/login`、`POST /user_api/register`
- `GET /user_api/mails`（`x-user-token`）

### 兼容接口相关环境变量

- `LEGACY_ADMIN_AUTH`: 对应 `x-admin-auth`
- `LEGACY_CUSTOM_AUTH`: 若非空，则兼容接口必须额外携带 `x-custom-auth`
- `LEGACY_ADDRESS_JWT_EXPIRE_HOURS`: `/api/new_address`、`/admin/new_address` 返回的地址 JWT 过期时间

## GitHub Actions

- Go 多平台编译: `.github/workflows/go-build.yml`
  - Linux/Windows/macOS + amd64/arm64（含 Linux arm）
- Docker 多架构构建: `.github/workflows/docker-multiarch.yml`
  - `linux/amd64`, `linux/arm64`, `linux/arm/v7`
  - 推送到 `ghcr.io/<owner>/tempmail`（单镜像）

## 生产建议

- 修改 `JWT_SECRET`、管理员默认密码
- 在反向代理层启用 HTTPS
- 开放并保护 SMTP 25 端口
- 定期备份 `data` 目录
