package auth

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Email        string    `json:"email" gorm:"unique;not null"`
	Password     string    `json:"-" gorm:"not null"`
	Role         string    `json:"role" gorm:"default:user"`
	StorageUsed  int64     `json:"storage_used" gorm:"default:0"`
	StorageQuota int64     `json:"storage_quota" gorm:"default:1073741824"` // 1GB default
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// APIKey represents an API key for programmatic access
type APIKey struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	UserID    uint       `json:"user_id"`
	User      User       `json:"user" gorm:"foreignKey:UserID"`
	Key       string     `json:"key" gorm:"unique;not null"`
	Name      string     `json:"name"`
	LastUsed  *time.Time `json:"last_used"`
	IsActive  bool       `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// FileOwnership tracks file ownership and storage usage
type FileOwnership struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id"`
	User      User      `json:"user" gorm:"foreignKey:UserID"`
	FileID    string    `json:"file_id" gorm:"unique;not null"`
	Filename  string    `json:"filename"`
	Size      int64     `json:"size"`
	Provider  string    `json:"provider"`
	MimeType  string    `json:"mime_type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Session represents user sessions for web interface
type Session struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id"`
	User      User      `json:"user" gorm:"foreignKey:UserID"`
	Token     string    `json:"token" gorm:"unique;not null"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// AuditLog tracks user actions for security
type AuditLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id"`
	User      User      `json:"user" gorm:"foreignKey:UserID"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	Success   bool      `json:"success"`
	Details   string    `json:"details"`
	CreatedAt time.Time `json:"created_at"`
}

// UserRole constants
const (
	RoleAdmin    = "admin"
	RoleUser     = "user"
	RoleReadOnly = "readonly"
)

// Default storage quotas (in bytes)
const (
	DefaultUserQuota     = 1 * 1024 * 1024 * 1024  // 1GB
	DefaultAdminQuota    = -1                       // Unlimited
	DefaultReadOnlyQuota = 0                        // No upload
)

// BeforeCreate hook for User
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// Set default quota based on role
	if u.StorageQuota == 0 {
		switch u.Role {
		case RoleAdmin:
			u.StorageQuota = DefaultAdminQuota
		case RoleReadOnly:
			u.StorageQuota = DefaultReadOnlyQuota
		default:
			u.StorageQuota = DefaultUserQuota
		}
	}
	return nil
}

// CanUpload checks if user can upload files
func (u *User) CanUpload() bool {
	return u.Role != RoleReadOnly && u.IsActive
}

// CanDelete checks if user can delete files
func (u *User) CanDelete() bool {
	return u.Role != RoleReadOnly && u.IsActive
}

// IsAdmin checks if user has admin privileges
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin && u.IsActive
}

// HasStorageSpace checks if user has enough storage space
func (u *User) HasStorageSpace(requiredSize int64) bool {
	if u.StorageQuota == -1 { // Unlimited
		return true
	}
	return u.StorageUsed+requiredSize <= u.StorageQuota
}

// GetStorageUsagePercent returns storage usage as percentage
func (u *User) GetStorageUsagePercent() float64 {
	if u.StorageQuota == -1 || u.StorageQuota == 0 {
		return 0
	}
	return float64(u.StorageUsed) / float64(u.StorageQuota) * 100
}