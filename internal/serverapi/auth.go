package serverapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"

	"github.com/dmad1989/urlcut/internal/config"
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

// Auth это middleware для регистрации и авторизации пользователей.
// Проверяет наличие и валидность токена в cookie "token".
// Если cookie нет - регистрируем нового пользователя: генерируем новый ID, токен и записываем в cookie.
// Полученный из токена или сгенерированный userid записываем в контекст вызова.
func (s Server) Auth(h http.Handler) http.Handler {
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
