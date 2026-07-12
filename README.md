# jgo

Java 版本管理器 — 管理多个 JDK 安装的命令行工具。

## 特性

- 从四个上游源下载 JDK：Alibaba Dragonwell、Amazon Corretto、Azul Zulu、Eclipse Temurin
- 从本地 `.zip` / `.tar.gz` 归档添加 JDK
- 切换 `JAVA_HOME` 和 `PATH` 环境变量
- 镜像插件系统（支持清华 TUNA 等镜像加速下载）
- 统一管理 Gradle / Maven 的代理和路径配置
- 代理支持（全局代理 + 单次命令临时代理）

## 安装

从源码编译：

```bash
go build -o jgo .
```

或使用 Makefile（推荐，包含版本信息注入）：

```bash
make build        # 编译当前平台，输出到 bin/
make build-all    # 交叉编译所有平台并打包
make test         # 运行测试
make clean        # 清理构建产物
make help         # 查看所有可用命令
```

支持的交叉编译平台：`windows/linux/darwin` × `amd64/arm64`。

将 `jgo` 放入 `PATH` 即可使用。

## 快速开始

```bash
jgo root                    # 查看 JDK 安装根目录（默认 ~/.jgo/jdks）
jgo install 21              # 下载并安装 JDK 21
jgo use Temurin-21          # 切换到 JDK 21
jgo list                    # 查看已安装的 JDK
```

## 命令速查

| 命令 | 说明 |
|------|------|
| `jgo list` | 列出已安装的 JDK |
| `jgo list available` | 列出所有可下载的 JDK |
| `jgo install [version]` | 下载并安装 JDK |
| `jgo add <path>` | 从本地归档添加 JDK |
| `jgo use [name]` | 切换当前使用的 JDK |
| `jgo remove [name]` | 从托管 JDKs 中移除（可选是否删除文件夹） |
| `jgo root [path]` | 查看/设置 JDK 安装根目录 |
| `jgo proxy [url]` | 查看/设置/移除下载代理 |
| `jgo mirror [source] [mirror-id]` | 查看/设置/移除镜像源 |
| `jgo sourcetest` | 测试所有 JDK 源的连通性 |
| `jgo gradle proxy [url]` | 查看/设置 Gradle 代理 |
| `jgo gradle path [path]` | 查看/设置 GRADLE_USER_HOME |
| `jgo maven proxy [url]` | 查看/设置 Maven 代理 |
| `jgo maven path [path]` | 查看/设置 MAVEN_HOME |

## 文档

- [快速上手](docs/getting-started.md)
- [命令参考](docs/commands.md)
- [JDK 源说明](docs/jdk-sources.md)
- [镜像系统](docs/mirror.md)
- [构建工具集成](docs/build-tools.md)

## 配置文件

配置文件位于 `~/.jgo/config.json`，格式如下：

```json
{
  "root_path": "/home/user/.jgo/jdks",
  "proxy": "",
  "mirrors": {},
  "jdks": {},
  "current": ""
}
```

## 许可证

本项目基于 [MIT License](LICENSE) 开源。
