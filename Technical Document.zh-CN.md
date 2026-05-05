# **技术文档**

### **目录**

- [1. 后端技术实现](#1-后端技术实现)
    - [1_1. 系统架构概述](#1_1-系统架构概述)
    - [1_2. 技术栈选型](#1_2-技术栈选型)
    - [1_3. 核心功能实现方案](#1_3-核心功能实现方案)
        - [1_3_1. 文件上传](#1_3_1-文件上传)
        - [1_3_2. 文件管理操作](#1_3_2-文件管理操作)
        - [1_3_3. 文件预览](#1_3_3-文件预览)
        - [1_3_4. Git 存储库管理（规划中）](#1_3_4-git-存储库管理规划中尚未实现)
        - [1_3_5. SonarQube 深度集成（规划中）](#1_3_5-sonarqube-深度集成规划中尚未实现)
    - [1_4. API 模块设计](#1_4-api-模块设计)
        - [1_4_1. 文件管理 API](#1_4_1-文件管理-api)
        - [1_4_2. 文件预览 API](#1_4_2-文件预览-api)
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
后端采用 **Gin** 框架构建 RESTful API，通过 **go-git** 库直接操作底层存储。由于涉及大文件处理和外部工具（SonarQube）集成，系统采用异步任务队列处理耗时操作。

当前已实现的核心模块包括：
- **文件管理**：支持简单上传与分块上传（断点续传），文件实体存储于 MinIO 对象存储，元数据由 PostgreSQL 管理。   
- **文件预览**：基于扩展名与 MIME 类型的双维度分类系统，支持代码/文本、图片、音视频、PDF、Office 等 7 种文件类型的在线预览。

规划中但尚未实现的模块：**go-git Git 版本管理**（分支、提交、差异对比）、**Asynq 异步任务队列**（文件合并、SonarQube 分析）、**SonarQube 深度集成**（扫描触发、Webhook 回调、质量门禁）。

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

#### **1_3_1. 文件上传**

- **简单上传**
    - **适用场景**：小文件直接通过 `multipart/form-data` 上传，无需分块。
    - **流程**：前端通过表单提交文件和目标路径，后端调用 MinIO `PutObject` 将文件写入主存储桶。
    - **覆盖策略**：若仓库内已存在同名路径文件，先删除旧 MinIO 对象，再上传新文件并更新数据库记录，实现同名覆盖。

- **分块上传与断点续传**
    - **三阶段流程**：
        1. **Init**：前端发送文件元信息（文件名、大小、分块大小、总分块数），后端基于 `repoID + uploaderID + fileName + fileSize + 时间戳` 生成 SHA-256 哈希作为 `upload_id`，在 `upload_tasks` 表中创建任务记录，返回 `upload_id` 和已上传分块数（用于续传）。
        2. **Chunk**：前端逐块发送分块数据，后端将分块暂存至 MinIO 临时桶（路径：`chunks/{uploadID}/{chunkIndex}`），并在数据库中更新已完成的分块索引（逗号分隔字符串，如 `"0,2,3"`）。若某分块已上传则跳过，实现幂等性。当全部分块上传完毕时，任务状态自动流转为 `merging`。
        3. **Complete**：前端确认上传完成，后端调用 MinIO `ComposeObject`（S3 Copy Part 协议）将所有分块按序合并为目标文件，合并完成后删除临时分块，并在 `files` 表中创建文件记录。
    - **断点续传**：若 Init 时发现相同 `upload_id` 的任务仍处于 `uploading` 状态，直接返回已上传的分块索引，前端据此跳过已完成的分块继续上传。
    - **UploadID 生成**：`SHA-256(repoID_uploaderID_fileName_fileSize_timestamp)` 取前 16 字节十六进制，确保同一用户对同一文件的多次上传具有唯一标识。

#### **1_3_2. 文件管理操作**

- **文件详情与列表**
    - `GetFileDetail`：通过文件 ID 查询 `files` 表，返回文件元数据。
    - `ListFiles`：通过仓库 ID 和父目录路径查询，返回该目录下的直接子文件/子目录列表。

- **文件下载**
    - 通过文件 ID 查询获取 `object_key`，调用 MinIO `GetObject` 返回流式数据，设置 `Content-Disposition: attachment` 触发浏览器下载。不支持目录下载。

- **创建目录**
    - 在 `files` 表中插入 `is_dir=true` 的记录，`mime_type` 固定为 `directory`，`file_size=0`，不涉及 MinIO 对象操作。若目录路径已存在则返回错误。

- **重命名文件/目录**
    - 更新数据库中的 `file_name` 和 `path`。对非目录文件，由于 MinIO 不支持原地重命名，需执行"下载旧对象 → 上传至新 ObjectKey → 删除旧对象"三步操作。仅仓库所有者有权操作。

- **移动文件/目录**
    - 与重命名逻辑类似，更新 `path` 和 `object_key`，非目录文件需在 MinIO 中拷贝到新位置并删除旧对象。目标路径同名检查防止覆盖。仅仓库所有者有权操作。

- **删除文件/目录**
    - 仅仓库所有者可操作。非目录文件同时删除 MinIO 对象和数据库记录；目录仅删除数据库记录（当前为单条删除，不递归删除子内容）。使用 GORM 软删除（`deleted_at`）。

#### **1_3_3. 文件预览**

- **文件分类系统**

    基于**文件扩展名**（主维度）和 **MIME 类型前缀**（次维度）双维度分类，将文件归为以下 7 种类型：

    | **类型** | **识别方式** | **支持格式** |
    |:---:|---|---|
    | **code** | 扩展名映射语言标识 | .go, .py, .js, .ts, .java, .c, .cpp, .rs, .swift, .kt, .rb, .php, .sql, .html, .css, .json, .yaml, .md, .vue, .proto 等 40+ 种扩展名 |
    | **image** | 扩展名匹配 | .jpg, .jpeg, .png, .gif, .bmp, .webp, .ico, .tiff, .avif, .heic |
    | **video** | 扩展名匹配 | .mp4, .webm, .ogg, .avi, .mov, .mkv, .flv, .wmv |
    | **audio** | 扩展名匹配 | .mp3, .wav, .flac, .aac, .m4a, .opus |
    | **pdf** | 扩展名匹配 | .pdf |
    | **office** | 扩展名匹配 | .doc, .docx, .xls, .xlsx, .ppt, .pptx, .odt, .ods, .odp |
    | **text** | 扩展名匹配 或 MIME 前缀 (text/, application/json 等) | .txt, .csv, .env, .log, .gitignore 等 |

    无法识别的文件归为 `binary` 类型。分类函数 `classifyFile(fileName, mimeType)` 优先使用扩展名匹配，扩展名未命中时回退到 MIME 前缀匹配。

- **预览内容返回策略**

    | **文件类型** | **返回内容** | **特殊处理** |
    |:---:|---|---|
    | code / text | 文件文本内容 (UTF-8) | 文件 ≤ 2MB 时返回完整内容；超过 2MB 仅返回元数据；读取后校验 UTF-8 有效性（检测替换字符 `\uFFFD`），非 UTF-8 文件不返回内容 |
    | image | 原始二进制流 | 设置 `Cache-Control: public, max-age=86400`（1 天缓存） |
    | video / audio | 原始二进制流 | 设置 `Accept-Ranges: bytes`（支持 Range 请求，前端可实现拖动进度条） |
    | pdf | 原始二进制流 | `Content-Type` 固定为 `application/pdf`，`Content-Disposition: inline`（浏览器内嵌显示） |
    | office | 仅分类元信息 | 后端返回 `file_type=office`，实际文档转换预览由前端处理（当前后端未实现 Office→PDF 转换） |
    | binary | 仅分类元信息 | 不返回内容，前端提示"该文件类型暂不支持在线预览" |

- **目录树构建**
    - 从 `files` 表查询仓库全部文件记录，通过路径层级关系递归构建树形 `TreeNode` 结构。
    - 每个 `TreeNode` 包含：`id, name, path, is_dir, mime_type, file_size, children`。
    - 支持按路径查询子目录内容（`ListDir`），前端可按需展开指定目录。

- **原始文件获取**
    - `GetRawFile`：直接从 MinIO 下载文件二进制数据，自动检测 `Content-Type`（优先使用 `http.DetectContentType`，回退到扩展名映射表）。用于前端直接渲染图片、音视频、PDF 等二进制内容。

#### **1_3_4. Git 存储库管理（规划中，尚未实现）**
    - **目录结构**：每个 Repository 映射到服务器磁盘上的一个 `.git` 裸仓库。
    - **分支与提交**：使用 `go-git` 实现分支创建、切换和提交记录查询。
    - **目录树**：后端递归扫描 Git Tree 对象，构建并缓存 JSON 结构的目录树以供前端渲染。

    - **配置管理**：在 Repository 表中存储 `project_key` 和 `auth_token`。
    - **分析触发**：通过 `os/exec` 调用 `sonar-scanner` 命令行工具，或通过 API 触发远程扫描。
    - **状态同步**：实现一个接收 **Webhook** 的 Endpoint。当 SonarQube 完成分析后，发送状态回执。后端更新数据库中的 Quality Gate 状态（Pass/Fail）。

### **1_4. API 模块设计**

- **鉴权中间件**

    所有受保护的 API 必须经过 `AuthMiddleware`。
    - 解析 `Authorization: Bearer <Token>`。
    - 验证 JWT 有效性，提取 `userID` 注入 Gin Context。

#### **1_4_1. 文件管理 API**

| **方法** | **路径** | **认证** | **请求参数** | **响应** | **说明** |
|:---:|---|:---:|---|---|---|
| POST | /api/v1/files/upload | ✅ | multipart/form-data: `repo_id`(form), `path`(form), `file`(file) | `FileResponse` | 简单文件上传，同名文件覆盖 |
| POST | /api/v1/files/upload/init | ✅ | JSON: `UploadInitRequest` | `UploadInitResponse` | 初始化分块上传任务 |
| POST | /api/v1/files/upload/chunk | ✅ | multipart + Query: `upload_id`, `chunk_index`; Body: `chunk`(file) | `UploadChunkResponse` | 上传单个分块 |
| POST | /api/v1/files/upload/complete | ✅ | JSON: `UploadCompleteRequest` | `FileResponse` | 完成分块上传并合并 |
| GET | /api/v1/files/:id | ✅ | 路径参数: `id` | `FileResponse` | 获取文件详情 |
| GET | /api/v1/files | ✅ | Query: `repo_id`, `path` | `FileListResponse` | 列出仓库目录下文件 |
| GET | /api/v1/files/:id/download | ✅ | 路径参数: `id` | Stream (binary) | 下载文件（流式传输） |
| DELETE | /api/v1/files/:id | ✅ | 路径参数: `id` | `{message}` | 删除文件（仅仓库所有者） |
| POST | /api/v1/files/dir | ✅ | JSON: `CreateDirRequest` | `FileResponse` | 创建目录 |
| PUT | /api/v1/files/:id/rename | ✅ | 路径参数: `id`; JSON: `RenameFileRequest` | `FileResponse` | 重命名文件/目录（仅仓库所有者） |
| PUT | /api/v1/files/:id/move | ✅ | 路径参数: `id`; JSON: `MoveFileRequest` | `FileResponse` | 移动文件/目录（仅仓库所有者） |

**文件模块请求/响应 DTO 定义**：

```go
// UploadInitRequest - 分块上传初始化请求
type UploadInitRequest struct {
    RepoID      uint64 `json:"repo_id" binding:"required"`
    FileName    string `json:"file_name" binding:"required,min=1,max=255"`
    FilePath    string `json:"file_path" binding:"required,min=1"`
    FileSize    int64  `json:"file_size" binding:"required,gt=0"`
    ChunkSize   int64  `json:"chunk_size" binding:"required,gt=0"`
    TotalChunks int    `json:"total_chunks" binding:"required,gt=0"`
    MimeType    string `json:"mime_type" binding:"max=100"`
}

// UploadInitResponse - 分块上传初始化响应
type UploadInitResponse struct {
    UploadID       string `json:"upload_id"`
    UploadedChunks int    `json:"uploaded_chunks"`  // 已上传分块数（续传时 > 0）
}

// UploadChunkRequest - 分块上传请求
type UploadChunkRequest struct {
    UploadID   string `form:"upload_id" binding:"required"`
    ChunkIndex int    `form:"chunk_index" binding:"required,gte=0"`
}

// UploadChunkResponse - 分块上传响应
type UploadChunkResponse struct {
    UploadID       string `json:"upload_id"`
    ChunkIndex     int    `json:"chunk_index"`
    UploadedChunks int    `json:"uploaded_chunks"`
    Completed      bool   `json:"completed"`  // 全部分块是否上传完毕
}

// UploadCompleteRequest - 完成分块上传请求
type UploadCompleteRequest struct {
    UploadID string `json:"upload_id" binding:"required"`
}

// FileResponse - 文件信息响应
type FileResponse struct {
    ID         uint64    `json:"id"`
    RepoID     uint64    `json:"repo_id"`
    FileName   string    `json:"file_name"`
    Path       string    `json:"path"`
    IsDir      bool      `json:"is_dir"`
    MimeType   string    `json:"mime_type"`
    FileSize   int64     `json:"file_size"`
    Status     string    `json:"status"`
    UploaderID uint64    `json:"uploader_id"`
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
}

// FileListResponse - 文件列表响应
type FileListResponse struct {
    Total int64          `json:"total"`
    Items []FileResponse `json:"items"`
}

// CreateDirRequest - 创建目录请求
type CreateDirRequest struct {
    RepoID  uint64 `json:"repo_id" binding:"required"`
    Path    string `json:"path" binding:"required,min=1"`
    DirName string `json:"dir_name" binding:"required,min=1,max=255"`
}

// RenameFileRequest - 重命名请求
type RenameFileRequest struct {
    NewName string `json:"new_name" binding:"required,min=1,max=255"`
}

// MoveFileRequest - 移动请求
type MoveFileRequest struct {
    NewPath string `json:"new_path" binding:"required,min=1"`
}
```

#### **1_4_2. 文件预览 API**

| **方法** | **路径** | **认证** | **请求参数** | **响应** | **说明** |
|:---:|---|:---:|---|---|---|
| GET | /api/v1/preview/:id | ✅ | 路径参数: `id` | `PreviewResponse` (含 content) | 预览文件（code/text 返回文本内容） |
| GET | /api/v1/preview/:id/info | ✅ | 路径参数: `id` | `PreviewResponse` (不含 content) | 获取预览元信息（类型、语言、大小） |
| GET | /api/v1/preview/:id/raw | ✅ | 路径参数: `id` | Stream (binary) | 获取原始文件内容（自动检测 Content-Type） |
| GET | /api/v1/preview/:id/image | ✅ | 路径参数: `id` | Stream (binary) | 图片预览（Cache-Control: 1 天） |
| GET | /api/v1/preview/:id/media | ✅ | 路径参数: `id` | Stream (binary) | 音视频预览（Accept-Ranges: bytes） |
| GET | /api/v1/preview/:id/pdf | ✅ | 路径参数: `id` | Stream (binary) | PDF 预览（Content-Type: application/pdf, inline） |
| GET | /api/v1/repos/:repo_id/tree | ✅ | 路径参数: `repo_id` | `TreeResponse` | 获取仓库完整目录树 |
| GET | /api/v1/repos/:repo_id/dir | ✅ | 路径参数: `repo_id`; Query: `path` | `{repo_id, path, items}` | 列出指定目录内容 |

**预览模块请求/响应 DTO 定义**：

```go
// PreviewResponse - 文件预览响应
type PreviewResponse struct {
    FileID   uint64 `json:"file_id"`
    FileName string `json:"file_name"`
    MimeType string `json:"mime_type"`
    FileSize int64  `json:"file_size"`
    FileType string `json:"file_type"`              // code/image/video/audio/pdf/office/text/binary
    Content  string `json:"content,omitempty"`       // code/text 类型且 ≤ 2MB 时返回
    Language string `json:"language,omitempty"`      // code 类型时返回语言标识
    Encoding string `json:"encoding,omitempty"`      // 固定 "utf-8"
}

// TreeNode - 目录树节点
type TreeNode struct {
    ID       uint64      `json:"id"`
    Name     string      `json:"name"`
    Path     string      `json:"path"`
    IsDir    bool        `json:"is_dir"`
    MimeType string      `json:"mime_type,omitempty"`
    FileSize int64       `json:"file_size,omitempty"`
    Children []*TreeNode `json:"children,omitempty"`
}

// TreeResponse - 目录树响应
type TreeResponse struct {
    RepoID uint64      `json:"repo_id"`
    Tree   []*TreeNode `json:"tree"`
}
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

#### **1_6_1. files 表**

文件与目录的统一存储表，同时记录文件的 MinIO 对象映射和上传状态。

| **字段** | **类型** | **约束** | **说明** |
|:---|:---|:---|:---|
| id | uint64 | PK, auto_increment | 主键 |
| repo_id | uint64 | NOT NULL, INDEX | 所属仓库 ID |
| file_name | varchar(255) | NOT NULL | 文件名或目录名 |
| path | varchar(500) | NOT NULL | 文件在仓库中的完整路径（如 `src/main.go`） |
| is_dir | bool | DEFAULT false | 是否为目录 |
| mime_type | varchar(100) | DEFAULT '' | MIME 类型（目录固定为 `directory`） |
| file_size | int64 | DEFAULT 0 | 文件大小（字节），目录为 0 |
| object_key | varchar(500) | DEFAULT '' | MinIO 对象存储键（目录为空） |
| upload_id | varchar(100) | DEFAULT '' | 关联的分块上传任务 ID |
| chunk_count | int | DEFAULT 0 | 分块数量（仅分块上传文件有值） |
| uploaded_chunks | int | DEFAULT 0 | 已上传分块数 |
| status | varchar(20) | DEFAULT 'completed' | 文件状态：`completed` / `uploading` / `merging` |
| uploader_id | uint64 | NOT NULL | 上传者用户 ID |
| created_at | timestamp | AUTO | 创建时间 |
| updated_at | timestamp | AUTO | 更新时间 |
| deleted_at | timestamp | INDEX | 软删除时间（GORM 软删除） |

**索引设计**：
- `(repo_id)`：按仓库查询文件列表。
- `(repo_id, path)` 组合查询：用于 `GetByPath` 查找同名文件（覆盖/重命名检查）。
- `(deleted_at)`：GORM 软删除索引。

#### **1_6_2. upload_tasks 表**

分块上传任务的中间状态表，支持断点续传。

| **字段** | **类型** | **约束** | **说明** |
|:---|:---|:---|:---|
| id | uint64 | PK, auto_increment | 主键 |
| upload_id | varchar(100) | UNIQUE, NOT NULL | 上传任务唯一标识（SHA-256 生成） |
| repo_id | uint64 | NOT NULL, INDEX | 目标仓库 ID |
| file_name | varchar(255) | NOT NULL | 文件名 |
| file_path | varchar(500) | NOT NULL | 目标存储路径 |
| file_size | int64 | NOT NULL | 文件总大小（字节） |
| chunk_size | int64 | NOT NULL | 单个分块大小（字节） |
| total_chunks | int | NOT NULL | 总分块数 |
| uploaded_chunks | int | DEFAULT 0 | 已上传完成的分块数 |
| uploaded_chunk_indices | text | DEFAULT '' | 已上传分块索引，逗号分隔（如 `"0,2,3"`），用于断点续传判断 |
| mime_type | varchar(100) | DEFAULT '' | 文件 MIME 类型 |
| status | varchar(20) | DEFAULT 'uploading' | 任务状态：`uploading` / `merging` / `completed` |
| uploader_id | uint64 | NOT NULL | 上传者用户 ID |
| created_at | timestamp | AUTO | 创建时间 |
| updated_at | timestamp | AUTO | 更新时间 |
| deleted_at | timestamp | INDEX | 软删除时间 |

**索引设计**：
- `(upload_id)` UNIQUE：通过 upload_id 快速查找任务，支持续传和去重。
- `(repo_id)`：按仓库查询上传任务。

### **1_7. MinIO 对象存储设计**

#### **1_7_1. 存储桶规划**

| **桶名** | **用途** | **生命周期** |
|:---|:---|:---|
| `code-repo-storage` | 主存储桶，存放最终的文件对象 | 永久存储，随文件删除而删除 |
| `upload-chunks` | 临时桶，存放分块上传的切片数据 | 临时存储，合并完成后由后端主动清理 |

#### **1_7_2. 对象键（Object Key）规则**

| **对象类型** | **Object Key 格式** | **示例** |
|:---|:---|:---|
| 最终文件 | `repos/{repoID}/{filePath}` | `repos/1/src/main.go` |
| 上传分块 | `chunks/{uploadID}/{chunkIndex}` | `chunks/a1b2c3.../0` |

**设计说明**：
- 最终文件的 Object Key 以 `repos/` 为前缀，按仓库 ID 隔离命名空间，确保不同仓库的同名文件不会冲突。
- 分块的 Object Key 以 `chunks/` 为前缀，存储在临时桶中，`uploadID` 确保不同上传任务之间隔离。
- 路径中的 `filePath` 经过 `filepath.Clean` 和去除前导 `/` 处理，保证路径规范化。

#### **1_7_3. 核心操作**

| **操作** | **MinIO API** | **说明** |
|:---|:---|:---|
| 简单上传 | `PutObject` → 主桶 | 单次写入完整文件 |
| 分块暂存 | `PutObject` → 临时桶 | 按分块索引写入临时对象 |
| 分块合并 | `ComposeObject` | 基于 S3 Copy Part 协议，将多个分块按序拼接为目标文件，写入主桶 |
| 分块清理 | `RemoveObject` → 临时桶 | 合并完成后逐个删除临时分块 |
| 文件下载 | `GetObject` → 主桶 | 返回流式读取器，支持大文件流式传输 |
| 文件删除 | `RemoveObject` → 主桶 | 删除文件时同步删除对象 |
| 文件重命名/移动 | `GetObject` + `PutObject` + `RemoveObject` | MinIO 不支持原地重命名，需"下载→上传至新Key→删除旧对象"三步完成 |
| 桶初始化 | `BucketExists` + `MakeBucket` | 服务启动时自动检查并创建缺失的桶 |

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
| 5 | 补全文件管理与预览功能技术文档：修正架构概述、补充分块上传实际流程、新增简单上传/文件管理/预览方案、补全 API 设计与 DTO 定义、新增数据模型与 MinIO 存储设计 | 刘咏 | 2026/05/05 |

</center>