# **技术文档**

### **目录**

- [1. 后端技术实现](#1-后端技术实现)
    - [1_1. 系统架构概述](#1_1-系统架构概述)
    - [1_2. 技术栈选型](#1_2-技术栈选型)
    - [1_3. 核心功能实现方案](#1_3-核心功能实现方案)
    - [1_4. API 模块设计](#1_4-api-模块设计)
    - [1_5. 容器化与部署设计](#1_5-容器化与部署设计)
    - [1_6. 数据模型设计](#1_6-数据模型设计)
    - [1_7. MinIO 对象存储设计](#1_7-minio-对象存储设计)
- [2. 前端技术实现](#2-前端技术实现)
    - [2_1. 前端技术栈选型](#2_1-前端技术栈选型)
    - [2_2. 核心功能前端实现方案](#2_2-核心功能前端实现方案)
    - [2_3. 前后端交互设计](#2_3-前后端交互设计)
- [3. 测试技术实现](#3-测试技术实现)
    - [3_1. 后端测试](#3_1-后端测试)

---

## **1. 后端技术实现**

- **语言**：Go 1.23.0

### **1_1. 系统架构概述**
后端采用 **Gin** 框架构建 RESTful API。在已实现的文件与预览模块中，文件内容存储于 **MinIO** 对象存储，文件元数据与上传任务状态由 **PostgreSQL**（通过 GORM）管理，鉴权基于 **JWT** 无状态会话。分块上传的合并采用同步方式（MinIO ComposeObject）完成。go-git Git 管理与 Asynq 异步任务队列为规划功能，尚未实现。

### **1_2. 技术栈选型**

| **组件** | **推荐技术** | **理由** |
|:---:|:---:|---|
| Web 框架 | Gin | 生态最成熟，中间件丰富，性能优秀且易于维护 |
| 认证 | JWT (golang-jwt) | 用于文档要求的无状态会话管理 |
| 数据库 (RDBMS) | PostgreSQL | 存储用户、项目元数据及文件属性（如 MIME、大小） |
| 缓存 | Redis | 存储分块上传状态、JWT 黑名单及高频访问的项目树数据 |
| Git 库 | go-git | 纯 Go 实现，支持内存操作，无需在服务器安装 Git 二进制文件 |
| 对象存储 | MinIO | 兼容 S3 协议，用于存储大文件分块和 Office 转换后的 PDF |
| 任务队列 | Asynq | 基于 Redis 的高性能任务队列，处理代码质量分析和文件压缩 |
| 容器化 | Docker | 统一运行环境，解决“在我机器上能跑”的问题，隔离应用与其依赖 |
| 编排工具 | Docker Compose | 能够通过单一配置文件 (docker-compose.yml) 一键启动后端的依赖服务栈（DB、缓存、对象存储、SonarQube）|

### **1_3. 核心功能实现方案**

- **简单文件上传**
    - 小文件通过 `multipart/form-data` 直接上传至 MinIO。
    - 若仓库内已存在同名路径文件，先删除 MinIO 中的旧对象再覆盖写入，并更新数据库记录；否则创建新记录。
    - 自动从请求头获取 `Content-Type`，缺失时默认 `application/octet-stream`。
    - 上传失败时执行回滚：删除已写入 MinIO 的对象，避免孤儿数据。

- **分块上传与断点续传**
    - **流程**：前端调用 `UploadInit` → 多次 `UploadChunk` → `UploadComplete`，三步完成。
    - **初始化 (UploadInit)**：基于 `repoID + uploaderID + fileName + fileSize + 时间戳` 生成 SHA-256 哈希作为 `upload_id`。若检测到同 ID 的未完成任务，直接返回已上传分块索引，实现断点续传。
    - **分块上传 (UploadChunk)**：将分块暂存至 MinIO 临时桶（路径 `chunks/{uploadID}/{chunkIndex}`），在 PostgreSQL 的 `uploaded_chunk_indices` 字段追加已完成索引（逗号分隔字符串）。重复上传同一分块时直接返回当前状态，不重复写入。
    - **完成合并 (UploadComplete)**：前端显式调用合并接口，后端使用 MinIO `ComposeObject`（S3 Copy Part 协议）将所有分块拼接为目标文件，合并后删除临时桶中的分块。若目标路径已有同名文件，先删除旧对象再覆盖。整个过程为同步执行。
    - **元数据记录**：在 PostgreSQL 中维护 `upload_tasks` 表，记录 `upload_id`、分块大小、总块数、已完成块数及索引、任务状态（uploading → merging → completed）。

- **文件管理操作**
    - **目录创建**：在 `files` 表插入 `is_dir=true`、`mime_type=directory` 的记录，不涉及对象存储操作。创建前校验仓库存在性和目录唯一性。
    - **文件重命名**：更新数据库中的 `file_name` 和 `path`。对非目录文件，因 MinIO 不支持原地重命名，需将旧对象下载后以新 ObjectKey 重新上传，再删除旧对象。
    - **文件移动**：更新数据库中的 `path`，非目录文件同样执行 MinIO 拷贝 + 删除旧对象。移动前校验目标路径无同名文件。
    - **文件删除**：仅仓库所有者可操作。非目录文件同时删除 MinIO 对象和数据库记录（GORM 软删除）。目录仅删除数据库记录。
    - **文件下载**：从 MinIO 流式读取文件内容，以 `Content-Disposition: attachment` 返回，支持大文件流式传输不占用过多内存。

- **在线预览**
    - **文件分类**：基于扩展名与 MIME 类型双维度分类为 7 种类型：`code`（40+ 种语言，映射语言标识供前端语法高亮）、`image`、`video`、`audio`、`pdf`、`office`、`text`；无法识别的归为 `binary`。扩展名优先，MIME 前缀作为兜底。
    - **代码/文本预览**：文件 ≤ 2MB 时从 MinIO 读取全文内容返回，超过 2MB 仅返回元数据（类型、语言、大小），由前端决定处理方式。读取后校验 UTF-8 有效性（检测替换字符 `\uFFFD`），非 UTF-8 文件返回错误提示。
    - **图片预览**：返回原始二进制流，设置 `Cache-Control: public, max-age=86400`（1 天客户端缓存）。
    - **音视频预览**：返回原始二进制流，设置 `Accept-Ranges: bytes`，支持浏览器 Range 请求实现拖动播放。
    - **PDF 预览**：返回原始二进制流，`Content-Type: application/pdf`，`Content-Disposition: inline`，浏览器内置 PDF 查看器直接渲染。
    - **Office 文档预览**：当前后端仅返回分类类型 `office`，实际文档转 PDF 由前端处理（后端未实现转换服务）。
    - **目录树**：查询仓库全部文件记录，在内存中递归构建树形结构（`TreeNode`），每个节点包含 ID、名称、路径、是否目录、子节点列表。前端按需展开子目录。

- **Git 存储库管理**（规划中，尚未实现）
    - **目录结构**：每个 Repository 映射到服务器磁盘上的一个 `.git` 裸仓库。
    - **分支与提交**：使用 `go-git` 实现分支创建、切换和提交记录查询。
    - **目录树**：后端递归扫描 Git Tree 对象，构建并缓存 JSON 结构的目录树以供前端渲染。

- **SonarQube 深度集成**（规划中，尚未实现）
    - **配置管理**：在 Repository 表中存储 `project_key` 和 `auth_token`。
    - **分析触发**：通过 `os/exec` 调用 `sonar-scanner` 命令行工具，或通过 API 触发远程扫描。
    - **状态同步**：实现一个接收 **Webhook** 的 Endpoint。当 SonarQube 完成分析后，发送状态回执。后端更新数据库中的 Quality Gate 状态（Pass/Fail）。

### **1_4. API 模块设计**

- **鉴权中间件**

    所有受保护的 API 必须经过 `AuthMiddleware`。
    - 解析 `Authorization: Bearer <Token>`。
    - 验证 JWT 有效性及用户权限（是否为仓库协作人员）。

#### **文件模块 API**

| **方法** | **路径** | **请求参数** | **响应** | **说明** |
|:---:|---|---|---|---|
| POST | `/api/v1/files/upload` | multipart: `repo_id`, `path`, `file` | `FileResponse` | 简单文件上传，同名文件覆盖 |
| POST | `/api/v1/files/upload/init` | JSON: `repo_id`, `file_name`, `file_path`, `file_size`, `chunk_size`, `total_chunks`, `mime_type` | `UploadInitResponse` | 初始化分块上传任务 |
| POST | `/api/v1/files/upload/chunk` | multipart `chunk` + Query: `upload_id`, `chunk_index` | `UploadChunkResponse` | 上传单个分块 |
| POST | `/api/v1/files/upload/complete` | JSON: `upload_id` | `FileResponse` | 完成分块上传并合并 |
| GET | `/api/v1/files/:id` | — | `FileResponse` | 获取文件详情 |
| GET | `/api/v1/files` | Query: `repo_id`, `path` | `FileListResponse` | 列出仓库文件（支持按目录过滤） |
| GET | `/api/v1/files/:id/download` | — | Stream (binary) | 下载文件（流式传输） |
| DELETE | `/api/v1/files/:id` | — | `{ "message": "文件已成功删除" }` | 删除文件（仅仓库所有者） |
| POST | `/api/v1/files/dir` | JSON: `repo_id`, `path`, `dir_name` | `FileResponse` | 创建目录 |
| PUT | `/api/v1/files/:id/rename` | JSON: `new_name` | `FileResponse` | 重命名文件/目录（仅仓库所有者） |
| PUT | `/api/v1/files/:id/move` | JSON: `new_path` | `FileResponse` | 移动文件/目录（仅仓库所有者） |

#### **预览模块 API**

| **方法** | **路径** | **请求参数** | **响应** | **说明** |
|:---:|---|---|---|---|
| GET | `/api/v1/preview/:id` | — | `PreviewResponse` | 预览文件（文本/代码返回内容） |
| GET | `/api/v1/preview/:id/info` | — | `PreviewResponse` | 获取预览元信息（不含文件内容） |
| GET | `/api/v1/preview/:id/raw` | — | Stream (binary) | 获取原始文件内容 |
| GET | `/api/v1/preview/:id/image` | — | Stream + `Cache-Control` | 图片预览（1 天缓存） |
| GET | `/api/v1/preview/:id/media` | — | Stream + `Accept-Ranges` | 音视频预览（支持 Range 请求） |
| GET | `/api/v1/preview/:id/pdf` | — | Stream + `Content-Type: application/pdf` | PDF 预览（inline 显示） |
| GET | `/api/v1/repos/:repo_id/tree` | — | `TreeResponse` | 获取仓库完整目录树 |
| GET | `/api/v1/repos/:repo_id/dir` | Query: `path` | `{ "repo_id", "path", "items" }` | 列出指定目录内容 |

#### **请求/响应 DTO 定义**

**文件模块：**

```
UploadInitRequest:
  repo_id      uint64  (required)  仓库ID
  file_name    string  (required, 1-255)  文件名
  file_path    string  (required, min=1)  文件路径
  file_size    int64   (required, >0)  文件大小(bytes)
  chunk_size   int64   (required, >0)  分块大小(bytes)
  total_chunks int     (required, >0)  总分块数
  mime_type    string  (max=100)  MIME类型

UploadInitResponse:
  upload_id        string  上传任务唯一标识
  uploaded_chunks  int     已上传分块数（断点续传时 >0）

UploadChunkRequest:
  upload_id    string  (required)  上传任务ID
  chunk_index  int     (required, >=0)  分块索引

UploadChunkResponse:
  upload_id        string  上传任务ID
  chunk_index      int     当前分块索引
  uploaded_chunks  int     已上传分块总数
  completed        bool    是否全部上传完成

UploadCompleteRequest:
  upload_id  string  (required)  上传任务ID

FileResponse:
  id          uint64    文件ID
  repo_id     uint64    所属仓库ID
  file_name   string    文件名
  path        string    完整路径
  is_dir      bool      是否目录
  mime_type   string    MIME类型
  file_size   int64     文件大小(bytes)
  status      string    状态 (completed/uploading/merging)
  uploader_id uint64    上传者ID
  created_at  time      创建时间
  updated_at  time      更新时间

FileListResponse:
  total  int64            文件总数
  items  []FileResponse   文件列表

CreateDirRequest:
  repo_id   uint64  (required)  仓库ID
  path      string  (required, min=1)  父目录路径
  dir_name  string  (required, 1-255)  目录名

RenameFileRequest:
  new_name  string  (required, 1-255)  新文件名

MoveFileRequest:
  new_path  string  (required, min=1)  新路径
```

**预览模块：**

```
PreviewResponse:
  file_id    uint64  文件ID
  file_name  string  文件名
  mime_type  string  MIME类型
  file_size  int64   文件大小(bytes)
  file_type  string  分类类型 (code/image/video/audio/pdf/office/text/binary)
  content    string  文本内容（仅 code/text 且 <=2MB 时返回）
  language   string  编程语言标识（仅 code 类型返回，如 go/python/javascript）
  encoding   string  编码（固定 utf-8）

TreeNode:
  id        uint64       文件/目录ID
  name      string       名称
  path      string       完整路径
  is_dir    bool         是否目录
  mime_type string       MIME类型
  file_size int64        文件大小
  children  []*TreeNode  子节点列表

TreeResponse:
  repo_id  uint64       仓库ID
  tree     []*TreeNode  目录树根节点列表

RawFileResponse:
  content      []byte   文件原始二进制内容
  content_type string   Content-Type
  file_name    string   文件名
  file_size    int64    文件大小
```

### **1_5. 容器化与部署设计**

为了提升开发效率并保障部署一致性，系统全面采用 Docker 进行容器化管理。

- **本地开发环境 (Development)**
    - 使用 `docker-compose.yml` 统一部署基础设施依赖，包括 PostgreSQL、Redis、MinIO 以及 SonarQube 及其自带的数据库。
    - 后端 Gin 服务与前端项目在本地开发时可直接运行于宿主机，以便于热重载和代码调试，同时通过网络端口与 Docker 容器内的依赖组件通信。

- **生产部署环境 (Production)**
    - **后端应用镜像**：编写 `Dockerfile` 基于轻量级的 `golang:alpine` 镜像进行多阶段构建 (Multi-stage Build)，最终生成仅包含编译后二进制文件的小体积运行时镜像。
    - **前端应用镜像**：使用 Node.js 镜像进行打包，并将构建产物复制到 Nginx 镜像中提供静态资源服务及反向代理。
    - **统一网络**：所有服务运行在同一个 Docker Bridge 网络下，通过容器名称（如 `postgres`, `redis`, `api-server`）进行内部 DNS 解析，确保外部无法直接访问核心数据库，提升系统安全性。

### **1_6. 数据模型设计**

#### **files 表**

存储文件与目录的元数据，每个记录对应仓库中的一个文件或目录。

| **字段** | **类型** | **约束** | **说明** |
|:---|:---|:---|:---|
| id | uint64 | PK, autoIncrement | 主键 |
| repo_id | uint64 | NOT NULL, INDEX | 所属仓库 ID |
| file_name | varchar(255) | NOT NULL | 文件/目录名 |
| path | varchar(500) | NOT NULL | 仓库内完整路径（如 `src/main.go`） |
| is_dir | bool | DEFAULT false | 是否为目录 |
| mime_type | varchar(100) | DEFAULT '' | MIME 类型（目录为 `directory`） |
| file_size | int64 | DEFAULT 0 | 文件大小（bytes），目录为 0 |
| object_key | varchar(500) | DEFAULT '' | MinIO 对象键（目录为空） |
| upload_id | varchar(100) | DEFAULT '' | 关联的分块上传任务 ID |
| chunk_count | int | DEFAULT 0 | 分块数量（仅分块上传文件有值） |
| uploaded_chunks | int | DEFAULT 0 | 已上传分块数（仅分块上传文件有值） |
| status | varchar(20) | DEFAULT 'completed' | 状态：`completed` / `uploading` / `merging` |
| uploader_id | uint64 | NOT NULL | 上传者用户 ID |
| created_at | timestamp | AUTO | 创建时间 |
| updated_at | timestamp | AUTO | 更新时间 |
| deleted_at | timestamp | INDEX | 软删除时间（GORM） |

#### **upload_tasks 表**

存储分块上传任务的中间状态，支持断点续传。

| **字段** | **类型** | **约束** | **说明** |
|:---|:---|:---|:---|
| id | uint64 | PK, autoIncrement | 主键 |
| upload_id | varchar(100) | UNIQUE INDEX, NOT NULL | 上传任务唯一标识（SHA-256 生成） |
| repo_id | uint64 | NOT NULL, INDEX | 所属仓库 ID |
| file_name | varchar(255) | NOT NULL | 目标文件名 |
| file_path | varchar(500) | NOT NULL | 目标文件路径 |
| file_size | int64 | NOT NULL | 文件总大小（bytes） |
| chunk_size | int64 | NOT NULL | 每个分块大小（bytes） |
| total_chunks | int | NOT NULL | 总分块数 |
| uploaded_chunks | int | DEFAULT 0 | 已完成上传的分块数 |
| uploaded_chunk_indices | text | DEFAULT '' | 已完成分块索引，逗号分隔（如 `0,2,3`），支持断点续传 |
| mime_type | varchar(100) | DEFAULT '' | 文件 MIME 类型 |
| status | varchar(20) | DEFAULT 'uploading' | 状态：`uploading` → `merging` → `completed` |
| uploader_id | uint64 | NOT NULL | 上传者用户 ID |
| created_at | timestamp | AUTO | 创建时间 |
| updated_at | timestamp | AUTO | 更新时间 |
| deleted_at | timestamp | INDEX | 软删除时间（GORM） |

### **1_7. MinIO 对象存储设计**

系统使用两个 MinIO 存储桶分别管理最终文件和临时分块：

| **存储桶** | **名称** | **用途** |
|:---|:---|:---|
| 主存储桶 | `code-repo-storage` | 存储仓库的最终文件（合并后或简单上传的完整文件） |
| 临时存储桶 | `upload-chunks` | 存储分块上传的临时分块，合并后自动清理 |

#### **对象键 (Object Key) 规则**

- **最终文件**：`repos/{repoID}/{filePath}`
    - 示例：仓库 ID 为 1，文件路径 `src/main.go` → 对象键 `repos/1/src/main.go`
- **临时分块**：`chunks/{uploadID}/{chunkIndex}`
    - 示例：上传任务 `abc123`，第 3 个分块 → 对象键 `chunks/abc123/3`

#### **核心操作**

| **操作** | **方法** | **说明** |
|:---|:---|:---|
| 简单上传 | `PutObject` → 主存储桶 | 通过 multipart 直接写入 |
| 分块上传 | `PutObject` → 临时存储桶 | 每个分块独立写入 |
| 分块合并 | `ComposeObject` | 利用 S3 Copy Part 协议将所有分块拼接为目标文件，写入主存储桶 |
| 分块清理 | `RemoveObject` × N | 合并成功后逐个删除临时桶中的分块 |
| 文件下载 | `GetObject` | 从主存储桶流式读取 |
| 文件重命名/移动 | `GetObject` + `PutObject` + `RemoveObject` | MinIO 不支持原地重命名，需拷贝到新键后删除旧键 |
| 文件删除 | `RemoveObject` | 删除主存储桶中的对象 |

#### **初始化与自动配置**

服务启动时自动检查两个存储桶是否存在，不存在则自动创建，无需手动预配置。

## **2. 前端技术实现**

- **语言**：TypeScript 
- **包管理器**：pnpm / npm

### **2_1. 前端技术栈选型**

| **组件**  |            **推荐技术**            | **理由**                                                     |
| :-------: | :--------------------------------: | ------------------------------------------------------------ |
| 核心框架  | React 18 / Vue 3 (Composition API) | 社区生态繁荣，组件化开发模式极大地提升代码复用率和可维护性。（注：可根据团队熟悉度任选其一） |
| 构建工具  |                Vite                | 提供极速的冷启动和热重载(HMR)，大幅提升前端开发体验。        |
| UI 组件库 |     Ant Design / Element Plus      | 提供完善的表格、树形控件、模态框、进度条等企业级组件，加速界面开发。 |
| 状态管理  |   Zustand (React) / Pinia (Vue)    | 轻量级状态管理，用于全局存储用户登录状态、JWT Token 以及全局主题配置。 |
| 路由管理  |     React Router / Vue Router      | 支持嵌套路由和路由守卫，用于实现页面鉴权拦截。               |
| 样式方案  |            Tailwind CSS            | 实用优先的 CSS 框架，方便快速编写高度定制化的响应式布局。    |
| 网络请求  |               Axios                | 配合请求/响应拦截器，统一处理 JWT 注入、Token 过期刷新及全局错误提示。 |

### **2_2. 核心功能前端实现方案**

- **大文件分块上传与断点续传**
    - **文件切片**：利用 HTML5 的 `File.slice()` API 将大文件在浏览器端切割为固定大小（如 5MB）的 Chunk。
    - **Hash 计算**：利用 `SparkMD5` 库结合 `Web Worker` 在后台线程计算文件整体的 MD5 或 SHA-256 Hash，避免阻塞主线程导致页面卡顿。
    - **并发与重试**：使用 `Promise.all` 控制并发上传请求数量（如最大 3-5 个），并对失败的切片实现自动重试逻辑。
    - **进度计算**：通过监听 Axios 的 `onUploadProgress` 事件结合已上传分块数，实时计算并更新总进度条。

- **复杂文件类型的在线预览**
    - **代码/文本预览**：集成 `Monaco Editor`（VS Code 的核心编辑器）并设置为只读模式，实现近乎原生 IDE 体验的语法高亮和行号显示。
    - **Markdown 渲染**：使用 `marked.js` 将 Markdown 转换为 HTML，配合 `DOMPurify` 进行 XSS 攻击过滤。
    - **PDF 预览**：集成 `pdf.js`，实现纯前端的 PDF 文件解析与分页渲染，提供缩放和跳转功能。
    - **Diff 差异对比**：对于 Commit 记录和 PR 差异，接入 `diff2html` 或 Monaco 的 Diff 模式，直观展示代码的增删改情况。

- **超大目录树的性能优化**
    - **虚拟列表 (Virtual Scrolling)**：由于部分底层仓库文件可能多达数千个，左侧目录树和文件列表必须采用虚拟化技术（如 `react-window` 或 `vue-virtual-scroller`），只渲染可视区域内的 DOM 节点，避免内存溢出和渲染卡顿。
    - **按需加载 (Lazy Load)**：目录树初始只加载顶层结构，用户点击展开文件夹时再发起网络请求获取子目录数据。

### **2_3. 前后端交互设计**

- **鉴权拦截机制**：前端在 Axios 实例中配置拦截器，将存储在 `localStorage` 或 `sessionStorage` 中的 JWT 附加在 HTTP Header (`Authorization: Bearer <token>`) 中。当后端返回 `401 Unauthorized` 时，自动跳转至登录页。
- **长连接与实时通知 (可选)**：针对 SonarQube 代码质量分析这种耗时较长的异步任务，前端可采用 `SSE (Server-Sent Events)` 或定时轮询 (Polling) 机制向后端查询分析进度，以便在界面上实时更新进度条或展示最终的门禁结果。

## **3. 测试技术实现**

### **3_1. 后端测试**

- **单元测试**：
    - 标准库 testing: Go 原生支持，用于编写基础逻辑测试
    - testify: 提供了更优雅的 assert 断言和 mock 功能，能显著提升测试代码的可读性
- **API/集成测试**：
    - httptest: Go 标准库，用于模拟 HTTP 请求，测试 Gin Handler 路由和中间件
    - sqlmock: 针对数据库层（PostgreSQL），在不连接真实数据库的情况下模拟 SQL 执行结果
- **模拟/打桩**：
    - GoMock: 用于模拟接口。由于项目涉及 MinIO 和 Git 文件系统操作，通过 Mock 接口可以避免测试时产生真实的磁盘读写或网络开销

---

<center>
 文档修订历史

| **编号** | **修订原因** | **作者** | **修订日期** |
|:---:|---|:---:|:---:|
| 1 | 文档创建 | 牛茂润 | 2026/04/03 |
| 2 | 完善后端技术实现 | 牛茂润 | 2026/04/03 |
| 3 | 引入 Docker 容器化方案与部署设计 | 牛茂润 | 2026/04/03 |
| 4 | 引入后端测试技术 | 牛茂润 | 2026/04/03 |
| 5 | 完善文件与预览模块：修正架构概述和分块上传方案，补充简单上传/文件管理/预览实现方案，补全 API 设计与 DTO 定义，新增数据模型与 MinIO 存储设计 | 刘咏 | 2026/05/05 |

</center>