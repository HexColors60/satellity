package models

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-pg/pg"
	"github.com/godiscourse/godiscourse/config"
	"github.com/godiscourse/godiscourse/session"
	"github.com/godiscourse/godiscourse/uuid"
	"golang.org/x/crypto/bcrypt"
)

const usersDDL = `
CREATE TABLE IF NOT EXISTS users (
	user_id               VARCHAR(36) PRIMARY KEY,
	email                 VARCHAR(512),
	username              VARCHAR(64) NOT NULL CHECK (username ~* '^[a-z0-9][a-z0-9_]{3,63}$'),
	nickname              VARCHAR(64) NOT NULL DEFAULT '',
	encrypted_password    VARCHAR(1024),
	github_id             VARCHAR(1024) UNIQUE,
	created_at            TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	updated_at            TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX ON users ((LOWER(email)));
CREATE UNIQUE INDEX ON users ((LOWER(username)));
CREATE INDEX ON users (created_at);
`

// User contains info of a register user
type User struct {
	UserID            string         `sql:"user_id,pk"`
	Email             sql.NullString `sql:"email"`
	Username          string         `sql:"username"`
	Nickname          string         `sql:"nickname"`
	EncryptedPassword sql.NullString `sql:"encrypted_password"`
	GithubID          sql.NullString `sql:"github_id"`
	CreatedAt         time.Time      `sql:"created_at"`
	UpdatedAt         time.Time      `sql:"updated_at"`

	SessionID string `sql:"-"`
	isNew     bool   `sql:"-"`
}

var userCols = []string{"user_id", "email", "username", "nickname", "encrypted_password", "github_id", "created_at", "updated_at"}

// CreateUser create a new user
func CreateUser(ctx context.Context, email, username, nickname, password string, sessionSecret string) (*User, error) {
	t := time.Now()
	data, err := hex.DecodeString(sessionSecret)
	if err != nil {
		return nil, session.BadDataError(ctx)
	}
	public, err := x509.ParsePKIXPublicKey(data)
	if err != nil {
		return nil, session.BadDataError(ctx)
	}
	switch public.(type) {
	case *ecdsa.PublicKey:
	default:
		return nil, session.BadDataError(ctx)
	}

	email = strings.TrimSpace(email)
	if err := validateEmailFormat(ctx, email); err != nil {
		return nil, err
	}
	if nickname == "" {
		nickname = username
	}
	if len(password) < 8 || len(password) > 64 {
		return nil, session.BadDataError(ctx)
	}
	password, err = validateAndEncryptPassword(ctx, password)
	if err != nil {
		return nil, err
	}

	user := &User{
		UserID:            uuid.NewV4().String(),
		Email:             sql.NullString{String: email, Valid: true},
		Username:          username,
		Nickname:          nickname,
		EncryptedPassword: sql.NullString{String: password, Valid: true},
		CreatedAt:         t,
		UpdatedAt:         t,
	}
	err = session.Database(ctx).RunInTransaction(func(tx *pg.Tx) error {
		if err := tx.Insert(user); err != nil {
			return err
		}
		sess, err := user.addSession(ctx, tx, sessionSecret)
		if err != nil {
			return err
		}
		user.SessionID = sess.SessionID
		return nil
	})
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return user, nil
}

// AuthenticateUser read a user by tokenString. tokenString is a jwt token, more
// about jwt: https://github.com/dgrijalva/jwt-go
func AuthenticateUser(ctx context.Context, tokenString string) (*User, error) {
	var user *User
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, nil
		}
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, nil
		}
		uid, sid := fmt.Sprint(claims["uid"]), fmt.Sprint(claims["sid"])
		u, err := findUserByID(ctx, uid)
		if err != nil {
			return nil, err
		} else if u == nil {
			return nil, nil
		}
		user = u
		sess, err := readSession(ctx, uid, sid)
		if err != nil {
			return nil, err
		} else if sess == nil {
			return nil, nil
		}
		user.SessionID = sess.SessionID
		pkix, err := hex.DecodeString(sess.Secret)
		if err != nil {
			return nil, err
		}
		return x509.ParsePKIXPublicKey(pkix)
	})
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	if !token.Valid {
		return nil, nil
	}
	return user, nil
}

// ReadUser read user by id.
func ReadUser(ctx context.Context, id string) (*User, error) {
	return findUserByID(ctx, id)
}

// ReadUserByUsernameOrEmail read user by identity, which is an email or username.
func ReadUserByUsernameOrEmail(ctx context.Context, identity string) (*User, error) {
	user := &User{}
	identity = strings.ToLower(strings.TrimSpace(identity))
	if len(identity) < 3 {
		return nil, nil
	}
	if err := session.Database(ctx).Model(user).Column(userCols...).Where("username = ? OR email = ?", identity, identity).Select(); err == pg.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return user, nil
}

// Role of an user, contains admin and member for now.
func (user *User) Role() string {
	if config.Operators[user.Email.String] {
		return "admin"
	}
	return "member"
}

func findUserByID(ctx context.Context, id string) (*User, error) {
	user := &User{UserID: id}
	if err := session.Database(ctx).Model(user).Column(userCols...).WherePK().Select(); err == pg.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return user, nil
}

func validateAndEncryptPassword(ctx context.Context, password string) (string, error) {
	password = strings.TrimSpace(password)
	if len(password) < 8 {
		return password, session.PasswordTooSimpleError(ctx)
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return password, session.ServerError(ctx, err)
	}
	return string(hashedPassword), nil
}
