package service

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"code-repo/internal/model/dto"
	"code-repo/internal/model/entity"
	"code-repo/internal/repository/db"
	"code-repo/pkg/git"
	"code-repo/pkg/utils"
)

type RepoService interface {
	CreateRepo(userID uint64, req dto.CreateRepoRequest) (*dto.RepoResponse, error)
	GetRepoDetail(repoID uint64) (*dto.RepoResponse, error)
	DeleteRepo(userID uint64, repoID uint64) error
	UpdateRepo(userID uint64, repoID uint64, req dto.UpdateRepoRequest) error
}

type repoService struct {
	repoDB db.RepoRepository
}

func NewRepoService(repoDB db.RepoRepository) RepoService {
	return &repoService{repoDB: repoDB}
}

// CreateRepo 实现创建仓库的业务逻辑
func (s *repoService) CreateRepo(userID uint64, req dto.CreateRepoRequest) (*dto.RepoResponse, error) {
	// 1. 业务准入检查：检查该用户下是否已有同名仓库
	existing, _ := s.repoDB.GetByNameAndOwner(req.Name, userID)
	if existing != nil && existing.ID > 0 {
		return nil, errors.New("该用户下已存在同名仓库")
	}

	// 2. 路径规划
	// 使用配置文件中的 git.root_path
	// 物理路径格式: /data/repositories/{userID}/{repoName}.git
	repoRelativePath := filepath.Join(fmt.Sprintf("%d", userID), req.Name+".git")
	absPath := filepath.Join(utils.Config.Git.RootPath, repoRelativePath)

	// 3. 物理持久化：初始化 Git 裸仓库
	if err := git.InitBareRepo(absPath); err != nil {
		return nil, fmt.Errorf("初始化物理仓库失败: %v", err)
	}

	// 4. 数据库持久化准备
	repo := &entity.Repository{
		Name:          req.Name,
		Description:   req.Description,
		OwnerID:       userID,
		IsPublic:      req.IsPublic,
		Path:          absPath,
		DefaultBranch: utils.Config.Git.DefaultBranch, // 从配置读取默认分支
	}

	// 5. 执行存库操作
	if err := s.repoDB.Create(repo); err != nil {
		// ⚠️ 补偿逻辑：如果数据库写入失败，必须回滚物理磁盘操作，防止产生脏文件夹
		_ = os.RemoveAll(absPath)
		return nil, fmt.Errorf("仓库数据存库失败: %v", err)
	}

	// 6. 重新加载完整数据（包含关联的 Owner 信息）用于返回
	fullRepo, err := s.repoDB.GetByID(repo.ID)
	if err != nil {
		return s.convertToResponse(repo), nil // 降级返回
	}

	return s.convertToResponse(fullRepo), nil
}

// GetRepoDetail 获取仓库详情
func (s *repoService) GetRepoDetail(repoID uint64) (*dto.RepoResponse, error) {
	repo, err := s.repoDB.GetByID(repoID)
	if err != nil {
		return nil, errors.New("仓库不存在")
	}
	return s.convertToResponse(repo), nil
}

// convertToResponse 将实体转换为返回给前端的 DTO
func (s *repoService) convertToResponse(repo *entity.Repository) *dto.RepoResponse {
	// 动态生成 CloneURL
	// 格式参考：http://localhost:8080/git/{userID}/{repoName}.git
	cloneURL := fmt.Sprintf("http://localhost:%d/git/%d/%s.git",
		utils.Config.Server.Port, repo.OwnerID, repo.Name)

	return &dto.RepoResponse{
		ID:            repo.ID,
		Name:          repo.Name,
		Description:   repo.Description,
		OwnerID:       repo.OwnerID,
		OwnerNickname: repo.Owner.Nickname,
		IsPublic:      repo.IsPublic,
		DefaultBranch: repo.DefaultBranch,
		CloneURL:      cloneURL,
		CreatedAt:     repo.CreatedAt,
	}
}

// DeleteRepo 删除仓库：不仅删除数据库记录，还要彻底删除磁盘上的 .git 目录
func (s *repoService) DeleteRepo(userID uint64, repoID uint64) error {
	repo, err := s.repoDB.GetByID(repoID)
	if err != nil {
		return errors.New("仓库不存在")
	}

	// 权限校验：只有所有者可以删除
	if repo.OwnerID != userID {
		return errors.New("权限不足，只有所有者可以删除仓库")
	}

	// 1. 先删除数据库记录（确保前端查不到）
	if err := s.repoDB.Delete(repoID); err != nil {
		return err
	}

	// 2. 异步或同步删除物理目录
	// 注意：os.RemoveAll 是不可逆的，实际生产中建议先移动到回收站（.trash）
	go func(path string) {
		_ = os.RemoveAll(path)
	}(repo.Path)

	return nil
}

// UpdateRepo 更新仓库元数据
func (s *repoService) UpdateRepo(userID uint64, repoID uint64, req dto.UpdateRepoRequest) error {
	updates := make(map[string]interface{})
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.IsPublic != nil {
		updates["is_public"] = *req.IsPublic
	}

	return s.repoDB.UpdateFields(repoID, updates)
}
