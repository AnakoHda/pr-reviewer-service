package storage

import (
	"pr-reviewer-service/internal/domain"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsersRepo_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := New(db)
	ctx := context.Background()

	t.Run("Create and Get User", func(t *testing.T) {
		// 1. Создаем команду с пользователями
		teamName := "backend"
		user1 := domain.User{UserId: "u1", Username: "alice", TeamName: teamName, IsActive: true}
		user2 := domain.User{UserId: "u2", Username: "bob", TeamName: teamName, IsActive: true}

		team := domain.Team{
			TeamName: teamName,
			Members:  []domain.User{user1, user2},
		}

		err := repo.CreateTeam(ctx, team)
		assert.NoError(t, err)

		// 2. Пытаемся достать пользователя
		gotUser, err := repo.GetUserByID(ctx, "u1")
		assert.NoError(t, err)
		assert.Equal(t, user1.UserId, gotUser.UserId)
		assert.Equal(t, user1.Username, gotUser.Username)
		assert.Equal(t, teamName, gotUser.TeamName)
		assert.True(t, gotUser.IsActive)
	})

	t.Run("Update User", func(t *testing.T) {
		teamName := "frontend-update"
		userID := domain.UserId("u3")

		err := repo.CreateTeam(ctx, domain.Team{
			TeamName: teamName,
			Members: []domain.User{
				{UserId: userID, Username: "charlie", TeamName: teamName, IsActive: true},
			},
		})
		assert.NoError(t, err)

		// 2. Обновляем пользователя
		updatedUser := domain.User{
			UserId:   userID,
			Username: "charlie_new",
			TeamName: teamName,
			IsActive: false,
		}
		err = repo.UpdateUser(ctx, updatedUser)
		assert.NoError(t, err)

		// 3. Проверяем изменения
		gotUser, err := repo.GetUserByID(ctx, userID)
		assert.NoError(t, err)
		if gotUser != nil {
			assert.Equal(t, "charlie_new", gotUser.Username)
			assert.False(t, gotUser.IsActive)
		}
	})
	t.Run("List Users By Team Name", func(t *testing.T) {
		teamName := "devops"
		users := []domain.User{
			{UserId: "d1", Username: "dave", TeamName: teamName, IsActive: true},
			{UserId: "d2", Username: "eve", TeamName: teamName, IsActive: true},
		}
		err := repo.CreateTeam(ctx, domain.Team{TeamName: teamName, Members: users})
		assert.NoError(t, err)

		gotUsers, err := repo.ListUsersByTeamName(ctx, teamName)
		assert.NoError(t, err)
		assert.Len(t, gotUsers, 2)

		usernames := []string{gotUsers[0].Username, gotUsers[1].Username}
		assert.Contains(t, usernames, "dave")
		assert.Contains(t, usernames, "eve")
	})

	t.Run("User Not Found", func(t *testing.T) {
		_, err := repo.GetUserByID(ctx, "non-existent")
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
	})
}
