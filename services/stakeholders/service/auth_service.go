package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"stakeholders/model"
	"stakeholders/repo"
)

const accessTokenTTL = 15 * time.Minute

type AccessClaims struct {
	Role  string `json:"role"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

type AuthService struct {
	repo        *repo.AccountRepo
	profileRepo *repo.ProfileRepo
	secret      []byte
}

func NewAuthService(r *repo.AccountRepo, profileRepo *repo.ProfileRepo, secret []byte) *AuthService {
	return &AuthService{repo: r, profileRepo: profileRepo, secret: secret}
}

func (s *AuthService) EnsureUniqueConstraints(ctx context.Context) error {
	return s.repo.EnsureUniqueConstraints(ctx)
}

func (s *AuthService) Register(ctx context.Context, req model.RegisterRequest) (*model.AccountResponse, error) {
	role, err := normalizeRegistrableRole(req.Role)
	if err != nil {
		return nil, err
	}

	exists, err := s.repo.ExistsByUsernameOrEmail(ctx, req.Username, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("username_or_email_exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	acc := model.Account{
		Username:     req.Username,
		Email:        req.Email,
		Role:         role,
		PasswordHash: string(hash),
	}
	if err = s.repo.CreateAccount(ctx, acc); err != nil {
		return nil, err
	}

	emptyProfile := model.Profile{Username: req.Username}
	if err = s.profileRepo.CreateProfile(ctx, emptyProfile); err != nil {
		return nil, err
	}

	return &model.AccountResponse{
		Username:  req.Username,
		Email:     req.Email,
		Role:      role,
		IsBlocked: false,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req model.LoginRequest) (string, time.Time, *model.Account, error) {
	acc, err := s.repo.FindByIdentity(ctx, req.UsernameOrEmail)
	if err != nil {
		return "", time.Time{}, nil, err
	}

	if acc.IsBlocked {
		return "", time.Time{}, nil, errors.New("account_blocked")
	}

	if err = bcrypt.CompareHashAndPassword([]byte(acc.PasswordHash), []byte(req.Password)); err != nil {
		return "", time.Time{}, nil, errors.New("invalid_credentials")
	}

	token, expiresAt, err := s.generateAccessToken(*acc)
	if err != nil {
		return "", time.Time{}, nil, err
	}

	return token, expiresAt, acc, nil
}

func (s *AuthService) GetAllAccounts(ctx context.Context) ([]map[string]any, error) {
	return s.repo.GetAllAccounts(ctx)
}

func (s *AuthService) BlockAccount(ctx context.Context, username string) error {
	return s.repo.BlockAccount(ctx, username)
}

func (s *AuthService) ParseAdminClaims(tokenString string) (*AccessClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessClaims{}, func(t *jwt.Token) (any, error) {
		return s.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid_token")
	}
	claims, ok := token.Claims.(*AccessClaims)
	if !ok {
		return nil, errors.New("invalid_claims")
	}
	if claims.Role != "admin" {
		return nil, errors.New("forbidden")
	}
	return claims, nil
}

func (s *AuthService) generateAccessToken(acc model.Account) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(accessTokenTTL)

	claims := AccessClaims{
		Role:  acc.Role,
		Email: acc.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   acc.Username,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	return signed, expiresAt, err
}

func normalizeRegistrableRole(role string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "guide", "vodic":
		return "guide", nil
	case "tourist", "turista":
		return "tourist", nil
	default:
		return "", errors.New("role must be guide or tourist")
	}
}

func (s *AuthService) ParseClaims(tokenString string) (*AccessClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessClaims{}, func(t *jwt.Token) (any, error) {
		return s.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid_token")
	}
	claims, ok := token.Claims.(*AccessClaims)
	if !ok {
		return nil, errors.New("invalid_claims")
	}
	return claims, nil
}
