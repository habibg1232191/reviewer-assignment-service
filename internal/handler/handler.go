// Package handler
package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"reviewer-assignment-service/internal/domain"
	"reviewer-assignment-service/internal/dto"
	"reviewer-assignment-service/internal/usecase"

	"github.com/gorilla/mux"
)

type Handler struct {
	service usecase.ReviewerSerivice
}

func (h *Handler) writeError(w http.ResponseWriter, status int, code string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(dto.ErrorResponse{
		Error: dto.ErrorResponseDetail{
			Code:    code,
			Message: message,
		},
	})
}

func (h *Handler) writeSuccess(w http.ResponseWriter, obj any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(obj)
}

func NewHandler(s usecase.ReviewerSerivice) *Handler {
	return &Handler{service: s}
}

func (h *Handler) RegisterPublicRoutes(r *mux.Router) {
	r.HandleFunc("/team/add", h.TeamAdd).Methods("POST")
	r.HandleFunc("/team/get", h.GetTeam).Methods("GET")

	r.HandleFunc("/user/setIsActive", h.UserSetIsActive).Methods("GET")
}

func (h *Handler) TeamAdd(w http.ResponseWriter, r *http.Request) {
	var teamReq dto.TeamReq
	if err := json.NewDecoder(r.Body).Decode(&teamReq); err != nil {
		h.writeError(w, http.StatusBadRequest, "", "bad request")
		return
	}

	if teamReq.TeamName == "" || len(teamReq.Members) <= 0 {
		h.writeError(w, http.StatusBadRequest, "", "bad request")
		return
	}

	teamMembers := make([]domain.TeamMember, 0, len(teamReq.Members))
	for _, member := range teamReq.Members {
		teamMembers = append(teamMembers, domain.TeamMember{
			UserID:   member.UserID,
			UserName: member.UserName,
			IsActive: member.IsActive,
		})
	}
	err := h.service.CreateTeam(r.Context(), &domain.Team{
		Name:    teamReq.TeamName,
		Members: teamMembers,
	})
	if err != nil {
		if errors.Is(err, pkgerrors.ErrTeamExist) {
			h.writeError(w, http.StatusBadRequest, dto.ErrTeamExist, "team_name already exists")
		} else {
			h.writeError(w, http.StatusInternalServerError, "", "failed create team")
		}
		return
	}

	h.writeSuccess(w, dto.TeamResponse{
		Team: teamReq,
	})
}

func (h *Handler) GetTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		h.writeError(w, http.StatusNotFound, dto.ErrNotFound, "resource not found")
		return
	}

	team, err := h.service.GetTeam(r.Context(), teamName)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrNotFound) {
			h.writeError(w, http.StatusNotFound, dto.ErrNotFound, "resource not found")
		} else {
			h.writeError(w, http.StatusInternalServerError, "", "internal server error")
		}
	}

	h.writeSuccess(w, *team)
}

func (h *Handler) UserSetIsActive(w http.ResponseWriter, r *http.Request) {
	var userSetIsActiveReq dto.UserSetIsActiveReq
	if err := json.NewDecoder(r.Body).Decode(&userSetIsActiveReq); err != nil {
		h.writeError(w, http.StatusBadRequest, "", "not valid request body")
	}

	user, err := h.service.UserSetIsActive(r.Context(), userSetIsActiveReq.UserID, userSetIsActiveReq.IsActive)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrNotFound) {
			h.writeError(w, http.StatusNotFound, dto.ErrNotFound, "resource not found")
		} else {
			h.writeError(w, http.StatusInternalServerError, "", "internal server error")
		}
	}

	h.writeSuccess(w, *user)
}
