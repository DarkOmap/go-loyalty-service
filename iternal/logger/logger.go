package logger

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type (
	loggingResponseWriter struct {
		http.ResponseWriter
		wroteHeader bool
		code        int
		bytes       int
		error       string
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}

	if r.code != http.StatusOK {
		r.error = string(b)
	}

	size, err := r.ResponseWriter.Write(b)
	r.bytes += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	if !r.wroteHeader {
		r.code = statusCode
		r.wroteHeader = true
		r.ResponseWriter.WriteHeader(statusCode)
	}
}

var Log *zap.Logger = zap.NewNop()

func Initialize(level string, outputPath string) error {
	lvl, err := zap.ParseAtomicLevel(level)

	if err != nil {
		return fmt.Errorf("parse level %s: %w", level, err)
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	cfg.OutputPaths = []string{outputPath}
	zl, err := cfg.Build()

	if err != nil {
		return fmt.Errorf("build logger: %w", err)
	}

	Log = zl

	return nil
}

func RequestLogger(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		Log.Info("Got incoming HTTP request",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
		)

		lw := loggingResponseWriter{
			ResponseWriter: w,
		}

		defer func() {
			duration := time.Since(start)

			Log.Info("Sending HTTP response",
				zap.String("duration", duration.String()),
				zap.Int("status", lw.code),
				zap.Int("size", lw.bytes),
				zap.String("error", lw.error),
			)
		}()

		h.ServeHTTP(&lw, r)
	}

	return http.HandlerFunc(logFn)
}
