package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTeam(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		members := []User{{UserId: "1"}}
		team, err := NewTeam("team-a", members)
		assert.NoError(t, err)
		assert.Equal(t, "team-a", team.TeamName)
		assert.Equal(t, members, team.Members)
	})

	t.Run("empty name", func(t *testing.T) {
		_, err := NewTeam("", nil)
		assert.ErrorIs(t, err, ErrEmptyTeamName)
	})
}

func TestTeam_UpdateTeam(t *testing.T) {
	team := &Team{TeamName: "old"}
	ok := team.UpdateTeam("new", nil)
	assert.True(t, ok)
	assert.Equal(t, "new", team.TeamName)

	ok = team.UpdateTeam("", nil)
	assert.False(t, ok)
}
