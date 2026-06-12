package domain

import "errors"

var ( //main errors
	ErrAuthorCannotBeReviewer   = errors.New("author cannot be a reviewer")
	ErrTooManyReviewersAssigned = errors.New("too many reviewers assigned to pull request (max:2)")
	ErrPullRequestMerged        = errors.New("pull request is already merged")

	ErrTeamAlreadyExists  = errors.New("team already exists")
	ErrTeamNotFound       = errors.New("team not found")
	ErrNoCandidatesInTeam = errors.New("no active replacement candidates in team")

	ErrUserNotFound = errors.New("user not found")

	ErrPullRequestAlreadyExists      = errors.New("pull request already exists")
	ErrPullRequestNotFound           = errors.New("pull request not found")
	ErrNotFoundReviewerInPullRequest = errors.New("reviewer not found in pull request")
	ErrNewReviewerAlreadyExist       = errors.New("newReviewer can`t be Actual Reviewer")
	ErrEmptyPullRequestID            = errors.New("pull request ID cannot be empty")
	ErrEmptyPullRequestName          = errors.New("pull request name cannot be empty")
	ErrEmptyAuthorID                 = errors.New("author ID cannot be empty")
	ErrEmptyUserID                   = errors.New("user ID cannot be empty")
	ErrEmptyUsername                 = errors.New("username cannot be empty")
	ErrEmptyTeamName                 = errors.New("team name cannot be empty")
)
