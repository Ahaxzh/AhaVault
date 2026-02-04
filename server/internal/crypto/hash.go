package crypto

import (
	"crypto/sha256"
	"fmt"
	"io"
)

// CalculateSHA256 计算数据的 SHA-256 哈希值
func CalculateSHA256(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

// CalculateSHA256Stream 计算流的 SHA-256 哈希值（用于大文件）
func CalculateSHA256Stream(reader io.Reader) (string, error) {
	hasher := sha256.New()

	// 使用 32KB 缓冲区
	buffer := make([]byte, 32*1024)

	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			hasher.Write(buffer[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read stream: %w", err)
		}
	}

	hash := hasher.Sum(nil)
	return fmt.Sprintf("%x", hash), nil
}

// VerifySHA256 验证数据的 SHA-256 哈希值
func VerifySHA256(data []byte, expectedHash string) bool {
	actualHash := CalculateSHA256(data)
	return actualHash == expectedHash
}

// VerifySHA256Stream 验证流的 SHA-256 哈希值
func VerifySHA256Stream(reader io.Reader, expectedHash string) (bool, error) {
	actualHash, err := CalculateSHA256Stream(reader)
	if err != nil {
		return false, err
	}
	return actualHash == expectedHash, nil
}
