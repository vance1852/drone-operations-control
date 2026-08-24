# 无人机机队任务编排与安全控制后端

Go 1.26 + PostgreSQL 16 的纯后端业务系统。用户登录后管理无人机设备、任务编排、执行批次、遥测上报、安全告警、人工复核和审计。入口位于 `cmd/server`，领域、用例、持久化、HTTP、登录、审计、幂等和 worker 分别位于 `internal` 下的独立 package。

```bash
docker compose up -d postgres
go test ./... -count=1
go run ./cmd/server
```

默认 DSN 为 `postgres://drone:drone@localhost:5432/drone_operations?sslmode=disable`，可通过 `DATABASE_URL` 覆盖。`GET /healthz` 检查进程存活，`GET /readyz` 实际探测数据库。

默认开发账号由迁移创建：`admin / admin123`。登录接口为 `POST /api/auth/login`；登录成功后可调用 `GET /api/auth/info`，业务 API 使用请求上下文中的操作者身份。

迁移 `migrations/001_initial.sql` 创建操作员、无人机设备、任务计划、设备分配、飞行任务、设备交接、执行批次、批次任务、遥测记录、安全告警、审计和幂等数据。迁移有版本记录和并发锁，应用重启后会从 PostgreSQL 恢复业务状态。

主要业务路径：

1. 登录并注册无人机设备，配置能力与工作空间。
2. 创建任务计划并分配无人机，经过安全策略后启动执行批次。
3. 上报遥测和异常，人工复核后关闭告警或隔离任务。
4. 通过审计查询追踪操作者、无人机、任务和安全结果。

核心验证：

```bash
go test ./... -count=1
go test -race ./... -count=1
go vet ./...
go build ./...
```

