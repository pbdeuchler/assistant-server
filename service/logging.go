package service

import (
	"log/slog"
	"net/http"
	"os"

	httplog "github.com/go-chi/httplog/v3"
)

// var log = zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

func httpLogger() func(http.Handler) http.Handler {
	isLocalhost := true
	logFormat := httplog.SchemaECS.Concise(isLocalhost)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: logFormat.ReplaceAttr,
	})).With(
		slog.String("app", "assistant-server"),
		slog.String("version", "v1.0.0-a1fa420"),
		slog.String("env", "production"),
	)

	return httplog.RequestLogger(logger, &httplog.Options{
		// Level defines the verbosity of the request logs:
		// slog.LevelDebug - log all responses (incl. OPTIONS)
		// slog.LevelInfo  - log responses (excl. OPTIONS)
		// slog.LevelWarn  - log 4xx and 5xx responses only (except for 429)
		// slog.LevelError - log 5xx responses only
		Level: slog.LevelInfo,

		// Set log output to Elastic Common Schema (ECS) format.
		Schema: httplog.SchemaECS,

		// RecoverPanics recovers from panics occurring in the underlying HTTP handlers
		// and middlewares. It returns HTTP 500 unless response status was already set.
		//
		// NOTE: Panics are logged as errors automatically, regardless of this setting.
		RecoverPanics: true,

		// Optionally, filter out some request logs.
		Skip: func(req *http.Request, respStatus int) bool {
			return respStatus == 404 || respStatus == 405
		},

		// Optionally, log selected request/response headers explicitly.
		LogRequestHeaders:  []string{"Origin"},
		LogResponseHeaders: []string{},

		// Optionally, enable logging of request/response body based on custom conditions.
		// Useful for debugging payload issues in development.
		LogRequestBody: func(req *http.Request) bool {
			return true
		},
		LogResponseBody: func(req *http.Request) bool {
			return false
		},
	})
}
