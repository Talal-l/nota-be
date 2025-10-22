package service

import (
	"net/http"
	"nota/types"

	"github.com/anotik/anocore/pkg/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/uptrace/bun"
)

func addRoutes(mux *http.ServeMux, cnf types.AppConfig, db *bun.DB) {

	mux.Handle("/nota/v1/docs/", httpSwagger.Handler(
		httpSwagger.URL("/nota/v1/docs/swagger/doc.json"),
	))
	mux.Handle("/nota/v1/docs/swagger/doc.json", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/swagger.json")
	}))
	mux.Handle("/nota/v1/health", middleware.Make(HealthHandler))

	mux.Handle("POST /nota/v1/user", middleware.Make(CreateUserHandler(db)))

}
