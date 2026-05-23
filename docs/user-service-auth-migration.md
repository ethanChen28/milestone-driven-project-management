# 用户服务鉴权迁移与回滚

## 迁移步骤

1. 启动 `user-service`，确认 `/api/v1/health` 返回 `ok`。
2. 运行身份引用报告：`go run ./cmd/migrate-identities -dsn "$MYSQL_DSN"`。
3. 处理 `unresolvedReferences`，确保项目 owner、participants、任务 owner、周报 author 都能映射到稳定用户 id。
4. 将项目服务配置从 `AUTH_MODE=dev-header` 切换为 `AUTH_MODE=token`，并设置与用户服务一致的 `AUTH_TOKEN_SECRET`。
5. 前端设置 `VITE_AUTH_MODE=token`，使用登录态和用户目录，不再显示手动角色/用户选择器。

## 回滚约束

- 只允许非生产环境回滚到 `AUTH_MODE=dev-header`。
- 生产环境启动时如果 `APP_ENV=production` 且 `AUTH_MODE=dev-header`，服务必须拒绝启动。
- 回滚不会删除用户服务数据；它只让项目服务临时恢复开发 header 模式。
