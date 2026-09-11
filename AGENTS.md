# Agent Instructions

本文档面向 AI 编码助手，介绍本项目的基本情况、构建方式与开发约定。

## 项目概览

**gitar**（Git Archive & Repository Tool）是一个用 Go 编写的命令行小工具，用于下载 Git 托管平台的仓库代码归档（archive），以及将仓库镜像（mirror）到本地。项目初衷是替代 `git clone` 全量克隆的方式，只保存需要的快照，节省存储空间。

主要功能：

- `gitar dl <url>`（别名 `download`）：解析 GitHub / Gitee 仓库 URL（支持仓库主页、`/releases/tag/<tag>`、`/tree/<branch>`、`/tree/<commit>` 等格式），通过 GitHub API 解析出对应的 commit，下载 `.tar.gz` 归档并转换为 `.tar.xz` 保存到本地；可选将归档通过邮件发送（依赖外部 `filemailer` 命令）。`--retry N` 指定下载/邮件最大重试次数（默认 20）。
- `gitar mi <url>`（别名 `mirror`）：将仓库镜像为本地 bare 仓库并持续 fetch 更新；remote 使用 mirror refspec `+refs/*:refs/*` 并开启 prune（等价 `git clone --mirror`，远端删除的引用会同步清理）。默认使用 go-git 内置实现，`--ssh` 时改用系统 `git fetch` 命令。`--retry N` 指定 fetch 最大重试次数（默认 20）。
- `gitar version`（或 `--version`）：显示版本号与 git SHA1。发布构建通过 `-ldflags -X` 注入 `pkg/app.Version` / `pkg/app.GitSHA`；本地直接 `go build` 时回退到 Go 内嵌的 VCS build info。

## 技术栈

- 语言：Go 1.25（`go.mod` 中声明 `go 1.25.0`），模块名 `gitar`
- CLI 框架：`github.com/urfave/cli/v2`
- 日志：`github.com/sirupsen/logrus`
- GitHub API：`github.com/google/go-github/v56`
- Git 操作：`github.com/go-git/go-git/v5`
- 配置：`gopkg.in/yaml.v3`（YAML 配置文件）
- 持久化：`github.com/jmoiron/sqlx` + SQLite，驱动按 build tag 自动选择：CGO 启用时用 `github.com/mattn/go-sqlite3`（`pkg/data/driver_cgo.go`），CGO 禁用时用纯 Go 的 `modernc.org/sqlite`（`pkg/data/driver_nocgo.go`）
- CI/CD：GitHub Actions（`.github/workflows/test.yml`：push 到 develop/master/main 跑三系统 × CGO 开关测试矩阵；`release.yml`：tag `v*.*.*` 构建 linux/windows/darwin 多架构纯 Go 二进制并附 checksum 发布到 Releases）
- 外部命令依赖：`curl`（下载）、`xz`（压缩格式转换）、`git`（SSH 模式 fetch）、`filemailer`（发邮件，可选）、`aria2c`（备用下载器，代码中存在但当前未使用）
