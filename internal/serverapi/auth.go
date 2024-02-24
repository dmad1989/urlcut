package serverapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

var (
	ErrorNoUser       = errors.New("no userid in auth token")
	ErrorInvalidToken = errors.New("auth token not valid")
)

const TOKEN_EXP = time.Hour * 6
const SECRET_KEY = "gopracticumshoretenersecretkey"

func checkToken(t string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(t, claims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		})
	if err != nil || !token.Valid {
		return "", ErrorInvalidToken
	}
	if claims.UserID == "" {
		return "", ErrorNoUser
	}
	return claims.UserID, nil
}

func generateToken(userId string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TOKEN_EXP)),
		},
		UserID: userId,
	})

	tokenString, err := token.SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", fmt.Errorf("generateToken: %w", err)
	}

	return tokenString, nil
}

func (s server) Auth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextW := w
		val, err := (*http.Request).Cookie(r, "token")

		if err != nil && !errors.Is(err, http.ErrNoCookie) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(fmt.Errorf("auth cookie: %w", err).Error()))
			return
		}
		userId := ""
		if !errors.Is(err, http.ErrNoCookie) {
			userId, err = checkToken(val.Value)
		}

		switch {
		case errors.Is(err, http.ErrNoCookie) || errors.Is(err, ErrorInvalidToken):
			userId = createUserId()
			token, err := generateToken(userId)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(fmt.Errorf("auth : %w", err).Error()))
				return
			}
			cookie := http.Cookie{Name: "token", Value: token}
			http.SetCookie(w, &cookie)
		case errors.Is(err, ErrorNoUser):
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(fmt.Errorf("auth : %w", err).Error()))
			return
		case err != nil:
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(fmt.Errorf("auth : %w", err).Error()))
			return
		}
		newCtx := context.WithValue(r.Context(), s.config.GetUserContextKey(), userId)
		h.ServeHTTP(nextW, r.WithContext(newCtx))
	})
}

func createUserId() string {
	u := uuid.New()
	return u.String()
}
