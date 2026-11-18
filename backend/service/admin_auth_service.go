package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"glance-wx/backend/config"
	"glance-wx/backend/model"
)

// 初始管理员账号（首次启动自动创建，登录页有提示）
const (
	defaultAdminUsername = "admin"
	defaultAdminPassword = "666666"
	adminTokenScope      = "glance-admin"
	adminTokenTTL        = 12 * time.Hour
)

// AdminActor 当前操作的管理员，用于写操作日志
type AdminActor struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

// AdminAuthService 后台管理员登录与令牌校验
type AdminAuthService struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewAdminAuthService(db *gorm.DB, cfg *config.Config) *AdminAuthService {
	return &AdminAuthService{db: db, cfg: cfg}
}

// AdminClaims 后台 JWT 载荷；Scope 与小程序用户令牌区分，互不通用
type AdminClaims struct {
	AdminID  int64  `json:"aid"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Scope    string `json:"scope"`
	jwt.RegisteredClaims
}

// EnsureDefaultAdmin 首次启动创建初始管理员
func (s *AdminAuthService) EnsureDefaultAdmin() error {
	var count int64
	if err := s.db.Model(&model.AdminUser{}).
		Where("username = ?", defaultAdminUsername).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(defaultAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.db.Create(&model.AdminUser{
		Username:     defaultAdminUsername,
		PasswordHash: string(hash),
		Role:         "admin",
	}).Error
}

// Login 校验账号密码并签发后台令牌
func (s *AdminAuthService) Login(username, password string) (string, *AdminActor, error) {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)
	if username == "" || password == "" {
		return "", nil, ErrAdminCredentials
	}

	var admin model.AdminUser
	err := s.db.Where("username = ?", username).Take(&admin).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil, ErrAdminCredentials
	}
	if err != nil {
		return "", nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)) != nil {
		return "", nil, ErrAdminCredentials
	}

	now := time.Now()
	if err := s.db.Model(&model.AdminUser{}).Where("id = ?", admin.ID).
		Update("last_login_at", now).Error; err != nil {
		return "", nil, err
	}

	token, err := s.signToken(admin, now)
	if err != nil {
		return "", nil, err
	}
	return token, &AdminActor{ID: admin.ID, Username: admin.Username}, nil
}

// ParseToken 校验后台令牌；小程序用户令牌无法通过
func (s *AdminAuthService) ParseToken(tokenString string) (*AdminClaims, error) {
	claims := &AdminClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("非预期的签名方法: %v", t.Header["alg"])
		}
		return []byte(s.secret()), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("token 无效")
	}
	if claims.Scope != adminTokenScope || claims.AdminID <= 0 {
		return nil, errors.New("token 载荷异常")
	}
	return claims, nil
}

func (s *AdminAuthService) signToken(admin model.AdminUser, now time.Time) (string, error) {
	claims := AdminClaims{
		AdminID:  admin.ID,
		Username: admin.Username,
		Role:     admin.Role,
		Scope:    adminTokenScope,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   admin.Username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(adminTokenTTL)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.secret()))
}

// secret 后台令牌使用独立密钥，避免与小程序用户令牌互相冒用
func (s *AdminAuthService) secret() string {
	return s.cfg.JWT.Secret + ":admin"
}
