package crypto

import (
	"bytes"
	"crypto/rand"
	"testing"
)

// TestCalculateSHA256 测试哈希计算
func TestCalculateSHA256(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected string
	}{
		{
			name:     "empty data",
			data:     []byte(""),
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "simple text",
			data:     []byte("hello"),
			expected: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
		},
		{
			name:     "AhaVault",
			data:     []byte("AhaVault"),
			expected: "4a8ab8bc67ca7f7318907f7f0e974c3f2b2b9039a76e055168348f61056d4bc2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateSHA256(tt.data)
			if got != tt.expected {
				t.Errorf("CalculateSHA256() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestCalculateSHA256Stream 测试流式哈希计算
func TestCalculateSHA256Stream(t *testing.T) {
	data := []byte("hello world")
	reader := bytes.NewReader(data)

	hash, err := CalculateSHA256Stream(reader)
	if err != nil {
		t.Fatalf("CalculateSHA256Stream() error = %v", err)
	}

	// 验证流式计算的结果与直接计算相同
	expectedHash := CalculateSHA256(data)
	if hash != expectedHash {
		t.Errorf("CalculateSHA256Stream() = %v, want %v", hash, expectedHash)
	}
}

// TestCalculateSHA256StreamLargeData 测试大文件流式哈希
func TestCalculateSHA256StreamLargeData(t *testing.T) {
	// 创建 10MB 的测试数据
	data := make([]byte, 10*1024*1024)
	rand.Read(data)

	// 直接计算
	expectedHash := CalculateSHA256(data)

	// 流式计算
	reader := bytes.NewReader(data)
	hash, err := CalculateSHA256Stream(reader)
	if err != nil {
		t.Fatalf("CalculateSHA256Stream() error = %v", err)
	}

	if hash != expectedHash {
		t.Error("CalculateSHA256Stream() result != direct calculation for large data")
	}
}

// TestVerifySHA256 测试哈希验证
func TestVerifySHA256(t *testing.T) {
	data := []byte("test data")
	correctHash := CalculateSHA256(data)
	wrongHash := "0000000000000000000000000000000000000000000000000000000000000000"

	// 正确的哈希
	if !VerifySHA256(data, correctHash) {
		t.Error("VerifySHA256() should return true for correct hash")
	}

	// 错误的哈希
	if VerifySHA256(data, wrongHash) {
		t.Error("VerifySHA256() should return false for wrong hash")
	}
}

// TestVerifySHA256Stream 测试流式哈希验证
func TestVerifySHA256Stream(t *testing.T) {
	data := []byte("test stream data")
	correctHash := CalculateSHA256(data)
	wrongHash := "1111111111111111111111111111111111111111111111111111111111111111"

	// 正确的哈希
	reader := bytes.NewReader(data)
	valid, err := VerifySHA256Stream(reader, correctHash)
	if err != nil {
		t.Fatalf("VerifySHA256Stream() error = %v", err)
	}
	if !valid {
		t.Error("VerifySHA256Stream() should return true for correct hash")
	}

	// 错误的哈希
	reader = bytes.NewReader(data)
	valid, err = VerifySHA256Stream(reader, wrongHash)
	if err != nil {
		t.Fatalf("VerifySHA256Stream() error = %v", err)
	}
	if valid {
		t.Error("VerifySHA256Stream() should return false for wrong hash")
	}
}

// TestHashConsistency 测试哈希一致性
func TestHashConsistency(t *testing.T) {
	data := []byte("consistency test")

	// 多次计算应该得到相同结果
	hash1 := CalculateSHA256(data)
	hash2 := CalculateSHA256(data)
	hash3 := CalculateSHA256(data)

	if hash1 != hash2 || hash2 != hash3 {
		t.Error("CalculateSHA256() should return consistent results")
	}
}

// BenchmarkCalculateSHA256 基准测试直接哈希计算
func BenchmarkCalculateSHA256(b *testing.B) {
	// 1MB 测试数据
	data := make([]byte, 1024*1024)
	rand.Read(data)

	b.ResetTimer()
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		_ = CalculateSHA256(data)
	}
}

// BenchmarkCalculateSHA256Stream 基准测试流式哈希计算
func BenchmarkCalculateSHA256Stream(b *testing.B) {
	// 1MB 测试数据
	data := make([]byte, 1024*1024)
	rand.Read(data)

	b.ResetTimer()
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(data)
		_, _ = CalculateSHA256Stream(reader)
	}
}
