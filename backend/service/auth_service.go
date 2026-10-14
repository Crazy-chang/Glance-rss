package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"glance-wx/backend/config"
	"glance-wx/backend/model"
)

// devOpenID 微信凭证未配置（本地开发）时使用的固定 openid，
// 保证模拟器联调时用户身份稳定，不会每次 wx.login 都新建用户
const devOpenID = "dev-openid"

// AuthService 登录与 JWT 签发/校验
type AuthService struct {
	db     *gorm.DB
	cfg    *config.Config
	client *http.Client
}

func NewAuthService(db *gorm.DB, cfg *config.Config) *AuthService {
	return &AuthService{
		db:  db,
		cfg: cfg,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Claims JWT 载荷
type Claims struct {
	UserID int64 `json:"uid"`
	jwt.RegisteredClaims
}

// LoginResult 登录返回
type LoginResult struct {
	Token string `json:"token"`
}

// Login 用 wx.login 的 code 换 openid，自动建用户与默认设置，签发 JWT
func (s *AuthService) Login(ctx context.Context, code string) (*LoginResult, error) {
	if strings.TrimSpace(code) == "" {
		return nil, ErrInvalidCode
	}

	openID, err := s.code2OpenID(ctx, code)
	if err != nil {
		return nil, err
	}

	// 查找或创建用户
	var user model.User
	err = s.db.Where("open_id = ?", openID).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		user = model.User{OpenID: openID}
		if err := s.db.Create(&user).Error; err != nil {
			return nil, err
		}
		// 随用户创建默认设置
		setting := model.DefaultSetting(user.ID)
		if err := s.db.Create(&setting).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	token, err := s.signToken(user.ID)
	if err != nil {
		return nil, err
	}
	return &LoginResult{Token: token}, nil
}

// ParseToken 校验 JWT 并返回 user_id
func (s *AuthService) ParseToken(tokenString string) (int64, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("非预期的签名方法: %v", t.Header["alg"])
		}
		return []byte(s.cfg.JWT.Secret), nil
	})
	if err != nil || !token.Valid {
		return 0, errors.New("token 无效")
	}
	if claims.UserID <= 0 {
		return 0, errors.New("token 载荷异常")
	}
	return claims.UserID, nil
}

func (s *AuthService) signToken(userID int64) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(s.cfg.JWT.ExpireDays) * 24 * time.Hour)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.cfg.JWT.Secret))
}

// code2SessionResp 微信 code2session 响应
type code2SessionResp struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// code2OpenID 调微信 code2session；凭证为占位值时走开发兜底
func (s *AuthService) code2OpenID(ctx context.Context, code string) (string, error) {
	if s.cfg.WeChat.AppID == "" || s.cfg.WeChat.Secret == "" ||
		s.cfg.WeChat.AppID == "YOUR_APPID" || s.cfg.WeChat.Secret == "YOUR_SECRET" {
		return devOpenID, nil
	}

	u, err := url.Parse("https://api.weixin.qq.com/sns/jscode2session")
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("appid", s.cfg.WeChat.AppID)
	q.Set("secret", s.cfg.WeChat.Secret)
	q.Set("js_code", code)
	q.Set("grant_type", "authorization_code")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return "", err
	}

	var data code2SessionResp
	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}
	if data.ErrCode != 0 {
		return "", fmt.Errorf("微信登录失败: %s", data.ErrMsg)
	}
	if data.OpenID == "" {
		return "", errors.New("微信未返回 openid")
	}
	return data.OpenID, nil
}
