#!/bin/bash

echo "🚀 启动后端+前端系统..."
echo ""

# 启动后端
echo "📦 启动后端服务器..."
cd /Users/op/Downloads/complex_ci_cd/backend
go run ./cmd/server &
BACKEND_PID=$!
echo "✅ 后端启动 (PID: $BACKEND_PID)"
sleep 2

# 启动前端
echo "🎨 启动前端开发服务器..."
cd /Users/op/Downloads/complex_ci_cd/frontend
npm run dev &
FRONTEND_PID=$!
echo "✅ 前端启动 (PID: $FRONTEND_PID)"
sleep 3

echo ""
echo "✅ 系统已启动！"
echo ""
echo "📍 后端: http://localhost:8080"
echo "📍 前端: http://localhost:5173"
echo ""
echo "按 Ctrl+C 停止所有服务..."

wait
