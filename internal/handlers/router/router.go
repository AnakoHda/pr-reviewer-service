package router

import (
	"pr-reviewer-service/internal/handlers/pullRequestHandler"
	"pr-reviewer-service/internal/handlers/teamHandler"
	"pr-reviewer-service/internal/handlers/userHandler"
	"pr-reviewer-service/internal/service/pullRequestService"
	"pr-reviewer-service/internal/service/teamService"
	"pr-reviewer-service/internal/service/userService"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
)

func RegisterRoutes(prServ *pullRequestService.Service, teamServ *teamService.Service, userServ *userService.Service) http.Handler {
	mux := http.NewServeMux()
	pullRequestHandler.New(prServ).Register(mux)
	teamHandler.New(teamServ).Register(mux)
	userHandler.New(userServ).Register(mux)

	// Раздаем файл спецификации
	mux.HandleFunc("/openapi.yml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "api/openapi.yml")
	})

	// Swagger UI
	mux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/openapi.yml"), // Ссылка на файл спецификации
	))

	return mux
}
