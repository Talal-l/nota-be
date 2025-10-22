package response

import "github.com/anotik/anocore/pkg/errs"

type APIResponse struct {
	Data any            `json:"data"`
	Err  *errs.ApiError `json:"error"`
}

func NewAPIResponse(data any, err *errs.ApiError) APIResponse {
	return APIResponse{Data: &data, Err: err}
}
