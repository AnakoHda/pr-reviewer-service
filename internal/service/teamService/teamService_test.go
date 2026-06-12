package teamService

import (
	"pr-reviewer-service/internal/domain"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUsersRepository struct {
	mock.Mock
}

func (m *MockUsersRepository) ListUsersByTeamName(ctx context.Context, teamName string) ([]domain.User, error) {
	args := m.Called(ctx, teamName)
	return args.Get(0).([]domain.User), args.Error(1)
}

type MockTeamRepository struct {
	mock.Mock
}

func (m *MockTeamRepository) CreateTeam(ctx context.Context, team domain.Team) error {
	args := m.Called(ctx, team)
	return args.Error(0)
}

func TestService_AddTeamWithUsers(t *testing.T) {
	ctx := context.Background()
	userRepo := new(MockUsersRepository)
	teamRepo := new(MockTeamRepository)
	svc := New(userRepo, teamRepo)

	team := domain.Team{TeamName: "team1"}
	users := []domain.User{{UserId: "user1", TeamName: "team1"}}

	t.Run("success", func(t *testing.T) {
		teamRepo.On("CreateTeam", ctx, team).Return(nil).Once()
		userRepo.On("ListUsersByTeamName", ctx, "team1").Return(users, nil).Once()

		res, err := svc.AddTeamWithUsers(ctx, team)

		assert.NoError(t, err)
		assert.Equal(t, "team1", res.TeamName)
		assert.Equal(t, users, res.Members)
		teamRepo.AssertExpectations(t)
		userRepo.AssertExpectations(t)
	})
}

func TestService_GetTeam(t *testing.T) {
	ctx := context.Background()
	userRepo := new(MockUsersRepository)
	teamRepo := new(MockTeamRepository)
	svc := New(userRepo, teamRepo)

	users := []domain.User{{UserId: "user1", TeamName: "team1"}}

	t.Run("success", func(t *testing.T) {
		userRepo.On("ListUsersByTeamName", ctx, "team1").Return(users, nil).Once()

		res, err := svc.GetTeam(ctx, "team1")

		assert.NoError(t, err)
		assert.Equal(t, "team1", res.TeamName)
		assert.Equal(t, users, res.Members)
		userRepo.AssertExpectations(t)
	})
}
