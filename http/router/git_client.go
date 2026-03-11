package router

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/richer421/q-deploy/domain/engine/gitops"
)

// 简单的 GitClient 实现，基于本地 git 命令
// 为了避免过度设计，这里只实现当前需要的 clone/pull + 写文件 + commit/push

type gitClient struct{}

type workingCopy struct {
	root string
}

func NewGitClient() gitops.GitClient {
	return &gitClient{}
}

func (c *gitClient) CloneOrPull(ctx context.Context, repoURL, branch string) (gitops.WorkingCopy, error) {
	// 这里简单起见，使用系统临时目录 + repo 名作为工作目录
	// 真正产品化可以抽到配置
	base := os.TempDir()
	name := filepath.Base(repoURL)
	root := filepath.Join(base, "q-deploy-gitops", name)

	if _, err := os.Stat(root); os.IsNotExist(err) {
		if err := os.MkdirAll(root, 0o755); err != nil {
			return nil, err
		}
		cmd := exec.CommandContext(ctx, "git", "clone", "-b", branch, repoURL, root)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return nil, err
		}
	} else {
		cmd := exec.CommandContext(ctx, "git", "pull", "origin", branch)
		cmd.Dir = root
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return nil, err
		}
	}

	return &workingCopy{root: root}, nil
}

func (w *workingCopy) WriteFile(path string, data []byte) error {
	abs := filepath.Join(w.root, path)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return err
	}
	return os.WriteFile(abs, data, 0o644)
}

func (w *workingCopy) CommitAndPush(message string) error {
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = w.root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("git", "commit", "-m", message)
	cmd.Dir = w.root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// 如果没有变更，commit 会失败，这里直接忽略错误
	_ = cmd.Run()

	cmd = exec.Command("git", "push")
	cmd.Dir = w.root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
