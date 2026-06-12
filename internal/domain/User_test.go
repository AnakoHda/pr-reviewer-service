package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		user, err := NewUser("user-1", "alice", "team-a", true)
		assert.NoError(t, err)
		assert.Equal(t, UserId("user-1"), user.UserId)
		assert.Equal(t, "alice", user.Username)
		assert.Equal(t, "team-a", user.TeamName)
		assert.True(t, user.IsActive)
	})

	t.Run("validation error", func(t *testing.T) {
		_, err := NewUser("", "alice", "team-a", true)
		assert.ErrorIs(t, err, ErrEmptyUserID)
	})
}

func TestUser_UpdateUser(t *testing.T) {
	user := &User{UserId: "user-1", Username: "alice", TeamName: "team-a", IsActive: true}

	t.Run("success", func(t *testing.T) {
		ok := user.UpdateUser("bob", "team-b", false)
		assert.True(t, ok)
		assert.Equal(t, "bob", user.Username)
		assert.Equal(t, "team-b", user.TeamName)
		assert.False(t, user.IsActive)
	})

	t.Run("validation failure", func(t *testing.T) {
		ok := user.UpdateUser("", "team-b", false)
		assert.False(t, ok)
	})
}
