# 构建工具集成

jgo 提供了对 Gradle 和 Maven 构建工具的统一配置管理，包括代理设置和路径管理。

## Gradle

### 代理配置

```bash
jgo gradle proxy                       # 查看当前代理
jgo gradle proxy host:port             # 设置代理
jgo gradle proxy user:pass@host:port   # 设置带认证的代理
jgo gradle proxy none                  # 移除代理
```

代理配置写入用户级 `~/.gradle/gradle.properties` 文件，设置以下属性：

```properties
systemProp.http.proxyHost=host
systemProp.http.proxyPort=port
systemProp.https.proxyHost=host
systemProp.https.proxyPort=port
```

如果指定了认证信息，还会设置 `systemProp.http.proxyUser` 和 `systemProp.http.proxyPassword`。

### GRADLE_USER_HOME

```bash
jgo gradle path                # 查看当前 GRADLE_USER_HOME
jgo gradle path /path/to/dir   # 设置 GRADLE_USER_HOME
```

`GRADLE_USER_HOME` 是 Gradle 的缓存和配置目录（默认 `~/.gradle`），不是 Gradle 可执行文件的安装目录。

---

## Maven

### 代理配置

```bash
jgo maven proxy                       # 查看当前代理
jgo maven proxy host:port             # 设置代理
jgo maven proxy user:pass@host:port   # 设置带认证的代理
jgo maven proxy none                  # 移除代理
```

代理配置写入用户级 `~/.m2/settings.xml` 文件的 `<proxies>` 节点。

### MAVEN_HOME

```bash
jgo maven path                # 查看当前 MAVEN_HOME
jgo maven path /path/to/dir   # 设置 MAVEN_HOME 并更新 PATH
```

设置 `MAVEN_HOME` 后，jgo 会自动将 `%MAVEN_HOME%\bin`（Windows）或 `$MAVEN_HOME/bin`（其他平台）添加到系统 `PATH` 中。
