// Package storage 提供存储引擎抽象层
//
// 本文件实现内存存储引擎，主要用于单元测试，
// 避免测试时依赖真实的文件系统。
//
// 作者: AhaVault Team
// 创建时间: 2026-02-04
package storage

import (
	"bytes"
	"fmt"
	"io"
	"sync"
)

// MemoryEngine 内存存储引擎（用于测试）
type MemoryEngine struct {
	mu    sync.RWMutex
	files map[string][]byte // hash -> file content
}

// NewMemoryEngine 创建内存存储引擎实例
func NewMemoryEngine() *MemoryEngine {
	return &MemoryEngine{
		files: make(map[string][]byte),
	}
}

// Put 存储文件到内存
func (e *MemoryEngine) Put(hash string, reader io.Reader) error {
	// 验证哈希
	if err := ValidateHash(hash); err != nil {
		return fmt.Errorf("invalid hash: %w", err)
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// 检查文件是否已存在
	if _, exists := e.files[hash]; exists {
		return fmt.Errorf("file already exists: %s", hash)
	}

	// 读取并存储数据
	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read data: %w", err)
	}

	if len(data) == 0 {
		return fmt.Errorf("no data to store")
	}

	e.files[hash] = data
	return nil
}

// Get 从内存读取文件
func (e *MemoryEngine) Get(hash string) (io.ReadCloser, error) {
	// 验证哈希
	if err := ValidateHash(hash); err != nil {
		return nil, fmt.Errorf("invalid hash: %w", err)
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	data, exists := e.files[hash]
	if !exists {
		return nil, fmt.Errorf("file not found: %s", hash)
	}

	// 返回数据的副本（避免外部修改）
	copyData := make([]byte, len(data))
	copy(copyData, data)

	return io.NopCloser(bytes.NewReader(copyData)), nil
}

// Delete 从内存删除文件
func (e *MemoryEngine) Delete(hash string) error {
	// 验证哈希
	if err := ValidateHash(hash); err != nil {
		return fmt.Errorf("invalid hash: %w", err)
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.files[hash]; !exists {
		return fmt.Errorf("file not found: %s", hash)
	}

	delete(e.files, hash)
	return nil
}

// Exists 检查文件是否存在于内存
func (e *MemoryEngine) Exists(hash string) (bool, error) {
	// 验证哈希
	if err := ValidateHash(hash); err != nil {
		return false, fmt.Errorf("invalid hash: %w", err)
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	_, exists := e.files[hash]
	return exists, nil
}

// Stat 获取内存中文件的信息
func (e *MemoryEngine) Stat(hash string) (*FileInfo, error) {
	// 验证哈希
	if err := ValidateHash(hash); err != nil {
		return nil, fmt.Errorf("invalid hash: %w", err)
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	data, exists := e.files[hash]
	if !exists {
		return nil, fmt.Errorf("file not found: %s", hash)
	}

	path, _ := GeneratePath(hash)
	return &FileInfo{
		Hash:      hash,
		Size:      int64(len(data)),
		StorePath: path,
	}, nil
}

// Clear 清空所有数据（测试辅助方法）
func (e *MemoryEngine) Clear() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.files = make(map[string][]byte)
}

// Count 返回存储的文件数量（测试辅助方法）
func (e *MemoryEngine) Count() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.files)
}
