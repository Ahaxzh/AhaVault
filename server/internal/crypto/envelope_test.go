package crypto

import (
	"bytes"
	"crypto/rand"
	"testing"
)

// TestGenerateDEK 测试 DEK 生成
func TestGenerateDEK(t *testing.T) {
	dek, err := GenerateDEK()
	if err != nil {
		t.Fatalf("GenerateDEK() error = %v", err)
	}

	if len(dek) != 32 {
		t.Errorf("GenerateDEK() length = %d, want 32", len(dek))
	}

	// 测试生成的 DEK 是随机的
	dek2, err := GenerateDEK()
	if err != nil {
		t.Fatalf("GenerateDEK() error = %v", err)
	}

	if bytes.Equal(dek, dek2) {
		t.Error("GenerateDEK() generated identical keys, should be random")
	}
}

// TestEncryptDecryptDEK 测试 DEK 加密和解密
func TestEncryptDecryptDEK(t *testing.T) {
	// 生成测试密钥
	kek := make([]byte, 32)
	rand.Read(kek)

	dek := make([]byte, 32)
	rand.Read(dek)

	// 加密
	encryptedDEK, err := EncryptDEK(dek, kek)
	if err != nil {
		t.Fatalf("EncryptDEK() error = %v", err)
	}

	// 验证密文长度 (12 nonce + 32 data + 16 tag = 60)
	if len(encryptedDEK) != 60 {
		t.Errorf("EncryptDEK() length = %d, want 60", len(encryptedDEK))
	}

	// 解密
	decryptedDEK, err := DecryptDEK(encryptedDEK, kek)
	if err != nil {
		t.Fatalf("DecryptDEK() error = %v", err)
	}

	// 验证解密后的 DEK 与原始 DEK 相同
	if !bytes.Equal(dek, decryptedDEK) {
		t.Error("DecryptDEK() result != original DEK")
	}
}

// TestEncryptDEKInvalidKEKLength 测试无效的 KEK 长度
func TestEncryptDEKInvalidKEKLength(t *testing.T) {
	dek := make([]byte, 32)
	rand.Read(dek)

	invalidKEK := make([]byte, 16) // 错误长度

	_, err := EncryptDEK(dek, invalidKEK)
	if err == nil {
		t.Error("EncryptDEK() should fail with invalid KEK length")
	}
}

// TestDecryptDEKWithWrongKEK 测试使用错误的 KEK 解密
func TestDecryptDEKWithWrongKEK(t *testing.T) {
	// 生成两个不同的 KEK
	kek1 := make([]byte, 32)
	rand.Read(kek1)

	kek2 := make([]byte, 32)
	rand.Read(kek2)

	dek := make([]byte, 32)
	rand.Read(dek)

	// 用 kek1 加密
	encryptedDEK, err := EncryptDEK(dek, kek1)
	if err != nil {
		t.Fatalf("EncryptDEK() error = %v", err)
	}

	// 用 kek2 解密（应该失败）
	_, err = DecryptDEK(encryptedDEK, kek2)
	if err == nil {
		t.Error("DecryptDEK() should fail with wrong KEK")
	}
}

// TestEncryptDEKToBase64 测试 Base64 编码
func TestEncryptDEKToBase64(t *testing.T) {
	kek := make([]byte, 32)
	rand.Read(kek)

	dek := make([]byte, 32)
	rand.Read(dek)

	// 加密并编码
	base64Str, err := EncryptDEKToBase64(dek, kek)
	if err != nil {
		t.Fatalf("EncryptDEKToBase64() error = %v", err)
	}

	if len(base64Str) == 0 {
		t.Error("EncryptDEKToBase64() returned empty string")
	}

	// 解码并解密
	decryptedDEK, err := DecryptDEKFromBase64(base64Str, kek)
	if err != nil {
		t.Fatalf("DecryptDEKFromBase64() error = %v", err)
	}

	if !bytes.Equal(dek, decryptedDEK) {
		t.Error("DecryptDEKFromBase64() result != original DEK")
	}
}

// TestEncryptDecryptFile 测试文件加密和解密
func TestEncryptDecryptFile(t *testing.T) {
	dek := make([]byte, 32)
	rand.Read(dek)

	plaintext := []byte("Hello, AhaVault! This is a secret message.")

	// 加密
	ciphertext, err := EncryptFile(plaintext, dek)
	if err != nil {
		t.Fatalf("EncryptFile() error = %v", err)
	}

	// 验证密文与明文不同
	if bytes.Equal(plaintext, ciphertext) {
		t.Error("EncryptFile() ciphertext == plaintext")
	}

	// 解密
	decrypted, err := DecryptFile(ciphertext, dek)
	if err != nil {
		t.Fatalf("DecryptFile() error = %v", err)
	}

	// 验证解密后与原文相同
	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("DecryptFile() = %s, want %s", string(decrypted), string(plaintext))
	}
}

// TestEncryptFileLargeData 测试大文件加密
func TestEncryptFileLargeData(t *testing.T) {
	dek := make([]byte, 32)
	rand.Read(dek)

	// 创建 10MB 的测试数据
	plaintext := make([]byte, 10*1024*1024)
	rand.Read(plaintext)

	// 加密
	ciphertext, err := EncryptFile(plaintext, dek)
	if err != nil {
		t.Fatalf("EncryptFile() error = %v", err)
	}

	// 解密
	decrypted, err := DecryptFile(ciphertext, dek)
	if err != nil {
		t.Fatalf("DecryptFile() error = %v", err)
	}

	// 验证
	if !bytes.Equal(plaintext, decrypted) {
		t.Error("DecryptFile() result != original plaintext for large file")
	}
}

// TestEncryptDecryptStream 测试流式加密
func TestEncryptDecryptStream(t *testing.T) {
	dek := make([]byte, 32)
	rand.Read(dek)

	plaintext := []byte("This is a test message for stream encryption.")
	reader := bytes.NewReader(plaintext)

	// 加密
	var encryptedBuf bytes.Buffer
	err := EncryptStream(reader, &encryptedBuf, dek)
	if err != nil {
		t.Fatalf("EncryptStream() error = %v", err)
	}

	// 解密
	var decryptedBuf bytes.Buffer
	err = DecryptStream(&encryptedBuf, &decryptedBuf, dek)
	if err != nil {
		t.Fatalf("DecryptStream() error = %v", err)
	}

	// 验证
	if !bytes.Equal(plaintext, decryptedBuf.Bytes()) {
		t.Errorf("DecryptStream() = %s, want %s", decryptedBuf.String(), string(plaintext))
	}
}

// TestEncryptStreamLargeData 测试大文件流式加密
func TestEncryptStreamLargeData(t *testing.T) {
	dek := make([]byte, 32)
	rand.Read(dek)

	// 创建 50MB 的测试数据
	plaintext := make([]byte, 50*1024*1024)
	rand.Read(plaintext)
	reader := bytes.NewReader(plaintext)

	// 加密
	var encryptedBuf bytes.Buffer
	err := EncryptStream(reader, &encryptedBuf, dek)
	if err != nil {
		t.Fatalf("EncryptStream() error = %v", err)
	}

	// 解密
	var decryptedBuf bytes.Buffer
	err = DecryptStream(&encryptedBuf, &decryptedBuf, dek)
	if err != nil {
		t.Fatalf("DecryptStream() error = %v", err)
	}

	// 验证（只比较前 1MB 以节省时间）
	if !bytes.Equal(plaintext[:1024*1024], decryptedBuf.Bytes()[:1024*1024]) {
		t.Error("DecryptStream() result != original plaintext for large stream")
	}
}

// TestZeroBytes 测试字节清零
func TestZeroBytes(t *testing.T) {
	data := []byte("sensitive data")
	ZeroBytes(data)

	// 验证所有字节都为 0
	for i, b := range data {
		if b != 0 {
			t.Errorf("ZeroBytes() failed at index %d: got %d, want 0", i, b)
		}
	}
}

// BenchmarkEncryptDEK 基准测试 DEK 加密
func BenchmarkEncryptDEK(b *testing.B) {
	kek := make([]byte, 32)
	rand.Read(kek)

	dek := make([]byte, 32)
	rand.Read(dek)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = EncryptDEK(dek, kek)
	}
}

// BenchmarkDecryptDEK 基准测试 DEK 解密
func BenchmarkDecryptDEK(b *testing.B) {
	kek := make([]byte, 32)
	rand.Read(kek)

	dek := make([]byte, 32)
	rand.Read(dek)

	encryptedDEK, _ := EncryptDEK(dek, kek)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = DecryptDEK(encryptedDEK, kek)
	}
}

// BenchmarkEncryptFile 基准测试文件加密
func BenchmarkEncryptFile(b *testing.B) {
	dek := make([]byte, 32)
	rand.Read(dek)

	// 1MB 测试数据
	plaintext := make([]byte, 1024*1024)
	rand.Read(plaintext)

	b.ResetTimer()
	b.SetBytes(int64(len(plaintext)))
	for i := 0; i < b.N; i++ {
		_, _ = EncryptFile(plaintext, dek)
	}
}

// BenchmarkDecryptFile 基准测试文件解密
func BenchmarkDecryptFile(b *testing.B) {
	dek := make([]byte, 32)
	rand.Read(dek)

	plaintext := make([]byte, 1024*1024)
	rand.Read(plaintext)

	ciphertext, _ := EncryptFile(plaintext, dek)

	b.ResetTimer()
	b.SetBytes(int64(len(plaintext)))
	for i := 0; i < b.N; i++ {
		_, _ = DecryptFile(ciphertext, dek)
	}
}
