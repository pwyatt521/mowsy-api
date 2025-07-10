package auth

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	t.Run("HashValidPassword", func(t *testing.T) {
		password := "mySecurePassword123!"

		hash, err := HashPassword(password)

		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.NotEqual(t, password, hash)
		assert.True(t, strings.HasPrefix(hash, "$2a$"))
	})

	t.Run("HashEmptyPassword", func(t *testing.T) {
		password := ""

		hash, err := HashPassword(password)

		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.True(t, strings.HasPrefix(hash, "$2a$"))
	})

	t.Run("HashLongPassword", func(t *testing.T) {
		password := strings.Repeat("a", 100) // Longer than bcrypt's 72-byte limit

		_, err := HashPassword(password)

		// bcrypt has a 72-byte limit, so this should error
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "password length exceeds 72 bytes")
	})

	t.Run("HashSpecialCharacters", func(t *testing.T) {
		password := "!@#$%^&*()_+-=[]{}|;:,.<>?"

		hash, err := HashPassword(password)

		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.True(t, strings.HasPrefix(hash, "$2a$"))
	})

	t.Run("HashUnicodePassword", func(t *testing.T) {
		password := "пароль123密码🔐"

		hash, err := HashPassword(password)

		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.True(t, strings.HasPrefix(hash, "$2a$"))
	})
}

func TestCheckPassword(t *testing.T) {
	t.Run("CheckCorrectPassword", func(t *testing.T) {
		password := "mySecurePassword123!"
		hash, err := HashPassword(password)
		require.NoError(t, err)

		isValid := CheckPassword(password, hash)

		assert.True(t, isValid)
	})

	t.Run("CheckIncorrectPassword", func(t *testing.T) {
		password := "mySecurePassword123!"
		wrongPassword := "wrongPassword123!"
		hash, err := HashPassword(password)
		require.NoError(t, err)

		isValid := CheckPassword(wrongPassword, hash)

		assert.False(t, isValid)
	})

	t.Run("CheckEmptyPassword", func(t *testing.T) {
		password := "mySecurePassword123!"
		hash, err := HashPassword(password)
		require.NoError(t, err)

		isValid := CheckPassword("", hash)

		assert.False(t, isValid)
	})

	t.Run("CheckPasswordWithEmptyHash", func(t *testing.T) {
		password := "mySecurePassword123!"

		isValid := CheckPassword(password, "")

		assert.False(t, isValid)
	})

	t.Run("CheckPasswordWithInvalidHash", func(t *testing.T) {
		password := "mySecurePassword123!"
		invalidHash := "invalid-hash-format"

		isValid := CheckPassword(password, invalidHash)

		assert.False(t, isValid)
	})

	t.Run("CheckCaseSensitivePassword", func(t *testing.T) {
		password := "mySecurePassword123!"
		caseChangedPassword := "MySecurePassword123!"
		hash, err := HashPassword(password)
		require.NoError(t, err)

		isValid := CheckPassword(caseChangedPassword, hash)

		assert.False(t, isValid)
	})

	t.Run("CheckSpecialCharacterPassword", func(t *testing.T) {
		password := "!@#$%^&*()_+-=[]{}|;:,.<>?"
		hash, err := HashPassword(password)
		require.NoError(t, err)

		isValid := CheckPassword(password, hash)

		assert.True(t, isValid)
	})

	t.Run("CheckUnicodePassword", func(t *testing.T) {
		password := "пароль123密码🔐"
		hash, err := HashPassword(password)
		require.NoError(t, err)

		isValid := CheckPassword(password, hash)

		assert.True(t, isValid)
	})
}

func TestPasswordRoundTrip(t *testing.T) {
	testCases := []string{
		"simplePassword",
		"ComplexP@ssw0rd!",
		"",
		strings.Repeat("a", 70), // within bcrypt limit
		"пароль123密码🔐",
		"!@#$%^&*()_+-=[]{}|;:,.<>?",
	}

	for _, password := range testCases {
		t.Run("RoundTrip_"+password, func(t *testing.T) {
			// Hash the password
			hash, err := HashPassword(password)
			require.NoError(t, err)

			// Verify the password matches
			isValid := CheckPassword(password, hash)
			assert.True(t, isValid)

			// Verify a different password doesn't match
			if password != "" {
				isValid = CheckPassword(password+"wrong", hash)
				assert.False(t, isValid)
			}
		})
	}
}

func TestHashPasswordConsistency(t *testing.T) {
	t.Run("DifferentHashesForSamePassword", func(t *testing.T) {
		password := "mySecurePassword123!"

		hash1, err1 := HashPassword(password)
		hash2, err2 := HashPassword(password)

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.NotEqual(t, hash1, hash2, "Hashes should be different due to salt")

		// Both should validate the same password
		assert.True(t, CheckPassword(password, hash1))
		assert.True(t, CheckPassword(password, hash2))
	})
}