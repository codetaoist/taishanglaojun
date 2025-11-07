package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/codetaoist/taishanglaojun/auth/internal/model"
)

// UserRepository interface defines user repository operations
type UserRepository interface {
	Create(user *model.User) error
	GetByID(id int) (*model.User, error)
	GetByUsername(username string) (*model.User, error)
	GetByEmail(email string) (*model.User, error)
	Update(user *model.User) error
	Delete(id int) error
	List(limit, offset int) ([]*model.User, error)
}

// SessionRepository interface defines session repository operations
type SessionRepository interface {
	Create(session *model.Session) error
	GetByID(id int) (*model.Session, error)
	GetByTokenHash(tokenHash string) (*model.Session, error)
	Update(session *model.Session) error
	Delete(id int) error
	DeleteExpired() error
}

// userRepository implements UserRepository
type userRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

// Create creates a new user
func (r *userRepository) Create(user *model.User) error {
	query := `
		INSERT INTO lao_users (username, email, password, role, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	err := r.db.QueryRow(
		query,
		user.Username,
		user.Email,
		user.Password,
		user.Role,
		user.Status,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID gets a user by ID
func (r *userRepository) GetByID(id int) (*model.User, error) {
	query := `
		SELECT id, username, email, password, role, status, created_at, updated_at
		FROM lao_users
		WHERE id = $1
	`

	user := &model.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetByUsername gets a user by username
func (r *userRepository) GetByUsername(username string) (*model.User, error) {
	query := `
		SELECT id, username, email, password, role, status, created_at, updated_at
		FROM lao_users
		WHERE username = $1
	`

	user := &model.User{}
	err := r.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetByEmail gets a user by email
func (r *userRepository) GetByEmail(email string) (*model.User, error) {
	query := `
		SELECT id, username, email, password, role, status, created_at, updated_at
		FROM lao_users
		WHERE email = $1
	`

	user := &model.User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// Update updates a user
func (r *userRepository) Update(user *model.User) error {
	query := `
		UPDATE lao_users
		SET username = $2, email = $3, password = $4, role = $5, status = $6, updated_at = $7
		WHERE id = $1
	`

	user.UpdatedAt = time.Now()

	_, err := r.db.Exec(
		query,
		user.ID,
		user.Username,
		user.Email,
		user.Password,
		user.Role,
		user.Status,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// Delete deletes a user
func (r *userRepository) Delete(id int) error {
	query := `DELETE FROM lao_users WHERE id = $1`

	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// List lists users with pagination
func (r *userRepository) List(limit, offset int) ([]*model.User, error) {
	query := `
		SELECT id, username, email, password, role, status, created_at, updated_at
		FROM lao_users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	users := []*model.User{}
	for rows.Next() {
		user := &model.User{}
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Password,
			&user.Role,
			&user.Status,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate users: %w", err)
	}

	return users, nil
}

// sessionRepository implements SessionRepository
type sessionRepository struct {
	db *sql.DB
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *sql.DB) SessionRepository {
	return &sessionRepository{db: db}
}

// Create creates a new session
func (r *sessionRepository) Create(session *model.Session) error {
	query := `
		INSERT INTO lao_sessions (user_id, token_hash, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	now := time.Now()
	session.CreatedAt = now
	session.UpdatedAt = now

	err := r.db.QueryRow(
		query,
		session.UserID,
		session.TokenHash,
		session.ExpiresAt,
		session.CreatedAt,
		session.UpdatedAt,
	).Scan(&session.ID)

	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// GetByID gets a session by ID
func (r *sessionRepository) GetByID(id int) (*model.Session, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, created_at, updated_at
		FROM lao_sessions
		WHERE id = $1
	`

	session := &model.Session{}
	err := r.db.QueryRow(query, id).Scan(
		&session.ID,
		&session.UserID,
		&session.TokenHash,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session, nil
}

// GetByTokenHash gets a session by token hash
func (r *sessionRepository) GetByTokenHash(tokenHash string) (*model.Session, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, created_at, updated_at
		FROM lao_sessions
		WHERE token_hash = $1
	`

	session := &model.Session{}
	err := r.db.QueryRow(query, tokenHash).Scan(
		&session.ID,
		&session.UserID,
		&session.TokenHash,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session, nil
}

// Update updates a session
func (r *sessionRepository) Update(session *model.Session) error {
	query := `
		UPDATE lao_sessions
		SET user_id = $2, token_hash = $3, expires_at = $4, updated_at = $5
		WHERE id = $1
	`

	session.UpdatedAt = time.Now()

	_, err := r.db.Exec(
		query,
		session.ID,
		session.UserID,
		session.TokenHash,
		session.ExpiresAt,
		session.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// Delete deletes a session
func (r *sessionRepository) Delete(id int) error {
	query := `DELETE FROM lao_sessions WHERE id = $1`

	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// DeleteExpired deletes all expired sessions
func (r *sessionRepository) DeleteExpired() error {
	query := `DELETE FROM lao_sessions WHERE expires_at < $1`

	_, err := r.db.Exec(query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	return nil
}