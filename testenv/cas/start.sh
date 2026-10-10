#!/bin/bash
# 启动本地 CAS 2.0/3.0 Mock Server
cd "$(dirname "$0")" || exit 1
echo "🚀 正在启动 CAS 统一身份认证测试服务 (http://127.0.0.1:8082/cas)..."
go run main.go
