package consts

// Value in raw header : value in logs to help logging tools parse it
var TRACE_HEADERS_MAP = map[string]string{
	"X-Request-Id": "x_request_id",
	"X-B3-TraceId": "x_b3_traceid",
	"X-B3-SpanId":  "x_b3_spanid",
	"X-Client-Id":  "x_client_id",
	"X-Api-User":   "x_api_user",
	"X-Client-Ip":  "x_client_ip",
}
