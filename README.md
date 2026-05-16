# Temp Mail Platform

一个可自托管的临时邮箱系统：

- 子域名绑定 MX 后即可收信
- 必须登录后才能创建邮箱
- 完整用户/角色/权限管理（RBAC）
- 前端可视化配置中心（管理员）
- 配置文件驱动（不依赖环境变量）
- 配置修改可实时生效（部分项会提示需重启）

## 技术栈

- 后端：Go + Gin + GORM + SQLite
- 前端：Vue 3 + Vite
- 邮件接收：SMTP（默认 `:2525`）
- 数据持久化：SQLite 文件 + EML 原文文件

## 配置方式（核心）

系统只使用 **YAML 配置文件**，默认启动参数：

```bash
./tempmail -config ./config.yaml
```

- 示例模板：`backend/config/config.example.yaml`
- Docker 默认路径：`/app/config/config.yaml`
- 前端配置中心：管理员登录后访问 `/config`

### 1) 配置文件怎么放

首次启动时，程序会按 `-config` 指定路径读取配置。

- 本地开发通常放在 `backend/../config/config.yaml`
- 二进制部署通常和可执行文件放在同级目录
- Docker Compose 默认挂载 `./config:/app/config`

### 2) 完整配置示例

```yaml
app_name: Temp Mail Service
http_addr: ":8080"
smtp_addr: ":2525"
web_dir: "./web"
jwt_secret: "change-me-in-production"
jwt_expire_hours: 24
legacy_admin_auth: "change-this-admin-auth"
legacy_custom_auth: ""
legacy_address_jwt_expire_hours: 720
legacy_allow_subdomain_match: true
db_path: "./data/tempmail.db"
data_dir: "./data/messages"
cors_origins:
  - "http://localhost:8080"
  - "http://localhost:5173"
default_admin_user: "admin"
default_admin_pass: "change-this-admin-password"
cleanup_interval_minutes: 10
```

### 3) 每个字段是什么意思

| 字段 | 作用 | 默认/建议 |
| --- | --- | --- |
| `app_name` | 后端服务名称 | 任意，展示用 |
| `http_addr` | Web/API 监听地址 | `:8080` |
| `smtp_addr` | SMTP 监听地址 | `:2525` |
| `web_dir` | 前端静态资源目录 | Docker 用 `./web` 或 `/app/web` |
| `jwt_secret` | 新版登录 Token 密钥 | 生产环境务必修改 |
| `jwt_expire_hours` | 新版登录有效期 | 默认 `24` |
| `legacy_admin_auth` | 旧版管理员鉴权串 | 兼容旧接口时需要 |
| `legacy_custom_auth` | 旧版自定义鉴权串 | 留空表示关闭 |
| `legacy_address_jwt_expire_hours` | 旧版邮箱 Token 有效期 | 默认 `720` |
| `legacy_allow_subdomain_match` | 是否允许子域名自动匹配 | `true`，支持多级域名 |
| `db_path` | SQLite 数据库文件 | `./data/tempmail.db` |
| `data_dir` | 邮件原文 `eml` 存放目录 | `./data/messages` |
| `cors_origins` | 允许的前端来源 | 按实际域名填写 |
| `default_admin_user` | 首次初始化管理员用户名 | 生产环境请改掉 |
| `default_admin_pass` | 首次初始化管理员密码 | 生产环境请改掉 |
| `cleanup_interval_minutes` | 过期邮箱清理周期 | 默认 `10` |

### 4) 改完会发生什么

- 修改 `cors_origins`、`jwt_secret`、`legacy_*`、`cleanup_interval_minutes`、`data_dir` 后可热生效
- 修改 `http_addr`、`smtp_addr`、`db_path`、`web_dir` 这类路径或监听项，通常需要重启服务
- 前端配置中心会自动写回同一份 `config.yaml`
- 配置保存后如果提示 `restartRequired=true`，说明需要重启进程才能完全生效

### 5) 常见部署建议

- SMTP 收信对外必须能访问 `25` 端口
- Docker Compose 默认把宿主机 `25` 映射到容器内 `2525`
- 数据库和邮件原文建议挂载到持久化目录
- 首次部署后先修改默认管理员密码和 `jwt_secret`

## 发布产物

GitHub Release 触发后会自动产出各平台压缩包并上传到对应 Release Assets。

每个压缩包包含：

- 可执行文件 `tempmail`（Windows 为 `tempmail.exe`）
- `config.yaml`（配置文件）
- `web/`（前端静态资源，可直接使用完整前端界面）

并额外上传：

- `SHA256SUMS.txt`

## 部署方式

### 1) 本地开发（前后端分离）

后端：

```bash
cd backend
go mod tidy
go run ./cmd/server -config ../config/config.yaml
```

前端：

```bash
cd frontend
npm install
npm run dev
```

- 后端 API：`http://localhost:8080/api/v1`
- 前端开发：`http://localhost:5173`

### 2) 二进制部署（推荐）

从 Release 下载对应平台压缩包，解压后：

```bash
./tempmail -config ./config.yaml
```

建议用 `systemd`/`supervisor` 托管进程。

### 3) Docker Compose（单容器）

项目根目录已提供：`docker-compose.yml`

```bash
docker compose up -d
```

默认直接拉取已编译好的发布镜像：

- `ghcr.io/clockclock1/tempmail:latest`
- 每次正式 Release 发布后，GitHub Actions 会同步更新 `latest`
- 如需固定到某个 Release 版本，可先设置环境变量：`TEMPMAIL_IMAGE=ghcr.io/clockclock1/tempmail:v1.0.10`

默认挂载：

- `./data/backend -> /app/data`
- `./config -> /app/config`

说明：

- 首次启动时会自动创建 `./config/config.yaml`
- `./data/backend` 和 `./config` 不需要提前手工建文件

访问：

- Web 控制台：`http://localhost:8080`
- API：`http://localhost:8080/api/v1`
- SMTP：`localhost:25`（容器内监听 `2525`）

### 4) Docker Run（单镜像）

```bash
docker build -f backend/Dockerfile -t tempmail:local .

docker run -d --name tempmail \
  --restart unless-stopped \
  -p 8080:8080 \
  -p 25:2525 \
  -v $(pwd)/data/backend:/app/data \
  -v $(pwd)/config:/app/config \
  tempmail:local
```

### 5) ClawCloud Run 部署

1. 进入 ClawCloud Run 控制台，创建应用
2. 选择镜像（建议使用 Release 产物镜像）：`ghcr.io/<owner>/tempmail:<tag>`
3. Container Port 填 `8080`
4. 挂载持久化存储到 `/app/data`
5. 挂载配置文件到 `/app/config/config.yaml`
6. 启动应用并绑定公网域名（HTTP/HTTPS）

说明：

- Web/API 走 `8080`
- SMTP 收信需要公网可达 SMTP 入口（通常端口 25）
- Compose 默认将宿主机 `25` 映射到容器内 `2525`
- 若 `25` 端口被占用，可设置 `TEMPMAIL_SMTP_PORT=2525` 等值覆盖宿主机端口
- 若云平台不直接开放 SMTP 25，建议增加一层 SMTP 网关/VPS 转发到本服务 `2525`

## MX 配置示例

假设收件域名为 `mail.example.com`：

1. `A` 记录：`mail.example.com -> 服务器公网 IP`
2. `MX` 记录：`mail.example.com -> mail.example.com`
3. 开放 SMTP 入站（公网 25 到宿主机 25，再映射到容器 `2525`）
4. 登录后在“域名管理”新增 `mail.example.com`
5. 创建地址如 `demo@mail.example.com` 验证收信

## API

### 新版 API（`/api/v1`）

- 认证：`POST /auth/login`、`GET /auth/me`
- 域名：`GET/POST/PUT/DELETE /domains`
- 邮箱：`GET/POST/DELETE /mailboxes`
- 邮件：`GET/DELETE /messages`、`GET /messages/:id/raw`
- 角色权限：
  - 用户：`GET /users`（支持 `q/active/roleId/page/pageSize`）、`GET /users/:id`、`POST /users`、`PATCH /users/:id`、`POST /users/:id/reset-password`、`DELETE /users/:id`
  - 角色：`GET /roles`、`GET /roles/:id/users`、`POST /roles`、`PUT /roles/:id`、`DELETE /roles/:id`
  - 权限：`GET /permissions`
- 统计：`GET /stats`
- 配置：`GET/PUT /system/config`、`POST /system/config/reload`

### 兼容旧版接口

已兼容你文档中的调用格式：

- `POST /admin/new_address`
- `POST /api/new_address`
- `GET /api/mails`
- `GET /admin/mails`
- `DELETE /admin/mails/:id`
- `DELETE /admin/delete_address/:id`
- `DELETE /admin/clear_inbox/:id`
- `DELETE /admin/clear_sent_items/:id`
- `POST /user_api/login`
- `POST /user_api/register`
- `GET /user_api/mails`

## GitHub Actions

- `go-build.yml`
  - 仅在 Release 发布时触发
  - 跨平台构建压缩包（可执行文件 + 配置文件 + 前端资源）
  - 自动上传到当前 Release Assets
- `docker-multiarch.yml`
  - 仅在 Release 发布时触发
  - 构建并推送多架构镜像

## 默认管理员

由配置文件控制：

- `default_admin_user`
- `default_admin_pass`

仅在“首次启动且管理员不存在”时用于初始化。

## 生产建议

- 修改 `jwt_secret`、`legacy_admin_auth`、默认管理员密码
- 为 Web/API 配置 HTTPS
- SMTP 入站增加限流与反滥用策略
- 定期备份 `/app/data`
- 用进程管理器保证服务自动拉起
