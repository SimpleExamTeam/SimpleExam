# 考试系统后端

基于 Gin 框架开发的在线考试系统后端

## 技术栈

- Gin: Web 框架
- GORM: ORM 框架
- JWT: 用户认证
- MySQL: 数据存储
- Redis: 缓存
- 微信SDK: 微信登录和支付

## 项目结构

```
.
├── cmd                     # 程序入口
├── config                  # 配置文件
├── internal               
│   ├── api                # API 处理器
│   ├── middleware         # 中间件
│   ├── model             # 数据模型
│   ├── repository        # 数据访问层
│   ├── service          # 业务逻辑层
│   └── pkg              # 内部公共包
├── pkg                    # 外部可用的公共包
└── scripts                # 脚本文件
```

## 快速开始

1. 安装依赖
```bash
go mod download
```

2. 配置环境变量
```bash
cp .env.example .env
```

3. 运行项目
```bash
go run cmd/main.go
```

## 数据库设计


## API 文档

