package service

import (
	"context"
	"net/http"
	"nota/service/db/queries"

	"github.com/anotik/anocore/pkg/errs"
	"github.com/anotik/anocore/pkg/logger"
	"github.com/anotik/anocore/pkg/middleware"
	"github.com/anotik/anocore/pkg/response"
	"github.com/anotik/anocore/pkg/util"
	"github.com/uptrace/bun"
)

// @Summary Health check endpoint
// @Description Check the health status of the email service
// @Tags health
// @Accept json
// @Produce json
// @Param fullTest query string false "Run full health check" Enums(true, false)
// @Success 200 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse

// @Router /health [get]
func HealthHandler(w http.ResponseWriter, r *http.Request) response.APIResponse {
	log, err := logger.WithTraceHeaders(r)

	if err != nil {
		return response.NewAPIResponse(nil, &errs.ApiError{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		})
	}
	fullTest := r.URL.Query().Get("fullTest")
	if fullTest == "true" {
		log.Info("full health check")
	}

	return response.NewAPIResponse("OK", nil)
}

type CreateUserReq struct {
	Email string `json:"email" validate:"required,email"`
	Name  string `json:"name" validate:"required"`
}

func (r CreateUserReq) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)

	if len(r.Email) == 0 {
		problems["email"] = "email is required"
	}

	if len(r.Name) == 0 {
		problems["name"] = "name is required"
	}

	return problems
}

// @Summary create user endpoint
// @Description Create new user
// @Tags User
// @Accept json
// @Produce json
// @Param request body CreateUserReq true "New user"
// @Success 200 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /user [post]
func CreateUserHandler(db *bun.DB) middleware.APIFunc {

	return func(w http.ResponseWriter, r *http.Request) response.APIResponse {
		_, err := logger.WithTraceHeaders(r)

		if err != nil {
			return response.NewAPIResponse(nil, &errs.ApiError{
				Message:    err.Error(),
				StatusCode: http.StatusInternalServerError,
			})
		}

		req, problems, err := util.DecodeValid[CreateUserReq](r)
		if err != nil {
			return response.NewAPIResponse(nil, &errs.ApiError{
				Message:    err.Error(),
				StatusCode: http.StatusInternalServerError,
			})
		}

		if len(problems) > 0 {
			return response.NewAPIResponse(nil, &errs.ApiError{
				Message:    "validation failed",
				StatusCode: http.StatusBadRequest,
				Details:    problems,
			})
		}

		user, err := queries.CreateUser(r.Context(), db, req.Email, req.Name)
		if err != nil {
			return response.NewAPIResponse(nil, &errs.ApiError{
				Message:    err.Error(),
				StatusCode: http.StatusInternalServerError,
			})
		}

		return response.NewAPIResponse(user, nil)
	}
}

type CreateContentReq struct {
	Content string
}

func (r CreateContentReq) Valid(ctx context.Context) map[string]string {
	if r.Content == "" {
		return map[string]string{"value": "value is required"}
	}

	problems := make(map[string]string)
	return problems
}

// @Summary create content endpoint
// @Description Create new content for current user
// @Tags Content
// @Accept json
// @Produce json
// @Param request body CreateContentReq true "New content"
// @Success 200 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /content [post]
func CreateContentHandler(db *bun.DB) middleware.APIFunc {

	return func(w http.ResponseWriter, r *http.Request) response.APIResponse {
		_, err := logger.WithTraceHeaders(r)

		if err != nil {
			return response.NewAPIResponse(nil, &errs.ApiError{
				Message:    err.Error(),
				StatusCode: http.StatusInternalServerError,
			})
		}

		req, problems, err := util.DecodeValid[CreateContentReq](r)
		if err != nil {
			return response.NewAPIResponse(nil, &errs.ApiError{
				Message:    err.Error(),
				StatusCode: http.StatusInternalServerError,
			})
		}

		if len(problems) > 0 {
			return response.NewAPIResponse(nil, &errs.ApiError{
				Message:    "validation failed",
				StatusCode: http.StatusBadRequest,
				Details:    problems,
			})
		}

		content, err := queries.CreateContent(r.Context(), db, queries.CreateContentArgs{
			UserID:  1,
			Content: req.Content,
		})
		if err != nil {
			return response.NewAPIResponse(nil, &errs.ApiError{
				Message:    err.Error(),
				StatusCode: http.StatusInternalServerError,
			})
		}

		return response.NewAPIResponse(content, nil)
	}
}

// @Summary get content endpoint
// @Description Get content for current user
// @Tags Content
// @Accept json
// @Produce json
// @Success 200 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /content [get]
func GetContentHandler(db *bun.DB) middleware.APIFunc {

	return func(w http.ResponseWriter, r *http.Request) response.APIResponse {
		_, err := logger.WithTraceHeaders(r)

		if err != nil {
			return response.NewAPIResponse(nil, &errs.ApiError{
				Message:    err.Error(),
				StatusCode: http.StatusInternalServerError,
			})
		}

		content, err := queries.GetContents(r.Context(), db)
		if err != nil {
			return response.NewAPIResponse(nil, &errs.ApiError{
				Message:    err.Error(),
				StatusCode: http.StatusInternalServerError,
			})
		}

		return response.NewAPIResponse(content, nil)
	}
}
