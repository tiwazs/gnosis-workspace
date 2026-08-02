package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Claims matches the payload signed by gnosis-main-service:
//
//	jwt.sign({ userId, email }, TOKEN_KEY, { expiresIn: "1h" })
type Claims struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// Auth verifies the JWT the same way main-service authenticator does:
// Authorization: Bearer <token> | x-access-token | ?token=
func Auth() gin.HandlerFunc {
	return func(context *gin.Context) {
		tokenString := extractToken(context)
		if tokenString == "" {
			context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "A token is required for authentication"})
			return
		}

		secret := os.Getenv("TOKEN_KEY")
		if secret == "" {
			context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "TOKEN_KEY is not configured"})
			return
		}

		claims := &Claims{}
		parsedToken, err := jwt.ParseWithClaims(tokenString, claims, func(parsed *jwt.Token) (interface{}, error) {
			if _, isHMAC := parsed.Method.(*jwt.SigningMethodHMAC); !isHMAC {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})
		if err != nil || !parsedToken.Valid {
			context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}
		if claims.UserID == "" {
			context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token missing userId claim"})
			return
		}

		context.Set("user_id", claims.UserID)
		context.Set("email", claims.Email)
		context.Next()
	}
}

func extractToken(context *gin.Context) string {
	if authorization := context.GetHeader("Authorization"); authorization != "" {
		if strings.HasPrefix(strings.ToLower(authorization), "bearer ") {
			return strings.TrimSpace(authorization[7:])
		}
		return authorization
	}
	if accessToken := context.GetHeader("x-access-token"); accessToken != "" {
		return accessToken
	}
	if queryToken := context.Query("token"); queryToken != "" {
		return queryToken
	}
	return ""
}
