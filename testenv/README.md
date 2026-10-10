# 认证测试环境总览 (Authentication Test Environments)

本目录为 `ang-admin` 及 `ang` 网关引擎提供各个认证协议的独立模拟服务，用于认证流、单因子及多因子登录联调测试。

各个认证测试环境**完全独立解耦**，各自拥有独立的配置文件、编排文件与说明文档：

---

## 目录结构索引

| 目录 | 协议类型 | 默认端口 | 核心说明文档 |
|------|---------|---------|-------------|
| [`testenv/radius/`](./radius/) | **RADIUS** (FreeRADIUS) | `1812/udp` (认证), `1813/udp` (计费) | [RADIUS 测试说明](./radius/README.md) |
| [`testenv/cas/`](./cas/) | **CAS 2.0/3.0** (Mock CAS Server) | `8082` (HTTP) | [CAS 测试说明](./cas/README.md) |

---

## 快速入口

### 1. RADIUS 测试环境
```bash
# 进入 RADIUS 目录
cd testenv/radius

# 启动 FreeRADIUS 测试容器
docker compose up -d

# 停止容器
docker compose down
```
详细配置、测试账号请查阅：[`testenv/radius/README.md`](./radius/README.md)

---

### 2. CAS 测试环境
```bash
# 进入 CAS 目录
cd testenv/cas

# 本地极速启动（推荐）
./start.sh

# 或通过 Docker 启动
docker compose up -d

# 停止 Docker 容器
docker compose down
```
详细端点、测试账号请查阅：[`testenv/cas/README.md`](./cas/README.md)
