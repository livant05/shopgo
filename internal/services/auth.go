package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"github.com/yourorg/shopgo/internal/domain"
	"github.com/yourorg/shopgo/internal/ports"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("credenciales inválidas")
	ErrMFARequired        = errors.New("se requiere código 2FA")
	ErrInvalidMFA         = errors.New("código 2FA incorrecto")
	ErrAccountDisabled    = errors.New("cuenta deshabilitada")
)

// Claims — minimalista y claro. Sin tenant_slug, sin slugs innecesarios.
type Claims struct {
	jwt.RegisteredClaims
	UserID   string `json:"uid"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	BranchID string `json:"bid,omitempty"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type AuthService struct {
	users      ports.UserRepository
	cache      ports.CacheService
	privKey    any
	pubKey     any
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewAuthService(users ports.UserRepository, cache ports.CacheService, priv, pub any, at, rt time.Duration) *AuthService {
	return &AuthService{users: users, cache: cache, privKey: priv, pubKey: pub, accessTTL: at, refreshTTL: rt}
}

func (s *AuthService) Login(ctx context.Context, email, password, totpCode string) (*TokenPair, error) {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		// Siempre ejecutar bcrypt para prevenir timing attacks
		bcrypt.CompareHashAndPassword([]byte("$2a$12$placeholder_hash_to_avoid_timing"), []byte(password))
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	if !user.IsActive {
		return nil, ErrAccountDisabled
	}
	if user.MFAEnabled {
		if totpCode == "" {
			return nil, ErrMFARequired
		}
		if !totp.Validate(totpCode, user.MFASecret) {
			return nil, ErrInvalidMFA
		}
	}
	return s.issue(user)
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	var revoked bool
	if s.cache.Get(ctx, "revoked:"+refreshToken, &revoked) == nil {
		return nil, errors.New("token revocado")
	}

	claims := &Claims{}
	t, err := jwt.ParseWithClaims(refreshToken, claims, func(_ *jwt.Token) (any, error) {
		return s.pubKey, nil
	})
	if err != nil || !t.Valid {
		return nil, errors.New("token inválido")
	}

	user, err := s.users.GetByID(ctx, claims.UserID)
	if err != nil || !user.IsActive {
		return nil, errors.New("usuario no válido")
	}
	return s.issue(user)
}

func (s *AuthService) SetupMFA(ctx context.Context, userID string) (secret, qrURL string, err error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return "", "", err
	}
	key, err := totp.Generate(totp.GenerateOpts{Issuer: "ShopGo", AccountName: user.Email})
	if err != nil {
		return "", "", err
	}
	if err := s.users.SetMFASecret(ctx, userID, key.Secret()); err != nil {
		return "", "", err
	}
	return key.Secret(), key.URL(), nil
}

func (s *AuthService) HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(b), err
}

func (s *AuthService) VerifyPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func (s *AuthService) issue(user *domain.User) (*TokenPair, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
		},
		UserID:   user.ID,
		Email:    user.Email,
		Role:     string(user.Role),
		BranchID: user.BranchID,
	}

	access, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(s.privKey)
	if err != nil {
		return nil, fmt.Errorf("firmar access token: %w", err)
	}

	claims.ID = uuid.New().String()
	claims.ExpiresAt = jwt.NewNumericDate(now.Add(s.refreshTTL))
	refresh, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(s.privKey)
	if err != nil {
		return nil, fmt.Errorf("firmar refresh token: %w", err)
	}

	return &TokenPair{AccessToken: access, RefreshToken: refresh, ExpiresIn: int(s.accessTTL.Seconds())}, nil
}
