package service

import (
	"context"
	"stakeholders/model"
	"stakeholders/repo"
)

type ProfileService struct {
	repo *repo.ProfileRepo
}

func NewProfileService(r *repo.ProfileRepo) *ProfileService {
	return &ProfileService{repo: r}
}

func (s *ProfileService) GetProfile(ctx context.Context, username string) (*model.ProfileResponse, error) {
	p, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	return &model.ProfileResponse{
		Username:  p.Username,
		FirstName: p.FirstName,
		LastName:  p.LastName,
		ImageURL:  p.ImageURL,
		Bio:       p.Bio,
		Motto:     p.Motto,
	}, nil
}

func (s *ProfileService) UpdateProfile(ctx context.Context, username string, req model.UpdateProfileRequest) (*model.ProfileResponse, error) {
	if err := s.repo.Update(ctx, username, req); err != nil {
		return nil, err
	}
	return s.GetProfile(ctx, username)
}
