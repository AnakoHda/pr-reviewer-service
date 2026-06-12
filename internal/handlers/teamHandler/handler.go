package teamHandler

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
	AddTeamWithUsers(ctx context.Context, team domain.Team) (*domain.Team, error)
	GetTeam(ctx context.Context, teamName string) (*domain.Team, error)
}
type Handler struct {
	svc Service
}

func New(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/team/add", h.POSTAddTeam)
	mux.HandleFunc("/team/get", h.GETGetTeam)
}

func (h *Handler) POSTAddTeam(w http.ResponseWriter, r *http.Request) {
	var req dto.PostTeamAddJSONRequestBody
	slog.Info("Touch ADD TEAM")
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.ResponseFormatError(w, http.StatusBadRequest, handlers.ErrBadRequest, "Decode error")
		return
	}
	tmpTeam, err := handlers.FromTeamDTOToTeam(req)
	if err != nil {
		handlers.ResponseFormatError(w, http.StatusBadRequest, handlers.ErrBadRequest, "Format error")
		return
	}

	ctx := r.Context()
	actualTeam, err := h.svc.AddTeamWithUsers(ctx, tmpTeam)

	if err != nil {
		if err.Error() == domain.ErrTeamAlreadyExists.Error() {
			handlers.ResponseFormatError(w, http.StatusBadRequest, dto.TEAMEXISTS, "team_name already exists")
			return
		}
		handlers.ResponseFormatError(w, http.StatusInternalServerError, handlers.ErrInternal, "service error")
		return
	}

	responseTeamDTO := handlers.FromTeamToDTO(*actualTeam)
	handlers.ResponseFormatOK(w, http.StatusCreated, handlers.TeamResponse{Team: responseTeamDTO})
}

func (h *Handler) GETGetTeam(w http.ResponseWriter, r *http.Request) {
	slog.Info("Touch GET TEAM")
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		handlers.ResponseFormatError(w, http.StatusBadRequest, handlers.ErrBadRequest, "Decode error")
		return
	}
	ctx := r.Context()
	team, err := h.svc.GetTeam(ctx, teamName)
	if err != nil {
		if err.Error() == domain.ErrTeamNotFound.Error() {
			handlers.ResponseFormatError(w, http.StatusNotFound, dto.NOTFOUND, "resource not found")
			return
		}
		handlers.ResponseFormatError(w, http.StatusInternalServerError, handlers.ErrInternal, "service error")
		return
	}
	responseTeamDTO := handlers.FromTeamToDTO(*team)
	handlers.ResponseFormatOK(w, http.StatusOK, responseTeamDTO)
}
