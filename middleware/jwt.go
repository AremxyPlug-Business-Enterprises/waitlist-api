package middleware

import (
	"crypto/rsa"
	"fmt"
	"log"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type AuthConn struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func NewAuthConn(privateKey, publicKey string) *AuthConn {

	priKey, err := generatePrivateKey(privateKey)
	if err != nil {
		fmt.Printf("Error generating private key: %s", err)
		log.Fatal(err)
	}

	pubKey, err := generatePublicKey(publicKey)
	if err != nil {
		fmt.Printf("Error generating public key: %s", err)
		log.Fatal(err)
	}

	return &AuthConn{
		privateKey: priKey,
		publicKey:  pubKey,
	}
}

func (a *AuthConn) GenerateJWT(email string) (string, error) {
	claims := jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(a.privateKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (a *AuthConn) ValidateJWT(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	return token, nil
}

func generatePublicKey(publicKey string) (*rsa.PublicKey, error) {
	tokenGeneratorPublicKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(publicKey))
	if err != nil {
		return nil, err
	}
	return tokenGeneratorPublicKey, nil
}

func generatePrivateKey(privateKey string) (*rsa.PrivateKey, error) {
	tokenGeneratorPrivateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKey))
	if err != nil {
		return nil, err
	}

	return tokenGeneratorPrivateKey, nil
}
