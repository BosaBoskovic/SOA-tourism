package repo

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// TokenRepo stores refresh tokens and password-reset tokens as opaque
// random strings - only their SHA-256 hash ever touches the database, same
// idea as a password hash, so a DB read alone can't be replayed as a token.
type TokenRepo struct {
	driver   neo4j.DriverWithContext
	database string
}

func NewTokenRepo(driver neo4j.DriverWithContext, database string) *TokenRepo {
	return &TokenRepo{driver: driver, database: database}
}

func (r *TokenRepo) EnsureUniqueConstraints(ctx context.Context) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: r.database})
	defer func() {
		if err := session.Close(ctx); err != nil {
			log.Printf("cannot close neo4j session: %v", err)
		}
	}()

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		queries := []string{
			`CREATE CONSTRAINT refresh_token_hash_unique IF NOT EXISTS
			 FOR (t:RefreshToken) REQUIRE t.tokenHash IS UNIQUE`,
			`CREATE CONSTRAINT reset_token_hash_unique IF NOT EXISTS
			 FOR (t:PasswordResetToken) REQUIRE t.tokenHash IS UNIQUE`,
		}
		for _, q := range queries {
			if _, err := tx.Run(ctx, q, nil); err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	return err
}

func newOpaqueToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// --- Refresh tokens -------------------------------------------------------

// CreateRefreshToken generates and stores a new refresh token for username,
// returning the raw token (never stored - only its hash is).
func (r *TokenRepo) CreateRefreshToken(ctx context.Context, username string, ttl time.Duration) (string, time.Time, error) {
	raw, err := newOpaqueToken()
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt := time.Now().UTC().Add(ttl)

	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: r.database})
	defer func() {
		if err := session.Close(ctx); err != nil {
			log.Printf("cannot close neo4j session: %v", err)
		}
	}()

	_, err = session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		_, err := tx.Run(ctx,
			`CREATE (t:RefreshToken {
				tokenHash: $tokenHash,
				username: $username,
				expiresAt: $expiresAt,
				createdAt: datetime(),
				revoked: false
			})`,
			map[string]any{
				"tokenHash": hashToken(raw),
				"username":  username,
				"expiresAt": expiresAt,
			},
		)
		return nil, err
	})
	if err != nil {
		return "", time.Time{}, err
	}
	return raw, expiresAt, nil
}

// ConsumeRefreshToken validates a raw refresh token and revokes it (single
// use - the caller is expected to issue a new one, i.e. rotation), returning
// the username it belonged to.
func (r *TokenRepo) ConsumeRefreshToken(ctx context.Context, raw string) (string, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: r.database})
	defer func() {
		if err := session.Close(ctx); err != nil {
			log.Printf("cannot close neo4j session: %v", err)
		}
	}()

	result, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		res, err := tx.Run(ctx,
			`MATCH (t:RefreshToken {tokenHash: $tokenHash})
			 WHERE t.revoked = false AND t.expiresAt > datetime()
			 RETURN t.username AS username`,
			map[string]any{"tokenHash": hashToken(raw)},
		)
		if err != nil {
			return nil, err
		}
		records, err := res.Collect(ctx)
		if err != nil {
			return nil, err
		}
		if len(records) == 0 {
			return nil, errors.New("invalid_refresh_token")
		}
		username := asString(mustGet(records[0], "username"))

		_, err = tx.Run(ctx,
			`MATCH (t:RefreshToken {tokenHash: $tokenHash}) SET t.revoked = true`,
			map[string]any{"tokenHash": hashToken(raw)},
		)
		if err != nil {
			return nil, err
		}
		return username, nil
	})
	if err != nil {
		return "", err
	}
	return result.(string), nil
}

// RevokeRefreshToken revokes a single refresh token (used on logout).
func (r *TokenRepo) RevokeRefreshToken(ctx context.Context, raw string) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: r.database})
	defer func() {
		if err := session.Close(ctx); err != nil {
			log.Printf("cannot close neo4j session: %v", err)
		}
	}()

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		_, err := tx.Run(ctx,
			`MATCH (t:RefreshToken {tokenHash: $tokenHash}) SET t.revoked = true`,
			map[string]any{"tokenHash": hashToken(raw)},
		)
		return nil, err
	})
	return err
}

// RevokeAllRefreshTokens revokes every refresh token for username (used
// after a password change/reset, so old sessions can't silently continue).
func (r *TokenRepo) RevokeAllRefreshTokens(ctx context.Context, username string) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: r.database})
	defer func() {
		if err := session.Close(ctx); err != nil {
			log.Printf("cannot close neo4j session: %v", err)
		}
	}()

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		_, err := tx.Run(ctx,
			`MATCH (t:RefreshToken {username: $username}) SET t.revoked = true`,
			map[string]any{"username": username},
		)
		return nil, err
	})
	return err
}

// --- Password reset tokens --------------------------------------------

func (r *TokenRepo) CreatePasswordResetToken(ctx context.Context, username string, ttl time.Duration) (string, time.Time, error) {
	raw, err := newOpaqueToken()
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt := time.Now().UTC().Add(ttl)

	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: r.database})
	defer func() {
		if err := session.Close(ctx); err != nil {
			log.Printf("cannot close neo4j session: %v", err)
		}
	}()

	_, err = session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		_, err := tx.Run(ctx,
			`CREATE (t:PasswordResetToken {
				tokenHash: $tokenHash,
				username: $username,
				expiresAt: $expiresAt,
				createdAt: datetime(),
				used: false
			})`,
			map[string]any{
				"tokenHash": hashToken(raw),
				"username":  username,
				"expiresAt": expiresAt,
			},
		)
		return nil, err
	})
	if err != nil {
		return "", time.Time{}, err
	}
	return raw, expiresAt, nil
}

// ConsumePasswordResetToken validates a raw reset token and marks it used
// (single use), returning the username it belonged to.
func (r *TokenRepo) ConsumePasswordResetToken(ctx context.Context, raw string) (string, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: r.database})
	defer func() {
		if err := session.Close(ctx); err != nil {
			log.Printf("cannot close neo4j session: %v", err)
		}
	}()

	result, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		res, err := tx.Run(ctx,
			`MATCH (t:PasswordResetToken {tokenHash: $tokenHash})
			 WHERE t.used = false AND t.expiresAt > datetime()
			 RETURN t.username AS username`,
			map[string]any{"tokenHash": hashToken(raw)},
		)
		if err != nil {
			return nil, err
		}
		records, err := res.Collect(ctx)
		if err != nil {
			return nil, err
		}
		if len(records) == 0 {
			return nil, errors.New("invalid_reset_token")
		}
		username := asString(mustGet(records[0], "username"))

		_, err = tx.Run(ctx,
			`MATCH (t:PasswordResetToken {tokenHash: $tokenHash}) SET t.used = true`,
			map[string]any{"tokenHash": hashToken(raw)},
		)
		if err != nil {
			return nil, err
		}
		return username, nil
	})
	if err != nil {
		return "", err
	}
	return result.(string), nil
}

func mustGet(rec *neo4j.Record, key string) any {
	v, _ := rec.Get(key)
	return v
}
