package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"stakeholders/model"
	"stakeholders/ratelimit"
	"stakeholders/repo"
)

const accessTokenTTL = 15 * time.Minute
const refreshTokenTTL = 7 * 24 * time.Hour
const passwordResetTokenTTL = 30 * time.Minute

// Login lockout: 5 failed attempts within a minute locks that identity out
// for 5 minutes. Blunts brute-forcing a specific account (there's no other
// throttling in front of /stakeholders/login).
const (
	loginMaxAttempts = 5
	loginWindow      = time.Minute
	loginLockout     = 5 * time.Minute
)

type AccessClaims struct {
	Role  string `json:"role"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// TokenPair is what every endpoint that authenticates a user (login,
// register, refresh) returns.
type TokenPair struct {
	AccessToken           string
	AccessTokenExpiresAt  time.Time
	RefreshToken          string
	RefreshTokenExpiresAt time.Time
}

type AuthService struct {
	repo         *repo.AccountRepo
	profileRepo  *repo.ProfileRepo
	tokenRepo    *repo.TokenRepo
	secret       []byte
	loginLimiter *ratelimit.LoginLimiter
}

func NewAuthService(r *repo.AccountRepo, profileRepo *repo.ProfileRepo, tokenRepo *repo.TokenRepo, secret []byte) *AuthService {
	return &AuthService{
		repo:         r,
		profileRepo:  profileRepo,
		tokenRepo:    tokenRepo,
		secret:       secret,
		loginLimiter: ratelimit.NewLoginLimiter(loginMaxAttempts, loginWindow, loginLockout),
	}
}

func (s *AuthService) EnsureUniqueConstraints(ctx context.Context) error {
	if err := s.repo.EnsureUniqueConstraints(ctx); err != nil {
		return err
	}
	return s.tokenRepo.EnsureUniqueConstraints(ctx)
}

// Register creates the account and, like Login, returns a token pair so the
// caller is immediately signed in - no separate login round-trip needed.
func (s *AuthService) Register(ctx context.Context, req model.RegisterRequest) (*model.AccountResponse, TokenPair, error) {
	role, err := normalizeRegistrableRole(req.Role)
	if err != nil {
		return nil, TokenPair{}, err
	}

	exists, err := s.repo.ExistsByUsernameOrEmail(ctx, req.Username, req.Email)
	if err != nil {
		return nil, TokenPair{}, err
	}
	if exists {
		return nil, TokenPair{}, errors.New("username_or_email_exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, TokenPair{}, err
	}

	acc := model.Account{
		Username:     req.Username,
		Email:        req.Email,
		Role:         role,
		PasswordHash: string(hash),
	}
	if err = s.repo.CreateAccount(ctx, acc); err != nil {
		return nil, TokenPair{}, err
	}

	emptyProfile := model.Profile{Username: req.Username}
	if err = s.profileRepo.CreateProfile(ctx, emptyProfile); err != nil {
		return nil, TokenPair{}, err
	}

	pair, err := s.issueTokenPair(ctx, acc)
	if err != nil {
		return nil, TokenPair{}, err
	}

	return &model.AccountResponse{
		Username:  req.Username,
		Email:     req.Email,
		Role:      role,
		IsBlocked: false,
		CreatedAt: time.Now().UTC(),
	}, pair, nil
}

func (s *AuthService) Login(ctx context.Context, req model.LoginRequest) (TokenPair, *model.Account, error) {
	limiterKey := strings.ToLower(strings.TrimSpace(req.UsernameOrEmail))
	if !s.loginLimiter.Allow(limiterKey) {
		return TokenPair{}, nil, errors.New("too_many_attempts")
	}

	acc, err := s.repo.FindByIdentity(ctx, req.UsernameOrEmail)
	if err != nil {
		s.loginLimiter.RecordFailure(limiterKey)
		return TokenPair{}, nil, err
	}

	if acc.IsBlocked {
		return TokenPair{}, nil, errors.New("account_blocked")
	}

	if err = bcrypt.CompareHashAndPassword([]byte(acc.PasswordHash), []byte(req.Password)); err != nil {
		s.loginLimiter.RecordFailure(limiterKey)
		return TokenPair{}, nil, errors.New("invalid_credentials")
	}

	pair, err := s.issueTokenPair(ctx, *acc)
	if err != nil {
		return TokenPair{}, nil, err
	}

	s.loginLimiter.RecordSuccess(limiterKey)
	return pair, acc, nil
}

// RefreshAccessToken exchanges a still-valid refresh token for a new token
// pair, rotating the refresh token (the old one is revoked - each refresh
// token is single-use).
func (s *AuthService) RefreshAccessToken(ctx context.Context, rawRefreshToken string) (TokenPair, *model.Account, error) {
	username, err := s.tokenRepo.ConsumeRefreshToken(ctx, rawRefreshToken)
	if err != nil {
		return TokenPair{}, nil, errors.New("invalid_refresh_token")
	}

	acc, err := s.repo.FindByIdentity(ctx, username)
	if err != nil {
		return TokenPair{}, nil, err
	}
	if acc.IsBlocked {
		return TokenPair{}, nil, errors.New("account_blocked")
	}

	pair, err := s.issueTokenPair(ctx, *acc)
	if err != nil {
		return TokenPair{}, nil, err
	}
	return pair, acc, nil
}

// Logout revokes a single refresh token. Best-effort - an already
// expired/unknown token is not an error from the caller's point of view.
func (s *AuthService) Logout(ctx context.Context, rawRefreshToken string) error {
	if strings.TrimSpace(rawRefreshToken) == "" {
		return nil
	}
	return s.tokenRepo.RevokeRefreshToken(ctx, rawRefreshToken)
}

// ChangePassword verifies the caller's current password before setting a
// new one, then revokes every refresh token so other sessions can't
// silently continue using the old credentials' trust.
func (s *AuthService) ChangePassword(ctx context.Context, username string, req model.ChangePasswordRequest) error {
	acc, err := s.repo.FindByIdentity(ctx, username)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(acc.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return errors.New("invalid_current_password")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.repo.SetPasswordHash(ctx, username, string(hash)); err != nil {
		return err
	}
	return s.tokenRepo.RevokeAllRefreshTokens(ctx, username)
}

// RequestPasswordReset issues a reset token if the identity resolves to a
// real account. There's no email service anywhere in this stack, so the
// raw token is returned directly to the caller (see the API-layer comment
// on the handler for why that's acceptable here and nowhere else).
// A nonexistent identity returns a nil token with no error, so the HTTP
// layer can reply with the same generic message either way and not leak
// which usernames/emails exist.
func (s *AuthService) RequestPasswordReset(ctx context.Context, usernameOrEmail string) (string, time.Time, error) {
	acc, err := s.repo.FindByIdentity(ctx, usernameOrEmail)
	if err != nil {
		return "", time.Time{}, nil
	}
	return s.tokenRepo.CreatePasswordResetToken(ctx, acc.Username, passwordResetTokenTTL)
}

// ConfirmPasswordReset consumes a reset token (single use) and sets the new
// password, then revokes every refresh token for that account.
func (s *AuthService) ConfirmPasswordReset(ctx context.Context, rawToken, newPassword string) error {
	username, err := s.tokenRepo.ConsumePasswordResetToken(ctx, rawToken)
	if err != nil {
		return errors.New("invalid_reset_token")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.repo.SetPasswordHash(ctx, username, string(hash)); err != nil {
		return err
	}
	return s.tokenRepo.RevokeAllRefreshTokens(ctx, username)
}

func (s *AuthService) GetAllAccounts(ctx context.Context) ([]map[string]any, error) {
	return s.repo.GetAllAccounts(ctx)
}

func (s *AuthService) BlockAccount(ctx context.Context, username string) error {
	return s.repo.BlockAccount(ctx, username)
}

// UnblockAccount lifts a block, the inverse of BlockAccount.
func (s *AuthService) UnblockAccount(ctx context.Context, username string) error {
	return s.repo.UnblockAccount(ctx, username)
}

// DeleteAccount removes an account entirely (admin-only).
func (s *AuthService) DeleteAccount(ctx context.Context, username string) error {
	if err := s.repo.DeleteAccount(ctx, username); err != nil {
		return err
	}
	return s.tokenRepo.RevokeAllRefreshTokens(ctx, username)
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

func (s *AuthService) issueTokenPair(ctx context.Context, acc model.Account) (TokenPair, error) {
	accessToken, accessExpiresAt, err := s.generateAccessToken(acc)
	if err != nil {
		return TokenPair{}, err
	}
	refreshToken, refreshExpiresAt, err := s.tokenRepo.CreateRefreshToken(ctx, acc.Username, refreshTokenTTL)
	if err != nil {
		return TokenPair{}, err
	}
	return TokenPair{
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessExpiresAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshExpiresAt,
	}, nil
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
