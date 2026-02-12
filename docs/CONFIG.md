# Simple Exam 配置文档

本文档详细说明 Simple Exam 项目的配置选项。

## 配置文件位置

配置文件位于 `config/config.yaml`，可以参考 `config/config.example.yaml` 创建。

## 配置项说明

### 服务器配置 (server)

```yaml
server:
  port: "8080"        # 服务监听端口
  mode: "release"     # 运行模式: debug 或 release
```

- `port`: HTTP 服务监听的端口号
- `mode`: 
  - `debug`: 开发模式，输出详细日志
  - `release`: 生产模式，优化性能

### 数据库配置 (database)

```yaml
database:
  driver: "mysql"           # 数据库驱动
  host: "127.0.0.1"        # 数据库主机地址
  port: "3306"             # 数据库端口
  username: "admin"        # 数据库用户名
  password: "123456"       # 数据库密码
  dbname: "simple_exam"    # 数据库名称
```

支持的数据库驱动：
- `mysql`: MySQL/MariaDB

### JWT 配置 (jwt)

```yaml
jwt:
  secret: "your-secret-key"    # JWT 签名密钥
  expire_time: 86400           # Token 过期时间（秒）
```

- `secret`: 用于签名 JWT Token 的密钥，建议使用强随机字符串
- `expire_time`: Token 有效期，单位为秒（86400 = 24小时）

### 日志配置 (log)

```yaml
log:
  level: "debug"                  # 日志级别
  format: "text"                  # 日志格式
  output: "both"                  # 输出方式
  file_path: "logs/app.log"       # 日志文件路径
  max_size: 100                   # 单个日志文件最大大小(MB)
  max_backups: 3                  # 保留的旧日志文件数量
  max_age: 28                     # 日志文件保留天数
  compress: true                  # 是否压缩旧日志文件
```

#### 日志级别 (level)
- `debug`: 调试信息（最详细）
- `info`: 一般信息
- `warn`: 警告信息
- `error`: 错误信息（最简洁）

#### 日志格式 (format)
- `json`: JSON 格式，便于日志分析工具处理
- `text`: 文本格式，便于人工阅读

#### 输出方式 (output)
- `console`: 仅输出到控制台
- `file`: 仅输出到文件
- `both`: 同时输出到控制台和文件

#### 日志轮转
- `max_size`: 单个日志文件达到此大小后自动轮转
- `max_backups`: 保留的历史日志文件数量
- `max_age`: 日志文件保留天数，超过后自动删除
- `compress`: 是否使用 gzip 压缩历史日志文件

### 微信配置 (wechat)

```yaml
wechat:
  app_id: "wx1111111111"              # 微信公众号 AppID
  app_secret: "ecf1111111111"         # 微信公众号 AppSecret
  mch_id: "1600000000"                # 微信商户号
  pay_key: "dfadsadsa"                # 商户支付密钥
  
  # 基础域名配置
  base_url: "https://example.com"
  
  # 回调地址配置（相对路径）
  notify_url: "/api/v1/payments/notify"
  oauth_redirect: "/wechat/callback"
  admin_oauth_redirect: "/admin/wechat-callback"
  qrcode_callback: "/wechat/qrcode/callback"
  admin_qrcode_callback: "/admin/wechat-qrcode-callback"
  refund_notify_url: "/api/v1/payments/refund/notify"
  
  # 微信API地址（无需修改）
  refund_url: "https://api.mch.weixin.qq.com/secapi/pay/refund"
  refund_query_url: "https://api.mch.weixin.qq.com/pay/refundquery"
  
  # 商户证书路径
  cert_path: "certs/apiclient_cert.p12"
```

#### 基础配置
- `app_id`: 微信公众号的 AppID，在微信公众平台获取
- `app_secret`: 微信公众号的 AppSecret，在微信公众平台获取
- `mch_id`: 微信支付商户号，在微信商户平台获取
- `pay_key`: 微信支付密钥（API密钥），在微信商户平台设置

#### 域名和回调配置
- `base_url`: 应用的基础域名，所有回调地址会自动拼接此域名
  - 示例：`https://example.com` 或 `https://exam.yourdomain.com`
  - 必须使用 HTTPS（微信要求）
  - 不要在末尾添加斜杠

- `notify_url`: 支付成功后的回调地址
- `oauth_redirect`: 用户微信授权登录回调地址
- `admin_oauth_redirect`: 管理员微信授权登录回调地址
- `qrcode_callback`: 用户扫码登录回调地址
- `admin_qrcode_callback`: 管理员扫码登录回调地址
- `refund_notify_url`: 退款结果通知回调地址

#### 证书配置
- `cert_path`: 微信商户证书文件路径
  - 用于退款等需要证书的接口
  - 证书文件从微信商户平台下载
  - 支持 `.p12` 格式

#### 获取微信配置信息

1. **公众号配置** (app_id, app_secret)
   - 登录 [微信公众平台](https://mp.weixin.qq.com/)
   - 进入"开发" -> "基本配置"
   - 获取 AppID 和 AppSecret

2. **商户配置** (mch_id, pay_key)
   - 登录 [微信商户平台](https://pay.weixin.qq.com/)
   - 进入"账户中心" -> "商户信息"获取商户号
   - 进入"账户中心" -> "API安全"设置 API 密钥

3. **商户证书** (cert_path)
   - 在微信商户平台"账户中心" -> "API安全"
   - 下载商户证书
   - 将证书文件放置到 `certs/` 目录

## 配置示例

### 开发环境配置

```yaml
server:
  port: "8080"
  mode: "debug"

database:
  driver: "mysql"
  host: "127.0.0.1"
  port: "3306"
  username: "root"
  password: "password"
  dbname: "simple_exam_dev"

jwt:
  secret: "dev-secret-key-change-in-production"
  expire_time: 86400

log:
  level: "debug"
  format: "text"
  output: "both"
  file_path: "logs/app.log"
  max_size: 100
  max_backups: 3
  max_age: 7
  compress: false

wechat:
  app_id: "your_test_app_id"
  app_secret: "your_test_app_secret"
  mch_id: "your_test_mch_id"
  pay_key: "your_test_pay_key"
  base_url: "https://dev.example.com"
  notify_url: "/api/v1/payments/notify"
  oauth_redirect: "/wechat/callback"
  admin_oauth_redirect: "/admin/wechat-callback"
  qrcode_callback: "/wechat/qrcode/callback"
  admin_qrcode_callback: "/admin/wechat-qrcode-callback"
  refund_notify_url: "/api/v1/payments/refund/notify"
  refund_url: "https://api.mch.weixin.qq.com/secapi/pay/refund"
  refund_query_url: "https://api.mch.weixin.qq.com/pay/refundquery"
  cert_path: "certs/apiclient_cert.p12"
```

### 生产环境配置

```yaml
server:
  port: "8080"
  mode: "release"

database:
  driver: "mysql"
  host: "db.example.com"
  port: "3306"
  username: "exam_user"
  password: "strong_password_here"
  dbname: "simple_exam"

jwt:
  secret: "very-strong-random-secret-key-here"
  expire_time: 86400

log:
  level: "info"
  format: "json"
  output: "file"
  file_path: "logs/app.log"
  max_size: 100
  max_backups: 10
  max_age: 30
  compress: true

wechat:
  app_id: "your_production_app_id"
  app_secret: "your_production_app_secret"
  mch_id: "your_production_mch_id"
  pay_key: "your_production_pay_key"
  base_url: "https://exam.yourdomain.com"
  notify_url: "/api/v1/payments/notify"
  oauth_redirect: "/wechat/callback"
  admin_oauth_redirect: "/admin/wechat-callback"
  qrcode_callback: "/wechat/qrcode/callback"
  admin_qrcode_callback: "/admin/wechat-qrcode-callback"
  refund_notify_url: "/api/v1/payments/refund/notify"
  refund_url: "https://api.mch.weixin.qq.com/secapi/pay/refund"
  refund_query_url: "https://api.mch.weixin.qq.com/pay/refundquery"
  cert_path: "certs/apiclient_cert.p12"
```

## 安全建议

1. **JWT Secret**: 使用强随机字符串，至少 32 字符
2. **数据库密码**: 使用复杂密码，定期更换
3. **微信密钥**: 妥善保管，不要提交到版本控制系统
4. **配置文件**: 
   - 将 `config/config.yaml` 添加到 `.gitignore`
   - 仅提交 `config/config.example.yaml` 作为模板
5. **生产环境**: 
   - 使用 `release` 模式
   - 日志级别设置为 `info` 或 `warn`
   - 启用日志压缩和轮转

## 环境变量

也可以通过环境变量覆盖配置文件中的设置：

```bash
export SERVER_PORT=8080
export DB_HOST=127.0.0.1
export DB_PASSWORD=your_password
export JWT_SECRET=your_secret
```

## 配置验证

启动应用时会自动验证配置：

```bash
./simpleexam -c config/config.yaml
```

如果配置有误，应用会输出错误信息并退出。

## 常见问题

### 1. 数据库连接失败
- 检查数据库服务是否运行
- 验证主机地址、端口、用户名和密码
- 确认数据库已创建

### 2. 微信支付回调失败
- 确认 `base_url` 使用 HTTPS
- 检查域名是否在微信公众平台配置
- 验证回调地址是否可公网访问

### 3. JWT Token 无效
- 检查 `jwt.secret` 是否正确
- 确认 Token 未过期
- 验证客户端和服务端时间同步

### 4. 日志文件过大
- 调整 `log.max_size` 减小单文件大小
- 减少 `log.max_age` 缩短保留时间
- 启用 `log.compress` 压缩历史日志
