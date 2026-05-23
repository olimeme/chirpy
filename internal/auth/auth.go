package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
    hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
    if err != nil {
        return "", err
    }
    return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
    match, err := argon2id.ComparePasswordAndHash(password, hash)
    if err != nil {
        return false, err
    }
    return match, nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	now := time.Now().UTC()

    claims := jwt.RegisteredClaims{
        Issuer:    "chirpy-access",
        IssuedAt:  jwt.NewNumericDate(now),
        ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
        Subject:   userID.String(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

    return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
    token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(tokenSecret), nil
    })
    if err != nil {
        return uuid.Nil, err
    }

    claims, ok := token.Claims.(*jwt.RegisteredClaims)
    if !ok {
        return uuid.Nil, fmt.Errorf("invalid claims")
    }

    userID, err := uuid.Parse(claims.Subject)
    if err != nil {
        return uuid.Nil, fmt.Errorf("invalid subject: %w", err)
    }

    return userID, nil
}

func GetBearerToken(headers http.Header) (string, error) {
    authHeader := headers.Get("Authorization")
    if authHeader == "" {
        return "", fmt.Errorf("authorization header missing")
    }

    token, found := strings.CutPrefix(authHeader, "Bearer ")
    if !found {
        return "", fmt.Errorf("authorization header must start with 'Bearer '")
    }

    return strings.TrimSpace(token), nil
}