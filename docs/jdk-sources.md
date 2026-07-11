# JDK 源说明

jgo 支持从以下四个上游源获取 JDK。每个源提供不同厂商的 OpenJDK 发行版。

## 1. Alibaba Dragonwell

Alibaba Dragonwell 是阿里巴巴维护的 OpenJDK 发行版，针对阿里云场景做了优化。

**API 端点：**

```
GET https://dragonwell-jdk.io/releases.json
```

返回所有可用版本的 JSON 列表。每个版本包含多平台的下载链接，推荐使用其中的 OSS 链接以获得更好的下载体验。

**官方地址：** https://dragonwell-jdk.io/

---

## 2. Amazon Corretto

Amazon Corretto 是 AWS 维护的免费、多平台 OpenJDK 发行版，提供长期支持。

**API 端点：**

```
GET https://downloads.corretto.aws/latest-release.json
```

返回各平台最新版本的下载信息。jgo 只取 macOS、Windows、Linux 三个平台的 JDK（不含 JRE），只取 `.zip` 和 `.tar.gz` 格式。

返回的 URL 为相对路径，需拼接前缀 `https://corretto.aws` 使用。

**官方文档：** https://docs.aws.amazon.com/corretto/

---

## 3. Azul Zulu

Azul Zulu 是 Azul Systems 提供的 OpenJDK 发行版，提供广泛的版本和平台支持。

**API 端点：**

```
前缀：https://api.azul.com/metadata
```

Azul 使用 OpenAPI 规范的 REST API。完整的 API 定义见 [azul-openapi.json](./openapi/azul-openapi.json)。

**官方文档：** https://docs.azul.com/core/install/metadata-api

---

## 4. Eclipse Temurin (Adoptium)

Eclipse Temurin 是 Eclipse Adoptium 项目提供的 OpenJDK 发行版，前身为 AdoptOpenJDK。

**API 端点：**

```
前缀：https://api.adoptium.net
```

Adoptium 使用 OpenAPI 规范的 REST API。完整的 API 定义见 [adoptium-openapi.yaml](./openapi/adoptium-openapi.yaml)。

**官方文档：** https://adoptium.net/zh-CN/installation/ci-scripts

---

## 镜像支持

以上源在中国大陆访问可能存在延迟。jgo 提供了镜像插件系统来加速下载，详见 [镜像系统](mirror.md)。

当前已实现的镜像：

| 镜像 | 插件 ID | 支持的源 | 说明 |
|------|---------|----------|------|
| 清华 TUNA | `tsinghua` | Temurin | `https://mirrors.tuna.tsinghua.edu.cn/Adoptium/` |

## API 参考文件

`./openapi/` 目录下包含各源的 API 参考文件：

| 文件 | 说明 |
|------|------|
| `azul-openapi.json` | Azul Metadata API 的 OpenAPI 定义 |
| `adoptium-openapi.yaml` | Adoptium API 的 OpenAPI 定义 |
