package handlers

import (
	"pr-reviewer-service/internal/domain"
	"pr-reviewer-service/internal/generated/api/dto"
	"encoding/json"
	"net/http"
	"time"
)

const (
	ErrBadRequest dto.ErrorResponseErrorCode = "BAD_REQUEST"
	ErrInternal   dto.ErrorResponseErrorCode = "Internal error"
)

type TeamResponse struct {
	Team dto.Team `json:"team"`
}
type UserResponse struct {
	User dto.User `json:"user"`
}
type PullRequestsShortResponse struct {
	UserId            string                 `json:"user_id"`
	PullRequestsShort []dto.PullRequestShort `json:"pull_requests"`
}

type PullRequestCreateResponse struct {
	PR PullRequestCreate `json:"pr"`
}
type PullRequestCreate struct {
	PullRequestId     string                `json:"pull_request_id"`
	PullRequestName   string                `json:"pull_request_name"`
	AuthorId          string                `json:"author_id"`
	Status            dto.PullRequestStatus `json:"status"`
	AssignedReviewers []string              `json:"assigned_reviewers"`
}
type PullRequestMergeResponse struct {
	PR PullRequestWithMergedAt `json:"pr"`
}
type PullRequestWithMergedAt struct {
	PullRequestId     string                `json:"pull_request_id"`
	PullRequestName   string                `json:"pull_request_name"`
	AuthorId          string                `json:"author_id"`
	Status            dto.PullRequestStatus `json:"status"`
	AssignedReviewers []string              `json:"assigned_reviewers"`
	MergedAt          *time.Time            `json:"mergedAt"`
}
type PullRequestReassignResponse struct {
	PR       PullRequestCreate `json:"pr"`
	Replaced string            `json:"replaced_by"`
}

func FromTeamDTOToTeam(dto dto.Team) (domain.Team, error) {
	var tmpTeam domain.Team
	tmpTeam.TeamName = dto.TeamName
	for _, member := range dto.Members {
		tmpUser, err := domain.NewUser(domain.UserId(member.UserId), member.Username, dto.TeamName, member.IsActive)
		if err != nil {
			return domain.Team{}, err
		}
		tmpTeam.Members = append(tmpTeam.Members, *tmpUser)
	}
	return tmpTeam, nil
}
func FromTeamToDTO(team domain.Team) dto.Team {
	var tmpTeam dto.Team
	tmpTeam.TeamName = team.TeamName
	tmpTeam.Members = make([]dto.TeamMember, 0, len(team.Members))
	for _, member := range team.Members {
		tmpTeam.Members = append(tmpTeam.Members, dto.TeamMember{
			UserId:   string(member.UserId),
			Username: member.Username,
			IsActive: member.IsActive,
		})
	}
	return tmpTeam
}

func ResponseFormatError(w http.ResponseWriter, httpStatus int, errResponseCode dto.ErrorResponseErrorCode, massage string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	jsonResponse := dto.ErrorResponse{Error: struct {
		Code    dto.ErrorResponseErrorCode `json:"code"`
		Message string                     `json:"message"`
	}{
		Code:    errResponseCode,
		Message: massage,
	},
	}
	_ = json.NewEncoder(w).Encode(jsonResponse)
}
func ResponseFormatOK(w http.ResponseWriter, httpStatus int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)

	_ = json.NewEncoder(w).Encode(data)
}

func FromUserToUserDTO(user domain.User) dto.User {
	return dto.User{
		IsActive: user.IsActive,
		TeamName: user.TeamName,
		UserId:   string(user.UserId),
		Username: user.Username,
	}
}
