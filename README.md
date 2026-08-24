# 木犀招新系统 v2

Nacos 的服务配置与基础设施配置约定见 [deploy/nacos/README.md](deploy/nacos/README.md)。

基于 `go-zero` 的木犀招新系统后端仓库

## 依赖说明：go-zero fork

本项目通过 `go.mod replace` 使用 go-zero 的团队 fork（**含 etcd 认证补丁**）。

- **为什么**：etcd 启用用户名密码认证后，token 默认 5 分钟过期，而 clientv3 的 watch 不自动刷新 token（[etcd#12385](https://github.com/etcd-io/etcd/issues/12385)），go-zero 又无限紧密重试同一 client，导致 `invalid auth token` 周期性刷屏、消耗 CPU 与日志磁盘。etcd 官方明确不修（[#17384](https://github.com/etcd-io/etcd/issues/17384)），go-zero 的修复 [PR #5709](https://github.com/zeromicro/go-zero/pull/5709) 未合并，只能 fork 打补丁
- **引用**：`replace github.com/zeromicro/go-zero => github.com/Muxi-X/go-zero v1.4.5-muxi`
- **补丁仓库**：[Muxi-X/go-zero](https://github.com/Muxi-X/go-zero)（分支 `muxi-patch`，tag `v1.4.5-muxi`，[diff](https://github.com/Muxi-X/go-zero/compare/v1.4.5...muxi-patch)）
- **补丁要点**：watch/keepalive 失败时移除缓存 client 惰性重建（新 token）+ 防 goroutine 泄漏/死锁，思路对齐上游 PR #5709，以 `[Muxi Patch]` 标记
- **升级注意**：升级 go-zero 需重新 apply 补丁；若上游 #5709 合并可评估替换回官方

## 服务

- auth：身份认证
- user：用户信息
- task：作业
- review：审阅
- schedule：进度
- form：报名表
- test：测验

## 运行

### 1. 配置

 复制 `~/etc/app-example.yaml` 文件为 `~/etc/app.yaml`，并根据需要进行配置。

### 2. 构建运行

- `rpc`服务

  进入`rpc`服务目录，执行

  ```go
  go run rpc.go -f etc/rpc.yaml
  ```

- `api`服务

  进入`api`服务目录，执行

  ```go
  go run api.go -f etc/api.yaml
  ```

ps：运行整个项目时，user服务需要在task，review，form，test服务之前启动。

