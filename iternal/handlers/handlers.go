package handlers

import (
	"bytes"
	"errors"
	"net/http"

	"github.com/Tomap-Tomap/go-loyalty-service/iternal/hasher"
	"github.com/Tomap-Tomap/go-loyalty-service/iternal/logger"
	"github.com/Tomap-Tomap/go-loyalty-service/iternal/models"
	"github.com/Tomap-Tomap/go-loyalty-service/iternal/storage"
	"github.com/Tomap-Tomap/go-loyalty-service/iternal/tokenworker"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

type Handlers struct {
	storage storage.Storage
	tw      tokenworker.TokenWorker
}

func NewHandlers(storage storage.Storage, tw tokenworker.TokenWorker) Handlers {
	return Handlers{storage: storage, tw: tw}
}

func (h *Handlers) register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	var buf bytes.Buffer
	_, err := buf.ReadFrom(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	u, err := models.NewUserByJSON(buf.Bytes())

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if u.Login == "" || u.Password == "" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	hU, err := hasher.GetHashedUser(*u)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.storage.CreateUser(r.Context(), *hU)

	var tError *pgconn.PgError
	if errors.As(err, &tError) && tError.Code == pgerrcode.UniqueViolation {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tokenString, err := h.tw.GetToken(u.Login)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cookie := &http.Cookie{
		Name:  "token",
		Value: tokenString,
	}

	http.SetCookie(w, cookie)

	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	var buf bytes.Buffer
	_, err := buf.ReadFrom(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	u, err := models.NewUserByJSON(buf.Bytes())

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if u.Login == "" || u.Password == "" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	uDB, err := h.storage.GetUser(r.Context(), u.Login)

	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logger.Log.Info("user", zap.String(uDB.Login, uDB.Password))
}

func ServiceMux(h Handlers) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/api/user/register", logger.RequestLogger(http.HandlerFunc(h.register)))
	mux.Handle("/api/user/login", logger.RequestLogger(http.HandlerFunc(h.login)))
	return mux
}
