// Package auth check and generate users and their JWT tokens
// Contains middleware for http  and interceptor for grpc
package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/dmad1989/urlcut/internal/config"
	iauth "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	ilog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
)

// Claims хранит в себе данные токена
type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

// Ошибки авторизации
var (
	ErrorNoUser       = errors.New("no userid in auth token") // в токене нет userId
	ErrorInvalidToken = errors.New("auth token not valid")    // токен не прошел валидацию
)

const (
	tokenExp  = time.Hour * 6
	secretKey = "gopracticumshoretenersecretkey"
)

// HTTP это middleware для регистрации и авторизации пользователей.
// Проверяет наличие и валидность токена в cookie "token".
// Если cookie нет - регистрируем нового пользователя: генерируем новый ID, токен и записываем в cookie.
// Полученный из токена или сгенерированный userid записываем в контекст вызова.
func HTTP(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextW := w
		tCookie, err := r.Cookie("token")
		if err != nil && !errors.Is(err, http.ErrNoCookie) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(fmt.Errorf("auth cookie: %w", err).Error()))
			return
		}
		userID := ""
		if !errors.Is(err, http.ErrNoCookie) {
			userID, err = checkToken(tCookie.Value)
		}
		switch {
		case errors.Is(err, http.ErrNoCookie) || errors.Is(err, ErrorInvalidToken):
			userID = createUserID()
			token, tokenErr := generateToken(userID)
			if tokenErr != nil {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(fmt.Errorf("auth : %w", tokenErr).Error()))
				return
			}
			cookie := http.Cookie{
				Name:  "token",
				Value: token,
				Path:  "/",
			}
			http.SetCookie(w, &cookie)
		case err != nil:
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(fmt.Errorf("auth : %w", err).Error()))
			return
		}
		ctx := context.WithValue(r.Context(), config.UserCtxKey, userID)
		ctx = context.WithValue(ctx, config.ErrorCtxKey, err)
		h.ServeHTTP(nextW, r.WithContext(ctx))
	})
}

// GRPC is interceptor for user's registration and authorization
// Check if token present in authorization metadata
// If authorization is empty ("beraer  ") - new user wil be created with new ID, token
// User id wil be written in context, token in response metadata
func GRPC(ctx context.Context) (context.Context, error) {
	token, err := iauth.AuthFromMD(ctx, "bearer")
	if err != nil && !strings.Contains(err.Error(), "Request unauthenticated with bearer") {
		return nil, err
	}
	userID := ""
	if token != "" {
		userID, err = checkToken(token)
	}

	switch {
	case token == "" || errors.Is(err, ErrorInvalidToken):
		userID = createUserID()
		if token, err = generateToken(userID); err != nil {
			return nil, fmt.Errorf("auth : %w", err)
		}
	case err != nil:
		return nil, fmt.Errorf("auth : %w", err)
	}

	tHeader := metadata.New(map[string]string{"authorization": "bearer " + token})
	if err := grpc.SendHeader(ctx, tHeader); err != nil {
		return nil, status.Errorf(codes.Internal, "unable to send 'authorization' header")
	}
	ctx = context.WithValue(ctx, config.TokenCtxKey, token)
	ctx = context.WithValue(ctx, config.UserCtxKey, userID)
	ctx = context.WithValue(ctx, config.ErrorCtxKey, err)

	ctx = ilog.InjectFields(ctx, ilog.Fields{"auth.sub", userID})

	// WARNING: In production define your own type to avoid context collisions.
	return ctx, nil
}

// checkToken проверяет токен на валидность.
func checkToken(t string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(t, claims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})
	if err != nil || !token.Valid {
		return "", ErrorInvalidToken
	}
	if claims.UserID == "" {
		return "", ErrorNoUser
	}
	return claims.UserID, nil
}

// generateToken генерирует HS256 - токен по userID.
func generateToken(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("generateToken: %w", err)
	}

	return tokenString, nil
}

func createUserID() string {
	u := uuid.New()
	return u.String()
}
