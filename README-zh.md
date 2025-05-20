# SyncDocs

## 项目描述

SyncDocs 是一项服务，可从 GitHub 仓库同步文档或文件，并提供一个 Web 界面来管理和查看同步的内容。它利用 GitHub API 访问仓库数据，将信息存储在 PostgreSQL 数据库中，并执行后台同步任务。

## 部署先决条件

*   Git
*   Docker
*   Docker Compose

## 部署步骤

1.  **克隆仓库：**
    ```bash
    git clone https://github.com/RedwindA/SyncDocs
    ```

2.  **导航到项目目录：**
    ```bash
    cd SyncDocs
    ```
    (或者您克隆仓库时使用的名称)

3.  **创建 `.env` 文件：**
    复制示例环境文件：
    ```bash
    cp .env.example .env
    ```

4.  **配置 `.env` 文件：**
    打开 `.env` 文件并使用您的特定设置更新以下变量：
    *   `SERVER_PORT`：应用程序服务器将侦听的端口 (默认为 `8080`)。
    *   `AUTH_USER`：基本身份验证的用户名 (例如 `admin`)。**请替换为强大且唯一的用户名。**
    *   `AUTH_PASS`：基本身份验证的密码 (例如 `changeme`)。**请替换为强大且唯一的密码。**
    *   `DATABASE_URL`：您的 PostgreSQL 数据库的连接字符串。
        *   示例：`postgres://user:password@host:port/dbname?sslmode=disable`
    *   `GITHUB_TOKEN`：您的 GitHub 个人访问令牌。此令牌需要 `repo` 范围才能访问仓库内容。您可以在 [https://github.com/settings/tokens](https://github.com/settings/tokens) 生成一个。
    *   `SYNC_INTERVAL`：后台同步任务的间隔 (例如 `1h` 表示 1 小时, `30m` 表示 30 分钟)。如果未设置或无效，则默认为 `1h`。

5.  **构建并运行应用程序：**
    使用 Docker Compose 拉取镜像并在分离模式下启动容器：
    ```bash
    docker compose up -d
    ```

6.  **访问应用程序：**
    容器运行后，您可以在 Web 浏览器中通过以下地址访问应用程序：
    `http://<your_host_ip_or_localhost>:${SERVER_PORT}`
    (将 `<your_host_ip_or_localhost>` 替换为您的服务器 IP 地址或 `localhost` (如果在本地运行)，并将 `${SERVER_PORT}` 替换为您在 `.env` 文件中配置的端口)。

## 项目结构 (概述)

*   `cmd/`：包含主要应用程序入口点 (例如 `cmd/server/main.go`)。
*   `internal/`：包含应用程序的核心逻辑，包括 API 处理程序、身份验证、数据库交互、GitHub 客户端和同步器服务。
*   `migrations/`：数据库迁移文件。
*   `web/frontend/`：包含 Vue.js 前端应用程序。
*   `docker-compose.yml`：定义 Docker 的服务、网络和卷。
*   `Dockerfile`：构建应用程序 Docker 镜像的说明。

## 贡献

欢迎贡献！请随时提交拉取请求或开启问题。
