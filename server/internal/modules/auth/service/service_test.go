package service_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agentteams/server/internal/modules/auth/service"
)

func TestJWTService(t *testing.T) {
	config := service.JWTConfig{
		Secret:        "test-secret-key",
		Expiry:        time.Hour,
		RefreshExpiry: 24 * time.Hour,
	}

	jwtService := service.NewJWTService(config)

	t.Run("generate and validate token", func(t *testing.T) {
		token, err := jwtService.GenerateToken("user-123", "testuser", "admin")
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		claims, err := jwtService.ValidateToken(token)
		require.NoError(t, err)
		assert.Equal(t, "user-123", claims.UserID)
		assert.Equal(t, "testuser", claims.Username)
		assert.Equal(t, "admin", claims.Role)
	})

	t.Run("generate and validate refresh token", func(t *testing.T) {
		token, err := jwtService.GenerateRefreshToken("user-123", "testuser", "admin")
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		claims, err := jwtService.ValidateToken(token)
		require.NoError(t, err)
		assert.Equal(t, "user-123", claims.UserID)
	})

	t.Run("invalid token", func(t *testing.T) {
		_, err := jwtService.ValidateToken("invalid-token")
		assert.Error(t, err)
		assert.Equal(t, service.ErrInvalidToken, err)
	})

	t.Run("wrong secret", func(t *testing.T) {
		token, err := jwtService.GenerateToken("user-123", "testuser", "admin")
		require.NoError(t, err)

		// Create another service with different secret
		wrongConfig := service.JWTConfig{
			Secret:        "wrong-secret",
			Expiry:        time.Hour,
			RefreshExpiry: 24 * time.Hour,
		}
		wrongService := service.NewJWTService(wrongConfig)

		_, err = wrongService.ValidateToken(token)
		assert.Error(t, err)
	})
}

func TestServiceErrors(t *testing.T) {
	// Test that error variables are defined
	assert.Equal(t, "user not found", service.ErrUserNotFound.Error())
	assert.Equal(t, "user already exists", service.ErrUserExists.Error())
	assert.Equal(t, "invalid credentials", service.ErrInvalidCredentials.Error())
	assert.Equal(t, "account is locked", service.ErrAccountLocked.Error())
	assert.Equal(t, "invalid token", service.ErrInvalidToken.Error())
	assert.Equal(t, "token expired", service.ErrTokenExpired.Error())
}
