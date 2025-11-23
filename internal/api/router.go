// Package api
package api

import (
	"github.com/gorilla/mux"
	"github.com/habibg1232191/reviewer-assignment-service/internal/handler"
	"github.com/habibg1232191/reviewer-assignment-service/internal/usecase"
)

func NewAPIRouter(s usecase.ReviewerService) *mux.Router {
	r := mux.NewRouter()
	h := handler.NewHandler(s)

	h.RegisterPublicRoutes(r)

	return r
}
