## Context

当前项目管理服务在前端通过 localStorage 选择当前角色和用户，并在 API 请求中发送 `X-Role` / `X-User`。后端将这些 header 作为权限判断依据，用于项目 owner、participants、任务 owner 和 RBAC 校验。

这个方式适合本地原型，但生产风险很明确：浏览器可以伪造 header，用户目录硬编码在前端，项目域直接承担身份职责，后续接入企业 SSO、LDAP/AD 或已有用户中心时会改动大量业务代码。

本设计将身份能力拆成独立用户微服务。项目管理服务只消费标准化身份声明和用户目录查询结果，不直接关心登录方式。这样可以先落地内置账号体系，后续通过 identity provider adapter 切换到外部用户系统。

## Goals / Non-Goals

**Goals:**
- 建立独立 `user-service` 微服务边界，包含账户、认证、令牌、成员、角色和外部身份映射。
- 用可信 token claims 替代生产环境的 `X-Role` / `X-User` header 身份来源。
- 让项目域保存稳定 user id，而不是依赖可变展示名或前端硬编码用户列表。
- 定义可替换身份提供方接口，支持内置用户、OIDC/OAuth2、LDAP/AD 或其他企业用户系统。
- 保留开发环境身份模拟，但通过配置强约束，避免误入生产。

**Non-Goals:**
- 不在本变更内实现完整企业 SSO 页面和所有外部协议细节。
- 不把项目、任务、里程碑权限规则迁移到用户服务；用户服务只提供身份、成员和角色事实。
- 不引入多租户计费、组织套餐或复杂 IAM policy language。
- 不要求一次性迁移所有历史 owner 字段；允许兼容迁移窗口。

## Decisions

### 1. 用户服务独立部署，项目服务通过 token + API 集成

选择：新增独立 `user-service`，拥有独立 HTTP API 和独立 MySQL schema。项目管理服务验证访问令牌，并在必要时调用用户服务查询用户、工作区成员和角色。

理由：身份是横切能力，不应继续耦合在项目域。独立部署可以让未来外接企业用户系统时只改用户服务或适配器，不改项目管理业务规则。

替代方案：把用户表直接加到现有项目服务。实现更快，但会把认证、用户目录、外部身份映射和项目业务混在一起，后续切换成本高。

### 2. 生产身份来源为 Bearer token claims，不信任浏览器身份 header

选择：生产环境 API 使用 `Authorization: Bearer <token>`。token 至少包含 `sub`、`workspace_id`、`roles`、`display_name`、`email`、`provider`、`version`。项目服务通过 JWKS 本地验签或 introspection 校验 token。

理由：权限判断必须基于服务端签发或可信外部 IdP 签发的声明。`X-User` / `X-Role` 只能作为开发模拟或内部网关转发字段，不能被浏览器直接控制。

替代方案：继续使用 header，并依赖网关注入。该方案可以用于内部服务，但当前前端直连 API，风险不可接受；如果未来有 API Gateway，也应由网关验证 token 后再注入内部 headers。

### 3. 内置身份提供方与外部身份提供方使用统一 adapter

选择：用户服务定义 `IdentityProvider` adapter，包含 authenticate、syncProfile、resolveExternalIdentity、refreshSession 等能力。内置账号是第一个 provider，OIDC/LDAP 等作为后续 provider 接入。

理由：用户系统的关键目标是“后续方便切入其他用户系统”。统一 adapter 可以避免把 OIDC、LDAP 或企业用户中心细节泄漏给项目服务。

替代方案：只做内置账号表。短期简单，但未来接入外部系统时会重构登录、用户映射和成员同步。

### 4. 项目域引用稳定 user id，展示信息按需解析

选择：项目 owner、participants、任务 owner、周报 author 等业务字段长期应保存 `user_id`。API 响应可补充 `displayName`、`avatarUrl` 等展示字段，前端成员选择从用户目录查询。

理由：用户名、邮箱、外部系统账号都可能变化；稳定 ID 可以保证历史审计和权限判断一致。

替代方案：继续保存字符串用户名。迁移成本低，但容易出现重名、改名、外部账号映射失效。

### 5. RBAC 保留在项目服务，用户服务提供角色事实

选择：用户服务管理 workspace membership 和 role assignment；项目服务仍持有 `PermManageProject`、`PermManageMilestone` 等业务权限矩阵，并结合项目 owner/participants 做授权。

理由：用户服务不应理解项目管理领域规则。它只回答“这个用户是谁、在哪个 workspace、拥有什么角色”。业务服务负责“这个角色在这个对象上能否做这个动作”。

替代方案：把所有权限规则集中到用户服务。集中化看似统一，但会让用户服务耦合项目、任务、里程碑等业务对象，演变成难维护的通用权限引擎。

### 6. 开发身份模拟保留，但配置隔离

选择：保留本地开发 role/user selector 或 dev token endpoint，但必须满足 `APP_ENV=development` 或显式 `AUTH_MODE=dev-header`，生产启动时若启用则失败。

理由：当前 E2E 和手工验证依赖快速切换角色。完全移除会降低开发效率；但必须在配置层防止生产误用。

替代方案：立即移除所有模拟能力。安全更强，但会使本地调试和 E2E 准备成本显著增加。

## Risks / Trade-offs

- [Risk] 微服务拆分增加部署和本地开发复杂度 -> 通过 Docker Compose、健康检查和开发 seed 用户降低启动成本。
- [Risk] token claims 与实时角色变更存在延迟 -> token 设置短有效期，并支持 token version / membership version 失效策略。
- [Risk] 历史数据 owner/participants 是字符串，无法立即映射到 user id -> 提供迁移脚本和兼容解析层，迁移失败项进入人工处理清单。
- [Risk] 外部身份系统不可用会影响登录 -> 内置服务保留本地 break-glass admin，且项目服务对已签发短期 token 可继续验签。
- [Risk] 过早引入复杂 IAM 会拖慢产品功能 -> 本阶段只做 workspace role + 项目对象授权，不做自定义 policy language。

## Migration Plan

1. 新增 user-service skeleton、数据库 schema、内置 seed 用户和登录/token API。
2. 项目服务新增 auth middleware，支持 `AUTH_MODE=dev-header` 与 `AUTH_MODE=token` 双模式。
3. 前端新增登录态和用户目录加载；开发模式保留 selector，生产模式隐藏。
4. 将项目 owner/participants、任务 owner、周报 author 的创建入口改为 user id；读取接口兼容旧字符串并返回解析状态。
5. 增加迁移脚本，把现有 `tester/alice/bob/carol/frontend-user` 映射为内置用户。
6. E2E 从 dev header 模式迁移为 dev token 模式，覆盖登录、成员选择和权限拒绝。
7. 生产切换到 token 模式后，禁止未认证请求和浏览器身份 header。

Rollback：保留 `AUTH_MODE=dev-header` 仅用于非生产环境；若 token 集成异常，可回滚项目服务配置到上一版本并保持 user-service 数据不参与权限判断。

## Open Questions

- 首个生产外部身份目标是 OIDC/OAuth2、LDAP/AD，还是已有公司用户中心 API？
- workspace 是否需要从一开始支持多组织，还是先做单 workspace + 可扩展字段？
- 用户服务是否需要管理邀请流程，还是仅由管理员创建成员？
- token 使用自签 JWT + JWKS，还是 opaque token + introspection？默认建议 JWT + JWKS，除非已有统一网关要求 opaque token。
