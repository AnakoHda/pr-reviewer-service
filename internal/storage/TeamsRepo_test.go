package storage

import (
	"pr-reviewer-service/internal/domain"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTeamsRepo_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := New(db)
	ctx := context.Background()

	t.Run("Create Team Success", func(t *testing.T) {
		teamName := "infrastructure"
		users := []domain.User{
			{UserId: "i1", Username: "ivan", IsActive: true},
			{UserId: "i2", Username: "igor", IsActive: true},
		}

		team := domain.Team{
			TeamName: teamName,
			Members:  users,
		}

		err := repo.CreateTeam(ctx, team)
		assert.NoError(t, err)

		// Проверяем, что пользователи создались и привязаны к команде
		gotUsers, err := repo.ListUsersByTeamName(ctx, teamName)
		assert.NoError(t, err)
		assert.Len(t, gotUsers, 2)

		usernames := []string{gotUsers[0].Username, gotUsers[1].Username}
		assert.Contains(t, usernames, "ivan")
		assert.Contains(t, usernames, "igor")
	})

	t.Run("Create Team Duplicate Name Error", func(t *testing.T) {
		teamName := "duplicate-team"
		team := domain.Team{TeamName: teamName}

		err := repo.CreateTeam(ctx, team)
		assert.NoError(t, err)

		// Повторное создание той же команды
		err = repo.CreateTeam(ctx, team)
		assert.ErrorIs(t, err, domain.ErrTeamAlreadyExists)
	})

	t.Run("Create Team Upsert Users", func(t *testing.T) {
		// 1. Создаем первую команду с пользователем
		team1Name := "team-alpha"
		user := domain.User{UserId: "alpha-1", Username: "user-old", IsActive: true}

		err := repo.CreateTeam(ctx, domain.Team{
			TeamName: team1Name,
			Members:  []domain.User{user},
		})
		assert.NoError(t, err)

		// 2. Создаем вторую команду и "перетягиваем" туда того же пользователя с обновлением данных
		team2Name := "team-beta"
		updatedUser := domain.User{UserId: "alpha-1", Username: "user-new", IsActive: false}

		err = repo.CreateTeam(ctx, domain.Team{
			TeamName: team2Name,
			Members:  []domain.User{updatedUser},
		})
		assert.NoError(t, err)

		// 3. Проверяем, что пользователь теперь в новой команде и данные обновились
		gotUser, err := repo.GetUserByID(ctx, "alpha-1")
		assert.NoError(t, err)
		assert.Equal(t, "user-new", gotUser.Username)
		assert.Equal(t, team2Name, gotUser.TeamName)
		assert.False(t, gotUser.IsActive)
	})
}
