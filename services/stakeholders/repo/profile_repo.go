package repo

import (
	"context"
	"errors"
	"stakeholders/model"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type ProfileRepo struct {
	driver   neo4j.DriverWithContext
	database string
}

func NewProfileRepo(driver neo4j.DriverWithContext, database string) *ProfileRepo {
	return &ProfileRepo{driver: driver, database: database}
}

func (r *ProfileRepo) CreateProfile(ctx context.Context, p model.Profile) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite, DatabaseName: r.database})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		_, err := tx.Run(ctx, `
            MATCH (a:Account {username: $username})
            CREATE (a)-[:HAS_PROFILE]->(p:Profile {
                username:  $username,
                firstName: "",
                lastName:  "",
                imageURL:  "",
                bio:       "",
                motto:     ""
            })
        `, map[string]any{
			"username": p.Username,
		})
		return nil, err
	})
	return err
}

func (r *ProfileRepo) GetByUsername(ctx context.Context, username string) (*model.Profile, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead, DatabaseName: r.database})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		rec, err := tx.Run(ctx, `
			MATCH (:Account {username: $username})-[:HAS_PROFILE]->(p:Profile)
			RETURN p.username AS username, p.firstName AS firstName, p.lastName AS lastName,
       		p.imageURL AS imageURL, p.bio AS bio, p.motto AS motto
        `, map[string]any{"username": username})
		if err != nil {
			return nil, err
		}
		if rec.Next(ctx) {
			return rec.Record(), nil
		}
		return nil, errors.New("profile_not_found")
	})
	if err != nil {
		return nil, err
	}
	rec := result.(*neo4j.Record)
	return &model.Profile{
		Username:  asString(rec.Values[0]),
		FirstName: asString(rec.Values[1]),
		LastName:  asString(rec.Values[2]),
		ImageURL:  asString(rec.Values[3]),
		Bio:       asString(rec.Values[4]),
		Motto:     asString(rec.Values[5]),
	}, nil
}

func (r *ProfileRepo) Update(ctx context.Context, username string, req model.UpdateProfileRequest) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite, DatabaseName: r.database})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		res, err := tx.Run(ctx, `
			MATCH (:Account {username: $username})-[:HAS_PROFILE]->(p:Profile)
            SET p.firstName = $firstName,
                p.lastName  = $lastName,
                p.imageURL  = $imageURL,
                p.bio       = $bio,
                p.motto     = $motto
		`, map[string]any{
			"username":  username,
			"firstName": req.FirstName,
			"lastName":  req.LastName,
			"imageURL":  req.ImageURL,
			"bio":       req.Bio,
			"motto":     req.Motto,
		})
		if err != nil {
			return nil, err
		}
		summary, err := res.Consume(ctx)
		if err != nil {
			return nil, err
		}
		if summary.Counters().PropertiesSet() == 0 {
			return nil, errors.New("profile_not_found")
		}
		return nil, nil
	})
	return err
}

func (r *ProfileRepo) GetPublicByUsername(ctx context.Context, username string) (*model.PublicProfileResponse, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead, DatabaseName: r.database})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		rec, err := tx.Run(ctx, `
			MATCH (a:Account {username: $username})-[:HAS_PROFILE]->(p:Profile)
			RETURN p.username AS username, p.firstName AS firstName, p.lastName AS lastName,
			       p.imageURL AS imageURL, p.bio AS bio, p.motto AS motto, a.role AS role
		`, map[string]any{"username": username})
		if err != nil {
			return nil, err
		}
		if rec.Next(ctx) {
			return rec.Record(), nil
		}
		return nil, errors.New("profile_not_found")
	})
	if err != nil {
		return nil, err
	}
	rec := result.(*neo4j.Record)
	return &model.PublicProfileResponse{
		Username:  asString(rec.Values[0]),
		FirstName: asString(rec.Values[1]),
		LastName:  asString(rec.Values[2]),
		ImageURL:  asString(rec.Values[3]),
		Bio:       asString(rec.Values[4]),
		Motto:     asString(rec.Values[5]),
		Role:      asString(rec.Values[6]),
	}, nil
}

func (r *ProfileRepo) Search(ctx context.Context, username, role string, limit int) ([]model.PublicProfileResponse, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead, DatabaseName: r.database})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		res, err := tx.Run(ctx, `
			MATCH (a:Account)-[:HAS_PROFILE]->(p:Profile)
			WHERE ($username = "" OR toLower(a.username) CONTAINS toLower($username))
			  AND ($role = "" OR toLower(a.role) = toLower($role))
			RETURN p.username AS username, p.firstName AS firstName, p.lastName AS lastName,
			       p.imageURL AS imageURL, p.bio AS bio, p.motto AS motto, a.role AS role
			ORDER BY a.username
			LIMIT $limit
		`, map[string]any{
			"username": username,
			"role":     role,
			"limit":    limit,
		})
		if err != nil {
			return nil, err
		}
		records, err := res.Collect(ctx)
		if err != nil {
			return nil, err
		}
		profiles := make([]model.PublicProfileResponse, 0, len(records))
		for _, rec := range records {
			profiles = append(profiles, model.PublicProfileResponse{
				Username:  asString(rec.Values[0]),
				FirstName: asString(rec.Values[1]),
				LastName:  asString(rec.Values[2]),
				ImageURL:  asString(rec.Values[3]),
				Bio:       asString(rec.Values[4]),
				Motto:     asString(rec.Values[5]),
				Role:      asString(rec.Values[6]),
			})
		}
		return profiles, nil
	})
	if err != nil {
		return nil, err
	}
	return result.([]model.PublicProfileResponse), nil
}
