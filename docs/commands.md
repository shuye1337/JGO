# 命令参考

## jgo list

列出已安装的 JDK。

```bash
jgo list
```

输出格式：

```
Name                 Version              Source       Path
--------------------------------------------------------------------------------
* Temurin-21         21.0.11+10           Temurin      /home/user/.jgo/jdks/Temurin-21
  Corretto-17        17.0.11.9.1          Corretto     /home/user/.jgo/jdks/Corretto-17
```

`*` 标记当前激活的 JDK。

### jgo list available

列出所有可下载的 JDK。

```bash
jgo list available
jgo list available --proxy http://host:port    # 使用临时代理
jgo list available --proxy none                # 不使用代理
```

结果按源分组显示，包含 Major 版本号、详细版本号和归档类型。

**Flags:**

| Flag | 说明 |
|------|------|
| `--available` | 等同于 `jgo list available` |
| `--proxy <url>` | 覆盖代理设置（`none` 表示禁用） |

---

## jgo install

下载并安装 JDK。

```bash
jgo install 21                     # 安装 JDK 21
jgo install                        # 列出所有可用 JDK 进行选择
jgo install 21 --proxy http://...  # 使用临时代理
```

流程：

1. 查询所有源（含镜像）获取可用 JDK 列表
2. 按版本号过滤（如指定了版本）
3. 交互式选择要安装的 JDK
4. 下载、解压、验证 JDK
5. 以 `{Source}-{Major}` 命名保存（如 `Temurin-21`）

如果同名 JDK 已存在，会提示是否覆盖。

**Flags:**

| Flag | 说明 |
|------|------|
| `--proxy <url>` | 覆盖代理设置 |

---

## jgo add

从本地归档文件添加 JDK。

```bash
jgo add ./openjdk-21.zip
jgo add /path/to/jdk.tar.gz
```

要求：

- 文件必须是 `.zip` 或 `.tar.gz` 格式
- 归档内必须包含有效的 JDK（含 `bin/java` 和 `bin/javac`）
- 系统会提示输入自定义名称

---

## jgo use

切换当前使用的 JDK，设置 `JAVA_HOME` 和 `PATH` 系统环境变量。

```bash
jgo use Temurin-21     # 按精确名称切换
jgo use 21             # 按版本号切换（唯一匹配时直接切换）
jgo use                # 交互式选择
```

---

## jgo remove

从托管的 JDKs 中移除指定的 JDK，可选择是否删除磁盘上的 JDK 文件夹。

```bash
jgo remove Temurin-21     # 按精确名称移除
jgo remove                # 交互式选择
```

流程：

1. 显示待移除 JDK 的名称、版本、来源和路径（如果是当前激活的 JDK 会予以警告）
2. 确认是否从托管列表中移除该 JDK
3. 询问是否同时删除磁盘上的 JDK 文件夹（此操作不可逆）
4. 执行移除，自动清理 `current` 标记（如果被移除的是当前使用的 JDK）

每次确认都需要输入 `y`，输入 `n` 可取消操作。

---

## jgo root

查看或设置 JDK 安装根目录。

```bash
jgo root               # 查看当前根目录
jgo root /path/to/dir  # 设置根目录
```

默认值：`~/.jgo/jdks`

---

## jgo proxy

查看、设置或移除全局下载代理。

```bash
jgo proxy                      # 查看当前代理
jgo proxy http://host:port     # 设置代理
jgo proxy none                 # 移除代理
```

代理影响所有网络请求：JDK 下载、源列表查询、源连通性测试。

---

## jgo mirror

查看、设置或移除 JDK 源的镜像。

```bash
jgo mirror                             # 查看已配置镜像和可用插件
jgo mirror Temurin tsinghua            # 为 Temurin 设置清华镜像
jgo mirror Temurin none                # 移除 Temurin 的镜像
```

详见 [镜像系统](mirror.md)。

---

## jgo sourcetest

测试所有 JDK 源的连通性和响应时间。

```bash
jgo sourcetest                         # 使用当前配置测试
jgo sourcetest --proxy http://host:port  # 使用临时代理测试
jgo sourcetest --proxy none            # 不使用代理测试
```

输出每个源的请求 URL、响应时间和状态（OK / 错误信息），以及总体统计。

**Flags:**

| Flag | 说明 |
|------|------|
| `--proxy <url>` | 覆盖代理设置 |

---

## jgo gradle

### jgo gradle proxy

查看、设置或移除 Gradle Wrapper 的代理配置。

```bash
jgo gradle proxy                       # 查看当前代理
jgo gradle proxy host:port             # 设置代理
jgo gradle proxy user:pass@host:port   # 设置带认证的代理
jgo gradle proxy none                  # 移除代理
```

配置写入 `~/.gradle/gradle.properties` 中的 `systemProp.*` 属性。

### jgo gradle path

查看或设置 `GRADLE_USER_HOME` 环境变量。

```bash
jgo gradle path                # 查看当前值
jgo gradle path /path/to/dir   # 设置
```

---

## jgo maven

### jgo maven proxy

查看、设置或移除 Maven Wrapper 的代理配置。

```bash
jgo maven proxy                       # 查看当前代理
jgo maven proxy host:port             # 设置代理
jgo maven proxy user:pass@host:port   # 设置带认证的代理
jgo maven proxy none                  # 移除代理
```

配置写入 `~/.m2/settings.xml` 中的 `<proxies>` 节点。

### jgo maven path

查看或设置 `MAVEN_HOME` 环境变量，并自动将 `%MAVEN_HOME%\bin` 添加到 `PATH`。

```bash
jgo maven path                # 查看当前值
jgo maven path /path/to/dir   # 设置并更新 PATH
```
