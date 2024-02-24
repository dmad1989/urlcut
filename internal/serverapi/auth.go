package serverapi

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dmad1989/urlcut/internal/logging"
	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

var ErrorNoUser = errors.New("no userid in auth token")
var ErrorInvalidToken = errors.New("auth token not valid")

const TOKEN_EXP = time.Hour * 6
const SECRET_KEY = "gopracticumshoretenersecretkey"

func checkToken(t string) (int, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(t, claims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		})
	if err != nil || !token.Valid {
		return 0, ErrorInvalidToken
	}
	if claims.UserID == 0 {
		return 0, ErrorNoUser
	}
	return claims.UserID, nil
}

func generateToken() (string, error) {
	//TODO create new user // set to context
	userId := 1

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

func Auth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextW := w
		val, err := (*http.Request).Cookie(r, "token")

		if err != nil && !errors.Is(err, http.ErrNoCookie) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Errorf("auth cookie: %w", err).Error()))
			return
		}
		userId := 0
		if !errors.Is(err, http.ErrNoCookie) {
			userId, err = checkToken(val.Value)
		}
		logging.Log.Info("error", err)
		is := errors.Is(err, ErrorInvalidToken)
		logging.Log.Info("errors.Is(err, ErrorInvalidToken)", is)
		switch {
		case errors.Is(err, http.ErrNoCookie) || errors.Is(err, ErrorInvalidToken):
			token, err := generateToken()
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
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
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Errorf("auth : %w", err).Error()))
			return
		default:
			fmt.Println("userId ", userId)
			//todo set to context
		}

		h.ServeHTTP(nextW, r)
	})
}
