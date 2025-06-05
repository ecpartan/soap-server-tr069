package login

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ecpartan/soap-server-tr069/internal/apperror"
	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/golang-jwt/jwt/v5"
)

// TODO: move to config
func getJWTsecret() string {
	return "secret"
}

func generateJWT(userID int, secretKey string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 1).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

func checkJWT(tokenString string) (*jwt.Token, error) {
	logger.LogDebug("Enter checkJWT")
	return jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(getJWTsecret()), nil
	})
}

func AuthMiddleware(next apperror.AppHandler) apperror.AppHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		logger.LogDebug("Enter AuthMiddleware")
		authHeader := r.Header.Get("Authorization")
		logger.LogDebug("authHeader", authHeader)

		if authHeader == "" {

			return fmt.Errorf("not authorized")

		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := checkJWT(tokenString)
		if err != nil || !token.Valid {
			return fmt.Errorf("not authorized %v", err)
		}

		next(w, r)
		return nil
	}
}
