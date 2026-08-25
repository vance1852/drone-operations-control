# BENZHI_README

这是一个 Go 后端服务，用于Go 1.26 + PostgreSQL 16 的纯后端业务系统。

## 项目说明

- 项目：vance1852/drone-operations-control
- 项目用途：Go 1.26 + PostgreSQL 16 的纯后端业务系统。用户登录后管理无人机设备、任务编排、执行批次、遥测上报、安全告警、人工复核和审计。入口位于 cmd/server，领域、用例、持久化、HTTP、登录、审计、幂等和 worker 分别位于 internal 下的独立 package。
- Go 工具链：`golang:1.26`
- 前端工具链：无

## 标准构建、运行和测试命令

进入容器后执行：

```bash
# 编译
cd '/app' && GOTOOLCHAIN=local go build ./...

# 启动
cd '/app' && GOTOOLCHAIN=local go run ./cmd/server

# 测试
cd '/app' && GOTOOLCHAIN=local go test ./...
```

## Docker 构建和进入容器

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh benzhi-task-330-amd64 linux/amd64
./build_benzhi_docker.sh benzhi-task-330-arm64 linux/arm64
docker run -it benzhi-task-330-amd64:latest
docker run -it --platform linux/arm64 benzhi-task-330-arm64:latest
```

## 题目验证命令

1. 预期退出码 0：`go test ./internal/worker -run '^TestReportWorkerDoesNotOverlapSlowGenerations$' -count=1`
2. 预期退出码 0：`go test ./...`
3. 预期退出码 0：`GOTOOLCHAIN=local go build -buildvcs=false ./... && GOTOOLCHAIN=local go vet ./...`
