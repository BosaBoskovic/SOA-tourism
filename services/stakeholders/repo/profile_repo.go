package repo

import (
	"context"
	"errors"
	"stakeholders/model"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type ProfileRepo struct {
	driver neo4j.DriverWithContext
}

func NewProfileRepo(driver neo4j.DriverWithContext) *ProfileRepo {
	return &ProfileRepo{driver: driver}
}

func (r *ProfileRepo) CreateProfile(ctx context.Context, p model.Profile) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
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
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
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
		Username:  rec.Values[0].(string),
		FirstName: rec.Values[1].(string),
		LastName:  rec.Values[2].(string),
		ImageURL:  rec.Values[3].(string),
		Bio:       rec.Values[4].(string),
		Motto:     rec.Values[5].(string),
	}, nil
}

func (r *ProfileRepo) Update(ctx context.Context, username string, req model.UpdateProfileRequest) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
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
