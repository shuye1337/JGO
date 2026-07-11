# 镜像插件系统

jgo 支持通过镜像插件将 JDK 源的请求路由到镜像站，加速在中国大陆的下载体验。

## 使用方法

### 查看镜像状态

```bash
jgo mirror
```

输出当前已配置的镜像映射，以及所有可用的镜像插件。

### 设置镜像

```bash
jgo mirror Temurin tsinghua
```

将 Temurin 源切换到清华 TUNA 镜像。设置后，`jgo list available`、`jgo install`、`jgo sourcetest` 对 Temurin 源的请求均走镜像站。

### 移除镜像

```bash
jgo mirror Temurin none
```

恢复 Temurin 源使用上游官方 API。

## 已实现的镜像插件

### 清华 TUNA Adoptium 镜像

- **插件 ID：** `tsinghua`
- **支持的源：** Temurin
- **镜像地址：** `https://mirrors.tuna.tsinghua.edu.cn/Adoptium/`
- **实现方式：** 解析镜像站的静态目录列表（纯 HTML，无 JSON API）
- **注意事项：**
  - 镜像不提供 checksum 校验文件
  - 只下载 `.zip` / `.tar.gz` 归档，不下载安装包（`.msi` 等）
  - 镜像中的版本目录是上游的子集，可能缺少某些 feature version
  - 每个叶子目录通常只包含该 major 版本的最新一个版本

### 镜像目录结构

```
/Adoptium/
├── 8/  11/  17/  18/  19/  20/  21/  25/     # major 版本目录
│   └── jdk/{arch}/{os}/
│       ├── OpenJDK21U-jdk_x64_windows_hotspot_21.0.11_10.zip
│       ├── OpenJDK21U-jdk_x64_linux_hotspot_21.0.11_10.tar.gz
│       └── ...
├── deb/                                        # 不使用
└── rpm/                                        # 不使用
```

- **os 目录名：** `windows`、`linux`、`mac`（注意不是 `macos`）、`alpine-linux`
- **arch 目录名：** `x64`、`aarch64`、`ppc64`、`ppc64le`、`riscv64`、`s390x`

## 扩展镜像

新增镜像只需在 `internal/mirror/` 下添加一个 `.go` 文件，实现 `Mirror` 接口并在 `init()` 中自注册，无需修改任何现有代码。

```go
func init() { Register(&myMirror{}) }

type myMirror struct{}

func (m *myMirror) ID() string                 { return "my-mirror" }
func (m *myMirror) DisplayName() string        { return "My Mirror" }
func (m *myMirror) SupportedSources() []string { return []string{"Temurin"} }
func (m *myMirror) ListAvailable(source, os, arch, proxy string) ([]provider.JDKAsset, error) {
    // 实现镜像目录解析逻辑
}
func (m *myMirror) TestSources(source, os, arch, proxy string) ([]provider.RequestRecord, error) {
    // 实现连通性测试逻辑
}
```

重新编译后即可通过 `jgo mirror <source> my-mirror` 使用。
