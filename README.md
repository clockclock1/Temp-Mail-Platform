# Temp Mail Platform (Go + Vue)

一个临时邮箱系统，支持子域名 MX 收信、登录后创建临时邮箱、完整 RBAC 权限管理。
生产部署默认采用单镜像单容器（Go API + SMTP + Vue 前端静态页面）。

## 功能特性

- 子域名 MX 即可接收邮件
- 登录后才能创建邮箱
- 用户/角色/权限（RBAC）
- SMTP 收件 + SQLite 文件数据库 + 原始 EML 保存
- 兼容旧接口格式（`/api/mails`、`/admin/new_address`、`/user_api/mails`）
- GitHub Actions 跨平台编译与多架构镜像构建

## 目录结构

- `backend`: Go 服务（HTTP API + SMTP）
- `frontend`: Vue 控制台源码（构建后由后端静态托管）
- `.github/workflows`: CI/CD

## 环境变量

可参考 `backend/.env.example`：

- `APP_NAME`: 应用名
- `HTTP_ADDR`: HTTP 监听地址，默认 `:8080`
- `SMTP_ADDR`: SMTP 监听地址，默认 `:2525`
- `WEB_DIR`: 前端静态文件目录，默认 `./web`
- `JWT_SECRET`: JWT 密钥（生产必须修改）
- `JWT_EXPIRE_HOURS`: 用户 JWT 过期小时
- `LEGACY_ADMIN_AUTH`: 旧接口 `x-admin-auth`
- `LEGACY_CUSTOM_AUTH`: 旧接口可选二次密钥（为空表示不启用）
- `LEGACY_ADDRESS_JWT_EXPIRE_HOURS`: 地址 JWT 过期小时
- `DB_PATH`: SQLite 文件路径
- `DATA_DIR`: 邮件原文目录
- `CORS_ORIGINS`: 允许跨域来源
- `DEFAULT_ADMIN_USER`: 默认管理员用户名
- `DEFAULT_ADMIN_PASS`: 默认管理员密码
- `CLEANUP_INTERVAL_MINUTES`: 过期邮箱清理周期

## 部署方式总览

- 方式 A: 本地开发（前后端分离）
- 方式 B: Linux 二进制直跑（systemd）
- 方式 C: Docker Compose 单容器
- 方式 D: Docker Run 单容器
- 方式 E: GitHub Release 自动构建镜像 + 拉取部署
- 方式 F: ClawCloud Run 部署

---

## 方式 A: 本地开发（前后端分离）

### 1) 启动后端

```bash
cd backend
cp .env.example .env
go mod tidy
go run ./cmd/server
```

默认地址：

- API: `http://localhost:8080/api/v1`
- Health: `http://localhost:8080/healthz`

### 2) 启动前端开发服务器

```bash
cd frontend
npm install
npm run dev
```

默认地址：`http://localhost:5173`

---

## 方式 B: Linux 二进制直跑（systemd）

### 1) 编译二进制

```bash
cd backend
go build -o tempmail ./cmd/server
```

### 2) 准备目录

```bash
sudo mkdir -p /opt/tempmail/{data,web}
sudo cp -r ../frontend/dist/* /opt/tempmail/web/
sudo cp tempmail /opt/tempmail/tempmail
sudo cp .env.example /opt/tempmail/.env
```

### 3) systemd 示例

`/etc/systemd/system/tempmail.service`

```ini
[Unit]
Description=TempMail Service
After=network.target

[Service]
Type=simple
WorkingDirectory=/opt/tempmail
EnvironmentFile=/opt/tempmail/.env
ExecStart=/opt/tempmail/tempmail
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now tempmail
```

---

## 方式 C: Docker Compose 单容器（推荐）

```bash
cp backend/.env.example backend/.env
docker compose up -d --build
```

访问：

- Web 控制台: `http://localhost:8080`
- API: `http://localhost:8080/api/v1`
- SMTP: `localhost:2525`

---

## 方式 D: Docker Run 单容器

### 1) 本地构建镜像

```bash
docker build -f backend/Dockerfile -t tempmail:local .
```

### 2) 启动容器

```bash
docker run -d --name tempmail \
  --restart unless-stopped \
  -p 8080:8080 \
  -p 2525:2525 \
  -v $(pwd)/data/backend:/app/data \
  --env-file backend/.env \
  tempmail:local
```

---

## 方式 E: GitHub Release 自动构建镜像 + 拉取部署

### 触发规则

当前 workflow 已设置为仅在 **发布新 Release（published）** 时触发。

- Go 多平台构建：`.github/workflows/go-build.yml`
- Docker 多架构镜像：`.github/workflows/docker-multiarch.yml`

镜像会推送到：

- `ghcr.io/<owner>/tempmail:<tag>`
- `ghcr.io/<owner>/tempmail:latest`（默认分支）

### 拉取并部署

```bash
docker pull ghcr.io/<owner>/tempmail:<tag>

docker run -d --name tempmail \
  --restart unless-stopped \
  -p 8080:8080 \
  -p 2525:2525 \
  -v /opt/tempmail/data:/app/data \
  --env-file /opt/tempmail/.env \
  ghcr.io/<owner>/tempmail:<tag>
```

---

## 方式 F: ClawCloud Run 部署

基于 ClawCloud Run 的 App Launchpad，从 Docker 镜像直接部署。

### 1) 准备镜像

优先使用 GitHub Release 自动产物：

- `ghcr.io/<owner>/tempmail:<release-tag>`

如果是私有镜像，在 ClawCloud Run 中配置镜像仓库账号密码。

### 2) 创建应用

在 ClawCloud Run 控制台：

1. 打开 `App Launchpad` -> `Create App`
2. Image Type 选择 `Public` 或 `Private`
3. Image Name 填写 `ghcr.io/<owner>/tempmail:<tag>`
4. 资源建议：至少 `0.5 CPU / 512MB`
5. Container Port 填 `8080`
6. 开启 Public Network（用于 Web/API）

### 3) 配置环境变量

至少配置：

- `JWT_SECRET`
- `DEFAULT_ADMIN_USER`
- `DEFAULT_ADMIN_PASS`
- `DB_PATH=/app/data/tempmail.db`
- `DATA_DIR=/app/data/messages`
- `WEB_DIR=/app/web`
- `HTTP_ADDR=:8080`
- `SMTP_ADDR=:2525`

### 4) 配置持久化存储

在 Local Storage / Volume 中挂载到：

- `/app/data`

否则容器重启后数据库和邮件会丢失。

### 5) 域名与 HTTPS

- 可直接使用平台分配的 Public Address（HTTPS）
- 可绑定自定义域名（在域名商添加 CNAME 后绑定）

### 6) SMTP/MX 注意事项

临时邮箱要真正收信，MX 必须能到达 SMTP 入口。

- 本服务 SMTP 监听端口是容器内 `2525`
- 若你的 ClawCloud Run 环境支持对外暴露 TCP SMTP 端口，可将 MX 指向该公网入口
- 若当前环境仅适合 HTTP 暴露，建议使用一台可开放 25 端口的网关/VPS 做 SMTP 转发到本服务

---

## 子域名 MX 配置示例

假设要使用 `mail.example.com` 收件：

1. `A` 记录：`mail.example.com` -> 服务器公网 IP
2. `MX` 记录：`mail.example.com` -> `mail.example.com`
3. 放行 SMTP 端口（生产通常公网 25 -> 容器 2525）
4. 登录系统后在域名管理中新增 `mail.example.com`
5. 创建地址如 `demo@mail.example.com` 测试收信

---

## API 文档（新接口）

基地址：`/api/v1`

- 认证
  - `POST /auth/login`
  - `GET /auth/me`
- 域名
  - `GET /domains/available`
  - `GET /domains`
  - `POST /domains`
  - `PUT /domains/:id`
  - `DELETE /domains/:id`
- 邮箱
  - `GET /mailboxes`
  - `POST /mailboxes`
  - `DELETE /mailboxes/:id`
  - `GET /mailboxes/:id/messages`
- 邮件
  - `GET /messages/:id`
  - `GET /messages/:id/raw`
  - `DELETE /messages/:id`
- 权限
  - `GET /users`、`POST /users`、`PATCH /users/:id`、`DELETE /users/:id`
  - `GET /roles`、`POST /roles`、`PUT /roles/:id`、`DELETE /roles/:id`
  - `GET /permissions`
- 统计
  - `GET /stats`

## API 文档（兼容旧格式）

已兼容 `查看邮件.md` 和 `新建邮箱地址.md` 中的调用方式：

- `POST /admin/new_address`（`x-admin-auth`）
- `POST /api/new_address`（`Authorization: Bearer <用户JWT>` 或 `x-user-token`）
- `GET /api/mails`（`Authorization: Bearer <address-jwt>`）
- `GET /admin/mails`（`x-admin-auth`，支持 `address` 过滤）
- `DELETE /admin/mails/:id`
- `DELETE /admin/delete_address/:id`
- `DELETE /admin/clear_inbox/:id`
- `DELETE /admin/clear_sent_items/:id`
- `POST /user_api/login`
- `POST /user_api/register`
- `GET /user_api/mails`（`x-user-token`）

---

## 默认管理员

- 用户名：`DEFAULT_ADMIN_USER`（默认 `admin`）
- 密码：`DEFAULT_ADMIN_PASS`（默认 `admin123456`）

生产环境请务必修改。

## 生产建议

- 修改 `JWT_SECRET`、默认管理员密码
- 给 Web/API 启用 HTTPS
- SMTP 入口做好防滥用策略（限流、防垃圾）
- 定期备份 `/app/data`
- 建议加监控与告警（CPU、内存、磁盘、SMTP 请求量）