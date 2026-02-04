package storage

import (
	"fmt"
	"io"
)

// Engine 存储引擎接口
type Engine interface {
	// Put 存储文件，hash 为 SHA-256 哈希值
	Put(hash string, reader io.Reader) error

	// Get 读取文件，返回可读流
	Get(hash string) (io.ReadCloser, error)

	// Delete 删除文件
	Delete(hash string) error

	// Exists 检查文件是否存在
	Exists(hash string) (bool, error)

	// Stat 获取文件信息
	Stat(hash string) (*FileInfo, error)
}

// FileInfo 文件信息
type FileInfo struct {
	Hash      string // SHA-256 哈希值
	Size      int64  // 文件大小（字节）
	StorePath string // 存储路径
}

// ValidateHash 验证哈希值格式
func ValidateHash(hash string) error {
	if len(hash) != 64 {
		return fmt.Errorf("invalid hash length: expected 64, got %d", len(hash))
	}

	// 检查是否为有效的十六进制字符串
	for _, c := range hash {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return fmt.Errorf("invalid hash character: %c (must be 0-9 or a-f)", c)
		}
	}

	return nil
}

// GeneratePath 生成两级目录路径
// 例如: hash = "aabbccdd..." -> path = "aa/bb/aabbccdd..."
func GeneratePath(hash string) (string, error) {
	if err := ValidateHash(hash); err != nil {
		return "", err
	}

	// 两级目录分片
	return fmt.Sprintf("%s/%s/%s", hash[:2], hash[2:4], hash), nil
}
