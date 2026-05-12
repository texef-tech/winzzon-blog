package sqlc

import (
	"context"

	"github.com/google/uuid"
)

const getAdminByUsername = `
SELECT id, username, password_hash, last_login_at, created_at
FROM admin_credentials WHERE username = $1
`

func (q *Queries) GetAdminByUsername(ctx context.Context, username string) (AdminCredential, error) {
	row := q.db.QueryRow(ctx, getAdminByUsername, username)
	var a AdminCredential
	err := row.Scan(&a.ID, &a.Username, &a.PasswordHash, &a.LastLoginAt, &a.CreatedAt)
	return a, err
}

const createAdmin = `
INSERT INTO admin_credentials (username, password_hash) VALUES ($1, $2)
RETURNING id, username, password_hash, last_login_at, created_at
`

type CreateAdminParams struct {
	Username     string
	PasswordHash string
}

func (q *Queries) CreateAdmin(ctx context.Context, arg CreateAdminParams) (AdminCredential, error) {
	row := q.db.QueryRow(ctx, createAdmin, arg.Username, arg.PasswordHash)
	var a AdminCredential
	err := row.Scan(&a.ID, &a.Username, &a.PasswordHash, &a.LastLoginAt, &a.CreatedAt)
	return a, err
}

const updateLastLogin = `UPDATE admin_credentials SET last_login_at = NOW() WHERE id = $1`

func (q *Queries) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.Exec(ctx, updateLastLogin, id)
	return err
}

const adminExists = `SELECT EXISTS(SELECT 1 FROM admin_credentials WHERE username = $1)`

func (q *Queries) AdminExists(ctx context.Context, username string) (bool, error) {
	row := q.db.QueryRow(ctx, adminExists, username)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}
