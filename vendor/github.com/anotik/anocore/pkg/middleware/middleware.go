package middleware

import (
	"net/http"
	"time"

	"github.com/anotik/anocore/pkg/consts"
	"github.com/anotik/anocore/pkg/logger"
	"github.com/anotik/anocore/pkg/response"
	"github.com/anotik/anocore/pkg/util"
)

func copyTraceHeaders(from, to http.Header) {
	for header, _ := range consts.TRACE_HEADERS_MAP {
		if from.Get(header) != "" {
			to.Set(header, from.Get(header))
		}
	}
}

type APIFunc func(w http.ResponseWriter, r *http.Request) response.APIResponse

func Make(h APIFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := logger.ContextWithTracHeaders(r)
		r = r.WithContext(ctx)
		ctxLogger, err := logger.WithTraceHeaders(r)
		if err != nil {
			panic("failed to create context logger: " + err.Error())
		}
		w.Header().Set("Content-Type", "application/json")

		ctxLogger.LogRequest(r)
		statusCode := http.StatusOK
		resp := h(w, r)
		// we do this here just in case the handler overrides the headers
		copyTraceHeaders(r.Header, w.Header())

		if resp.Err != nil {
			e := resp.Err
			ctxLogger.Error(e.Message, "data", resp.Data, "details", e.Details)
			statusCode = e.StatusCode
		}

		jsonRes, err := util.ToJSON(resp)
		if err != nil {
			ctxLogger.Error("Failed to convert response to JSON", "error", err.Error(), "response", resp)
			statusCode = http.StatusInternalServerError
			w.WriteHeader(statusCode)
			w.Write([]byte(`{"data": null, "error": "Failed to convert response to JSON"}`))
			return
		}

		w.WriteHeader(statusCode)

		bytesWritten, err := w.Write(jsonRes)
		if err != nil {
			ctxLogger.LogFailedResponse(w, resp, r, time.Now(), bytesWritten, err)
		} else {
			ctxLogger.LogResponse(w, resp, r, time.Now())
		}
	}
}
