package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeAndValidateJWT(t *testing.T) {
    userID := uuid.New()
    secret := "test-secret"

    token, err := MakeJWT(userID, secret, time.Hour)
    if err != nil {
        t.Fatalf("MakeJWT failed: %v", err)
    }

    gotID, err := ValidateJWT(token, secret)
    if err != nil {
        t.Fatalf("ValidateJWT failed: %v", err)
    }
    if gotID != userID {
        t.Errorf("expected userID %v, got %v", userID, gotID)
    }
}

func TestExpiredToken(t *testing.T) {
    userID := uuid.New()
    secret := "test-secret"

    token, err := MakeJWT(userID, secret, -time.Hour)
    if err != nil {
        t.Fatalf("MakeJWT failed: %v", err)
    }

    _, err = ValidateJWT(token, secret)
    if err == nil {
        t.Fatal("expected error for expired token, got nil")
    }
}

func TestWrongSecret(t *testing.T) {
    userID := uuid.New()

    token, err := MakeJWT(userID, "correct-secret", time.Hour)
    if err != nil {
        t.Fatalf("MakeJWT failed: %v", err)
    }

    _, err = ValidateJWT(token, "wrong-secret")
    if err == nil {
        t.Fatal("expected error for wrong secret, got nil")
    }
}

func TestGetBearerToken(t *testing.T) {
    tests := []struct {
        name        string
        header      string
        wantToken   string
        wantErr     bool
    }{
        {
            name:      "valid bearer token",
            header:    "Bearer my-token-string",
            wantToken: "my-token-string",
            wantErr:   false,
        },
        {
            name:    "missing authorization header",
            header:  "",
            wantErr: true,
        },
        {
            name:    "wrong scheme",
            header:  "Basic my-token-string",
            wantErr: true,
        },
        {
            name:      "extra whitespace",
            header:    "Bearer   my-token-string  ",
            wantToken: "my-token-string",
            wantErr:   false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            headers := http.Header{}
            if tt.header != "" {
                headers.Set("Authorization", tt.header)
            }

            got, err := GetBearerToken(headers)
            if (err != nil) != tt.wantErr {
                t.Errorf("GetBearerToken() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.wantToken {
                t.Errorf("GetBearerToken() = %q, want %q", got, tt.wantToken)
            }
        })
    }
}