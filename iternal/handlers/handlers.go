package handlers

import (
	"bytes"
	"errors"
	"net/http"

	"github.com/Tomap-Tomap/go-loyalty-service/iternal/models"
	"github.com/Tomap-Tomap/go-loyalty-service/iternal/storage"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type Handlers struct {
	storage storage.Storage
}

func NewHandlers(storage storage.Storage) Handlers {
	return Handlers{storage: storage}
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
	}

	u, err := models.NewUserByJSON(buf.Bytes())

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	if u.Login == "" || len(u.Password) == 0 {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	err = h.storage.CreateUser(r.Context(), *u)

	var tError *pgconn.PgError
	if errors.As(err, &tError) && tError.Code == pgerrcode.UniqueViolation {
		http.Error(w, err.Error(), http.StatusConflict)
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

func ServiceMux(h Handlers) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/apu/user/register", h.register)

	return mux
}
