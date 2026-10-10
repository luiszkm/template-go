package httpx

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
)

const HeaderRequestID = "X-Request-ID"

var acceptedRequestID = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

type ctxKey struct{}

type causeKey struct{}

type causeSlot struct {
	err   error
	stack string
}

func recordCause(ctx context.Context, err error) {
	if slot, ok := ctx.Value(causeKey{}).(*causeSlot); ok && err != nil {
		slot.err = err
	}
}

func recordPanic(ctx context.Context, v any, stack string) bool {
	slot, ok := ctx.Value(causeKey{}).(*causeSlot)
	if ok {
		slot.err = fmt.Errorf("panic: %v", v)
		slot.stack = stack
	}
	return ok
}

func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(ctxKey{}).(string)
	return id
}

func Chain(h http.Handler, log *slog.Logger) http.Handler {
	return RequestID(AccessLog(Recover(h, log), log))
}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(HeaderRequestID)
		if !acceptedRequestID.MatchString(id) {
			id = uuid.NewString()
		}
		w.Header().Set(HeaderRequestID, id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, id)))
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	if s.status == 0 {
		s.status = code
	}
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	if s.status == 0 {
		s.status = http.StatusOK
	}
	return s.ResponseWriter.Write(b)
}

func (s *statusRecorder) Unwrap() http.ResponseWriter { return s.ResponseWriter }

func AccessLog(next http.Handler, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w}
		slot := &causeSlot{}
		next.ServeHTTP(rec, r.WithContext(context.WithValue(r.Context(), causeKey{}, slot)))
		if rec.status == 0 {
			rec.status = http.StatusOK
		}
		attrs := []slog.Attr{
			slog.String("request_id", RequestIDFrom(r.Context())),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", rec.status),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
		}
		level := slog.LevelInfo
		if rec.status >= http.StatusInternalServerError {
			level = slog.LevelError
		}
		if slot.err != nil {
			attrs = append(attrs, slog.String("error", slot.err.Error()))
		}
		if slot.stack != "" {
			attrs = append(attrs, slog.String("stack", slot.stack))
		}
		log.LogAttrs(r.Context(), level, "request", attrs...)
	})
}

func Recover(next http.Handler, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			v := recover()
			if v == nil {
				return
			}
			if err, ok := v.(error); ok && errors.Is(err, http.ErrAbortHandler) {
				panic(v)
			}
			stack := string(debug.Stack())
			if !recordPanic(r.Context(), v, stack) {
				log.LogAttrs(r.Context(), slog.LevelError, "panic",
					slog.String("request_id", RequestIDFrom(r.Context())),
					slog.String("panic", fmt.Sprint(v)),
					slog.String("stack", stack),
				)
			}
			WriteProblem(w, r, http.StatusInternalServerError, "internal server error")
		}()
		next.ServeHTTP(w, r)
	})
}
