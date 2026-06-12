package userService

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

func (m *MockUsersRepository) GetUserByID(ctx context.Context, userID domain.UserId) (*domain.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUsersRepository) UpdateUser(ctx context.Context, user domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

type MockPRRepository struct {
	mock.Mock
}

func (m *MockPRRepository) ListPullRequestsByReviewerID(ctx context.Context, reviewerID domain.UserId) ([]domain.PullRequest, error) {
	args := m.Called(ctx, reviewerID)
	return args.Get(0).([]domain.PullRequest), args.Error(1)
}

func TestService_SetIsActive(t *testing.T) {
	ctx := context.Background()
	userRepo := new(MockUsersRepository)
	prRepo := new(MockPRRepository)
	svc := New(userRepo, prRepo)

	userID := domain.UserId("user1")

	t.Run("success change", func(t *testing.T) {
		user := &domain.User{UserId: userID, IsActive: true}
		userRepo.On("GetUserByID", ctx, userID).Return(user, nil).Once()
		userRepo.On("UpdateUser", ctx, mock.MatchedBy(
			func(u domain.User) bool {
				return u.IsActive == false
			})).Return(nil).Once()

		res, err := svc.SetIsActive(ctx, userID, false)

		assert.NoError(t, err)
		assert.False(t, res.IsActive)
		userRepo.AssertExpectations(t)
	})

	t.Run("no change", func(t *testing.T) {
		user := &domain.User{UserId: userID, IsActive: true}
		userRepo.On("GetUserByID", ctx, userID).Return(user, nil).Once()

		res, err := svc.SetIsActive(ctx, userID, true)

		assert.NoError(t, err)
		assert.True(t, res.IsActive)
		userRepo.AssertExpectations(t)
	})
}

func TestService_GetReview(t *testing.T) {
	ctx := context.Background()
	userRepo := new(MockUsersRepository)
	prRepo := new(MockPRRepository)
	svc := New(userRepo, prRepo)

	userID := domain.UserId("user1")
	prs := []domain.PullRequest{{PullRequestId: "pr-1"}}

	t.Run("success", func(t *testing.T) {
		prRepo.On("ListPullRequestsByReviewerID", ctx, userID).Return(prs, nil).Once()

		res, err := svc.GetReview(ctx, userID)

		assert.NoError(t, err)
		assert.Equal(t, prs, res)
		prRepo.AssertExpectations(t)
	})
}
