.PHONY: help install dev build preview lint format clean

help:
	@echo "发布控制系统 - 前端项目命令"
	@echo ""
	@echo "Available commands:"
	@echo "  make install   - 安装依赖"
	@echo "  make dev       - 启动开发服务器 (http://localhost:5173)"
	@echo "  make build     - 生产构建"
	@echo "  make preview   - 预览构建结果"
	@echo "  make lint      - 代码检查和格式化"
	@echo "  make format    - 格式化代码"
	@echo "  make clean     - 清理构建产物"

install:
	npm install

dev:
	npm run dev

build:
	npm run build

preview:
	npm run preview

lint:
	npm run lint

format:
	npm run format

clean:
	rm -rf node_modules
	rm -rf dist
	rm -rf .vite
