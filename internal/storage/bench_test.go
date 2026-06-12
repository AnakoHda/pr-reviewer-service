package storage

import (
	"pr-reviewer-service/internal/domain"
	"context"
	"fmt"
	"testing"
)

// BenchmarkCreateTeam измеряет производительность создания команды с разным количеством участников.
func BenchmarkCreateTeam(b *testing.B) {
	db, cleanup := setupTestDB(nil)
	defer cleanup()
	repo := New(db)
	ctx := context.Background()

	sizes := []int{1, 10, 100, 1000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				teamName := fmt.Sprintf("team-%d-%d-%d", size, i, b.N)
				users := make([]domain.User, size)
				for j := 0; j < size; j++ {
					users[j] = domain.User{
						UserId:   domain.UserId(fmt.Sprintf("u-%d-%d-%d-%d", size, i, j, b.N)),
						Username: fmt.Sprintf("name-%d", j),
						IsActive: true,
					}
				}

				_ = repo.CreateTeam(ctx, domain.Team{
					TeamName: teamName,
					Members:  users,
				})
			}
		})
	}
}

// BenchmarkListUsersByTeamName измеряет скорость получения списка пользователей разного размера.
func BenchmarkListUsersByTeamName(b *testing.B) {
	db, cleanup := setupTestDB(nil)
	defer cleanup()
	repo := New(db)
	ctx := context.Background()

	sizes := []int{100, 200, 300, 400}

	for _, size := range sizes {
		teamName := fmt.Sprintf("bench-team-%d", size)
		users := make([]domain.User, size)
		for i := 0; i < size; i++ {
			users[i] = domain.User{
				UserId:   domain.UserId(fmt.Sprintf("u-%d-%d", size, i)),
				Username: fmt.Sprintf("user-%d", i),
				IsActive: true,
			}
		}
		_ = repo.CreateTeam(ctx, domain.Team{TeamName: teamName, Members: users})

		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = repo.ListUsersByTeamName(ctx, teamName)
			}
		})
	}
}

// BenchmarkListPullRequestsByAuthorID измеряет скорость получения списка PR по автору.
func BenchmarkListPullRequestsByAuthorID(b *testing.B) {
	db, cleanup := setupTestDB(nil)
	defer cleanup()
	repo := New(db)
	ctx := context.Background()

	// Setup: Создаем автора и ревьювера
	teamName := "pr-bench-team"
	authorID := domain.UserId("bench-author")
	revID := domain.UserId("bench-rev")
	_ = repo.CreateTeam(ctx, domain.Team{
		TeamName: teamName,
		Members: []domain.User{
			{UserId: authorID, Username: "author", IsActive: true},
			{UserId: revID, Username: "reviewer", IsActive: true},
		},
	})

	sizes := []int{100, 200, 300, 400}

	for _, size := range sizes {
		// Предварительно создаем нужное количество PR
		for i := 0; i < size; i++ {
			prID := domain.PullRequestId(fmt.Sprintf("pr-%d-%d", size, i))
			pr, _ := domain.NewPullRequest(prID, "Bench PR", authorID, []domain.UserId{revID})
			_ = repo.CreatePullRequest(ctx, *pr)
		}

		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = repo.ListPullRequestsByAuthorID(ctx, authorID)
			}
		})
	}
}

// BenchmarkListPullRequestsByReviewerID измеряет скорость получения списка PR, где пользователь является ревьювером.
func BenchmarkListPullRequestsByReviewerID(b *testing.B) {
	db, cleanup := setupTestDB(nil)
	defer cleanup()
	repo := New(db)
	ctx := context.Background()

	// Setup: Создаем автора и ревьювера
	teamName := "rev-bench-team"
	authorID := domain.UserId("bench-author-rev")
	revID := domain.UserId("bench-rev-target")
	_ = repo.CreateTeam(ctx, domain.Team{
		TeamName: teamName,
		Members: []domain.User{
			{UserId: authorID, Username: "author", IsActive: true},
			{UserId: revID, Username: "reviewer", IsActive: true},
		},
	})

	sizes := []int{100, 200, 300, 400}

	for _, size := range sizes {
		// Предварительно создаем нужное количество PR, где наш целевой пользователь - ревьювер
		for i := 0; i < size; i++ {
			prID := domain.PullRequestId(fmt.Sprintf("pr-rev-%d-%d", size, i))
			pr, _ := domain.NewPullRequest(prID, "Bench PR", authorID, []domain.UserId{revID})
			_ = repo.CreatePullRequest(ctx, *pr)
		}

		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = repo.ListPullRequestsByReviewerID(ctx, revID)
			}
		})
	}
}

// BenchmarkDatabaseScale измеряет скорость получения одного пользователя при разном общем количестве записей в базе.
func BenchmarkDatabaseScale(b *testing.B) {
	db, cleanup := setupTestDB(nil)
	defer cleanup()
	repo := New(db)
	ctx := context.Background()

	scales := []int{100, 1000, 10000}

	for _, scale := range scales {
		// Наполняем базу до нужного масштаба
		teamName := fmt.Sprintf("scale-team-%d", scale)
		users := make([]domain.User, scale)
		for i := 0; i < scale; i++ {
			users[i] = domain.User{
				UserId:   domain.UserId(fmt.Sprintf("scale-u-%d-%d", scale, i)),
				Username: fmt.Sprintf("user-%d", i),
				IsActive: true,
			}
		}
		_ = repo.CreateTeam(ctx, domain.Team{TeamName: teamName, Members: users})

		// Будем искать последнего добавленного пользователя
		targetID := domain.UserId(fmt.Sprintf("scale-u-%d-%d", scale, scale-1))

		b.Run(fmt.Sprintf("DB-Size-%d", scale), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = repo.GetUserByID(ctx, targetID)
			}
		})
	}
}

// BenchmarkListUsersByTeamsVolume измеряет скорость получения списка пользователей одной команды
// при разном общем количестве команд в базе данных.
func BenchmarkListUsersByTeamsVolume(b *testing.B) {
	db, cleanup := setupTestDB(nil)
	defer cleanup()
	repo := New(db)
	ctx := context.Background()

	volumes := []int{100, 200, 300, 400}

	for _, volume := range volumes {
		// Наполняем базу командами
		for i := 0; i < volume; i++ {
			teamName := fmt.Sprintf("vol-team-%d-%d", volume, i)
			users := []domain.User{
				{UserId: domain.UserId(fmt.Sprintf("u-%d-%d", volume, i)), Username: "user", IsActive: true},
			}
			_ = repo.CreateTeam(ctx, domain.Team{TeamName: teamName, Members: users})
		}

		targetTeam := fmt.Sprintf("vol-team-%d-%d", volume, volume-1)

		b.Run(fmt.Sprintf("Teams-Count-%d", volume), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = repo.ListUsersByTeamName(ctx, targetTeam)
			}
		})
	}
}
