# Portable Chat

Portable Chat 是一个类似 QQ 的轻量 IM Web 项目，后端基于 Go + Gin，前端为内嵌静态页面，开箱即用。

当前版本已经支持：

- 账号注册
- 邮箱验证码发送
- 登录 / 登出
- 联系人列表查询与搜索
- 点击联系人进入一对一聊天房间
- WebSocket 实时收发消息
- 默认展示最近 30 天历史消息
- 已读状态与已读回执
- SSE 联系人列表通知
- 新消息声音提醒
- 移动端适配页面

## 技术栈

- Go 1.25
- Gin
- GORM
- SQLite
- Gorilla WebSocket
- SSE

## 项目结构

```text
cmd/api/main.go             # 服务启动入口
internal/router/           # 路由注册
internal/handler/          # HTTP / WebSocket / SSE 处理层
internal/service/          # 业务逻辑
internal/repository/       # 数据访问层
pkg/sqlitecli/             # SQLite 初始化与迁移
pkg/storecli/              # 轻量内存 KV（验证码、登录 token）
static/index.html          # 内嵌前端页面
static/code.html           # 邮箱验证码模板
portable-chat.db           # SQLite 数据库文件（运行时自动创建）
```

## 功能说明

### 1. 注册与验证码

- 用户先输入邮箱，请求验证码
- 验证码默认 10 分钟有效
- 同一邮箱 60 秒内限制重复发送
- 如果本地没有配置 SMTP，接口会直接返回 `debugCode`，方便调试

### 2. 登录与会话

- 登录成功后返回 token
- token 默认有效期 24 小时
- 前端会自动保存 token
- token 失效时，页面会自动清空会话并跳回登录页

### 3. 联系人列表

- 联系人列表基于已注册用户生成
- 支持关键字搜索
- 自动显示：
  - 联系人名称
  - 最近一条消息
  - 最近消息时间
  - 未读数量

### 4. 聊天

- 点击联系人后进入对应 WebSocket 聊天房间
- 发送消息后对方实时收到
- 历史消息默认只加载最近 30 天
- 当前联系人窗口打开时，消息会自动标记已读

### 5. SSE 通知

除了当前聊天窗口的 WebSocket 外，系统还会为登录用户建立一条 SSE 通道，用于：

- 联系人列表自动刷新
- 新消息到达时刷新未读数
- 已读回执后同步刷新状态

这样即使没有手动点“刷新”，联系人列表也会自动变化。

### 6. 声音提醒

- 收到非当前会话的新消息时，前端会播放提示音
- 浏览器通常要求用户先与页面有一次交互（点击 / 键盘），声音才能正常播放

## 本地运行

### 方式一：直接运行

确保本机安装 Go 1.25 或更高版本。

```bash
go mod download
go run ./cmd/api
```

启动后访问：

```text
http://localhost:9191/
```

## 构建

```bash
go build -o portable-chat ./cmd/api/main.go
./portable-chat
```

## Docker 运行

项目包含 `Dockerfile`，可以直接构建镜像：

```bash
docker build -t portable-chat .
docker run --rm -p 9191:9191 portable-chat
```

启动后访问：

```text
http://localhost:9191/
```

## 邮箱配置

如果你希望真实发送验证码邮件，需要配置以下环境变量：

```bash
SEND_EMAIL=your_email@example.com
SEND_PWD=your_smtp_password
SMTP_SERVER=smtp.example.com
SMTP_PORT=587
```

例如 QQ 邮箱：

```bash
SEND_EMAIL=example@qq.com
SEND_PWD=your_smtp_auth_code
SMTP_SERVER=smtp.qq.com
SMTP_PORT=587
```

如果不配置这些变量，系统会进入本地调试模式：

- 不真正发邮件
- `/portable/sendCode` 响应里返回 `debugCode`

## 数据存储说明

### SQLite

- 数据库文件默认位于项目根目录：`portable-chat.db`
- 服务启动时会自动建表和迁移

### 内存 KV

项目内部使用轻量内存 KV 保存：

- 邮箱验证码
- 验证码限流
- 登录 token

注意：这部分数据重启服务后会丢失，适合当前演示与本地开发场景。

## 页面使用说明

### 注册

1. 打开首页
2. 切换到“注册”
3. 输入用户名、密码、邮箱
4. 点击“发送验证码”
5. 输入验证码并提交

如果本地未配置 SMTP，可以直接使用接口返回的 `debugCode` 注册。

### 登录

1. 切换到“登录”
2. 输入用户名和密码
3. 登录成功后进入聊天页

### 聊天

1. 从左侧联系人列表中选择一个联系人
2. 页面会自动拉取最近 30 天聊天记录
3. 输入消息并发送
4. 对方在线时会实时收到消息
5. 对方查看消息后，发送侧会显示“已读”

### 通知

- 联系人列表新消息、未读数变化会自动刷新
- 非当前聊天窗口收到新消息时会弹出提示并播放声音

## 主要路由

### 公共接口

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/` | 前端页面 |
| GET | `/portable/sendCode` | 发送邮箱验证码 |
| POST | `/portable/register` | 注册 |
| POST | `/portable/login` | 登录 |
| GET | `/portable/exist` | 检查用户名是否存在 |

### 需要登录的接口

请求头：

```text
Authorization: Bearer <token>
```

或在 WebSocket / SSE 场景使用：

```text
access_token=<token>
```

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/portable/root/info` | 当前用户信息 |
| POST | `/portable/root/logout` | 登出 |
| GET | `/portable/root/contacts` | 联系人列表 |
| GET | `/portable/root/history` | 聊天历史 |
| GET | `/portable/root/events` | SSE 事件流 |
| GET | `/portable/root/ws` | WebSocket 聊天连接 |

## WebSocket 说明

连接地址：

```text
ws://localhost:9191/portable/root/ws?contactId=<联系人ID>&access_token=<token>
```

支持消息类型：

- `message`：发送消息
- `read`：标记已读
- `ping`：心跳

服务端可能返回：

- `connected`
- `message`
- `read`
- `pong`
- `error`

## SSE 说明

连接地址：

```text
GET /portable/root/events?access_token=<token>
```

可能收到的事件：

- `connected`
- `ping`
- `incoming_message`
- `contacts_refresh`
- `read_receipt`

## 测试与校验

编译检查：

```bash
go build ./...
```

项目包含集成测试，可执行：

```bash
go test ./...
```

## 当前实现边界

这个项目目前更偏向演示 / 原型版本，已能完整跑通 IM 主流程，但还有一些适合后续增强的点：

- token 与验证码目前是内存存储，重启会失效
- 目前是一对一聊天，不包含群聊
- 暂无文件上传、图片消息、表情消息
- 暂无消息撤回、删除、置顶等高级能力
- 暂无真正的离线推送能力

## 建议的后续增强

- 接入 Redis 存储 token、验证码与会话状态
- 增加用户头像上传
- 增加群聊与群成员管理
- 支持图片、文件、语音消息
- 增加消息分页加载与更长历史查询
- 增加消息通知权限与桌面通知

## 启动后建议的体验路径

1. 启动服务并打开 `http://localhost:9191/`
2. 注册两个账号
3. 分别登录两个浏览器窗口
4. 打开同一个联系人开始聊天
5. 观察：
   - WebSocket 实时消息
   - 联系人列表自动刷新
   - 新消息提示音
   - 已读状态变化
   - token 失效后自动回登录页
