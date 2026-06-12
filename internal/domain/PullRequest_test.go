package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPullRequest(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		id := PullRequestId("pr-1")
		name := "Fix bug"
		author := UserId("user-1")
		reviewers := []UserId{"user-2"}

		pr, err := NewPullRequest(id, name, author, reviewers)

		assert.NoError(t, err)
		assert.NotNil(t, pr)
		assert.Equal(t, id, pr.PullRequestId)
		assert.Equal(t, name, pr.PullRequestName)
		assert.Equal(t, author, pr.AuthorId)
		assert.Equal(t, reviewers, pr.AssignedReviewers)
		assert.Equal(t, PullRequestStatusOPEN, pr.Status)
		assert.False(t, pr.CreatedAt.IsZero())
		assert.True(t, pr.MergedAt.IsZero())
	})

	t.Run("empty id", func(t *testing.T) {
		_, err := NewPullRequest("", "Fix bug", "user-1", []UserId{"user-2"})
		assert.ErrorIs(t, err, ErrEmptyPullRequestID)
	})

	t.Run("too many reviewers", func(t *testing.T) {
		_, err := NewPullRequest("pr-1", "Fix bug", "user-1", []UserId{"user-2", "user-3", "user-4"})
		assert.ErrorIs(t, err, ErrTooManyReviewersAssigned)
	})

	t.Run("author as reviewer", func(t *testing.T) {
		_, err := NewPullRequest("pr-1", "Fix bug", "user-1", []UserId{"user-1"})
		assert.ErrorIs(t, err, ErrAuthorCannotBeReviewer)
	})
}

func TestPullRequest_Merge(t *testing.T) {
	t.Run("success merge", func(t *testing.T) {
		pr := &PullRequest{
			Status: PullRequestStatusOPEN,
		}
		changed := pr.Merge()
		assert.True(t, changed)
		assert.Equal(t, PullRequestStatusMERGED, pr.Status)
		assert.False(t, pr.MergedAt.IsZero())
	})

	t.Run("already merged", func(t *testing.T) {
		pr := &PullRequest{
			Status: PullRequestStatusMERGED,
		}
		changed := pr.Merge()
		assert.False(t, changed)
		assert.Equal(t, PullRequestStatusMERGED, pr.Status)
	})
}

func TestPullRequest_ReplaceReviewer(t *testing.T) {
	pr := &PullRequest{
		AuthorId:          "author",
		Status:            PullRequestStatusOPEN,
		AssignedReviewers: []UserId{"reviewer-1"},
	}

	t.Run("success", func(t *testing.T) {
		err := pr.ReplaceReviewer("reviewer-1", "reviewer-2")
		assert.NoError(t, err)
		assert.Contains(t, pr.AssignedReviewers, UserId("reviewer-2"))
		assert.NotContains(t, pr.AssignedReviewers, UserId("reviewer-1"))
	})

	t.Run("new reviewer is author", func(t *testing.T) {
		err := pr.ReplaceReviewer("reviewer-2", "author")
		assert.ErrorIs(t, err, ErrAuthorCannotBeReviewer)
	})

	t.Run("new reviewer already assigned", func(t *testing.T) {
		pr.AssignedReviewers = []UserId{"reviewer-2", "reviewer-3"}
		err := pr.ReplaceReviewer("reviewer-2", "reviewer-3")
		assert.ErrorIs(t, err, ErrNewReviewerAlreadyExist)
	})

	t.Run("old reviewer not found", func(t *testing.T) {
		err := pr.ReplaceReviewer("non-existent", "reviewer-4")
		assert.ErrorIs(t, err, ErrNotFoundReviewerInPullRequest)
	})

	t.Run("pr already merged", func(t *testing.T) {
		pr.Status = PullRequestStatusMERGED
		err := pr.ReplaceReviewer("reviewer-2", "reviewer-4")
		assert.ErrorIs(t, err, ErrPullRequestMerged)
	})
}
