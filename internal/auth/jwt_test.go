package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTManager_GenerateToken(t *testing.T) {
	secretKey := "test-secret-key-at-least-32-chars-long"
	manager := NewJWTManager(secretKey, 1*time.Hour)

	token, err := manager.GenerateToken("user123", "user@example.com", "basic")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Error("Generated token is empty")
	}

	parts := 0
	for _, ch := range token {
		if ch == '.' {
			parts++
		}
	}
	if parts != 2 {
		t.Errorf("Token should have 3 parts, got %d parts", parts+1)
	}
}

func TestJWTManager_ValidateToken(t *testing.T) {
	secretKey := "test-secret-key-at-least-32-chars-long"
	manager := NewJWTManager(secretKey, 1*time.Hour)

	userID := "user123"
	email := "user@example.com"
	authType := "oauth2"

	token, err := manager.GenerateToken(userID, email, authType)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := manager.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, claims.UserID)
	}

	if claims.Email != email {
		t.Errorf("Expected Email %s, got %s", email, claims.Email)
	}

	if claims.AuthType != authType {
		t.Errorf("Expected AuthType %s, got %s", authType, claims.AuthType)
	}

	if claims.Issuer != "qaynaq" {
		t.Errorf("Expected Issuer 'qaynaq', got %s", claims.Issuer)
	}
}

func TestJWTManager_ValidateToken_InvalidSignature(t *testing.T) {
	secretKey := "test-secret-key-at-least-32-chars-long"
	manager := NewJWTManager(secretKey, 1*time.Hour)

	token, err := manager.GenerateToken("user123", "user@example.com", "basic")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	differentManager := NewJWTManager("different-secret-key-32-chars-min", 1*time.Hour)

	_, err = differentManager.ValidateToken(token)
	if err == nil {
		t.Error("Expected error when validating token with wrong secret, got nil")
	}
}

func TestJWTManager_ValidateToken_Expired(t *testing.T) {
	secretKey := "test-secret-key-at-least-32-chars-long"
	manager := NewJWTManager(secretKey, 1*time.Millisecond)

	token, err := manager.GenerateToken("user123", "user@example.com", "basic")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	_, err = manager.ValidateToken(token)
	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}
}

func TestJWTManager_ValidateToken_InvalidFormat(t *testing.T) {
	secretKey := "test-secret-key-at-least-32-chars-long"
	manager := NewJWTManager(secretKey, 1*time.Hour)

	invalidTokens := []string{
		"",
		"invalid",
		"invalid.token",
		"invalid.token.format.extra",
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.signature",
	}

	for _, token := range invalidTokens {
		_, err := manager.ValidateToken(token)
		if err == nil {
			t.Errorf("Expected error for invalid token '%s', got nil", token)
		}
	}
}

func TestJWTManager_RefreshToken(t *testing.T) {
	secretKey := "test-secret-key-at-least-32-chars-long"
	manager := NewJWTManager(secretKey, 1*time.Hour)

	userID := "user123"
	email := "user@example.com"
	authType := "basic"

	originalToken, err := manager.GenerateToken(userID, email, authType)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	newToken, err := manager.RefreshToken(originalToken)
	if err != nil {
		t.Fatalf("Failed to refresh token: %v", err)
	}

	if newToken == "" {
		t.Error("Refreshed token is empty")
	}

	claims, err := manager.ValidateToken(newToken)
	if err != nil {
		t.Fatalf("Failed to validate refreshed token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected UserID %s in refreshed token, got %s", userID, claims.UserID)
	}

	if claims.Email != email {
		t.Errorf("Expected Email %s in refreshed token, got %s", email, claims.Email)
	}

	if claims.AuthType != authType {
		t.Errorf("Expected AuthType %s in refreshed token, got %s", authType, claims.AuthType)
	}
}

func TestJWTManager_RefreshToken_Expired(t *testing.T) {
	secretKey := "test-secret-key-at-least-32-chars-long"
	manager := NewJWTManager(secretKey, 1*time.Millisecond)

	token, err := manager.GenerateToken("user123", "user@example.com", "basic")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	_, err = manager.RefreshToken(token)
	if err == nil {
		t.Error("Expected error when refreshing expired token, got nil")
	}
}

func TestJWTManager_DefaultDuration(t *testing.T) {
	secretKey := "test-secret-key-at-least-32-chars-long"
	manager := NewJWTManager(secretKey, 0) // 0 should use default

	token, err := manager.GenerateToken("user123", "user@example.com", "basic")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := manager.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	expectedExpiration := time.Now().Add(24 * time.Hour)
	actualExpiration := claims.ExpiresAt.Time

	diff := actualExpiration.Sub(expectedExpiration)
	if diff < -1*time.Minute || diff > 1*time.Minute {
		t.Errorf("Expected expiration around %v, got %v (diff: %v)", expectedExpiration, actualExpiration, diff)
	}
}

func TestJWTManager_TokenContainsExpectedClaims(t *testing.T) {
	secretKey := "test-secret-key-at-least-32-chars-long"
	manager := NewJWTManager(secretKey, 1*time.Hour)

	token, err := manager.GenerateToken("user123", "user@example.com", "basic")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	parsedToken, _, err := jwt.NewParser().ParseUnverified(token, &JWTClaims{})
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}

	claims, ok := parsedToken.Claims.(*JWTClaims)
	if !ok {
		t.Fatal("Failed to cast claims")
	}

	if claims.UserID == "" {
		t.Error("UserID claim is empty")
	}

	if claims.Email == "" {
		t.Error("Email claim is empty")
	}

	if claims.AuthType == "" {
		t.Error("AuthType claim is empty")
	}

	if claims.Issuer != "qaynaq" {
		t.Errorf("Expected Issuer 'qaynaq', got '%s'", claims.Issuer)
	}

	if claims.IssuedAt == nil {
		t.Error("IssuedAt claim is nil")
	}

	if claims.ExpiresAt == nil {
		t.Error("ExpiresAt claim is nil")
	}
}

func TestJWTManager_MultipleAuthTypes(t *testing.T) {
	secretKey := "test-secret-key-at-least-32-chars-long"
	manager := NewJWTManager(secretKey, 1*time.Hour)

	authTypes := []string{"basic", "oauth2", "custom"}

	for _, authType := range authTypes {
		token, err := manager.GenerateToken("user123", "user@example.com", authType)
		if err != nil {
			t.Fatalf("Failed to generate token for auth type %s: %v", authType, err)
		}

		claims, err := manager.ValidateToken(token)
		if err != nil {
			t.Fatalf("Failed to validate token for auth type %s: %v", authType, err)
		}

		if claims.AuthType != authType {
			t.Errorf("Expected AuthType %s, got %s", authType, claims.AuthType)
		}
	}
}
