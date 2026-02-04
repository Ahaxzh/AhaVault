package services

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"gorm.io/gorm"
)

// PickupCodeGenerator 取件码生成器
type PickupCodeGenerator struct {
	length  int
	charset string
}

// NewPickupCodeGenerator 创建取件码生成器
// 默认生成 8 位取件码，字符集为 2-9 和 A-Z（排除 0/O/1/I 防止混淆）
func NewPickupCodeGenerator(length int) *PickupCodeGenerator {
	// 字符集：2-9 (8个数字) + A-Z 排除 O 和 I (24个字母) = 32个字符
	// 总组合数：32^8 = 1,099,511,627,776 (约1.1万亿)
	charset := "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"

	return &PickupCodeGenerator{
		length:  length,
		charset: charset,
	}
}

// Generate 生成随机取件码
func (g *PickupCodeGenerator) Generate() (string, error) {
	code := make([]byte, g.length)
	charsetLen := big.NewInt(int64(len(g.charset)))

	for i := 0; i < g.length; i++ {
		// 使用加密安全的随机数生成器
		randomIndex, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", fmt.Errorf("failed to generate random index: %w", err)
		}
		code[i] = g.charset[randomIndex.Int64()]
	}

	return string(code), nil
}

// GenerateUnique 生成唯一的取件码（检查数据库防止碰撞）
func (g *PickupCodeGenerator) GenerateUnique(db *gorm.DB) (string, error) {
	maxAttempts := 10 // 最多尝试10次

	for attempt := 0; attempt < maxAttempts; attempt++ {
		code, err := g.Generate()
		if err != nil {
			return "", err
		}

		// 检查数据库中是否已存在
		var count int64
		err = db.Table("share_sessions").Where("pickup_code = ?", code).Count(&count).Error
		if err != nil {
			return "", fmt.Errorf("failed to check code uniqueness: %w", err)
		}

		if count == 0 {
			return code, nil
		}

		// 碰撞，重试
	}

	return "", fmt.Errorf("failed to generate unique code after %d attempts", maxAttempts)
}

// ValidatePickupCode 验证取件码格式
func ValidatePickupCode(code string, expectedLength int) error {
	if len(code) != expectedLength {
		return fmt.Errorf("invalid code length: expected %d, got %d", expectedLength, len(code))
	}

	// 验证字符集
	validChars := "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"
	for _, char := range code {
		valid := false
		for _, validChar := range validChars {
			if char == validChar {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid character in code: %c", char)
		}
	}

	return nil
}

// DefaultPickupCodeGenerator 默认的取件码生成器（8位）
var DefaultPickupCodeGenerator = NewPickupCodeGenerator(8)
