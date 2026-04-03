# **Development Document**

### **Table**

- [1. Project Overview](#1-project-overview)
- [2. Requirements Analysis](#2-requirements-analysis)
    - [2_1. Functional requirements](#2_1-functional-requirements)
    - [2_2. Interface requirements](#2_2-interface-requirements)
- [3. High-Level Design](#3-high-level-design)
- [4. Database Design](#4-database-design)

---

## **1. Project Overview**

**Code-Repository** is an online code and file hosting platform for developers and general users, similar to a lightweight Git repository management system. The platform supports user registration and login via email, and uses JWT for secure authentication and session management.

Core features are centered around four modules: repository, file, preview, and code quality:

- **Repository Management**: Users can create public or private repositories, manage branches, view commit history, and set repository permissions and collaborators.

- **File Management**: Supports uploading files and folders via drag-and-drop or file picker. Large files use chunked upload with resumable support. Users can create, rename, move, and delete directories/files in repositories, and download single files or batch ZIP archives.

- **Online Preview**: Provides rich preview capabilities, including syntax-highlighted text/code preview, image zoom preview, PDF paginated preview, and audio/video playback and Office document preview (converted to PDF/HTML). A collapsible left-side directory tree helps browsing repository structure.

- **Code Quality**: Deep integration with SonarQube. Users can trigger code quality analysis with one click, view gate status (pass/fail), bug counts, vulnerabilities, and code smells. Analysis results display in the repository view, and merge requests automatically run quality checks as a merge precondition.

The platform uses a frontend-backend separation architecture, the backend offers RESTful APIs, and the frontend provides a responsive UI. It aims to provide an easy-to-use, efficient, and extensible code/file management solution for individuals, teams, and small businesses, while helping users continuously improve code quality via SonarQube.

## **2. Requirements Analysis**

### **2_1. Functional requirements**

<center>

| **Function Module** | **Function** | **Description** | **Priority** |
|:---:|---|---|:---:|
| **User** | Registration | Register with email | High |
| | Login | Login with password or email verification | High |
| | Authentication | JWT-based session management to protect APIs | High |
| | Profile Management | Update avatar, nickname, etc. | Medium |
| | Repository Permissions | Set repository visibility (public/private) and manage collaborators | High |
| **Repository** | Create Repository | Create a repository with name, description, visibility | High |
| | Repository List | List owned, contributed, starred repositories with pagination and search | High |
| | Repository Settings | Modify description, visibility, delete repository | Medium |
| | Branch Management | Create / switch / delete branches (Git-style) | Low (optional) |
| | Commit History | Show commit records with changed file lists per commit | Medium |
| **File** | Upload (Drag & Drop / Folder) | Drag files/folders from local machine to upload | High |
| | Upload (File Picker) | Upload via system file picker, support multiple selection | High |
| | Chunked Upload | Automatically split large files into chunks with resumable support | Medium |
| | Directory Management | Create, rename, delete directories inside a repository | High |
| | File Management | Rename, move, delete files | High |
| | File Metadata | Store file size, MIME type, uploader, upload time, last modified time | High |
| | Download | Download a single file; optionally download multiple files or entire directories as a ZIP archive | High (single file) / Medium (batch) |
| **Preview** | Text / Code Preview | Syntax highlighting based on extension, line numbers | High |
| | Image Preview | Display images inline with zoom capability | High |
| | PDF Preview | Embedded PDF.js viewer with pagination, zoom, download | High |
| | Office Documents Preview | Convert Word/Excel/PPT to PDF or HTML for preview | Medium |
| | Audio / Video Playback | Built-in player for common formats (MP4, MP3, WebM) | Medium |
| | Directory Tree | Display repository folder structure on the left with collapse/expand | High |
| | Diff Comparison | Show differences between two versions of a text file | Low (optional) |
| **Code Quality** | SonarQube Integration Configuration | Configure SonarQube service URL, project key, and authentication token for repositories | Medium |
| | Manual Analysis Trigger | Allow users to trigger SonarQube scan via API or CI button | High |
| | Automatic Analysis Trigger | Automatically trigger scans on pushes or pull request creation | High |
| | Quality Gate Display | Show quality gate status (pass/fail) and trend charts on repo home | High |
| | Issue Detail View | Show list of bugs, vulnerabilities, and code smells with SonarQube jump links | Medium |
| | Pull Request Quality Check | Use quality gate as merge condition and block merge on fail | High |
| | Webhook Event Receiver | Receive SonarQube scan completion callbacks and update local analysis status | Medium |
</center>

**Git Management**

<center>

| **Function Module** | **Function** | **Description** | **Priority** |
|:---:|---|---|:---:|
| **Base Management** | Bare Repository | Initialize a bare repository on server backend without working directory on repository creation, unified Git object management | High |
| | Storage Quota Monitoring | Real-time calculation of `.git` directory size to prevent unlimited repository growth exhausting disk space | Medium |
| **Branch Management** | Flexible Branch Operations | Support creating new branches from specific commits, tags or existing branches; support branch renaming | High |
| | Default Branch Strategy | Allow users to customize default branch (e.g. `main`); prevent deletion of default branch when deleting branches | Medium |
| | Protected Branches | Allow setting protection rules: prohibit force push (Force Push) or restrict specific people with push rights | High |
| **Version Tracking** | Commit Details View | Click commit hash to view all file changes, author, date and detailed message of that commit | High |
| | Diff Comparison | Implement line-level diff display (Side-by-side or Inline view), supporting text highlighting | High |
| | Audit Log | Record all sensitive operations such as push, delete, merge via Git protocol or web frontend | Medium |
| **Workflow Support** | Merge Request | Support initiating merge request from source branch to target branch, automatically detect conflicts | High |
| | Code Quality Enforcement | **Mandatory requirement**: Must pass SonarQube Quality Gate before merge, otherwise lock merge button | High |
| | Web IDE Online Editing | Support modifying file code directly in browser and create new commit to specified branch | Medium |
| **External Integration** | Webhooks | Send JSON format notifications to configured external server when repository `push`, `merge` or `tag_push` occurs | Medium |
| | Deploy Keys | Allow users to configure read-only or read-write SSH/HTTP keys for automated script pulling | Low |
</center>

### **2_2. Interface requirements**

## **3. High-Level Design**

## **4. Database Design**

---

<center>
 Document Revision History

| **Number** | **Reason for revision** | **Author** | **Revision Date** |
|:---:|---|:---:|:---:|
| 1 | Document Creation | Vito Lonius | 2026/04/01 |
| 2 | Improve functional requirements | Vito Lonius | 2026/04/01 |
| 3 | Improve project overview and functional requirements | Vito Lonius | 2026/04/02 |
| 4 | Add Git management requirements | Vito Lonius | 2026/04/03 |

</center>