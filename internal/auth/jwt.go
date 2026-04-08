package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret []byte

func init() {
	// #region agent log
	debugLog := func(msg string, data map[string]interface{}) {
		f, _ := os.OpenFile("/home/user/git/ai-for-developers-project-386/.cursor/debug-536161.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if f != nil {
			entry := map[string]interface{}{"id": "log_"+fmt.Sprint(time.Now().UnixNano()), "timestamp": time.Now().UnixMilli(), "location": "jwt.go:init", "message": msg, "data": data, "runId": "debug-run-1", "sessionId": "536161"}
			json.NewEncoder(f).Encode(entry)
			f.Close()
		}
	}
	// #endregion
	secret := os.Getenv("JWT_SECRET")
	// #region agent log
	debugLog("jwt_init_start", map[string]interface{}{"secret_from_env": secret != "", "secret_length": len(secret), "hypothesisId": "B"})
	// #endregion
	if secret == "" {
		// Use a default for development only
		secret = "dev-secret-key-minimum-32-characters-long"
	}
	// #region agent log
	debugLog("jwt_secret_final", map[string]interface{}{"final_length": len(secret), "will_panic": len(secret) < 32, "hypothesisId": "B"})
	// #endregion
	if len(secret) < 32 {
		panic("JWT_SECRET must be at least 32 characters for security")
	}
	jwtSecret = []byte(secret)
	// #region agent log
	debugLog("jwt_init_success", map[string]interface{}{"jwt_secret_length": len(jwtSecret), "hypothesisId": "B"})
	// #endregion
}

// Claims represents JWT claims
type Claims struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateToken creates a new JWT token for a user
func GenerateToken(userID, email string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateToken validates a JWT token and returns claims
func ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// SetSecret allows setting a custom JWT secret (for testing)
func SetSecret(secret string) {
	jwtSecret = []byte(secret)
}
