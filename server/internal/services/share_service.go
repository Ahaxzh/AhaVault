package services

import (
	"errors"
	"fmt"
	"time"

	"ahavault/server/internal/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ShareService 分享服务
type ShareService struct {
	db          *gorm.DB
	codeGen     *PickupCodeGenerator
	fileService *FileService
}

// NewShareService 创建分享服务实例
func NewShareService(db *gorm.DB, fileService *FileService) *ShareService {
	return &ShareService{
		db:          db,
		codeGen:     DefaultPickupCodeGenerator,
		fileService: fileService,
	}
}

// CreateShareRequest 创建分享请求
type CreateShareRequest struct {
	FileIDs      []uuid.UUID
	ExpiresIn    time.Duration
	MaxDownloads int
	Password     string
}

// CreateShare 创建分享
func (s *ShareService) CreateShare(userID uuid.UUID, req *CreateShareRequest) (*models.ShareSession, error) {
	if len(req.FileIDs) == 0 {
		return nil, errors.New("no files selected")
	}

	// 验证所有文件属于该用户
	var count int64
	err := s.db.Model(&models.FileMetadata{}).
		Where("id IN ? AND user_id = ? AND deleted_at IS NULL", req.FileIDs, userID).
		Count(&count).Error
	if err != nil {
		return nil, fmt.Errorf("failed to verify files: %w", err)
	}

	if count != int64(len(req.FileIDs)) {
		return nil, errors.New("some files not found or access denied")
	}

	// 生成唯一取件码
	pickupCode, err := s.codeGen.GenerateUnique(s.db)
	if err != nil {
		return nil, fmt.Errorf("failed to generate pickup code: %w", err)
	}

	// 处理密码
	var passwordHash string
	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		passwordHash = string(hash)
	}

	// 计算过期时间
	expiresAt := time.Now().Add(req.ExpiresIn)

	// 开启事务
	tx := s.db.Begin()
	defer tx.Rollback()

	// 创建分享会话
	session := &models.ShareSession{
		PickupCode:       pickupCode,
		CreatorID:        userID,
		PasswordHash:     passwordHash,
		MaxDownloads:     req.MaxDownloads,
		CurrentDownloads: 0,
		ExpiresAt:        expiresAt,
	}

	if err := tx.Create(session).Error; err != nil {
		return nil, fmt.Errorf("failed to create share session: %w", err)
	}

	// 创建文件关联
	for _, fileID := range req.FileIDs {
		shareFile := &models.ShareFile{
			ShareID: session.ID,
			FileID:  fileID,
		}
		if err := tx.Create(shareFile).Error; err != nil {
			return nil, fmt.Errorf("failed to create share file: %w", err)
		}
	}

	tx.Commit()

	return session, nil
}

// GetShareByCode 通过取件码获取分享
func (s *ShareService) GetShareByCode(pickupCode string, password string) (*models.ShareSession, []models.FileMetadata, error) {
	// 验证取件码格式
	if err := ValidatePickupCode(pickupCode, 8); err != nil {
		return nil, nil, err
	}

	// 查询分享会话
	var session models.ShareSession
	err := s.db.Where("pickup_code = ?", pickupCode).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errors.New("invalid pickup code")
		}
		return nil, nil, fmt.Errorf("failed to get share: %w", err)
	}

	// 检查访问权限
	if err := session.CanAccess(); err != nil {
		return nil, nil, err
	}

	// 验证密码
	if session.HasPassword() {
		if password == "" {
			return nil, nil, errors.New("password required")
		}
		err = bcrypt.CompareHashAndPassword([]byte(session.PasswordHash), []byte(password))
		if err != nil {
			return nil, nil, errors.New("invalid password")
		}
	}

	// 获取关联的文件
	var shareFiles []models.ShareFile
	if err := s.db.Where("share_id = ?", session.ID).Find(&shareFiles).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to get share files: %w", err)
	}

	fileIDs := make([]uuid.UUID, len(shareFiles))
	for i, sf := range shareFiles {
		fileIDs[i] = sf.FileID
	}

	var files []models.FileMetadata
	if err := s.db.Where("id IN ? AND deleted_at IS NULL", fileIDs).Find(&files).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to get files: %w", err)
	}

	return &session, files, nil
}

// IncrementDownload 增加下载次数
func (s *ShareService) IncrementDownload(shareID uuid.UUID) error {
	var session models.ShareSession
	if err := s.db.First(&session, shareID).Error; err != nil {
		return fmt.Errorf("failed to get share: %w", err)
	}

	if err := session.IncrementDownloadCount(s.db); err != nil {
		return fmt.Errorf("failed to increment download count: %w", err)
	}

	return nil
}

// StopShare 停止分享
func (s *ShareService) StopShare(shareID uuid.UUID, userID uuid.UUID) error {
	var session models.ShareSession
	err := s.db.Where("id = ? AND creator_id = ?", shareID, userID).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("share not found")
		}
		return fmt.Errorf("failed to get share: %w", err)
	}

	if err := session.Stop(s.db); err != nil {
		return fmt.Errorf("failed to stop share: %w", err)
	}

	return nil
}

// SaveToVault 转存到文件柜
func (s *ShareService) SaveToVault(pickupCode string, password string, fileIDs []uuid.UUID, userID uuid.UUID) ([]uuid.UUID, error) {
	// 验证分享
	session, files, err := s.GetShareByCode(pickupCode, password)
	if err != nil {
		return nil, err
	}

	// 验证文件ID
	fileMap := make(map[uuid.UUID]models.FileMetadata)
	for _, file := range files {
		fileMap[file.ID] = file
	}

	savedIDs := make([]uuid.UUID, 0, len(fileIDs))

	// 转存每个文件
	for _, fileID := range fileIDs {
		file, ok := fileMap[fileID]
		if !ok {
			continue // 跳过不存在的文件
		}

		// 执行秒传（逻辑复制）
		newMetadata, err := s.fileService.CreateFileMetadata(userID, file.FileBlobHash, file.Filename, file.Size)
		if err != nil {
			return savedIDs, fmt.Errorf("failed to save file %s: %w", file.Filename, err)
		}

		savedIDs = append(savedIDs, newMetadata.ID)
	}

	// 增加下载计数
	s.IncrementDownload(session.ID)

	return savedIDs, nil
}

// ListMyShares 获取我的分享列表
func (s *ShareService) ListMyShares(userID uuid.UUID, page int, pageSize int) ([]models.ShareSession, int64, error) {
	var shares []models.ShareSession
	var total int64

	// 查询总数
	if err := s.db.Model(&models.ShareSession{}).
		Where("creator_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count shares: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := s.db.Where("creator_id = ?", userID).
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&shares).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list shares: %w", err)
	}

	return shares, total, nil
}
