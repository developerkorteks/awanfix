package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseManager handles database operations for authentication
type DatabaseManager struct {
	db              *gorm.DB
	passwordManager *PasswordManager
}

// NewDatabaseManager creates a new database manager
func NewDatabaseManager(dbPath string) (*DatabaseManager, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	dm := &DatabaseManager{
		db:              db,
		passwordManager: NewPasswordManager(),
	}

	// Auto-migrate the schema
	if err := dm.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// Create default admin user if not exists
	if err := dm.createDefaultAdmin(); err != nil {
		return nil, fmt.Errorf("failed to create default admin: %w", err)
	}

	return dm, nil
}

// migrate runs database migrations
func (dm *DatabaseManager) migrate() error {
	return dm.db.AutoMigrate(
		&User{},
		&APIKey{},
		&FileOwnership{},
		&Session{},
		&AuditLog{},
	)
}

// createDefaultAdmin creates a default admin user
func (dm *DatabaseManager) createDefaultAdmin() error {
	var count int64
	dm.db.Model(&User{}).Where("role = ?", RoleAdmin).Count(&count)
	
	if count > 0 {
		return nil // Admin already exists
	}

	hashedPassword, err := dm.passwordManager.HashPassword("Admin123!")
	if err != nil {
		return err
	}

	admin := &User{
		Email:        "admin@rclonestorage.local",
		Password:     hashedPassword,
		Role:         RoleAdmin,
		StorageQuota: DefaultAdminQuota,
		IsActive:     true,
	}

	return dm.db.Create(admin).Error
}

// CreateUser creates a new user
func (dm *DatabaseManager) CreateUser(email, password, role string) (*User, error) {
	if err := ValidateEmail(email); err != nil {
		return nil, err
	}

	hashedPassword, err := dm.passwordManager.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &User{
		Email:    email,
		Password: hashedPassword,
		Role:     role,
		IsActive: true,
	}

	if err := dm.db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// AuthenticateUser authenticates a user with email and password
func (dm *DatabaseManager) AuthenticateUser(email, password string) (*User, error) {
	var user User
	if err := dm.db.Where("email = ? AND is_active = ?", email, true).First(&user).Error; err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := dm.passwordManager.CheckPassword(password, user.Password); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (dm *DatabaseManager) GetUserByID(id uint) (*User, error) {
	var user User
	if err := dm.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (dm *DatabaseManager) GetUserByEmail(email string) (*User, error) {
	var user User
	if err := dm.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates user information
func (dm *DatabaseManager) UpdateUser(user *User) error {
	return dm.db.Save(user).Error
}

// DeleteUser soft deletes a user
func (dm *DatabaseManager) DeleteUser(id uint) error {
	return dm.db.Model(&User{}).Where("id = ?", id).Update("is_active", false).Error
}

// ListUsers lists all users with pagination
func (dm *DatabaseManager) ListUsers(offset, limit int) ([]User, int64, error) {
	var users []User
	var total int64

	dm.db.Model(&User{}).Count(&total)
	err := dm.db.Offset(offset).Limit(limit).Find(&users).Error

	return users, total, err
}

// CreateAPIKey creates a new API key for a user
func (dm *DatabaseManager) CreateAPIKey(userID uint, name string) (*APIKey, error) {
	// Generate random API key
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}
	key := "rcs_" + hex.EncodeToString(bytes)

	apiKey := &APIKey{
		UserID:   userID,
		Key:      key,
		Name:     name,
		IsActive: true,
	}

	if err := dm.db.Create(apiKey).Error; err != nil {
		return nil, err
	}

	return apiKey, nil
}

// ValidateAPIKey validates an API key and returns the associated user
func (dm *DatabaseManager) ValidateAPIKey(key string) (*User, error) {
	var apiKey APIKey
	if err := dm.db.Preload("User").Where("key = ? AND is_active = ?", key, true).First(&apiKey).Error; err != nil {
		return nil, fmt.Errorf("invalid API key")
	}

	if !apiKey.User.IsActive {
		return nil, fmt.Errorf("user account is disabled")
	}

	// Update last used timestamp
	now := time.Now()
	apiKey.LastUsed = &now
	dm.db.Save(&apiKey)

	return &apiKey.User, nil
}

// ListAPIKeys lists API keys for a user
func (dm *DatabaseManager) ListAPIKeys(userID uint) ([]APIKey, error) {
	var apiKeys []APIKey
	err := dm.db.Where("user_id = ? AND is_active = ?", userID, true).Find(&apiKeys).Error
	return apiKeys, err
}

// DeleteAPIKey deletes an API key
func (dm *DatabaseManager) DeleteAPIKey(id uint, userID uint) error {
	return dm.db.Model(&APIKey{}).Where("id = ? AND user_id = ?", id, userID).Update("is_active", false).Error
}

// CreateFileOwnership creates a file ownership record
func (dm *DatabaseManager) CreateFileOwnership(userID uint, fileID, filename, provider string, size int64, mimeType string) error {
	ownership := &FileOwnership{
		UserID:   userID,
		FileID:   fileID,
		Filename: filename,
		Size:     size,
		Provider: provider,
		MimeType: mimeType,
	}

	if err := dm.db.Create(ownership).Error; err != nil {
		return err
	}

	// Update user storage usage
	return dm.db.Model(&User{}).Where("id = ?", userID).Update("storage_used", gorm.Expr("storage_used + ?", size)).Error
}

// DeleteFileOwnership deletes a file ownership record
func (dm *DatabaseManager) DeleteFileOwnership(fileID string, userID uint) error {
	var ownership FileOwnership
	if err := dm.db.Where("file_id = ? AND user_id = ?", fileID, userID).First(&ownership).Error; err != nil {
		return err
	}

	// Delete ownership record
	if err := dm.db.Delete(&ownership).Error; err != nil {
		return err
	}

	// Update user storage usage
	return dm.db.Model(&User{}).Where("id = ?", userID).Update("storage_used", gorm.Expr("storage_used - ?", ownership.Size)).Error
}

// CheckFileOwnership checks if a user owns a file
func (dm *DatabaseManager) CheckFileOwnership(fileID string, userID uint) (*FileOwnership, error) {
	var ownership FileOwnership
	err := dm.db.Where("file_id = ? AND user_id = ?", fileID, userID).First(&ownership).Error
	if err != nil {
		return nil, err
	}
	return &ownership, nil
}

// ListUserFiles lists files owned by a user
func (dm *DatabaseManager) ListUserFiles(userID uint, offset, limit int) ([]FileOwnership, int64, error) {
	var files []FileOwnership
	var total int64

	dm.db.Model(&FileOwnership{}).Where("user_id = ?", userID).Count(&total)
	err := dm.db.Where("user_id = ?", userID).Offset(offset).Limit(limit).Find(&files).Error

	return files, total, err
}

// LogAudit logs an audit event
func (dm *DatabaseManager) LogAudit(userID uint, action, resource, ipAddress, userAgent string, success bool, details string) error {
	audit := &AuditLog{
		UserID:    userID,
		Action:    action,
		Resource:  resource,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Success:   success,
		Details:   details,
	}

	return dm.db.Create(audit).Error
}

// GetDatabase returns the underlying database connection
func (dm *DatabaseManager) GetDatabase() *gorm.DB {
	return dm.db
}

// Close closes the database connection
func (dm *DatabaseManager) Close() error {
	sqlDB, err := dm.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}