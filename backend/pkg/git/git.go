package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// InitBareRepo 在指定路径初始化一个 Git 裸仓库
// 裸仓库（Bare Repository）不包含工作区，仅存储 Git 对象的版本历史，
// 这是服务器端托管仓库的标准方式。
func InitBareRepo(path string) error {
	// 1. 确保父级目录存在
	parentDir := filepath.Dir(path)
	if _, err := os.Stat(parentDir); os.IsNotExist(err) {
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return fmt.Errorf("创建仓库目录失败: %v", err)
		}
	}

	// 2. 检查目录是否已存在且不为空
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		files, _ := os.ReadDir(path)
		if len(files) > 0 {
			return fmt.Errorf("路径 %s 已存在且不为空", path)
		}
	}

	// 3. 执行 git init --bare 命令
	// 相当于在终端运行: git init --bare /path/to/repo.git
	cmd := exec.Command("git", "init", "--bare", path)

	// 捕获错误输出
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git init 失败: %v, 输出: %s", err, string(output))
	}

	return nil
}

// GetRepoSize 获取仓库在磁盘上的物理大小（单位：字节）
func GetRepoSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}
