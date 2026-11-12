package middleware

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"time"

	"anhgelus.world/ljus"
)

// Log takes a [slog.Logger] to log every requests.
func Log(logger *slog.Logger, logNotFound, logBadRequest bool) ljus.Middleware {
	return func(next http.Handler, w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := logger.With("uri", ctx.Value(ljus.KeyRealURI))
		name, ok := ctx.Value(ljus.KeyRouteName).(string)
		if ok {
			log = log.With("route_name", name)
		}
		now := time.Now()
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Error("panic!", "error", err, "duration", time.Since(now))
				cancelCause, ok := ctx.Value(ljus.KeyCancelContext).(context.CancelCauseFunc)
				if !ok || cancelCause == nil {
					return
				}
				switch e := err.(type) {
				case error:
					cancelCause(e)
				case string:
					cancelCause(errors.New(e))
				default:
					log.Warn(
						"cannot set cancel cause, because error type is not supported",
						"type", fmt.Sprintf("%T", e),
					)
				}
			}
		}()

		next.ServeHTTP(w, r)

		code := w.(*ljus.StatusWriter).Code
		if code == 0 {
			code = http.StatusOK
		}

		log = log.With("status", code, "duration", time.Since(now))
		if code < 400 {
			log.Debug("handled")
		} else if code < 500 {
			level := slog.LevelDebug
			if (code == http.StatusNotFound && logNotFound) ||
				(code == http.StatusBadRequest && logBadRequest) ||
				(code != http.StatusNotFound &&
					code != http.StatusBadRequest &&
					code != http.StatusTooManyRequests) {
				level = slog.LevelWarn
			}
			log.Log(context.Background(), level, "invalid request")
		} else {
			log.Error("error while handling request")
		}
	}
}

// SecurityHeaders adds common security-related headers to the request.
func SecurityHeaders(domain string, cacheDur time.Duration) ljus.Middleware {
	if cacheDur == 0 {
		cacheDur = 15 * time.Minute
	}
	return func(next http.Handler, w http.ResponseWriter, r *http.Request) {
		// prevent tracking
		w.Header().Add("Referrer-Policy", "strict-origin-when-cross-origin")
		// prevent iframe
		w.Header().Add("X-Frame-Options", "deny")
		// prevent bad content being parsed
		w.Header().Add("X-Content-Type-Options", "nosniff")
		w.Header().Add("X-Permitted-Cross-Domain-Policies", "none")
		// content security, cors & co
		w.Header().Add("Content-Security-Policy", "default-src 'self' https://*."+domain+"; object-src 'none';")
		w.Header().Add("Access-Control-Allow-Origin", "https://"+domain)
		w.Header().Add("Cross-Origin-Resource-Policy", "same-origin")
		// caching
		w.Header().Add("Access-Control-Max-Age", fmt.Sprintf("%.0f", math.Floor(cacheDur.Seconds())))

		next.ServeHTTP(w, r)
	}
}
