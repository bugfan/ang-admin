# RADIUS 测试环境 (FreeRADIUS)

本目录为独立完整的 **FreeRADIUS** 模拟测试环境。

---

## 1. 启停服务

```bash
# 启动（后台运行）
docker compose up -d

# 启动（控制台前台调试模式，可实时看 RADIUS 报文）
docker compose up

# 停止并清理容器
docker compose down
```

---

## 2. 预设测试账号

定义在 [`config/users`](file:///Users/sean/Desktop/go-proj/src/github.com/bugfan/ang-admin/testenv/radius/config/users)：

| 用户名 | 密码 | 角色说明 |
|--------|------|---------|
| `testuser` | `testpass` | 普通测试用户 |
| `admin` | `admin123` | 管理员账号 |
| `alice` | `alice@2024` | 员工账号 |
| `bob` | `bob@2024` | 员工账号 |

---

## 3. RADIUS 客户端对接参数

- **服务器地址 (Host)**: `127.0.0.1`
- **认证端口 (Port)**: `1812` (UDP)
- **共享密钥 (Secret)**: `testing123`
- **认证协议**: PAP
- **NAS 标识符**: `ang-admin`

管理端配置 JSON 示例：
```json
{
  "host": "127.0.0.1",
  "port": 1812,
  "secret": "testing123",
  "timeout": 5,
  "retries": 3,
  "nas_identifier": "ang-admin"
}
```

---

## 4. 配置文件说明

- [`config/clients.conf`](file:///Users/sean/Desktop/go-proj/src/github.com/bugfan/ang-admin/testenv/radius/config/clients.conf): 放行来自所有来源的客户端请求，共享密钥配置为 `testing123`。
- [`config/users`](file:///Users/sean/Desktop/go-proj/src/github.com/bugfan/ang-admin/testenv/radius/config/users): 定义用户及其 Cleartext-Password。
