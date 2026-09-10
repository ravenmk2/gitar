# 🐹 适合仓鼠症的代码下载小工具

[![test](https://github.com/ravenmk2/gitar/actions/workflows/test.yml/badge.svg)](https://github.com/ravenmk2/gitar/actions/workflows/test.yml)
[![release](https://github.com/ravenmk2/gitar/actions/workflows/release.yml/badge.svg)](https://github.com/ravenmk2/gitar/actions/workflows/release.yml)

### 📜 使用示例

```bash
# 自动选择 Release 或 Branch 下载
gitar dl https://github.com/kubernetes/kubernetes

# 从 Release 下载
gitar dl https://github.com/kubernetes/kubernetes/releases/tag/v1.25.15

# 从 Branch 下载
gitar dl https://github.com/kubernetes/kubernetes/tree/master

# 从 Commit ID 下载
gitar dl https://github.com/kubernetes/kubernetes/tree/6d6d7b6fbf41ed539edf21944a92f61f52929660
```

### 👀 为什么不用 `git clone` ?

使用 Git 可以增量更新，但也需要在本地保存所有的历史记录，对于一些比较大的仓库是非常浪费存储空间的。
