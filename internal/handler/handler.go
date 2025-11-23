// Package handler
package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/habibg1232191/reviewer-assignment-service/internal/domain"
	"github.com/habibg1232191/reviewer-assignment-service/internal/dto"
	"github.com/habibg1232191/reviewer-assignment-service/internal/usecase"

	pkgerrors "github.com/habibg1232191/reviewer-assignment-service/pgk/errors"
)

var (
	GET  string = "GET"
	POST string = "POST"
)

type Handler struct {
	service usecase.ReviewerService
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
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(obj)
}

func NewHandler(s usecase.ReviewerService) *Handler {
	return &Handler{service: s}
}

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) RegisterPublicRoutes(r *mux.Router) {
	r.Use(recoverMiddleware)
	r.HandleFunc("/team/add", h.TeamAdd).Methods(POST)
	r.HandleFunc("/team/get", h.GetTeam).Methods(GET)

	r.HandleFunc("/users/setIsActive", h.UserSetIsActive).Methods(POST)
	r.HandleFunc("/users/getReview", h.UserGetReview).Methods(GET)

	r.HandleFunc("/pullRequest/create", h.PullRequestCreate).Methods(POST)
	r.HandleFunc("/pullRequest/merge", h.PullRequestMerge).Methods(POST)
	r.HandleFunc("/pullRequest/reassign", h.PullRequestReassign).Methods(POST)
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

	slog.Info("team add", "team request", teamReq)
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
		return
	}
	teamMembers := make([]dto.TeamMember, 0, len(team.Members))
	for _, member := range team.Members {
		teamMembers = append(teamMembers, dto.TeamMember{
			UserID:   member.UserID,
			UserName: member.UserName,
			IsActive: member.IsActive,
		})
	}
	slog.Info("len members", "members", len(team.Members), "teamMembers", len(teamMembers))
	teamDto := dto.TeamReq{
		TeamName: team.Name,
		Members:  teamMembers,
	}

	h.writeSuccess(w, teamDto)
}

func (h *Handler) UserSetIsActive(w http.ResponseWriter, r *http.Request) {
	var userSetIsActiveReq dto.UserSetIsActiveReq
	if err := json.NewDecoder(r.Body).Decode(&userSetIsActiveReq); err != nil {
		h.writeError(w, http.StatusBadRequest, "", "not valid request body")
		return
	}

	user, err := h.service.UserSetIsActive(r.Context(), userSetIsActiveReq.UserID, userSetIsActiveReq.IsActive)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrNotFound) {
			h.writeError(w, http.StatusNotFound, dto.ErrNotFound, "resource not found")
		} else {
			h.writeError(w, http.StatusInternalServerError, "", "internal server error")
		}
		return
	}

	h.writeSuccess(w, *user)
}

func (h *Handler) UserGetReview(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		h.writeError(w, http.StatusBadRequest, "", "user_id is not correct")
		return
	}

	prs, err := h.service.UserGetPR(r.Context(), userID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "", "")
		return
	}
	prsDto := make([]dto.PullRequestShort, 0, len(prs))
	for _, pr := range prs {
		prsDto = append(prsDto, dto.PullRequestShort{
			PullRequestID:   pr.ID,
			PullRequestName: pr.Name,
			AuthorID:        pr.AuthorID,
			Status:          pr.Status,
		})
	}

	h.writeSuccess(w, dto.UserGetPRReq{UserID: userID, PullRequest: prsDto})
}

func (h *Handler) PullRequestCreate(w http.ResponseWriter, r *http.Request) {
	var pullRequestReq dto.PullRequestReq

	if err := json.NewDecoder(r.Body).Decode(&pullRequestReq); err != nil {
		h.writeError(w, http.StatusBadRequest, "", "not valid request body")
		return
	}

	prCreated, err := h.service.CreatePullRequest(r.Context(), pullRequestReq.PullRequestID, pullRequestReq.PullRequestName, pullRequestReq.AuthorID)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrPrExist) {
			h.writeError(w, http.StatusConflict, dto.ErrPrExist, "PR id already exists")
		} else if errors.Is(err, pkgerrors.ErrNotFound) {
			h.writeError(w, http.StatusNotFound, dto.ErrNotFound, "resource not found")
		} else {
			h.writeError(w, http.StatusInternalServerError, dto.ErrUndefined, "internal server error")
		}
		return
	}

	pr1 := prCreated.PR

	pr := dto.PullRequest{
		PullRequestID:     pr1.ID,
		PullRequestName:   pr1.Name,
		AuthorID:          pr1.AuthorID,
		Status:            pr1.Status,
		AssignedReviewers: prCreated.AssignmentReviewers,
		CreatedAt:         pr1.CreatedAt,
		MergedAt:          pr1.MergedAt,
	}

	h.writeSuccess(w, dto.PullRequestResponse{PR: pr})
}

func (h *Handler) PullRequestMerge(w http.ResponseWriter, r *http.Request) {
	type prReq struct {
		PullRequestID string `json:"pull_request_id"`
	}

	var pr prReq
	if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
		h.writeError(w, http.StatusBadRequest, "", "bad request")
		return
	}

	prs, err := h.service.MarkMerged(r.Context(), pr.PullRequestID)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrNotFound) {
			h.writeError(w, http.StatusNotFound, dto.ErrNotFound, "resource not found")
			return
		}
		h.writeError(w, http.StatusInternalServerError, "", "internal server error")
		return
	}

	h.writeSuccess(w, dto.PullRequestResponse{
		PR: dto.PullRequest{
			PullRequestID:     prs.PR.ID,
			PullRequestName:   prs.PR.Name,
			AuthorID:          prs.PR.AuthorID,
			Status:            prs.PR.Status,
			CreatedAt:         prs.PR.CreatedAt,
			MergedAt:          prs.PR.MergedAt,
			AssignedReviewers: prs.AssignmentReviewers,
		},
	})
}

func (h *Handler) PullRequestReassign(w http.ResponseWriter, r *http.Request) {
	var req dto.PullRequestReassignReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "", "invalid request body")
		return
	}

	if req.PullRequestID == "" || req.OldReviewerID == "" {
		h.writeError(w, http.StatusBadRequest, "", "pull_request_id and old_reviewer_id are required")
		return
	}

	pr, newReviewerID, err := h.service.ReassignReviewer(r.Context(), req.PullRequestID, req.OldReviewerID)
	if err != nil {
		switch {
		case errors.Is(err, pkgerrors.ErrNotFound):
			h.writeError(w, http.StatusNotFound, dto.ErrNotFound, "PR or user not found")
		case errors.Is(err, pkgerrors.ErrPrMerged):
			h.writeError(w, http.StatusConflict, dto.ErrPrMerged, "cannot reassign on merged PR")
		case errors.Is(err, pkgerrors.ErrNotAssigned):
			h.writeError(w, http.StatusConflict, dto.ErrNotAssigned, "reviewer is not assigned to this PR")
		case errors.Is(err, pkgerrors.ErrNoCandidate):
			h.writeError(w, http.StatusConflict, dto.ErrNoCandidate, "no active replacement candidate in team")
		default:
			h.writeError(w, http.StatusInternalServerError, dto.ErrUndefined, "internal server error")
		}
		return
	}

	resp := dto.PullRequestReassignResp{
		PR:         pr.ToDTO(),
		ReplacedBy: newReviewerID,
	}

	h.writeSuccess(w, resp)
}
