package repo

import (
	"context"
	"errors"
	"log"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"stakeholders/model"
)

type AccountRepo struct {
	driver   neo4j.DriverWithContext
	database string
}

func NewAccountRepo(driver neo4j.DriverWithContext, database string) *AccountRepo {
	return &AccountRepo{driver: driver, database: database}
}

func (r *AccountRepo) EnsureUniqueConstraints(ctx context.Context) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: r.database})
	defer func() {
		if err := session.Close(ctx); err != nil {
			log.Printf("cannot close neo4j session: %v", err)
		}
	}()

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		queries := []string{
			`CREATE CONSTRAINT account_username_unique IF NOT EXISTS
			 FOR (u:Account) REQUIRE u.username IS UNIQUE`,
			`CREATE CONSTRAINT account_email_unique IF NOT EXISTS
			 FOR (u:Account) REQUIRE u.email IS UNIQUE`,
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

func (r *AccountRepo) ExistsByUsernameOrEmail(ctx context.Context, username, email string) (bool, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: r.database})
	defer func() {
		if err := session.Close(ctx); err != nil {
			log.Printf("cannot close neo4j session: %v", err)
		}
	}()

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		res, err := tx.Run(ctx,
			`MATCH (u:Account)
			 WHERE toLower(u.username) = toLower($username) OR toLower(u.email) = toLower($email)
			 RETURN count(u) AS total`,
			map[string]any{"username": username, "email": email},
		)
		if err != nil {
			return nil, err
		}
		record, err := res.Single(ctx)
		if err != nil {
			return nil, err
		}
		total, _ := record.Get("total")
		count, _ := total.(int64)
		return count > 0, nil
	})
	if err != nil {
		return false, err
	}
	return result.(bool), nil
}

func (r *AccountRepo) CreateAccount(ctx context.Context, acc model.Account) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: r.database})
	defer func() {
		if err := session.Close(ctx); err != nil {
			log.Printf("cannot close neo4j session: %v", err)
		}
	}()

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		_, err := tx.Run(ctx,
			`CREATE (u:Account {
				username: $username,
				passwordHash: $passwordHash,
				email: $email,
				role: $role,
				isBlocked: false,
				createdAt: datetime()
			})`,
			map[string]any{
				"username":     acc.Username,
				"passwordHash": acc.PasswordHash,
				"email":        acc.Email,
				"role":         acc.Role,
			},
		)
		return nil, err
	})
	return err
}

func (r *AccountRepo) FindByIdentity(ctx context.Context, identity string) (*model.Account, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: r.database})
	defer func() {
		if err := session.Close(ctx); err != nil {
			log.Printf("cannot close neo4j session: %v", err)
		}
	}()

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		res, err := tx.Run(ctx,
			`MATCH (u:Account)
			 WHERE toLower(u.username) = toLower($identity) OR toLower(u.email) = toLower($identity)
			 RETURN u.username AS username, u.email AS email, u.role AS role,
			        u.isBlocked AS isBlocked, u.passwordHash AS passwordHash
			 LIMIT 1`,
			map[string]any{"identity": identity},
		)
		if err != nil {
			return nil, err
		}
		records, err := res.Collect(ctx)
		if err != nil {
			return nil, err
		}
		if len(records) == 0 {
			return nil, errors.New("invalid_credentials")
		}

		rec := records[0]
		username, _ := rec.Get("username")
		email, _ := rec.Get("email")
		role, _ := rec.Get("role")
		isBlocked, _ := rec.Get("isBlocked")
		passwordHash, _ := rec.Get("passwordHash")

		return &model.Account{
			Username:     username.(string),
			Email:        email.(string),
			Role:         role.(string),
			IsBlocked:    isBlocked.(bool),
			PasswordHash: passwordHash.(string),
		}, nil
	})
	if err != nil {
		return nil, err
	}
	return result.(*model.Account), nil
}

func (r *AccountRepo) GetAllAccounts(ctx context.Context) ([]map[string]any, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: r.database})
	defer func() {
		if err := session.Close(ctx); err != nil {
			log.Printf("cannot close neo4j session: %v", err)
		}
	}()

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		res, err := tx.Run(ctx,
			`MATCH (u:Account)
			 RETURN u.username AS username, u.email AS email, u.role AS role,
			        coalesce(u.isBlocked, false) AS isBlocked, u.createdAt AS createdAt
			 ORDER BY u.createdAt DESC`,
			nil,
		)
		if err != nil {
			return nil, err
		}
		records, err := res.Collect(ctx)
		if err != nil {
			return nil, err
		}

		accounts := make([]map[string]any, 0, len(records))
		for _, rec := range records {
			username, _ := rec.Get("username")
			email, _ := rec.Get("email")
			role, _ := rec.Get("role")
			isBlocked, _ := rec.Get("isBlocked")
			createdAt, _ := rec.Get("createdAt")
			accounts = append(accounts, map[string]any{
				"username":  username,
				"email":     email,
				"role":      role,
				"isBlocked": isBlocked,
				"createdAt": createdAt,
			})
		}
		return accounts, nil
	})
	if err != nil {
		return nil, err
	}
	return result.([]map[string]any), nil
}

func (r *AccountRepo) BlockAccount(ctx context.Context, username string) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: r.database})
	defer func() {
		if err := session.Close(ctx); err != nil {
			log.Printf("cannot close neo4j session: %v", err)
		}
	}()

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		res, err := tx.Run(ctx,
			`MATCH (u:Account)
			 WHERE toLower(u.username) = toLower($username)
			 RETURN u.role AS role LIMIT 1`,
			map[string]any{"username": username},
		)
		if err != nil {
			return nil, err
		}
		records, err := res.Collect(ctx)
		if err != nil {
			return nil, err
		}
		if len(records) == 0 {
			return nil, errors.New("account_not_found")
		}
		role, _ := records[0].Get("role")
		if role.(string) == "admin" {
			return nil, errors.New("cannot_block_admin")
		}

		_, err = tx.Run(ctx,
			`MATCH (u:Account)
			 WHERE toLower(u.username) = toLower($username)
			 SET u.isBlocked = true`,
			map[string]any{"username": username},
		)
		return nil, err
	})
	return err
}
