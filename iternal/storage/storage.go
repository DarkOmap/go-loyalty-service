package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Tomap-Tomap/go-loyalty-service/iternal/hasher"
	"github.com/Tomap-Tomap/go-loyalty-service/iternal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type retryPolicy struct {
	retryCount int
	duration   int
	increment  int
}

type Storage struct {
	conn        *pgx.Conn
	retryPolicy retryPolicy
}

func NewStorage(conn *pgx.Conn) (*Storage, error) {
	rp := retryPolicy{3, 1, 2}
	s := &Storage{conn: conn, retryPolicy: rp}

	if err := s.createTables(); err != nil {
		return nil, fmt.Errorf("create tables in database: %w", err)
	}

	return s, nil
}

func (s *Storage) createTables() error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	createUserQuery := `
		CREATE TABLE IF NOT EXISTS users (
			Id UUID PRIMARY KEY,
			Login VARCHAR(150) UNIQUE,
			Password CHAR(64),
			Salt VARCHAR(150)
		);
		CREATE UNIQUE INDEX IF NOT EXISTS user_idx ON users (Login);
	`

	err := pgx.BeginFunc(ctx, s.conn, func(tx pgx.Tx) error {
		_, err := retry2(ctx, s.retryPolicy, func() (pgconn.CommandTag, error) {
			return s.conn.Exec(ctx, createUserQuery)
		})

		if err != nil {
			return fmt.Errorf("create users table: %w", err)
		}

		return nil
	})

	return err
}

func (s *Storage) CreateUser(ctx context.Context, u models.User) error {
	query := `
		INSERT INTO users (Id, Login, Password, Salt) VALUES (gen_random_uuid(), $1, $2, $3);
	`

	hU, err := hasher.GetHashedUser(u)

	if err != nil {
		return fmt.Errorf("get hashed user: %w", err)
	}

	_, err = retry2(ctx, s.retryPolicy, func() (pgconn.CommandTag, error) {
		return s.conn.Exec(ctx, query, hU.Login, hU.Password, hU.Salt)
	})

	return err
}

// func retry(ctx context.Context, rp retryPolicy, fn func() error) error {
// 	fnWithReturn := func() (struct{}, error) {
// 		return struct{}{}, fn()
// 	}

// 	_, err := retry2(ctx, rp, fnWithReturn)
// 	return err
// }

func retry2[T any](ctx context.Context, rp retryPolicy, fn func() (T, error)) (T, error) {
	if val1, err := fn(); err == nil || !isonnectionException(err) {
		return val1, err
	}

	var err error
	var ret1 T
	duration := rp.duration
	for i := 0; i < rp.retryCount; i++ {
		select {
		case <-time.NewTimer(time.Duration(duration) * time.Second).C:
			ret1, err = fn()
			if err == nil || !isonnectionException(err) {
				return ret1, err
			}
		case <-ctx.Done():
			return ret1, err
		}

		duration += rp.increment
	}

	return ret1, err
}

func isonnectionException(err error) bool {
	var tError *pgconn.PgError
	if errors.As(err, &tError) && pgerrcode.IsConnectionException(tError.Code) {
		return true
	}

	return false
}
