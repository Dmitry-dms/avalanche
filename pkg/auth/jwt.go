package auth

import (
	//"errors"
	//"fmt"

	"time"

	"github.com/dgrijalva/jwt-go"
)

// TokenManager provides logic for JWT & Refresh tokens generation and parsing
type TokenManager interface {
	NewJWT(companyId string, ttl time.Duration) (string, error)
	Parse(accessToken string) (string, error)
	//NewRefreshToken() (string, error)
}
type Manager struct {
	key string // TODO: add extra time to duration
}
type tokenClaims struct {
	jwt.StandardClaims
	CompanyName string 
}

func NewManager(key string) (*Manager, error) {
	if key == "" {
		key = "test"
	}
	return &Manager{key}, nil
}
func (m *Manager) NewJWT(companyId string, ttl time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims{jwt.StandardClaims{
		ExpiresAt: time.Now().Add(ttl).Unix(),
		Subject:   companyId,
	}, companyId})

	return token.SignedString([]byte(m.key))
}
func (m *Manager) Parse(accessToken string) (string, error) {
	// token, err := jwt.ParseWithClaims(accessToken, &tokenClaims{}, func(token *jwt.Token) (i interface{}, err error) {
	// 	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
	// 		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	// 	}
	// 	return []byte(m.key), nil
	// })
	// if err != nil {
	// 	return "", err
	// }
	
	// claims, ok := token.Claims.(*tokenClaims)
	// if !ok {
	// 	return "", fmt.Errorf("error get user claims from token")
	// }


	//return claims.CompanyName, nil
	return "test",nil
}
