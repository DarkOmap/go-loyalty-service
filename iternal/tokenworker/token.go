package tokenworker

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type TokenWorker struct {
	secret string
	exp    time.Duration
}

func NewToken(secret string, exp time.Duration) *TokenWorker {
	return &TokenWorker{secret: secret, exp: exp}
}

func (t *TokenWorker) GetToken(sub string) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			Subject:   sub,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(t.exp)),
		},
	)
	tokenString, err := token.SignedString([]byte(t.secret))

	if err != nil {
		return "", fmt.Errorf("signed token: %w", err)
	}

	return tokenString, nil
}

func (t *TokenWorker) TokenValidation(token string) bool {
	jwtToken, err := jwt.Parse(token, func(jwtT *jwt.Token) (interface{}, error) {
		if _, ok := jwtT.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", jwtT.Header["alg"])
		}

		return []byte(t.secret), nil
	})

	if err != nil {
		return false
	}

	if !jwtToken.Valid {
		return false
	}

	return true
}

// func (t *TokenWorker) RequestToken(h http.Handler) http.Handler {
// 	logFn := func(w http.ResponseWriter, r *http.Request) {
// 		tokenCookie, err := r.Cookie("token")

// 		if err != nil {
// 			h.ServeHTTP(w, r)
// 		}

// 		tokenValid := t.TokenValidation(tokenCookie.Value)

// 		if !tokenValid{}
// 		start := time.Now()

// 		Log.Info("Got incoming HTTP request",
// 			zap.String("uri", r.RequestURI),
// 			zap.String("method", r.Method),
// 		)

// 		lw := loggingResponseWriter{
// 			ResponseWriter: w,
// 		}

// 		defer func() {
// 			duration := time.Since(start)

// 			Log.Info("Sending HTTP response",
// 				zap.String("duration", duration.String()),
// 				zap.Int("status", lw.code),
// 				zap.Int("size", lw.bytes),
// 				zap.String("error", lw.error),
// 			)
// 		}()

// 		h.ServeHTTP(&lw, r)
// 	}

// 	return http.HandlerFunc(logFn)
// }
