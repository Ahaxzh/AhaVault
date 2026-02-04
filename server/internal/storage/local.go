package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LocalEngine 本地文件系统存储引擎
type LocalEngine struct {
	basePath string // 基础存储路径，如 /data/storage
}

// NewLocalEngine 创建本地存储引擎实例
func NewLocalEngine(basePath string) (*LocalEngine, error) {
	// 确保基础目录存在
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &LocalEngine{
		basePath: basePath,
	}, nil
}

// Put 存储文件
func (e *LocalEngine) Put(hash string, reader io.Reader) error {
	// 验证哈希
	if err := ValidateHash(hash); err != nil {
		return fmt.Errorf("invalid hash: %w", err)
	}

	// 生成存储路径
	relativePath, err := GeneratePath(hash)
	if err != nil {
		return err
	}
	fullPath := filepath.Join(e.basePath, relativePath)

	// 检查文件是否已存在
	if _, err := os.Stat(fullPath); err == nil {
		return fmt.Errorf("file already exists: %s", hash)
	}

	// 创建目录
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 创建临时文件
	tempFile := fullPath + ".tmp"
	file, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer file.Close()

	// 写入数据
	written, err := io.Copy(file, reader)
	if err != nil {
		os.Remove(tempFile) // 清理临时文件
		return fmt.Errorf("failed to write file: %w", err)
	}

	if written == 0 {
		os.Remove(tempFile)
		return fmt.Errorf("no data written")
	}

	// 关闭文件
	if err := file.Close(); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("failed to close file: %w", err)
	}

	// 原子性重命名
	if err := os.Rename(tempFile, fullPath); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("failed to rename file: %w", err)
	}

	return nil
}

// Get 读取文件
func (e *LocalEngine) Get(hash string) (io.ReadCloser, error) {
	// 验证哈希
	if err := ValidateHash(hash); err != nil {
		return nil, fmt.Errorf("invalid hash: %w", err)
	}

	// 生成存储路径
	relativePath, err := GeneratePath(hash)
	if err != nil {
		return nil, err
	}
	fullPath := filepath.Join(e.basePath, relativePath)

	// 打开文件
	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", hash)
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

// Delete 删除文件
func (e *LocalEngine) Delete(hash string) error {
	// 验证哈希
	if err := ValidateHash(hash); err != nil {
		return fmt.Errorf("invalid hash: %w", err)
	}

	// 生成存储路径
	relativePath, err := GeneratePath(hash)
	if err != nil {
		return err
	}
	fullPath := filepath.Join(e.basePath, relativePath)

	// 删除文件
	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", hash)
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// 尝试删除空目录（忽略错误）
	dir := filepath.Dir(fullPath)
	os.Remove(dir) // 二级目录
	os.Remove(filepath.Dir(dir)) // 一级目录

	return nil
}

// Exists 检查文件是否存在
func (e *LocalEngine) Exists(hash string) (bool, error) {
	// 验证哈希
	if err := ValidateHash(hash); err != nil {
		return false, fmt.Errorf("invalid hash: %w", err)
	}

	// 生成存储路径
	relativePath, err := GeneratePath(hash)
	if err != nil {
		return false, err
	}
	fullPath := filepath.Join(e.basePath, relativePath)

	// 检查文件
	_, err = os.Stat(fullPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("failed to stat file: %w", err)
}

// Stat 获取文件信息
func (e *LocalEngine) Stat(hash string) (*FileInfo, error) {
	// 验证哈希
	if err := ValidateHash(hash); err != nil {
		return nil, fmt.Errorf("invalid hash: %w", err)
	}

	// 生成存储路径
	relativePath, err := GeneratePath(hash)
	if err != nil {
		return nil, err
	}
	fullPath := filepath.Join(e.basePath, relativePath)

	// 获取文件信息
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", hash)
		}
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	return &FileInfo{
		Hash:      hash,
		Size:      info.Size(),
		StorePath: relativePath,
	}, nil
}

// GetBasePath 获取基础路径
func (e *LocalEngine) GetBasePath() string {
	return e.basePath
}
