package storage

import (
	"bytes"
	"crypto/rand"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// TestNewLocalEngine 测试本地存储引擎创建
func TestNewLocalEngine(t *testing.T) {
	tempDir := t.TempDir()

	engine, err := NewLocalEngine(tempDir)
	if err != nil {
		t.Fatalf("NewLocalEngine() error = %v", err)
	}

	if engine.basePath != tempDir {
		t.Errorf("NewLocalEngine() basePath = %v, want %v", engine.basePath, tempDir)
	}

	// 验证目录已创建
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Error("NewLocalEngine() did not create base directory")
	}
}

// TestLocalEnginePutGet 测试文件存储和读取
func TestLocalEnginePutGet(t *testing.T) {
	tempDir := t.TempDir()
	engine, _ := NewLocalEngine(tempDir)

	hash := "aabbccddeeff1122334455667788990011223344556677889900aabbccddeeff"
	data := []byte("Hello, AhaVault!")

	// 存储文件
	err := engine.Put(hash, bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Put() error = %v", err)
	}

	// 读取文件
	reader, err := engine.Get(hash)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	defer reader.Close()

	// 验证内容
	retrieved, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	if !bytes.Equal(data, retrieved) {
		t.Errorf("Get() returned %s, want %s", string(retrieved), string(data))
	}
}

// TestLocalEngineExists 测试文件存在性检查
func TestLocalEngineExists(t *testing.T) {
	tempDir := t.TempDir()
	engine, _ := NewLocalEngine(tempDir)

	hash := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	data := []byte("test data")

	// 文件不存在
	exists, err := engine.Exists(hash)
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if exists {
		t.Error("Exists() returned true for non-existent file")
	}

	// 存储文件
	engine.Put(hash, bytes.NewReader(data))

	// 文件存在
	exists, err = engine.Exists(hash)
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if !exists {
		t.Error("Exists() returned false for existing file")
	}
}

// TestLocalEngineDelete 测试文件删除
func TestLocalEngineDelete(t *testing.T) {
	tempDir := t.TempDir()
	engine, _ := NewLocalEngine(tempDir)

	hash := "fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210"
	data := []byte("delete test")

	// 存储文件
	engine.Put(hash, bytes.NewReader(data))

	// 验证文件存在
	exists, _ := engine.Exists(hash)
	if !exists {
		t.Fatal("File should exist before deletion")
	}

	// 删除文件
	err := engine.Delete(hash)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// 验证文件已删除
	exists, _ = engine.Exists(hash)
	if exists {
		t.Error("File should not exist after deletion")
	}
}

// TestLocalEngineStat 测试文件信息获取
func TestLocalEngineStat(t *testing.T) {
	tempDir := t.TempDir()
	engine, _ := NewLocalEngine(tempDir)

	hash := "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
	data := []byte("stat test data")

	// 存储文件
	engine.Put(hash, bytes.NewReader(data))

	// 获取文件信息
	info, err := engine.Stat(hash)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}

	if info.Hash != hash {
		t.Errorf("Stat() hash = %v, want %v", info.Hash, hash)
	}

	if info.Size != int64(len(data)) {
		t.Errorf("Stat() size = %v, want %v", info.Size, len(data))
	}

	expectedPath := "ab/cd/abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
	if info.StorePath != expectedPath {
		t.Errorf("Stat() path = %v, want %v", info.StorePath, expectedPath)
	}
}

// TestLocalEnginePutDuplicate 测试重复存储
func TestLocalEnginePutDuplicate(t *testing.T) {
	tempDir := t.TempDir()
	engine, _ := NewLocalEngine(tempDir)

	hash := "d0d0d0d0d0d0d0d0d0d0d0d0d0d0d0d0d0d0d0d0d0d0d0d0d0d0d0d0d0d0d0d0"
	data := []byte("duplicate test")

	// 第一次存储
	err := engine.Put(hash, bytes.NewReader(data))
	if err != nil {
		t.Fatalf("First Put() error = %v", err)
	}

	// 第二次存储（应该失败）
	err = engine.Put(hash, bytes.NewReader(data))
	if err == nil {
		t.Error("Put() should fail for duplicate file")
	}
}

// TestLocalEngineInvalidHash 测试无效哈希
func TestLocalEngineInvalidHash(t *testing.T) {
	tempDir := t.TempDir()
	engine, _ := NewLocalEngine(tempDir)

	tests := []struct {
		name string
		hash string
	}{
		{"too short", "abc"},
		{"too long", "abc1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"},
		{"invalid chars", "xyz1234567890abcdef1234567890abcdef1234567890abcdef1234567890"},
		{"uppercase", "AABBCCDDEEFF1122334455667788990011223344556677889900AABBCCDDEE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := []byte("test")
			err := engine.Put(tt.hash, bytes.NewReader(data))
			if err == nil {
				t.Error("Put() should fail with invalid hash")
			}
		})
	}
}

// TestLocalEngineLargeFile 测试大文件存储
func TestLocalEngineLargeFile(t *testing.T) {
	tempDir := t.TempDir()
	engine, _ := NewLocalEngine(tempDir)

	hash := "1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a"

	// 创建 10MB 的测试数据
	data := make([]byte, 10*1024*1024)
	rand.Read(data)

	// 存储
	err := engine.Put(hash, bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Put() error = %v", err)
	}

	// 读取
	reader, err := engine.Get(hash)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	defer reader.Close()

	retrieved, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	// 验证大小
	if len(retrieved) != len(data) {
		t.Errorf("Retrieved file size = %d, want %d", len(retrieved), len(data))
	}

	// 验证内容（只比较前 1MB）
	if !bytes.Equal(data[:1024*1024], retrieved[:1024*1024]) {
		t.Error("Retrieved file content does not match original")
	}
}

// TestLocalEngineDirectoryStructure 测试目录结构
func TestLocalEngineDirectoryStructure(t *testing.T) {
	tempDir := t.TempDir()
	engine, _ := NewLocalEngine(tempDir)

	hash := "aabbccddeeff1122334455667788990011223344556677889900aabbccddeeff"
	data := []byte("directory test")

	engine.Put(hash, bytes.NewReader(data))

	// 验证目录结构
	expectedPath := filepath.Join(tempDir, "aa", "bb", hash)
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("File not found at expected path: %s", expectedPath)
	}

	// 验证一级目录存在
	dir1 := filepath.Join(tempDir, "aa")
	if _, err := os.Stat(dir1); os.IsNotExist(err) {
		t.Error("First level directory not created")
	}

	// 验证二级目录存在
	dir2 := filepath.Join(tempDir, "aa", "bb")
	if _, err := os.Stat(dir2); os.IsNotExist(err) {
		t.Error("Second level directory not created")
	}
}

// TestValidateHash 测试哈希验证
func TestValidateHash(t *testing.T) {
	tests := []struct {
		name    string
		hash    string
		wantErr bool
	}{
		{
			name:    "valid hash",
			hash:    "aabbccddeeff1122334455667788990011223344556677889900aabbccddeeff",
			wantErr: false,
		},
		{
			name:    "too short",
			hash:    "aabbcc",
			wantErr: true,
		},
		{
			name:    "too long",
			hash:    "aabbccddeeff1122334455667788990011223344556677889900aabbccddeeff00",
			wantErr: true,
		},
		{
			name:    "invalid character",
			hash:    "ggbbccddeeff1122334455667788990011223344556677889900aabbccddeeff",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHash(tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateHash() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestGeneratePath 测试路径生成
func TestGeneratePath(t *testing.T) {
	tests := []struct {
		name     string
		hash     string
		expected string
		wantErr  bool
	}{
		{
			name:     "valid hash",
			hash:     "aabbccddeeff1122334455667788990011223344556677889900aabbccddeeff",
			expected: "aa/bb/aabbccddeeff1122334455667788990011223344556677889900aabbccddeeff",
			wantErr:  false,
		},
		{
			name:     "another valid hash",
			hash:     "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			expected: "12/34/1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GeneratePath(tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("GeneratePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("GeneratePath() = %v, want %v", got, tt.expected)
			}
		})
	}
}
