// Package services 提供业务逻辑服务层
//
// 本文件为 UserService 的单元测试，覆盖以下功能：
//   - 用户注册（Register）
//   - 用户登录（Login）
//   - JWT Token 生成与验证（GenerateToken, ValidateToken）
//   - 获取用户信息（GetUserByID）
//   - 邮箱格式验证
//   - 密码强度验证
//
// 作者: AhaVault Team
// 创建时间: 2026-02-04
package services

import (
	"testing"

	"ahavault/server/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// setupUserTestEnv 创建用户测试环境
func setupUserTestEnv(t *testing.T) (*UserService, *gorm.DB) {
	db := setupTestDB(t)
	jwtSecret := "test-jwt-secret-key-for-testing-only"
	userService := NewUserService(db, jwtSecret)
	return userService, db
}

// TestRegister 测试用户注册功能
//
// 测试场景：
//  1. 正常注册 - 有效邮箱和密码
//  2. 邮箱已存在 - 返回错误
//  3. 无效邮箱格式
//  4. 密码太短（< 8 字符）
//  5. 密码缺少字母
//  6. 密码缺少数字
func TestRegister(t *testing.T) {
	userService, db := setupUserTestEnv(t)

	// 预先创建一个用户
	existingUser := &models.User{
		Email:        "existing@example.com",
		Password:     "hashedpassword",
		Role:         models.RoleUser,
		Status:       models.StatusActive,
		StorageQuota: 10 * 1024 * 1024 * 1024,
		StorageUsed:  0,
	}
	require.NoError(t, db.Create(existingUser).Error)

	tests := []struct {
		name               string
		req                *RegisterRequest
		requireInviteCode  bool
		validInviteCode    string
		wantErr            bool
		errContains        string
		checkPasswordHash  bool
		checkToken         bool
	}{
		{
			name: "正常注册 - 有效邮箱和密码",
			req: &RegisterRequest{
				Email:    "newuser@example.com",
				Password: "password123",
			},
			requireInviteCode: false,
			validInviteCode:   "",
			wantErr:           false,
			checkPasswordHash: true,
			checkToken:        true,
		},
		{
			name: "邮箱已存在",
			req: &RegisterRequest{
				Email:    "existing@example.com",
				Password: "password123",
			},
			requireInviteCode: false,
			validInviteCode:   "",
			wantErr:           true,
			errContains:       "already exists",
		},
		{
			name: "无效邮箱格式 - 缺少@",
			req: &RegisterRequest{
				Email:    "invalidemail",
				Password: "password123",
			},
			requireInviteCode: false,
			validInviteCode:   "",
			wantErr:           true,
			errContains:       "invalid email",
		},
		{
			name: "无效邮箱格式 - 缺少域名",
			req: &RegisterRequest{
				Email:    "user@",
				Password: "password123",
			},
			requireInviteCode: false,
			validInviteCode:   "",
			wantErr:           true,
			errContains:       "invalid email",
		},
		{
			name: "密码太短",
			req: &RegisterRequest{
				Email:    "shortpw@example.com",
				Password: "pass1",
			},
			requireInviteCode: false,
			validInviteCode:   "",
			wantErr:           true,
			errContains:       "at least 8 characters",
		},
		{
			name: "密码缺少字母",
			req: &RegisterRequest{
				Email:    "noletters@example.com",
				Password: "12345678",
			},
			requireInviteCode: false,
			validInviteCode:   "",
			wantErr:           true,
			errContains:       "letters and digits",
		},
		{
			name: "密码缺少数字",
			req: &RegisterRequest{
				Email:    "nodigits@example.com",
				Password: "passwordonly",
			},
			requireInviteCode: false,
			validInviteCode:   "",
			wantErr:           true,
			errContains:       "letters and digits",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := userService.Register(tt.req, tt.requireInviteCode, tt.validInviteCode)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.User)

			// 验证用户信息
			assert.Equal(t, tt.req.Email, resp.User.Email)
			assert.Empty(t, resp.User.Password) // 密码应该被隐藏

			// 验证密码哈希
			if tt.checkPasswordHash {
				var user models.User
				err := db.Where("email = ?", tt.req.Email).First(&user).Error
				require.NoError(t, err)

				err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(tt.req.Password))
				assert.NoError(t, err)
			}

			// 验证 Token
			if tt.checkToken {
				assert.NotEmpty(t, resp.Token)

				claims, err := userService.ValidateToken(resp.Token)
				require.NoError(t, err)
				assert.Equal(t, resp.User.ID.String(), (*claims)["user_id"])
				assert.Equal(t, tt.req.Email, (*claims)["email"])
			}
		})
	}
}

// TestLogin 测试用户登录功能
//
// 测试场景：
//  1. 正常登录 - 正确的邮箱和密码
//  2. 错误的密码
//  3. 用户不存在
//  4. 账户被禁用
func TestLogin(t *testing.T) {
	userService, db := setupUserTestEnv(t)

	// 创建测试用户
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("correctpassword123"), bcrypt.DefaultCost)
	activeUser := &models.User{
		Email:        "active@example.com",
		Password:     string(passwordHash),
		Role:         models.RoleUser,
		Status:       models.StatusActive,
		StorageQuota: 10 * 1024 * 1024 * 1024,
		StorageUsed:  0,
	}
	require.NoError(t, db.Create(activeUser).Error)

	// 创建被禁用的用户
	disabledUser := &models.User{
		Email:        "disabled@example.com",
		Password:     string(passwordHash),
		Role:         models.RoleUser,
		Status:       models.StatusDisabled,
		StorageQuota: 10 * 1024 * 1024 * 1024,
		StorageUsed:  0,
	}
	require.NoError(t, db.Create(disabledUser).Error)

	tests := []struct {
		name        string
		req         *LoginRequest
		wantErr     bool
		errContains string
		checkToken  bool
	}{
		{
			name: "正常登录 - 正确的邮箱和密码",
			req: &LoginRequest{
				Email:    "active@example.com",
				Password: "correctpassword123",
			},
			wantErr:    false,
			checkToken: true,
		},
		{
			name: "错误的密码",
			req: &LoginRequest{
				Email:    "active@example.com",
				Password: "wrongpassword",
			},
			wantErr:     true,
			errContains: "invalid email or password",
		},
		{
			name: "用户不存在",
			req: &LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			wantErr:     true,
			errContains: "invalid email or password",
		},
		{
			name: "账户被禁用",
			req: &LoginRequest{
				Email:    "disabled@example.com",
				Password: "correctpassword123",
			},
			wantErr:     true,
			errContains: "disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := userService.Login(tt.req)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.User)

			// 验证用户信息
			assert.Equal(t, tt.req.Email, resp.User.Email)
			assert.Empty(t, resp.User.Password) // 密码应该被隐藏

			// 验证 Token
			if tt.checkToken {
				assert.NotEmpty(t, resp.Token)

				claims, err := userService.ValidateToken(resp.Token)
				require.NoError(t, err)
				assert.Equal(t, resp.User.ID.String(), (*claims)["user_id"])
			}
		})
	}
}

// TestGenerateToken 测试 JWT Token 生成
func TestGenerateToken(t *testing.T) {
	userService, _ := setupUserTestEnv(t)

	user := &models.User{
		Email:        "token@example.com",
		Password:     "hashedpassword",
		Role:         models.RoleUser,
		Status:       models.StatusActive,
		StorageQuota: 10 * 1024 * 1024 * 1024,
		StorageUsed:  0,
	}
	user.ID = uuid.New()

	token, err := userService.GenerateToken(user)

	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// 验证 Token 内容
	claims, err := userService.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, user.ID.String(), (*claims)["user_id"])
	assert.Equal(t, user.Email, (*claims)["email"])
	assert.Equal(t, false, (*claims)["is_admin"])
}

// TestGenerateToken_Admin 测试管理员 Token 生成
func TestGenerateToken_Admin(t *testing.T) {
	userService, _ := setupUserTestEnv(t)

	admin := &models.User{
		Email:        "admin@example.com",
		Password:     "hashedpassword",
		Role:         models.RoleAdmin,
		Status:       models.StatusActive,
		StorageQuota: 10 * 1024 * 1024 * 1024,
		StorageUsed:  0,
	}
	admin.ID = uuid.New()

	token, err := userService.GenerateToken(admin)

	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// 验证 Token 内容
	claims, err := userService.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, true, (*claims)["is_admin"])
}

// TestValidateToken 测试 JWT Token 验证
func TestValidateToken(t *testing.T) {
	userService, _ := setupUserTestEnv(t)

	user := &models.User{
		Email:        "validate@example.com",
		Password:     "hashedpassword",
		Role:         models.RoleUser,
		Status:       models.StatusActive,
		StorageQuota: 10 * 1024 * 1024 * 1024,
		StorageUsed:  0,
	}
	user.ID = uuid.New()

	// 生成有效 Token
	validToken, err := userService.GenerateToken(user)
	require.NoError(t, err)

	// 生成无效 Token（使用错误的密钥）
	wrongSecret := []byte("wrong-secret-key")
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	invalidToken, _ := token.SignedString(wrongSecret)

	tests := []struct {
		name        string
		token       string
		wantErr     bool
		errContains string
	}{
		{
			name:    "有效 Token",
			token:   validToken,
			wantErr: false,
		},
		{
			name:        "无效 Token - 错误签名",
			token:       invalidToken,
			wantErr:     true,
			errContains: "signature",
		},
		{
			name:        "无效 Token - 格式错误",
			token:       "invalid.token.format",
			wantErr:     true,
		},
		{
			name:        "空 Token",
			token:       "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := userService.ValidateToken(tt.token)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, claims)
			assert.Equal(t, user.ID.String(), (*claims)["user_id"])
			assert.Equal(t, user.Email, (*claims)["email"])
		})
	}
}

// TestGetUserByID 测试通过 ID 获取用户
func TestGetUserByID(t *testing.T) {
	userService, db := setupUserTestEnv(t)

	// 创建测试用户
	user := &models.User{
		Email:        "getbyid@example.com",
		Password:     "hashedpassword",
		Role:         models.RoleUser,
		Status:       models.StatusActive,
		StorageQuota: 10 * 1024 * 1024 * 1024,
		StorageUsed:  0,
	}
	require.NoError(t, db.Create(user).Error)

	tests := []struct {
		name        string
		userID      string
		wantErr     bool
		errContains string
		checkUser   bool
	}{
		{
			name:      "正常获取用户",
			userID:    user.ID.String(),
			wantErr:   false,
			checkUser: true,
		},
		{
			name:        "用户不存在",
			userID:      uuid.New().String(),
			wantErr:     true,
			errContains: "not found",
		},
		{
			name:        "无效的 UUID",
			userID:      "invalid-uuid",
			wantErr:     true,
			errContains: "invalid user ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetchedUser, err := userService.GetUserByID(tt.userID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, fetchedUser)

			if tt.checkUser {
				assert.Equal(t, user.ID, fetchedUser.ID)
				assert.Equal(t, user.Email, fetchedUser.Email)
			}
		})
	}
}

// TestValidateEmail 测试邮箱格式验证
func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"有效邮箱 - 标准格式", "user@example.com", false},
		{"有效邮箱 - 带点", "user.name@example.com", false},
		{"有效邮箱 - 带加号", "user+tag@example.com", false},
		{"有效邮箱 - 带下划线", "user_name@example.com", false},
		{"有效邮箱 - 带连字符", "user-name@example.com", false},
		{"有效邮箱 - 子域名", "user@mail.example.com", false},
		{"无效邮箱 - 缺少@", "userexample.com", true},
		{"无效邮箱 - 缺少域名", "user@", true},
		{"无效邮箱 - 缺少用户名", "@example.com", true},
		{"无效邮箱 - 缺少顶级域", "user@example", true},
		{"无效邮箱 - 空字符串", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEmail(tt.email)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidatePassword 测试密码强度验证
func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "有效密码 - 字母+数字",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "有效密码 - 混合大小写",
			password: "Password123",
			wantErr:  false,
		},
		{
			name:     "有效密码 - 带特殊字符",
			password: "Pass@word123!",
			wantErr:  false,
		},
		{
			name:        "无效密码 - 太短",
			password:    "pass1",
			wantErr:     true,
			errContains: "at least 8 characters",
		},
		{
			name:        "无效密码 - 只有字母",
			password:    "passwordonly",
			wantErr:     true,
			errContains: "letters and digits",
		},
		{
			name:        "无效密码 - 只有数字",
			password:    "12345678",
			wantErr:     true,
			errContains: "letters and digits",
		},
		{
			name:        "无效密码 - 空字符串",
			password:    "",
			wantErr:     true,
			errContains: "at least 8 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePassword(tt.password)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
