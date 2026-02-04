package services

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"ahavault/server/internal/crypto"
	"ahavault/server/internal/models"
	"ahavault/server/internal/storage"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FileService 文件服务
type FileService struct {
	db      *gorm.DB
	storage storage.Engine
	kek     []byte
}

// NewFileService 创建文件服务实例
func NewFileService(db *gorm.DB, storageEngine storage.Engine, kek []byte) *FileService {
	return &FileService{
		db:      db,
		storage: storageEngine,
		kek:     kek,
	}
}

// CheckInstantUpload 秒传检测
func (s *FileService) CheckInstantUpload(hash string, userID uuid.UUID) (bool, *models.FileBlob, error) {
	// 验证哈希格式
	if err := storage.ValidateHash(hash); err != nil {
		return false, nil, err
	}

	// 查询文件是否存在
	var blob models.FileBlob
	err := s.db.Where("hash = ?", hash).First(&blob).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil, nil // 文件不存在，需要上传
		}
		return false, nil, fmt.Errorf("failed to check file: %w", err)
	}

	// 检查文件是否被禁止
	if blob.IsBanned {
		return false, nil, errors.New("file has been banned")
	}

	return true, &blob, nil
}

// CreateFileMetadata 创建文件元数据（秒传）
func (s *FileService) CreateFileMetadata(userID uuid.UUID, hash string, filename string, size int64) (*models.FileMetadata, error) {
	// 检查用户存储空间
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if !user.HasStorageSpace(size) {
		return nil, errors.New("insufficient storage space")
	}

	// 开启事务
	tx := s.db.Begin()
	defer tx.Rollback()

	// 增加引用计数
	var blob models.FileBlob
	if err := tx.Where("hash = ?", hash).First(&blob).Error; err != nil {
		return nil, fmt.Errorf("file blob not found: %w", err)
	}

	if err := blob.IncrementRefCount(tx); err != nil {
		return nil, fmt.Errorf("failed to increment ref count: %w", err)
	}

	// 创建文件元数据
	metadata := &models.FileMetadata{
		UserID:       userID,
		FileBlobHash: hash,
		Filename:     filename,
		Size:         size,
	}

	if err := tx.Create(metadata).Error; err != nil {
		return nil, fmt.Errorf("failed to create metadata: %w", err)
	}

	// 更新用户存储使用量
	if err := tx.Model(&user).Update("storage_used", gorm.Expr("storage_used + ?", size)).Error; err != nil {
		return nil, fmt.Errorf("failed to update storage usage: %w", err)
	}

	tx.Commit()

	return metadata, nil
}

// UploadFile 上传新文件
func (s *FileService) UploadFile(userID uuid.UUID, filename string, size int64, reader io.Reader) (*models.FileMetadata, error) {
	// 检查用户存储空间
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if !user.HasStorageSpace(size) {
		return nil, errors.New("insufficient storage space")
	}

	// 计算哈希
	var buf bytes.Buffer
	tee := io.TeeReader(reader, &buf)
	hash, err := crypto.CalculateSHA256Stream(tee)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate hash: %w", err)
	}

	// 检查是否已存在（二次秒传检测）
	exists, _, err := s.CheckInstantUpload(hash, userID)
	if err != nil {
		return nil, err
	}
	if exists {
		// 文件已存在，执行秒传
		return s.CreateFileMetadata(userID, hash, filename, size)
	}

	// 生成 DEK
	dek, err := crypto.GenerateDEK()
	if err != nil {
		return nil, fmt.Errorf("failed to generate DEK: %w", err)
	}
	defer crypto.ZeroBytes(dek)

	// 加密文件内容
	plaintext, err := io.ReadAll(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	ciphertext, err := crypto.EncryptFile(plaintext, dek)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt file: %w", err)
	}

	// 加密 DEK
	encryptedDEK, err := crypto.EncryptDEKToBase64(dek, s.kek)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt DEK: %w", err)
	}

	// 生成存储路径
	storePath, err := storage.GeneratePath(hash)
	if err != nil {
		return nil, err
	}

	// 开启事务
	tx := s.db.Begin()
	defer tx.Rollback()

	// 存储物理文件
	if err := s.storage.Put(hash, bytes.NewReader(ciphertext)); err != nil {
		return nil, fmt.Errorf("failed to store file: %w", err)
	}

	// 创建 FileBlob 记录
	blob := &models.FileBlob{
		Hash:         hash,
		StorePath:    storePath,
		EncryptedDEK: encryptedDEK,
		Size:         size,
		RefCount:     1,
	}

	if err := tx.Create(blob).Error; err != nil {
		// 删除已存储的文件
		s.storage.Delete(hash)
		return nil, fmt.Errorf("failed to create blob: %w", err)
	}

	// 创建文件元数据
	metadata := &models.FileMetadata{
		UserID:       userID,
		FileBlobHash: hash,
		Filename:     filename,
		Size:         size,
	}

	if err := tx.Create(metadata).Error; err != nil {
		s.storage.Delete(hash)
		return nil, fmt.Errorf("failed to create metadata: %w", err)
	}

	// 更新用户存储使用量
	if err := tx.Model(&user).Update("storage_used", gorm.Expr("storage_used + ?", size)).Error; err != nil {
		s.storage.Delete(hash)
		return nil, fmt.Errorf("failed to update storage usage: %w", err)
	}

	tx.Commit()

	return metadata, nil
}

// DownloadFile 下载文件
func (s *FileService) DownloadFile(fileID uuid.UUID, userID uuid.UUID) (io.ReadCloser, *models.FileMetadata, error) {
	// 获取文件元数据
	var metadata models.FileMetadata
	err := s.db.Where("id = ? AND user_id = ? AND deleted_at IS NULL", fileID, userID).
		First(&metadata).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errors.New("file not found")
		}
		return nil, nil, fmt.Errorf("failed to get file: %w", err)
	}

	// 检查文件是否过期
	if metadata.IsExpired() {
		return nil, nil, errors.New("file has expired")
	}

	// 获取物理文件
	var blob models.FileBlob
	if err := s.db.Where("hash = ?", metadata.FileBlobHash).First(&blob).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to get blob: %w", err)
	}

	// 解密 DEK
	dek, err := crypto.DecryptDEKFromBase64(blob.EncryptedDEK, s.kek)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decrypt DEK: %w", err)
	}
	defer crypto.ZeroBytes(dek)

	// 读取加密文件
	encryptedReader, err := s.storage.Get(blob.Hash)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get file from storage: %w", err)
	}
	defer encryptedReader.Close()

	// 读取并解密
	ciphertext, err := io.ReadAll(encryptedReader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read encrypted file: %w", err)
	}

	plaintext, err := crypto.DecryptFile(ciphertext, dek)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decrypt file: %w", err)
	}

	// 返回解密后的内容
	reader := io.NopCloser(bytes.NewReader(plaintext))
	return reader, &metadata, nil
}

// DeleteFile 删除文件（软删除）
func (s *FileService) DeleteFile(fileID uuid.UUID, userID uuid.UUID) error {
	tx := s.db.Begin()
	defer tx.Rollback()

	// 获取文件元数据
	var metadata models.FileMetadata
	err := tx.Where("id = ? AND user_id = ? AND deleted_at IS NULL", fileID, userID).
		First(&metadata).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("file not found")
		}
		return fmt.Errorf("failed to get file: %w", err)
	}

	// 软删除
	if err := metadata.SoftDelete(tx); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// 减少引用计数
	var blob models.FileBlob
	if err := tx.Where("hash = ?", metadata.FileBlobHash).First(&blob).Error; err != nil {
		return fmt.Errorf("failed to get blob: %w", err)
	}

	if err := blob.DecrementRefCount(tx); err != nil {
		return fmt.Errorf("failed to decrement ref count: %w", err)
	}

	// 更新用户存储使用量
	var user models.User
	if err := tx.First(&user, userID).Error; err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if err := tx.Model(&user).Update("storage_used", gorm.Expr("storage_used - ?", metadata.Size)).Error; err != nil {
		return fmt.Errorf("failed to update storage usage: %w", err)
	}

	tx.Commit()

	return nil
}

// ListFiles 获取用户文件列表
func (s *FileService) ListFiles(userID uuid.UUID, page int, pageSize int) ([]models.FileMetadata, int64, error) {
	var files []models.FileMetadata
	var total int64

	// 查询总数
	if err := s.db.Model(&models.FileMetadata{}).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count files: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := s.db.Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&files).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list files: %w", err)
	}

	return files, total, nil
}
