package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// EncryptDEK 使用 KEK 加密 DEK (信封加密)
// KEK: Key Encryption Key (主密钥，从环境变量获取)
// DEK: Data Encryption Key (文件专属密钥，随机生成)
// 返回格式: [Nonce(12) + Ciphertext(32) + AuthTag(16)] = 60 bytes
func EncryptDEK(dek []byte, kek []byte) ([]byte, error) {
	// 验证密钥长度
	if len(kek) != 32 {
		return nil, fmt.Errorf("KEK must be 32 bytes, got %d bytes", len(kek))
	}
	if len(dek) != 32 {
		return nil, fmt.Errorf("DEK must be 32 bytes, got %d bytes", len(dek))
	}

	// 创建 AES 加密块
	block, err := aes.NewCipher(kek)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// 创建 GCM 模式
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// 生成随机 Nonce
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// 加密 DEK
	// Seal 会将 nonce 追加到密文前面，并在末尾添加认证标签
	ciphertext := aead.Seal(nonce, nonce, dek, nil)

	return ciphertext, nil
}

// DecryptDEK 使用 KEK 解密 DEK
// encryptedDEK 格式: [Nonce(12) + Ciphertext(32) + AuthTag(16)]
func DecryptDEK(encryptedDEK []byte, kek []byte) ([]byte, error) {
	// 验证 KEK 长度
	if len(kek) != 32 {
		return nil, fmt.Errorf("KEK must be 32 bytes, got %d bytes", len(kek))
	}

	// 创建 AES 加密块
	block, err := aes.NewCipher(kek)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// 创建 GCM 模式
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// 验证加密数据长度
	nonceSize := aead.NonceSize()
	if len(encryptedDEK) < nonceSize {
		return nil, fmt.Errorf("encrypted DEK too short: %d bytes", len(encryptedDEK))
	}

	// 分离 Nonce 和密文
	nonce := encryptedDEK[:nonceSize]
	ciphertext := encryptedDEK[nonceSize:]

	// 解密并验证
	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt DEK: %w", err)
	}

	return plaintext, nil
}

// GenerateDEK 生成随机的 32 字节 DEK
func GenerateDEK() ([]byte, error) {
	dek := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, dek); err != nil {
		return nil, fmt.Errorf("failed to generate DEK: %w", err)
	}
	return dek, nil
}

// EncryptDEKToBase64 加密 DEK 并返回 Base64 编码字符串（用于数据库存储）
func EncryptDEKToBase64(dek []byte, kek []byte) (string, error) {
	encryptedDEK, err := EncryptDEK(dek, kek)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encryptedDEK), nil
}

// DecryptDEKFromBase64 从 Base64 字符串解密 DEK
func DecryptDEKFromBase64(encryptedDEKBase64 string, kek []byte) ([]byte, error) {
	encryptedDEK, err := base64.StdEncoding.DecodeString(encryptedDEKBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}
	return DecryptDEK(encryptedDEK, kek)
}

// EncryptFile 使用 DEK 加密文件内容
// 返回格式: [Nonce(12) + Ciphertext(N) + AuthTag(16)]
func EncryptFile(plaintext []byte, dek []byte) ([]byte, error) {
	if len(dek) != 32 {
		return nil, fmt.Errorf("DEK must be 32 bytes, got %d bytes", len(dek))
	}

	block, err := aes.NewCipher(dek)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := aead.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// DecryptFile 使用 DEK 解密文件内容
func DecryptFile(ciphertext []byte, dek []byte) ([]byte, error) {
	if len(dek) != 32 {
		return nil, fmt.Errorf("DEK must be 32 bytes, got %d bytes", len(dek))
	}

	block, err := aes.NewCipher(dek)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short: %d bytes", len(ciphertext))
	}

	nonce := ciphertext[:nonceSize]
	encryptedData := ciphertext[nonceSize:]

	plaintext, err := aead.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt file: %w", err)
	}

	return plaintext, nil
}

// EncryptStream 使用 DEK 加密数据流（用于大文件）
func EncryptStream(reader io.Reader, writer io.Writer, dek []byte) error {
	if len(dek) != 32 {
		return fmt.Errorf("DEK must be 32 bytes, got %d bytes", len(dek))
	}

	block, err := aes.NewCipher(dek)
	if err != nil {
		return fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// 生成 IV (16 bytes for CTR mode)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return fmt.Errorf("failed to generate IV: %w", err)
	}

	// 先写入 IV
	if _, err := writer.Write(iv); err != nil {
		return fmt.Errorf("failed to write IV: %w", err)
	}

	// 使用 CTR 模式进行流式加密
	stream := cipher.NewCTR(block, iv)
	streamWriter := &cipher.StreamWriter{S: stream, W: writer}

	// 复制并加密数据
	if _, err := io.Copy(streamWriter, reader); err != nil {
		return fmt.Errorf("failed to encrypt stream: %w", err)
	}

	return nil
}

// DecryptStream 使用 DEK 解密数据流（用于大文件）
func DecryptStream(reader io.Reader, writer io.Writer, dek []byte) error {
	if len(dek) != 32 {
		return fmt.Errorf("DEK must be 32 bytes, got %d bytes", len(dek))
	}

	block, err := aes.NewCipher(dek)
	if err != nil {
		return fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// 读取 IV (16 bytes for CTR mode)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(reader, iv); err != nil {
		return fmt.Errorf("failed to read IV: %w", err)
	}

	// 使用 CTR 模式解密
	stream := cipher.NewCTR(block, iv)
	streamReader := &cipher.StreamReader{S: stream, R: reader}

	// 复制并解密数据
	if _, err := io.Copy(writer, streamReader); err != nil {
		return fmt.Errorf("failed to decrypt stream: %w", err)
	}

	return nil
}

// ZeroBytes 安全地清零字节切片（防止敏感数据留在内存中）
func ZeroBytes(data []byte) {
	for i := range data {
		data[i] = 0
	}
}
