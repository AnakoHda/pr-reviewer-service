package pullRequestService

import (
	"pr-reviewer-service/internal/domain"
	"context"
	"errors"
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

func (m *MockUsersRepository) ListUsersByTeamName(ctx context.Context, teamName string) ([]domain.User, error) {
	args := m.Called(ctx, teamName)
	return args.Get(0).([]domain.User), args.Error(1)
}

type MockPRRepository struct {
	mock.Mock
}

func (m *MockPRRepository) GetPullRequestByID(ctx context.Context, pullRequestId domain.PullRequestId) (*domain.PullRequest, error) {
	args := m.Called(ctx, pullRequestId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PullRequest), args.Error(1)
}

func (m *MockPRRepository) CreatePullRequest(ctx context.Context, pullRequest domain.PullRequest) error {
	args := m.Called(ctx, pullRequest)
	return args.Error(0)
}

func (m *MockPRRepository) UpdatePullRequest(ctx context.Context, pullRequest domain.PullRequest) error {
	args := m.Called(ctx, pullRequest)
	return args.Error(0)
}

func TestService_CreatePullRequest(t *testing.T) {
	ctx := context.Background()
	userRepo := new(MockUsersRepository)
	prRepo := new(MockPRRepository)
	svc := New(userRepo, prRepo)

	authorID := domain.UserId("author")
	author := &domain.User{UserId: authorID, TeamName: "team1", IsActive: true}
	teamMembers := []domain.User{
		{UserId: authorID, TeamName: "team1", IsActive: true},
		{UserId: "reviewer1", TeamName: "team1", IsActive: true},
		{UserId: "reviewer2", TeamName: "team1", IsActive: true},
		{UserId: "reviewer3", TeamName: "team1", IsActive: false},
	}

	t.Run("success", func(t *testing.T) {
		userRepo.On("GetUserByID", ctx, authorID).Return(author, nil).Once()
		userRepo.On("ListUsersByTeamName", ctx, "team1").Return(teamMembers, nil).Once()
		prRepo.On("CreatePullRequest", ctx, mock.Anything).Return(nil).Once()

		pr, err := svc.CreatePullRequest(ctx, "pr-1", "My PR", authorID)

		assert.NoError(t, err)
		assert.NotNil(t, pr)
		assert.Equal(t, 2, len(pr.AssignedReviewers))
		assert.Contains(t, pr.AssignedReviewers, domain.UserId("reviewer1"))
		assert.Contains(t, pr.AssignedReviewers, domain.UserId("reviewer2"))
		userRepo.AssertExpectations(t)
		prRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		userRepo.On("GetUserByID", ctx, authorID).Return(nil, errors.New("not found")).Once()
		_, err := svc.CreatePullRequest(ctx, "pr-1", "My PR", authorID)
		assert.Error(t, err)
		userRepo.AssertExpectations(t)
	})
}

func TestService_Merge(t *testing.T) {
	ctx := context.Background()
	userRepo := new(MockUsersRepository)
	prRepo := new(MockPRRepository)
	svc := New(userRepo, prRepo)

	prID := domain.PullRequestId("pr-1")
	pr := &domain.PullRequest{PullRequestId: prID, Status: domain.PullRequestStatusOPEN}

	t.Run("success", func(t *testing.T) {
		prRepo.On("GetPullRequestByID", ctx, prID).Return(pr, nil).Once()
		prRepo.On("UpdatePullRequest", ctx, mock.MatchedBy(func(p domain.PullRequest) bool {
			return p.Status == domain.PullRequestStatusMERGED
		})).Return(nil).Once()

		res, err := svc.Merge(ctx, prID)
		assert.NoError(t, err)
		assert.Equal(t, domain.PullRequestStatusMERGED, res.Status)
		prRepo.AssertExpectations(t)
	})

	t.Run("already merged", func(t *testing.T) {
		mergedPR := &domain.PullRequest{PullRequestId: prID, Status: domain.PullRequestStatusMERGED}
		prRepo.On("GetPullRequestByID", ctx, prID).Return(mergedPR, nil).Once()

		res, err := svc.Merge(ctx, prID)
		assert.NoError(t, err)
		assert.Equal(t, domain.PullRequestStatusMERGED, res.Status)
		prRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		prRepo.On("GetPullRequestByID", ctx, prID).Return(nil, errors.New("not found")).Once()
		_, err := svc.Merge(ctx, prID)
		assert.ErrorIs(t, err, domain.ErrPullRequestNotFound)
		prRepo.AssertExpectations(t)
	})
}

func TestService_Reassign(t *testing.T) {
	ctx := context.Background()
	userRepo := new(MockUsersRepository)
	prRepo := new(MockPRRepository)
	svc := New(userRepo, prRepo)

	prID := domain.PullRequestId("pr-1")
	authorID := domain.UserId("author")
	oldReviewerID := domain.UserId("reviewer1")
	newReviewerID := domain.UserId("reviewer2")

	pr := &domain.PullRequest{
		PullRequestId:     prID,
		AuthorId:          authorID,
		Status:            domain.PullRequestStatusOPEN,
		AssignedReviewers: []domain.UserId{oldReviewerID},
	}

	oldReviewer := &domain.User{UserId: oldReviewerID, TeamName: "team1"}
	teamMembers := []domain.User{
		{UserId: oldReviewerID, TeamName: "team1", IsActive: true},
		{UserId: authorID, TeamName: "team1", IsActive: true},
		{UserId: newReviewerID, TeamName: "team1", IsActive: true},
	}

	t.Run("success", func(t *testing.T) {
		prRepo.On("GetPullRequestByID", ctx, prID).Return(pr, nil).Once()
		userRepo.On("GetUserByID", ctx, oldReviewerID).Return(oldReviewer, nil).Once()
		userRepo.On("ListUsersByTeamName", ctx, "team1").Return(teamMembers, nil).Once()
		prRepo.On("UpdatePullRequest", ctx, mock.Anything).Return(nil).Once()

		updatedPR, reassignedID, err := svc.Reassign(ctx, prID, oldReviewerID)

		assert.NoError(t, err)
		assert.Equal(t, newReviewerID, *reassignedID)
		assert.Contains(t, updatedPR.AssignedReviewers, newReviewerID)
		assert.NotContains(t, updatedPR.AssignedReviewers, oldReviewerID)
		prRepo.AssertExpectations(t)
		userRepo.AssertExpectations(t)
	})

	t.Run("no candidates", func(t *testing.T) {
		prRepo.On("GetPullRequestByID", ctx, prID).Return(pr, nil).Once()
		userRepo.On("GetUserByID", ctx, oldReviewerID).Return(oldReviewer, nil).Once()
		userRepo.On("ListUsersByTeamName", ctx, "team1").Return([]domain.User{*oldReviewer}, nil).Once()

		_, _, err := svc.Reassign(ctx, prID, oldReviewerID)
		assert.ErrorIs(t, err, domain.ErrNoCandidatesInTeam)
		prRepo.AssertExpectations(t)
		userRepo.AssertExpectations(t)
	})
}
