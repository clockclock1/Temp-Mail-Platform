# Temp Mail Platform

一个可自托管的临时邮箱系统：

- 固定根域邮箱 + 泛子域收信
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
db_path: "./data/tempmail.db"
data_dir: "./data/messages"
cors_origins:
  - "http://localhost:8080"
  - "http://localhost:5173"
dns_resolvers:
  - "1.1.1.1:53"
  - "8.8.8.8:53"
  - "9.9.9.9:53"
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
| `jwt_secret` | 登录 Token 密钥 | 生产环境务必修改 |
| `jwt_expire_hours` | 登录有效期 | 默认 `24` |
| `db_path` | SQLite 数据库文件 | `./data/tempmail.db` |
| `data_dir` | 邮件原文 `eml` 存放目录 | `./data/messages` |
| `cors_origins` | 允许的前端来源 | 按实际域名填写 |
| `dns_resolvers` | MX 校验失败时使用的回退 DNS 服务器列表 | 建议保留 `1.1.1.1:53`、`8.8.8.8:53` |
| `default_admin_user` | 首次初始化管理员用户名 | 生产环境请改掉 |
| `default_admin_pass` | 首次初始化管理员密码 | 生产环境请改掉 |
| `cleanup_interval_minutes` | 过期邮箱清理周期 | 默认 `10` |

### 4) 改完会发生什么

- 修改 `cors_origins`、`jwt_secret`、`cleanup_interval_minutes`、`data_dir` 后可热生效
- 修改 `http_addr`、`smtp_addr`、`db_path`、`web_dir` 这类路径或监听项，通常需要重启服务
- 前端配置中心会自动写回同一份 `config.yaml`
- 配置保存后如果提示 `restartRequired=true`，说明需要重启进程才能完全生效

### 5) 常见部署建议

- SMTP 收信对外必须能访问 `25` 端口
- Docker Compose 默认把宿主机 `25` 映射到容器内 `2525`
- 根域名新增时会先验证 MX；若容器内出现 `127.0.0.11` 解析失败，可通过 `dns_resolvers` 使用公共 DNS 回退
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
- 如需固定到某个 Release 版本，可先设置环境变量：`TEMPMAIL_IMAGE=ghcr.io/clockclock1/tempmail:v1.0.12`

默认挂载：

- `./data/backend -> /app/data`
- `./config -> /app/config`

说明：

- 首次启动时会自动创建 `./config/config.yaml`
- `./data/backend` 和 `./config` 不需要提前手工建文件
- MX 校验会先走容器默认 DNS，再自动回退到 `dns_resolvers` 中配置的解析器

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

## 固定根域 + 泛子域收信

当前版本的实际收信模式是：

1. 在系统里只维护固定根域，例如 `example.com`
2. 创建邮箱后，主地址固定为 `alice@example.com`
3. 如果 DNS 已配置通配收信，系统也会接受：
   `alice@foo.example.com`
4. 同样也接受更深层的子域：
   `alice@bar.foo.example.com`

也就是说，系统内的“邮箱归属域”固定是根域；泛子域能力由 DNS 和 SMTP 投递规则负责，后端只做统一路由匹配。

### 推荐 DNS 记录

假设你的固定根域是 `example.com`，收信主机是 `mail.example.com`：

1. `mail A 服务器公网 IP`
2. `@ MX 10 mail.example.com`
3. `* MX 10 mail.example.com`
4. `* A 服务器公网 IP`
5. 开放 SMTP 入站（公网 `25` 到宿主机 `25`，再映射到容器 `2525`）

推荐流程：

1. 先确认 `mail.example.com` 能解析到公网 IP
2. 再确认根域 `example.com` 的 MX 已生效
3. 然后在系统“域名管理”里添加 `example.com`
4. 创建邮箱 `alice@example.com`
5. 测试主地址 `alice@example.com`
6. 再测试泛子域地址 `alice@test.example.com`

如果新增根域时看到类似：

```text
lookup mx for example.com: lookup example.com on 127.0.0.11:53: no such host
```

通常不是系统功能缺失，而是 Docker 当前 DNS 无法解析该域名。现在服务会自动继续尝试 `dns_resolvers` 中的公共 DNS；如果仍失败，再检查：

1. 根域本身是否已生效
2. `@ MX` 是否已正确指向你的收信主机
3. 如果要收任意子域邮箱，`* MX` 是否也已经生效
4. `mail.example.com` 是否有可用的 `A/AAAA` 记录
5. 宿主机公网 `25` 端口是否已放通并映射到容器 `2525`

### 为什么必须做 `* MX`

外部邮件服务器投递到 `user.foo.example.com` 时，会先按收件域本身查 MX，不会自动继承父域 `example.com` 的 MX。  
因此，如果你想让 `user@任意子域.example.com` 都能真正送到这台服务器，最稳妥的做法是显式配置：

- `@ MX 10 mail.example.com`
- `* MX 10 mail.example.com`

`* A` 记录不是 MX 的替代品，但建议同时加上，便于子域解析和部分隐式投递场景。

## API

### API（`/api/v1`）

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

- 修改 `jwt_secret`、默认管理员密码
- 为 Web/API 配置 HTTPS
- SMTP 入站增加限流与反滥用策略
- 定期备份 `/app/data`
- 用进程管理器保证服务自动拉起
