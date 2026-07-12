# 快速上手

## 1. 编译安装

```bash
make build        # 编译当前平台，输出到 bin/
make build-all    # 交叉编译所有平台并打包
make test         # 运行测试
make clean        # 清理构建产物
make help         # 查看所有可用命令
```

将生成压缩包解压为 `jgo`（或 `jgo.exe`）放入系统 `PATH` 中。

## 2. 设置 JDK 安装目录

jgo 默认将 JDK 安装在 `~/.jgo/jdks`。如需更改：

```bash
jgo root /path/to/your/jdks
```

查看当前设置：

```bash
jgo root
```

## 3. 配置代理（可选）

如果网络环境需要代理才能访问 JDK 源：

```bash
jgo proxy http://127.0.0.1:7890
```

也可以为单次命令指定临时代理，不影响全局配置：

```bash
jgo list available --proxy http://127.0.0.1:7890
jgo install 21 --proxy none    # 本次不使用代理
```

## 4. 配置镜像（可选）

如果访问上游 API 较慢，可以配置镜像加速：

```bash
jgo mirror Temurin tsinghua    # Temurin 使用清华 TUNA 镜像
```

详见 [镜像系统](mirror.md)。

## 5. 安装 JDK

```bash
jgo install 21                 # 安装 JDK 21（交互式选择源）
jgo install                    # 列出所有可用 JDK 进行选择
```

安装完成后，JDK 会被解压到 `root_path` 下，命名格式为 `{Source}-{Major}`，如 `Temurin-21`。

## 6. 切换 JDK

```bash
jgo use Temurin-21             # 按名称切换
jgo use 21                     # 按版本号切换（如有多个匹配会交互选择）
jgo use                        # 交互式选择
```

切换后，`JAVA_HOME` 和 `PATH` 会被更新为系统级环境变量。

## 7. 添加本地 JDK

如果已有 JDK 归档文件（`.zip` 或 `.tar.gz`）：

```bash
jgo add ./openjdk-21.zip
```

系统会提示输入自定义名称，并自动检测 JDK 版本。

## 8. 测试源连通性

```bash
jgo sourcetest                 # 测试所有源的连通性和响应时间
jgo sourcetest --proxy none    # 不使用代理测试
```
