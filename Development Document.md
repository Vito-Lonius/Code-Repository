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

## **2. Requirements Analysis**

### **2_1. Functional requirements**

<center>
Table 1 Function Module Description

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

</center>