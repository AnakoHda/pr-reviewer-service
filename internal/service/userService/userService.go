package userService

import (
	"pr-reviewer-service/internal/domain"
	"context"
)

type UsersRepository interface {
	GetUserByID(ctx context.Context, userID domain.UserId) (*domain.User, error)
	UpdateUser(ctx context.Context, user domain.User) error
}
type PRRepository interface {
	ListPullRequestsByReviewerID(ctx context.Context, reviewerID domain.UserId) ([]domain.PullRequest, error)
}
type Service struct {
	userRepo UsersRepository
	prRepo   PRRepository
}

func New(usersRepository UsersRepository, prRepository PRRepository) *Service {
	return &Service{
		userRepo: usersRepository,
		prRepo:   prRepository,
	}
}

func (s *Service) SetIsActive(ctx context.Context, userId domain.UserId, isActive bool) (*domain.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userId)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}
	if user.IsActive != isActive {
		user.IsActive = isActive
		if err := s.userRepo.UpdateUser(ctx, *user); err != nil {
			return nil, err
		}
	}
	return user, nil
}

func (s *Service) GetReview(ctx context.Context, userId domain.UserId) ([]domain.PullRequest, error) {
	return s.prRepo.ListPullRequestsByReviewerID(ctx, userId)
}
