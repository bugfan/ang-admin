# CAS 测试环境 (Mock CAS Server 2.0/3.0)

本目录为独立完整的 **CAS 2.0/3.0/1.0** 统一身份认证测试服务器。采用 Go 编写，零依赖，秒级启动。

---

## 1. 启停服务

### 方式 A：本地直接运行（推荐，极简秒启）
```bash
# 启动服务
./start.sh
# 按 Ctrl+C 即可退出停止
```

### 方式 B：Docker 运行
```bash
# 构建并启动容器
docker compose up -d

# 停止并清理容器
docker compose down
```

服务默认监听在：`http://127.0.0.1:8082/cas`

---

## 2. 预设测试账号

| 用户名 / 学号 | 密码 | 预设邮箱 | 角色说明 |
|--------------|------|---------|---------|
| `testuser` | `testpass` | testuser@ang.local | 普通用户 |
| `admin` | `admin123` | admin@ang.local | 管理员账号 |
| `alice` | `alice@2024` | alice@ang.local | 员工账号 |
| `casuser` | `cas123` | casuser@ang.local | 专属 CAS 测试用户 |

---

## 3. CAS 客户端对接参数

- **Base URL**: `http://127.0.0.1:8082/cas`
- **协议版本**: `2` (或 `3`)
- **登录页路径 (login_path)**: `/login`
- **票据验证路径 (validate_path)**: `/serviceValidate` (或 `/p3/serviceValidate`)
- **跳过 TLS 校验**: `false`

管理端配置 JSON 示例：
```json
{
  "base_url": "http://127.0.0.1:8082/cas",
  "version": 2,
  "login_path": "/login",
  "validate_path": "/serviceValidate",
  "skip_tls_verify": false
}
```

---

## 4. 端点与协议说明

- `GET /cas/login?service={service}`: 渲染美观的统一身份认证登录界面。
- `POST /cas/login`: 校验用户名密码，验证通过生成一次性 Service Ticket (`ST-xxxx`)，并 302 重定向到 `{service}?ticket={ST-xxxx}`。
- `GET /cas/serviceValidate?ticket={ticket}&service={service}`: 标准 CAS 2.0/3.0 XML 验票接口，验证成功返回 `<cas:authenticationSuccess>` 及 `<cas:user>`，一次性消费票据。
- `GET /cas/p3/serviceValidate`: CAS 3.0 别名端点。
- `GET /cas/validate`: CAS 1.0 纯文本验票端点 (`yes\n{user}\n`)。
- `GET /cas/ping`: 探活接口，返回 200 OK。
- `GET /cas/logout`: 会话注销接口。
