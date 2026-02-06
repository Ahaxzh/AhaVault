package services

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"ahavault/server/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService 用户服务
type UserService struct {
	db        *gorm.DB
	jwtSecret []byte
}

// NewUserService 创建用户服务实例
func NewUserService(db *gorm.DB, jwtSecret string) *UserService {
	return &UserService{
		db:        db,
		jwtSecret: []byte(jwtSecret),
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Email    string
	Password string
}

// LoginRequest 登录请求
type LoginRequest struct {
	Email    string
	Password string
}

// AuthResponse 认证响应
type AuthResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

// Register 用户注册
func (s *UserService) Register(req *RegisterRequest, requireInviteCode bool, validInviteCode string) (*AuthResponse, error) {
	// 验证邮箱格式
	if err := validateEmail(req.Email); err != nil {
		return nil, err
	}

	// 验证密码强度
	if err := validatePassword(req.Password); err != nil {
		return nil, err
	}

	// 检查邮箱是否已存在
	var count int64
	if err := s.db.Model(&models.User{}).Where("email = ?", req.Email).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if count > 0 {
		return nil, errors.New("email already exists")
	}

	// 哈希密码
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Determine role (First user is Admin)
	var totalUsers int64
	if err := s.db.Model(&models.User{}).Count(&totalUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	role := models.RoleUser
	if totalUsers == 0 {
		role = models.RoleAdmin
	}

	// Create User
	user := &models.User{
		Email:        req.Email,
		Password:     string(passwordHash),
		Role:         role,
		Status:       models.StatusActive,
		StorageQuota: 10 * 1024 * 1024 * 1024, // Default 10GB
		StorageUsed:  0,
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 生成 token
	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	// 隐藏密码
	user.Password = ""

	return &AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

// Login 用户登录
func (s *UserService) Login(req *LoginRequest) (*AuthResponse, error) {
	// 查询用户
	var user models.User
	err := s.db.Where("email = ?", req.Email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// 检查账户状态
	if !user.IsActive() {
		return nil, errors.New("account is disabled")
	}

	// 生成 token
	token, err := s.GenerateToken(&user)
	if err != nil {
		return nil, err
	}

	// 隐藏密码
	user.Password = ""

	return &AuthResponse{
		Token: token,
		User:  &user,
	}, nil
}

// GenerateToken 生成 JWT token
func (s *UserService) GenerateToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID.String(),
		"email":    user.Email,
		"is_admin": user.IsAdmin(),
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// ValidateToken 验证 JWT token
func (s *UserService) ValidateToken(tokenString string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &claims, nil
	}

	return nil, errors.New("invalid token")
}

// GetUserByID 通过 ID 获取用户
func (s *UserService) GetUserByID(userID string) (*models.User, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	var user models.User
	if err := s.db.First(&user, userUUID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// validateEmail 验证邮箱格式
func validateEmail(email string) error {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`, email)
	if !matched {
		return errors.New("invalid email format")
	}
	return nil
}

// validatePassword 验证密码强度
func validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)

	if !hasLetter || !hasDigit {
		return errors.New("password must contain both letters and digits")
	}

	return nil
}
