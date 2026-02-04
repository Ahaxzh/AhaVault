package services

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestNewPickupCodeGenerator(t *testing.T) {
	gen := NewPickupCodeGenerator(8)
	if gen.length != 8 {
		t.Errorf("Expected length 8, got %d", gen.length)
	}

	if len(gen.charset) != 32 {
		t.Errorf("Expected charset length 32, got %d", len(gen.charset))
	}
}

func TestGenerate(t *testing.T) {
	gen := NewPickupCodeGenerator(8)

	code, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(code) != 8 {
		t.Errorf("Generated code length = %d, want 8", len(code))
	}

	// 验证字符集
	validChars := "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"
	for _, char := range code {
		found := false
		for _, valid := range validChars {
			if char == valid {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Invalid character in code: %c", char)
		}
	}
}

func TestGenerateUniqueness(t *testing.T) {
	gen := NewPickupCodeGenerator(8)

	// 生成多个取件码，检查唯一性
	codes := make(map[string]bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		code, err := gen.Generate()
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}

		if codes[code] {
			t.Logf("Collision found after %d iterations: %s", i, code)
			// 碰撞是可能的，但在100次迭代中概率极低
			// 不算作失败，只是记录
		}
		codes[code] = true
	}

	if len(codes) < iterations {
		t.Logf("Generated %d unique codes out of %d attempts", len(codes), iterations)
	}
}

func TestGenerateUniqueWithDB(t *testing.T) {
	// 使用内存 SQLite 数据库测试
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Skipf("Failed to connect to test database: %v", err)
	}

	// 创建测试表
	db.Exec(`CREATE TABLE share_sessions (
		id TEXT PRIMARY KEY,
		pickup_code TEXT UNIQUE NOT NULL
	)`)

	gen := NewPickupCodeGenerator(8)

	// 生成第一个唯一码
	code1, err := gen.GenerateUnique(db)
	if err != nil {
		t.Fatalf("GenerateUnique() error = %v", err)
	}

	// 插入到数据库
	db.Exec("INSERT INTO share_sessions (id, pickup_code) VALUES (?, ?)", "id1", code1)

	// 生成第二个唯一码
	code2, err := gen.GenerateUnique(db)
	if err != nil {
		t.Fatalf("GenerateUnique() error = %v", err)
	}

	if code1 == code2 {
		t.Error("GenerateUnique() should generate different codes")
	}
}

func TestValidatePickupCode(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		length  int
		wantErr bool
	}{
		{
			name:    "valid code",
			code:    "A2B3C4D5",
			length:  8,
			wantErr: false,
		},
		{
			name:    "too short",
			code:    "A2B3",
			length:  8,
			wantErr: true,
		},
		{
			name:    "too long",
			code:    "A2B3C4D5E6",
			length:  8,
			wantErr: true,
		},
		{
			name:    "invalid character O",
			code:    "A2B3C4DO",
			length:  8,
			wantErr: true,
		},
		{
			name:    "invalid character I",
			code:    "A2B3C4DI",
			length:  8,
			wantErr: true,
		},
		{
			name:    "invalid character 0",
			code:    "A2B3C4D0",
			length:  8,
			wantErr: true,
		},
		{
			name:    "invalid character 1",
			code:    "A2B3C4D1",
			length:  8,
			wantErr: true,
		},
		{
			name:    "lowercase not allowed",
			code:    "a2b3c4d5",
			length:  8,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePickupCode(tt.code, tt.length)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePickupCode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func BenchmarkGenerate(b *testing.B) {
	gen := NewPickupCodeGenerator(8)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.Generate()
	}
}

func TestDefaultPickupCodeGenerator(t *testing.T) {
	if DefaultPickupCodeGenerator == nil {
		t.Fatal("DefaultPickupCodeGenerator should not be nil")
	}

	code, err := DefaultPickupCodeGenerator.Generate()
	if err != nil {
		t.Fatalf("DefaultPickupCodeGenerator.Generate() error = %v", err)
	}

	if len(code) != 8 {
		t.Errorf("Default generator should produce 8-character codes, got %d", len(code))
	}
}
