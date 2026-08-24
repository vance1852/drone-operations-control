# BENZHI_README

这是一个用于无人机任务编排与执行监管的 Go 后端系统，管理设备、批次、遥测、安全告警、复核和审计。

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
./build_benzhi_docker.sh benzhi-task-333-amd64 linux/amd64
./build_benzhi_docker.sh benzhi-task-333-arm64 linux/arm64
docker run -it benzhi-task-333-amd64:latest
docker run -it --platform linux/arm64 benzhi-task-333-arm64:latest
```
