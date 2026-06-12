package storage

import (
	"pr-reviewer-service/internal/domain"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPullRequestsRepo_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := New(db)
	ctx := context.Background()

	// Setup: Создаем команду и пользователей для PR
	teamName := "core"
	author := domain.User{UserId: "author1", Username: "Alice", TeamName: teamName, IsActive: true}
	reviewer := domain.User{UserId: "rev1", Username: "Bob", TeamName: teamName, IsActive: true}
	err := repo.CreateTeam(ctx, domain.Team{TeamName: teamName, Members: []domain.User{author, reviewer}})
	assert.NoError(t, err)

	t.Run("Create and Get PR", func(t *testing.T) {
		prID := domain.PullRequestId("pr-123")
		pr, _ := domain.NewPullRequest(prID, "Initial PR", author.UserId, []domain.UserId{reviewer.UserId})

		err := repo.CreatePullRequest(ctx, *pr)
		assert.NoError(t, err)

		gotPR, err := repo.GetPullRequestByID(ctx, prID)
		assert.NoError(t, err)
		assert.Equal(t, pr.PullRequestId, gotPR.PullRequestId)
		assert.Equal(t, pr.PullRequestName, gotPR.PullRequestName)
		assert.Equal(t, pr.AuthorId, gotPR.AuthorId)
		assert.Contains(t, gotPR.AssignedReviewers, reviewer.UserId)
	})

	t.Run("Update PR (Assign new reviewers)", func(t *testing.T) {
		prID := domain.PullRequestId("pr-123")
		
		// Создаем еще одного ревьювера в НОВОЙ команде
		reviewer2 := domain.User{UserId: "rev2", Username: "Charlie", TeamName: "core-2", IsActive: true}
		err := repo.CreateTeam(ctx, domain.Team{TeamName: "core-2", Members: []domain.User{reviewer2}})
		assert.NoError(t, err)

		// Обновляем PR: меняем ревьювера
		pr, err := repo.GetPullRequestByID(ctx, prID)
		assert.NoError(t, err)
		pr.AssignedReviewers = []domain.UserId{reviewer2.UserId}
		
		err = repo.UpdatePullRequest(ctx, *pr)
		assert.NoError(t, err)

		// Проверяем, что старый удален, новый добавлен
		gotPR, err := repo.GetPullRequestByID(ctx, prID)
		assert.NoError(t, err)
		if gotPR != nil {
			assert.Len(t, gotPR.AssignedReviewers, 1)
			assert.Contains(t, gotPR.AssignedReviewers, reviewer2.UserId)
		}
	})

	t.Run("List PRs By Reviewer", func(t *testing.T) {
		reviewerID := domain.UserId("rev2")
		prs, err := repo.ListPullRequestsByReviewerID(ctx, reviewerID)
		assert.NoError(t, err)
		if assert.NotEmpty(t, prs) {
			assert.Equal(t, domain.PullRequestId("pr-123"), prs[0].PullRequestId)
		}
	})
}
