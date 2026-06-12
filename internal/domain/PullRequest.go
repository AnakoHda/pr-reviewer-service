package domain

import (
	"time"
)

type PullRequestStatus string

const (
	PullRequestStatusOPEN   PullRequestStatus = "OPEN"
	PullRequestStatusMERGED PullRequestStatus = "MERGED"
)

type PullRequestId string
type PullRequest struct {
	PullRequestId     PullRequestId
	PullRequestName   string
	AuthorId          UserId
	Status            PullRequestStatus
	AssignedReviewers []UserId
	CreatedAt         time.Time
	MergedAt          time.Time
}

func NewPullRequest(pullRequestId PullRequestId, pullRequestName string, authorId UserId, assignedReviewers []UserId) (*PullRequest, error) {
	if err := ValidatePullRequestFields(pullRequestId, pullRequestName, authorId, assignedReviewers); err != nil {
		return nil, err
	}
	return &PullRequest{
		PullRequestId:     pullRequestId,
		PullRequestName:   pullRequestName,
		AuthorId:          authorId,
		Status:            PullRequestStatusOPEN,
		AssignedReviewers: assignedReviewers,
		CreatedAt:         time.Now(),
		MergedAt:          time.Time{},
	}, nil
}

// если изменился статус -> true
func (pr *PullRequest) Merge() bool {
	if pr.Status == PullRequestStatusMERGED {
		return false
	}

	pr.Status = PullRequestStatusMERGED
	pr.MergedAt = time.Now()
	return true
}

func (pr *PullRequest) ReplaceReviewer(oldReviewerId, newReviewerId UserId) error {
	if pr.Status == PullRequestStatusMERGED {
		return ErrPullRequestMerged
	}
	if newReviewerId == pr.AuthorId {
		return ErrAuthorCannotBeReviewer
	}
	for _, reviewerId := range pr.AssignedReviewers {
		if reviewerId == newReviewerId {
			return ErrNewReviewerAlreadyExist
		}
	}
	for i, findUserId := range pr.AssignedReviewers {
		if findUserId == oldReviewerId {
			pr.AssignedReviewers[i] = newReviewerId
			return nil
		}
	}

	return ErrNotFoundReviewerInPullRequest
}

func ValidatePullRequestFields(pullRequestId PullRequestId, pullRequestName string, authorId UserId, assignedReviewers []UserId) error {
	if pullRequestId == "" {
		return ErrEmptyPullRequestID
	}
	if pullRequestName == "" {
		return ErrEmptyPullRequestName
	}
	if authorId == "" {
		return ErrEmptyAuthorID
	}
	if len(assignedReviewers) > 2 {
		return ErrTooManyReviewersAssigned
	}
	for _, reviewer := range assignedReviewers {
		if reviewer == authorId {
			return ErrAuthorCannotBeReviewer
		}
	}
	return nil
}
