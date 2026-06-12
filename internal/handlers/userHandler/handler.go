package userHandler

import (
	"pr-reviewer-service/internal/domain"
	"pr-reviewer-service/internal/generated/api/dto"
	"pr-reviewer-service/internal/handlers"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
)

type Service interface {
	SetIsActive(ctx context.Context, userId domain.UserId, isActive bool) (*domain.User, error)
	GetReview(ctx context.Context, userId domain.UserId) ([]domain.PullRequest, error)
}
type Handler struct {
	svc Service
}

func New(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/users/setIsActive", h.POSTSetIsActiveUser)
	mux.HandleFunc("/users/getReview", h.GETGetReviewUser)
}

func (h *Handler) POSTSetIsActiveUser(w http.ResponseWriter, r *http.Request) {
	slog.Info("Touch SetIsActive USERS")
	var req dto.PostUsersSetIsActiveJSONBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.ResponseFormatError(w, http.StatusBadRequest, handlers.ErrBadRequest, "Decode error")
		return
	}
	ctx := r.Context()
	user, err := h.svc.SetIsActive(ctx, domain.UserId(req.UserId), req.IsActive)
	if err != nil {
		if err.Error() == domain.ErrUserNotFound.Error() {
			handlers.ResponseFormatError(w, http.StatusNotFound, dto.NOTFOUND, "resource not found")
			return
		}
		handlers.ResponseFormatError(w, http.StatusInternalServerError, handlers.ErrInternal, "service error")
		return
	}
	dtoUser := handlers.FromUserToUserDTO(*user)
	handlers.ResponseFormatOK(w, http.StatusOK, handlers.UserResponse{User: dtoUser})
}

func (h *Handler) GETGetReviewUser(w http.ResponseWriter, r *http.Request) {
	slog.Info("Touch GET USERS Reviews")
	userId := r.URL.Query().Get("user_id")
	if userId == "" {
		handlers.ResponseFormatError(w, http.StatusBadRequest, handlers.ErrBadRequest, "Decode error")
		return
	}
	ctx := r.Context()
	massPR, err := h.svc.GetReview(ctx, domain.UserId(userId))
	if err != nil {
		if err.Error() == domain.ErrUserNotFound.Error() {
			handlers.ResponseFormatError(w, http.StatusNotFound, dto.NOTFOUND, "resource not found")
			return
		}
		handlers.ResponseFormatError(w, http.StatusInternalServerError, handlers.ErrInternal, "service error")
		return
	}

	var res handlers.PullRequestsShortResponse
	res.UserId = userId
	res.PullRequestsShort = make([]dto.PullRequestShort, 0, len(massPR))
	for _, pr := range massPR {
		res.PullRequestsShort = append(res.PullRequestsShort, dto.PullRequestShort{
			PullRequestId:   string(pr.PullRequestId),
			PullRequestName: pr.PullRequestName,
			AuthorId:        string(pr.AuthorId),
			Status:          dto.PullRequestShortStatus(pr.Status),
		})
	}
	handlers.ResponseFormatOK(w, http.StatusOK, res)
}
