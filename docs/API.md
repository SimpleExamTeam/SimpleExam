# SimpleExam API 文档

**框架**: Gin  
**基础路径**: `/api/v1` (用户端), `/api/v1/admin` (管理端)  
**认证方式**: Bearer Token (JWT)  
**通用响应格式**:

```json
// 成功
{"code": 200, "data": {...}}
// 失败
{"code": 4xx/5xx, "msg": "错误描述"}
```

---

## 目录

- [1. 公共接口](#1-公共接口)
- [2. 用户认证](#2-用户认证)
- [3. 微信登录/授权](#3-微信登录授权)
- [4. 支付](#4-支付)
- [5. 用户](#5-用户)
- [6. 课程](#6-课程)
- [7. 题目](#7-题目)
- [8. 练习](#8-练习)
- [9. 考试](#9-考试)
- [10. 订单](#10-订单)
- [11. 卡券](#11-卡券)
- [12. 管理端 - 认证](#12-管理端---认证)
- [13. 管理端 - 系统管理](#13-管理端---系统管理)
- [14. 管理端 - 用户管理](#14-管理端---用户管理)
- [15. 管理端 - 订单管理](#15-管理端---订单管理)
- [16. 管理端 - 课程管理](#16-管理端---课程管理)
- [17. 管理端 - 题库管理](#17-管理端---题库管理)
- [18. 管理端 - 卡券管理](#18-管理端---卡券管理)
- [19. 管理端 - 反馈管理](#19-管理端---反馈管理)
- [20. 前端 SPA 路由](#20-前端-spa-路由)

---

## 1. 公共接口

### 1.1 健康检查

```
GET /api/v1/health
```

**响应示例**:
```json
{"code": 200, "data": {"status": "ok"}}
```

---

## 2. 用户认证

### 2.1 登录

```
POST /api/v1/auth/login
```

**请求体**:
```json
{
  "username": "string (必填)",
  "password": "string (必填)"
}
```

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": 1,
      "username": "zhangsan",
      "nickname": "张三"
    }
  }
}
```

### 2.2 注册

```
POST /api/v1/auth/register
```

**请求体**:
```json
{
  "username": "string (必填)",
  "password": "string (必填)",
  "nickname": "string (必填)"
}
```

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "id": 1,
    "username": "zhangsan",
    "nickname": "张三"
  }
}
```

### 2.3 微信小程序登录

```
POST /api/v1/auth/wx/login
```

**请求体**:
```json
{"code": "string (必填)"}
```

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {"id": 1, "username": "", "nickname": "张三", "avatar": "https://..."}
  }
}
```

---

## 3. 微信登录/授权

### 3.1 获取微信网页授权 URL

```
GET /api/v1/wechat/oauth/url?state=STATE
```

**查询参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| state | string | 否 | 回调状态，默认 `STATE` |

**响应示例**:
```json
{"code": 200, "data": {"url": "https://open.weixin.qq.com/..."}}
```

### 3.2 微信网页授权回调

```
GET /api/v1/wechat/oauth/callback?code=xxx&state=xxx
```

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {"id": 1, "username": "", "nickname": "张三", "avatar": "https://..."}
  }
}
```

### 3.3 创建登录二维码

```
POST /api/v1/wechat/qrcode/create
```

**响应示例**:
```json
{
  "code": 200,
  "data": {"scene_str": "qr_xxx", "url": "https://mp.weixin.qq.com/..."}
}
```

### 3.4 检查二维码状态

```
GET /api/v1/wechat/qrcode/check?scene_str=xxx
```

**状态轮询机制**: 客户端需轮询此接口，当 `status` 变为 `confirmed` 时返回 token。

**响应示例 (未扫码)**:
```json
{"code": 200, "data": {"status": "pending"}}
```

**响应示例 (已扫码确认)**:
```json
{
  "code": 200,
  "data": {
    "status": "confirmed",
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

### 3.5 扫码登录回调 (微信服务器回调)

```
GET /api/v1/wechat/qrcode/callback?code=xxx&state=xxx
```

**响应示例**:
```json
{
  "code": 200,
  "msg": "登录成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {"id": 1, "nickname": "张三", "avatar": "https://..."}
  }
}
```

### 3.6 管理员微信相关

通过 `/api/v1/admin/wechat/` 路径，接口签名与用户端微信接口类似:

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| POST | `/admin/wechat/login` | WXAdminLogin | 管理员微信小程序登录 |
| GET | `/admin/wechat/oauth/url` | GetWXAdminOAuthURL | 获取管理员微信网页授权 URL |
| GET | `/admin/wechat/oauth/callback` | WXAdminOAuthCallback | 管理员微信授权回调 |
| POST | `/admin/wechat/qrcode/create` | CreateAdminLoginQRCode | 创建管理员扫码登录二维码 |
| GET | `/admin/wechat/qrcode/check` | CheckAdminQRCodeStatus | 检查管理员二维码状态 |
| GET | `/admin/wechat/qrcode/callback` | AdminQRCodeCallback | 管理员扫码登录回调 |

**管理员扫码登录回调响应示例**:
```json
{
  "code": 200,
  "msg": "登录成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {"id": 1, "username": "admin", "nickname": "管理员", "avatar": "", "is_admin": true}
  }
}
```

---

## 4. 支付

### 4.1 支付回调通知 (无需认证, 微信服务器回调)

```
POST /api/v1/payments/notify
```

支持 JSON 和 XML 格式。JSON 请求示例:
```json
{"order_no": "202403011200001", "status": "success"}
```

XML 请求为微信标准支付回调格式。

**响应示例 (JSON)**:
```json
{"code": 200, "msg": "success"}
```

### 4.2 退款回调通知 (无需认证, 微信服务器回调)

```
POST /api/v1/payments/refund/notify
```

### 4.3 创建支付 (需 JWT)

```
POST /api/v1/payments/create
```

**请求头**: `Authorization: Bearer <token>`

**请求体**:
```json
{
  "course_id": "string (必填)",
  "total_fee": 0,
  "open_id": "string (必填)",
  "order_no": "string (可选，重新发起支付时填写)"
}
```

> 当 `total_fee = 0` 时视为免费课程，直接开通并返回 `{"status": "paid"}`。

**响应示例 (免费课程)**:
```json
{
  "code": 200,
  "data": {
    "orderNo": "202403011200001",
    "status": "paid",
    "message": "免费课程已开通"
  }
}
```

**响应示例 (付费课程)**:
```json
{
  "code": 200,
  "data": {
    "orderNo": "202403011200001",
    "params": { /* 微信小程序调起支付参数 */ }
  }
}
```

### 4.4 查询支付结果 (需 JWT)

```
GET /api/v1/payments/query/:order_no
```

**响应示例**:
```json
{"code": 200, "data": { /* 支付结果 */ }}
```

### 4.5 取消支付 (需 JWT)

```
POST /api/v1/payments/cancel/:order_no
```

**响应示例**:
```json
{"code": 200, "msg": "取消支付成功"}
```

### 4.6 卡券兑换 (需 JWT)

```
POST /api/v1/payments/redeem-card
```

**请求体**:
```json
{
  "card_no": "string (必填)",
  "course_id": 1
}
```

**响应示例**:
```json
{
  "code": 200,
  "msg": "兑换成功",
  "data": {
    "order_id": 1,
    "order_no": "202403011200001",
    "course_id": 1,
    "amount": 0,
    "status": "paid"
  }
}
```

### 4.7 申请退款 (管理端, 需 JWT + AdminAuth)

```
POST /api/v1/admin/orders/refund
```

**请求体**:
```json
{
  "order_no": "string (必填)",
  "refund_fee": 99.99,
  "refund_reason": "string (可选)"
}
```

**响应示例**:
```json
{"code": 200, "data": { /* 退款结果 */ }, "msg": "退款申请成功"}
```

### 4.8 查询退款状态 (管理端, 需 JWT + AdminAuth)

```
GET /api/v1/admin/orders/refund/:order_no
```

**响应示例**:
```json
{"code": 200, "data": { /* 退款状态 */ }}
```

---

## 5. 用户 (需 JWT)

### 5.1 获取个人信息

```
GET /api/v1/user/profile
```

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "id": 1,
    "username": "zhangsan",
    "nickname": "张三",
    "avatar": "https://...",
    "open_id": "oXx..."
  }
}
```

### 5.2 更新个人信息

```
PUT /api/v1/user/profile/update
```

**请求体**:
```json
{"nickname": "新昵称", "avatar": "https://...", "open_id": "oXx..."}
```

**响应示例**:
```json
{"code": 200, "msg": "更新成功"}
```

### 5.3 更新微信用户信息

```
POST /api/v1/user/wx/update-info
```

**请求体**:
```json
{"userInfo": { /* 微信用户信息 */ }}
```

### 5.4 获取 Token 过期时间

```
GET /api/v1/user/token/expire-time
```

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "expire_time": 1709876543,
    "expire_time_formatted": "2024-03-08 12:34:56"
  }
}
```

### 5.5 获取用户反馈列表

```
GET /api/v1/user/feedback?page=1&size=10
```

**查询参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| size | int | 否 | 每页条数，默认 10，最大 100 |

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "total": 5,
    "items": [
      {
        "ID": 1,
        "UserID": 1,
        "FeedbackContent": "建议增加更多练习题",
        "Status": 0,
        "ReplyContent": "",
        "CreatedAt": "2024-03-01T12:00:00+08:00"
      }
    ]
  }
}
```

### 5.6 提交反馈

```
POST /api/v1/user/feedback
```

**请求体**:
```json
{"feedback_content": "string (必填)"}
```

**响应示例**:
```json
{
  "code": 200,
  "msg": "反馈提交成功",
  "data": { /* 创建的反馈对象 */ }
}
```

---

## 6. 课程 (需 JWT)

### 6.1 获取课程分类列表

```
GET /api/v1/courses
```

**响应示例**:
```json
{"code": 200, "data": [ /* 分类树 */ ]}
```

### 6.2 获取分类详情

```
GET /api/v1/courses/category/:id
```

**响应示例**:
```json
{"code": 200, "data": { /* 分类详情 */ }}
```

### 6.3 获取课程详情

```
GET /api/v1/courses/:id
```

**响应示例**:
```json
{"code": 200, "data": { /* 课程详情 */ }}
```

### 6.4 获取课程模拟考试

```
GET /api/v1/courses/:id/exam
```

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "questions": [ /* 随机题目列表 */ ],
    "duration": 60,
    "total_score": 100,
    "pass_score": 60
  }
}
```

### 6.5 提交课程考试

```
POST /api/v1/courses/:id/exam/submit
```

**请求体**:
```json
{
  "user_id": 1,
  "course_id": 1,
  "score": 85.5,
  "wrong_answers": [3, 7, 15]
}
```

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "id": 1,
    "score": 85.5,
    "passed": true
  }
}
```

---

## 7. 题目 (需 JWT)

### 7.1 获取课程题目

```
GET /api/v1/questions/:course_id?type=single
```

**查询参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | string | 否 | 题目类型过滤: `single`, `multiple`, `judge` |

**响应示例**:
```json
{"code": 200, "data": [ /* 题目列表 */ ]}
```

---

## 8. 练习 (需 JWT)

### 8.1 获取错题统计

```
GET /api/v1/practice/wrong-questions
```

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "courses": [ /* 各课程错题统计 */ ],
    "total": 15
  }
}
```

### 8.2 获取指定课程错题

```
GET /api/v1/practice/wrong-questions/:course_id
```

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "questions": [ /* 错题列表 */ ],
    "total": 5
  }
}
```

### 8.3 清空错题

```
DELETE /api/v1/practice/wrong-questions
```

**响应示例**:
```json
{"code": 200, "msg": "错题已清空"}
```

### 8.4 提交练习

```
POST /api/v1/practice/submit
```

**请求体**:
```json
{
  "question_id": 1,
  "answer": "A"
}
```

**响应示例**:
```json
{
  "code": 200,
  "data": {"correct": true}
}
```

### 8.5 生成题目 AI 解析

```
POST /api/v1/practice/question/:id/explanation
```

**请求体**:
```json
{"force": false}
```

**响应示例**:
```json
{
  "code": 200,
  "data": {"explanation": "本题考察的是..."}
}
```

---

## 9. 考试 (需 JWT)

### 9.1 获取考试成绩

```
GET /api/v1/exams/result
```

**响应示例**:
```json
{"code": 200, "data": [ /* 考试结果列表 */ ]}
```

---

## 10. 订单 (需 JWT)

### 10.1 创建订单

```
POST /api/v1/orders
```

**请求体**:
```json
{"courseId": 1}
```

**响应示例**:
```json
{"code": 200, "data": { /* 订单对象 */ }}
```

### 10.2 获取订单列表

```
GET /api/v1/orders
```

**响应示例**:
```json
{"code": 200, "data": [ /* 订单列表 */ ]}
```

### 10.3 获取订单详情

```
GET /api/v1/orders/:id
```

**响应示例**:
```json
{"code": 200, "data": { /* 订单详情 */ }}
```

---

## 11. 卡券 (需 JWT)

### 11.1 卡券兑换

见 [4.6 卡券兑换](#46-卡券兑换-需-jwt)

---

## 12. 管理端 - 认证

### 12.1 管理员登录

```
POST /api/v1/admin/login
```

**请求体**:
```json
{
  "username": "string (必填)",
  "password": "string (必填)"
}
```

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": 1,
      "username": "admin",
      "nickname": "管理员",
      "avatar": ""
    }
  }
}
```

---

## 13. 管理端 - 系统管理 (需 JWT + AdminAuth)

### 13.1 获取登录日志

```
GET /api/v1/admin/system/login-logs?page=1&size=10&username=&status=success&start_time=&end_time=
```

**查询参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 默认 1 |
| size | int | 否 | 默认 10 |
| username | string | 否 | 用户名模糊搜索 |
| status | string | 否 | `success` 或 `fail` |
| start_time | string | 否 | 格式 `2006-01-02 15:04:05` |
| end_time | string | 否 | 格式 `2006-01-02 15:04:05` |

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "total": 100,
    "items": [
      {
        "id": 1,
        "username": "admin",
        "ip": "192.168.1.1",
        "user_agent": "Mozilla/5.0...",
        "is_success": true,
        "fail_reason": "",
        "login_time": "2024-03-01T12:00:00+08:00"
      }
    ]
  }
}
```

### 13.2 获取销售统计数据

```
GET /api/v1/admin/system/sales-statistics?dimension=day&start_time=&end_time=
```

**查询参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| dimension | string | 否 | 统计维度: `day`(默认), `month`, `year` |
| start_time | string | 否 | 格式 `2006-01-02 15:04:05`，默认当前前 30 天 |
| end_time | string | 否 | 格式 `2006-01-02 15:04:05`，默认当前时间 |

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "statistics": [ /* 销售统计数据 */ ],
    "query": {"dimension": "day", "start_time": "...", "end_time": "..."}
  }
}
```

### 13.3 获取系统信息统计

```
GET /api/v1/admin/system/system-info
```

**响应示例**:
```json
{"code": 200, "data": { /* 系统统计数据 */ }}
```

### 13.4 获取管理员个人信息

```
GET /api/v1/admin/system/profile
```

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "id": 1,
    "username": "admin",
    "nickname": "管理员",
    "avatar": "",
    "sex": 1,
    "is_admin": true,
    "created_at": "2024-01-01T00:00:00+08:00"
  }
}
```

### 13.5 更新管理员个人信息

```
PUT /api/v1/admin/system/profile
```

**请求体**:
```json
{
  "nickname": "新昵称",
  "avatar": "https://...",
  "sex": 1,
  "password": "新密码"
}
```

**响应示例**:
```json
{"code": 200, "msg": "更新成功"}
```

---

## 14. 管理端 - 用户管理 (需 JWT + AdminAuth)

### 14.1 获取用户列表

```
GET /api/v1/admin/users?page=1&size=10&keyword=&is_admin=
```

**查询参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 默认 1 |
| size | int | 否 | 默认 10 |
| keyword | string | 否 | 用户名/昵称模糊搜索 |
| is_admin | string | 否 | `true` 或 `false` |

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "total": 50,
    "items": [
      {
        "id": 1,
        "username": "zhangsan",
        "nickname": "张三",
        "avatar": "",
        "sex": 1,
        "country": "中国",
        "province": "广东",
        "city": "深圳",
        "is_admin": false,
        "open_id": "oXx...",
        "union_id": "uXx...",
        "created_at": "2024-01-01T00:00:00+08:00",
        "updated_at": "2024-03-01T00:00:00+08:00"
      }
    ]
  }
}
```

### 14.2 获取用户详情

```
GET /api/v1/admin/users/:id
```

**响应示例**:
```json
{"code": 200, "data": { /* 用户详情，字段同列表项 */ }}
```

### 14.3 创建用户

```
POST /api/v1/admin/users
```

**请求体**:
```json
{
  "username": "string (必填)",
  "password": "string (必填)",
  "nickname": "string (必填)",
  "avatar": "string (可选)",
  "sex": 1,
  "country": "中国",
  "province": "广东",
  "city": "深圳",
  "is_admin": false
}
```

**响应示例**:
```json
{"code": 200, "data": {"id": 1}}
```

### 14.4 更新用户

```
PUT /api/v1/admin/users/:id
```

**请求体**:
```json
{
  "nickname": "新昵称",
  "password": "新密码",
  "avatar": "https://...",
  "sex": 1,
  "country": "中国",
  "province": "广东",
  "city": "深圳",
  "is_admin": true
}
```

**响应示例**:
```json
{"code": 200, "msg": "更新成功"}
```

### 14.5 删除用户

```
DELETE /api/v1/admin/users/:id
```

**限制**: 用户有关联订单或考试记录时无法删除。

**响应示例**:
```json
{"code": 200, "msg": "删除成功"}
```

---

## 15. 管理端 - 订单管理 (需 JWT + AdminAuth)

### 15.1 获取订单列表

```
GET /api/v1/admin/orders?page=1&size=10&order_no=&username=&user_id=&status=&payment_type=&start_time=&end_time=
```

**查询参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 默认 1 |
| size | int | 否 | 默认 10 |
| order_no | string | 否 | 订单号模糊搜索 |
| username | string | 否 | 用户名搜索 (关联用户表) |
| user_id | uint | 否 | 用户 ID |
| status | string | 否 | 订单状态 |
| payment_type | string | 否 | 支付类型 |
| start_time | string | 否 | 格式 `2006-01-02 15:04:05` |
| end_time | string | 否 | 格式 `2006-01-02 15:04:05` |

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "total": 100,
    "items": [
      {
        "id": 1,
        "order_no": "202403011200001",
        "user": {"id": 1, "username": "zhangsan", "nickname": "张三"},
        "course_id": 1,
        "course_name": "分类-课程名",
        "amount": 99.99,
        "status": "paid",
        "payment_type": "wxpay",
        "pay_time": "2024-03-01T12:00:00+08:00",
        "expire_time": "2025-03-01T12:00:00+08:00",
        "created_at": "2024-03-01T12:00:00+08:00"
      }
    ]
  }
}
```

### 15.2 获取订单详情

```
GET /api/v1/admin/orders/:id
```

**响应示例**:
```json
{"code": 200, "data": { /* 订单详情，字段同列表项 */ }}
```

### 15.3 创建订单

```
POST /api/v1/admin/orders
```

**请求体**:
```json
{
  "user_id": 1,
  "course_id": 1,
  "amount": 99.99
}
```

**响应示例**:
```json
{"code": 200, "data": {"id": 1, "order_no": "202403011200001"}}
```

### 15.4 更新订单

```
PUT /api/v1/admin/orders/:id
```

**请求体**:
```json
{
  "status": "paid",
  "payment_type": "wxpay",
  "pay_time": "2024-03-01 12:00:00",
  "expire_time": "2025-03-01 12:00:00",
  "amount": 99.99
}
```

**响应示例**:
```json
{"code": 200, "msg": "更新成功"}
```

### 15.5 删除订单

```
DELETE /api/v1/admin/orders/:id
```

**响应示例**:
```json
{"code": 200, "msg": "删除成功"}
```

---

## 16. 管理端 - 课程管理 (需 JWT + AdminAuth)

### 16.1 获取课程列表

```
GET /api/v1/admin/courses?page=1&size=10&keyword=
```

**查询参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 默认 1 |
| size | int | 否 | 默认 10 |
| keyword | string | 否 | 课程名/分类搜索 |

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "total": 10,
    "items": [
      {
        "id": 1,
        "name": "课程名称",
        "cover": "https://...",
        "category_level1": "一级分类",
        "category_level2": "二级分类",
        "price": 99.99,
        "description": "课程描述",
        "expire_days": 365,
        "sort": 1,
        "category_sort1": 1,
        "category_sort2": 1
      }
    ]
  }
}
```

### 16.2 获取课程详情

```
GET /api/v1/admin/courses/:id
```

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "id": 1,
    "name": "课程名称",
    "cover": "https://...",
    "category_level1": "一级分类",
    "category_level2": "二级分类",
    "price": 99.99,
    "description": "课程描述",
    "expire_days": 365,
    "sort": 1,
    "category_sort1": 1,
    "category_sort2": 1,
    "exam_config": [ /* 考试配置 */ ],
    "mock_exam_config": { /* 模拟考试配置 */ }
  }
}
```

### 16.3 创建课程

```
POST /api/v1/admin/courses
```

**请求体**:
```json
{
  "name": "string (必填)",
  "cover": "string (必填)",
  "category_level1": "string (必填)",
  "category_level2": "string (必填)",
  "price": 99.99,
  "description": "string",
  "expire_days": 365,
  "sort": 1,
  "category_sort1": 1,
  "category_sort2": 1,
  "exam_config": [],
  "mock_exam_config": {}
}
```

**响应示例**:
```json
{"code": 200, "data": {"id": 1}}
```

### 16.4 更新课程

```
PUT /api/v1/admin/courses/:id
```

**请求体字段同创建课程，均为可选**。

**响应示例**:
```json
{"code": 200, "msg": "更新成功"}
```

### 16.5 删除课程

```
DELETE /api/v1/admin/courses/:id
```

**限制**: 有关联订单时无法删除。

**响应示例**:
```json
{"code": 200, "msg": "删除成功"}
```

---

## 17. 管理端 - 题库管理 (需 JWT + AdminAuth)

### 17.1 获取题目列表

```
GET /api/v1/admin/questions?page=1&size=10&type=single&question=&course_id=
```

**查询参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 默认 1 |
| size | int | 否 | 默认 10 |
| type | string | 否 | `single`, `multiple`, `judge` |
| question | string | 否 | 题目内容模糊搜索 |
| course_id | uint | 否 | 课程 ID |

**题目类型说明**:

| type 值 | 类型说明 | 答案格式 |
|---------|---------|---------|
| single | 单选题 | 单个大写字母，如 `A` |
| multiple | 多选题 | 多个大写字母组合，如 `ABC` |
| judge | 判断题 | `A`(正确) 或 `B`(错误) |

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "total": 100,
    "items": [
      {
        "id": 1,
        "type": "single",
        "type_desc": "单选题",
        "question": "题目内容",
        "options": [
          {"label": "A", "text": "选项A"},
          {"label": "B", "text": "选项B"},
          {"label": "C", "text": "选项C"},
          {"label": "D", "text": "选项D"}
        ],
        "answer": "A",
        "explanation": "解析内容",
        "course_id": 1,
        "course_name": "课程名",
        "created_at": "2024-03-01T12:00:00+08:00"
      }
    ]
  }
}
```

### 17.2 获取题目详情

```
GET /api/v1/admin/questions/:id
```

**响应示例**: 同上列表项结构。

### 17.3 创建题目

```
POST /api/v1/admin/questions
```

**请求体**:
```json
{
  "type": "single (必填, single/multiple/judge)",
  "question": "题目内容 (必填)",
  "options": [
    {"label": "A", "text": "选项A"},
    {"label": "B", "text": "选项B"},
    {"label": "C", "text": "选项C"},
    {"label": "D", "text": "选项D"}
  ],
  "answer": "A (必填)",
  "explanation": "解析 (可选)",
  "course_id": 1
}
```

**响应示例**:
```json
{"code": 200, "data": {"id": 1}}
```

### 17.4 更新题目

```
PUT /api/v1/admin/questions/:id
```

**请求体**: 同创建题目。

**响应示例**:
```json
{"code": 200, "msg": "更新成功"}
```

### 17.5 删除题目

```
DELETE /api/v1/admin/questions/:id
```

**响应示例**:
```json
{"code": 200, "msg": "删除成功"}
```

### 17.6 批量删除题目

```
POST /api/v1/admin/questions/batch-delete
```

**请求体**:
```json
{"ids": [1, 2, 3]}
```

**响应示例**:
```json
{
  "code": 200,
  "msg": "删除成功",
  "data": {"deleted_count": 3}
}
```

### 17.7 一键清空指定课程的全部题目

```
DELETE /api/v1/admin/questions/clear-by-course/:course_id
```

**路径参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| course_id | uint | 课程 ID |

**说明**: 删除指定课程下的所有题目，先校验课程存在性，不存在返回 404。

**响应示例**:
```json
{
  "code": 200,
  "msg": "清空成功",
  "data": {"deleted_count": 30}
}
```

### 17.8 导出题库

```
GET /api/v1/admin/questions/export?course_id=1
```

导出为 CSV 文件下载。表头: `ID, 题目类型, 题目内容, 选项, 答案, 解析, 课程ID, 题目类型说明`

### 17.8 导入题库

```
POST /api/v1/admin/questions/import
```

**请求体**: `multipart/form-data`，字段名 `file`，上传 CSV 文件。

CSV 格式: `ID, 题目类型(single/multiple/judge), 题目内容, 选项(JSON数组字符串), 答案, 解析, 课程ID`

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "import_count": 95,
    "error_count": 5
  },
  "msg": "导入完成",
  "errors": ["第3行: 课程ID格式错误", "..."]
}
```

---

## 18. 管理端 - 卡券管理 (需 JWT + AdminAuth)

### 18.1 获取卡券列表

```
GET /api/v1/admin/cards?page=1&size=10&card_no=&course_id=
```

**查询参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 默认 1 |
| size | int | 否 | 默认 10 |
| card_no | string | 否 | 卡号模糊搜索 |
| course_id | uint | 否 | 课程 ID |

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "total": 20,
    "items": [
      {
        "id": 1,
        "card_no": "L3xVn9Kp2QrA5sD8Fg",
        "course_id": null,
        "course_name": "全部课程",
        "amount": 0,
        "total": 100,
        "used": 35,
        "expire_days": 365,
        "expire_time": "2025-03-01T12:00:00+08:00",
        "is_expired": false,
        "created_at": "2024-03-01T12:00:00+08:00"
      }
    ]
  }
}
```

### 18.2 获取卡券详情

```
GET /api/v1/admin/cards/:id
```

包含该卡券的兑换记录列表。

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "id": 1,
    "card_no": "L3xVn9Kp2QrA5sD8Fg",
    "course_id": null,
    "course_name": "全部课程",
    "amount": 0,
    "total": 100,
    "used": 35,
    "expire_days": 365,
    "expire_time": "2025-03-01T12:00:00+08:00",
    "is_expired": false,
    "created_at": "2024-03-01T12:00:00+08:00",
    "records": [
      {
        "id": 1,
        "user_id": 1,
        "username": "zhangsan",
        "nickname": "张三",
        "order_id": 1,
        "order_no": "202403011200001",
        "course_id": 1,
        "amount": 0,
        "created_at": "2024-03-01T12:00:00+08:00"
      }
    ]
  }
}
```

### 18.3 创建卡券

```
POST /api/v1/admin/cards
```

**请求体**:
```json
{
  "course_id": null,
  "amount": 0,
  "total": 100,
  "expire_days": 365
}
```

> `course_id` 为 `null` 表示可兑换任意课程，指定则仅可兑换该课程。

**响应示例**:
```json
{"code": 200, "data": {"id": 1, "card_no": "L3xVn9Kp2QrA5sD8Fg"}}
```

### 18.4 更新卡券

```
PUT /api/v1/admin/cards/:id
```

**请求体**:
```json
{
  "course_id": null,
  "amount": 0,
  "total": 200,
  "expire_days": 365
}
```

**响应示例**:
```json
{"code": 200, "msg": "更新成功"}
```

### 18.5 删除卡券

```
DELETE /api/v1/admin/cards/:id
```

**响应示例**:
```json
{"code": 200, "msg": "删除成功"}
```

### 18.6 获取所有兑换记录

```
GET /api/v1/admin/cards/records?page=1&size=10&card_no=&username=&course_id=
```

**查询参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 默认 1 |
| size | int | 否 | 默认 10 |
| card_no | string | 否 | 卡号搜索 |
| username | string | 否 | 用户名搜索 |
| course_id | uint | 否 | 课程 ID |

### 18.7 获取指定卡券兑换记录

```
GET /api/v1/admin/cards/:id/records
```

---

## 19. 管理端 - 反馈管理 (需 JWT + AdminAuth)

### 19.1 获取所有反馈

```
GET /api/v1/admin/users/feedback?page=1&size=10&username=&status=&start_time=&end_time=
```

**查询参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 默认 1 |
| size | int | 否 | 默认 10 |
| username | string | 否 | 用户名搜索 |
| status | int | 否 | 状态码过滤 |
| start_time | string | 否 | 开始时间 |
| end_time | string | 否 | 结束时间 |

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "total": 10,
    "items": [
      {
        "id": 1,
        "user": {"id": 1, "username": "zhangsan", "nickname": "张三"},
        "feedback_content": "建议增加更多练习题",
        "status": 0,
        "reply_content": "",
        "created_at": "2024-03-01T12:00:00+08:00",
        "updated_at": "2024-03-01T12:00:00+08:00"
      }
    ]
  }
}
```

### 19.2 获取反馈详情

```
GET /api/v1/admin/users/feedback/:id
```

**响应示例**: 同上列表项结构。

### 19.3 更新反馈 (回复/更改状态)

```
PUT /api/v1/admin/users/feedback/:id
```

**请求体**:
```json
{
  "status": 1,
  "reply_content": "感谢您的建议，我们已添加更多练习题"
}
```

**响应示例**:
```json
{"code": 200, "msg": "更新成功"}
```

### 19.4 删除反馈

```
DELETE /api/v1/admin/users/feedback/:id
```

**响应示例**:
```json
{"code": 200, "msg": "删除成功"}
```

---

## 20. 前端 SPA 路由

| 路径 | 说明 |
|------|------|
| `GET /admin` | 返回管理端 `index.html` |
| `GET /admin/*path` | 尝试匹配管理端静态文件，不存在则回退 `index.html` |
| `GET /*` (NoRoute) | 非 `/api/` `/static/` 的请求，由用户端 `index.html` 处理 |

---

## 中间件栈

| 中间件 | 所属分组 | 说明 |
|--------|---------|------|
| Logger | 所有 API 分组 | 请求日志记录 |
| Recovery | 所有 API 分组 | Panic 恢复，避免进程崩溃 |
| Cors | 所有 API 分组 | 跨域资源共享 |
| JWT | 需认证的路由 | 从 `Authorization: Bearer <token>` 提取 JWT，解析 HS256 签名，校验用户存在且未被软删除，将 `userId` 注入 Context |
| AdminAuth | 管理端需认证路由 | 从 Context 读取 `userId`，查询 `user.IsAdmin` 是否为 `true`，否则返回 403 |
