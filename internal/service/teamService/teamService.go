package teamService

import (
	"pr-reviewer-service/internal/domain"
	"context"
)

type UsersRepository interface {
	ListUsersByTeamName(ctx context.Context, teamName string) ([]domain.User, error)
}
type TeamRepository interface {
	CreateTeam(ctx context.Context, team domain.Team) error
}
type Service struct {
	userRepo UsersRepository
	teamRepo TeamRepository
}

func New(usersRepository UsersRepository, teamRepository TeamRepository) *Service {
	return &Service{
		userRepo: usersRepository,
		teamRepo: teamRepository,
	}
}

func (s *Service) AddTeamWithUsers(ctx context.Context, team domain.Team) (*domain.Team, error) {
	// создание команды
	if err := s.teamRepo.CreateTeam(ctx, team); err != nil {
		return nil, err
	}

	recordedUsers, err := s.userRepo.ListUsersByTeamName(ctx, team.TeamName)
	if err != nil {
		return nil, err
	}
	return domain.NewTeam(team.TeamName, recordedUsers)
}

func (s *Service) GetTeam(ctx context.Context, teamName string) (*domain.Team, error) {
	// существует ли команда и получить скисок
	existedUsersInTeam, err := s.userRepo.ListUsersByTeamName(ctx, teamName)
	if err != nil {
		return nil, err
	}
	return domain.NewTeam(teamName, existedUsersInTeam)
}
