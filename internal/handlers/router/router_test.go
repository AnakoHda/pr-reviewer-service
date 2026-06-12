package router_test

import (
	"pr-reviewer-service/internal/generated/api/dto"
	"pr-reviewer-service/internal/handlers"
	"pr-reviewer-service/internal/handlers/router"
	"pr-reviewer-service/internal/service/pullRequestService"
	"pr-reviewer-service/internal/service/teamService"
	"pr-reviewer-service/internal/service/userService"
	"pr-reviewer-service/internal/storage"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// setupTestEnv поднимает БД и собирает весь роутер со всеми сервисами
func setupTestEnv(t *testing.T) (http.Handler, func()) {
	ctx := context.Background()

	// 1. Поднимаем базу
	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	require.NoError(t, err)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sqlx.Connect("postgres", connStr)
	require.NoError(t, err)

	// 2. Накатываем миграции
	_, b, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(b), "../../../migrations")

	require.NoError(t, goose.SetDialect("postgres"))
	require.NoError(t, goose.Up(db.DB, migrationsDir))

	// 3. Собираем приложение "по-настоящему"
	repo := storage.New(db)
	teamSvc := teamService.New(repo, repo)
	userSvc := userService.New(repo, repo) // Добавили второй аргумент
	prSvc := pullRequestService.New(repo, repo)

	mux := router.RegisterRoutes(prSvc, teamSvc, userSvc)

	cleanup := func() {
		_ = db.Close()
		_ = pgContainer.Terminate(ctx)
	}

	return mux, cleanup
}

func TestRouter_Integration_Team(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	mux, cleanup := setupTestEnv(t)
	defer cleanup()

	t.Run("Create and Get Team", func(t *testing.T) {
		// --- 1. POST /team/add ---
		reqBody := dto.Team{
			TeamName: "ops",
			Members: []dto.TeamMember{
				{UserId: "ops1", Username: "Alice", IsActive: true}, // Убрали TeamName
			},
		}
		bodyBytes, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		// Проверяем, что хендлер вернул 201 Created (согласно вашему коду)
		assert.Equal(t, http.StatusCreated, rec.Code)

		// --- 2. GET /team/get ---
		// Теперь проверяем, что данные реально сохранились и отдаются
		reqGet := httptest.NewRequest(http.MethodGet, "/team/get?team_name=ops", nil)
		recGet := httptest.NewRecorder()

		mux.ServeHTTP(recGet, reqGet)

		assert.Equal(t, http.StatusOK, recGet.Code)

		var getResp dto.Team
		require.NoError(t, json.Unmarshal(recGet.Body.Bytes(), &getResp))

		assert.Equal(t, "ops", getResp.TeamName)
		if assert.Len(t, getResp.Members, 1, "Response body was: %s", recGet.Body.String()) {
			assert.Equal(t, "Alice", getResp.Members[0].Username)
		}
	})

	t.Run("Create Duplicate Team (Error Case)", func(t *testing.T) {
		// Пытаемся создать команду, которая уже была создана в предыдущем подтесте
		reqBody := dto.Team{
			TeamName: "ops",
			Members:  []dto.TeamMember{}, // Изменили тип на TeamMember
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		// Проверяем, что вернулся статус 400 и кастомный код ошибки
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "team_name already exists")
	})

	t.Run("Get Non-Existent Team", func(t *testing.T) {
		reqGet := httptest.NewRequest(http.MethodGet, "/team/get?team_name=nowhere", nil)
		recGet := httptest.NewRecorder() // Исправили rec -> recGet

		mux.ServeHTTP(recGet, reqGet)

		assert.Equal(t, http.StatusNotFound, recGet.Code)
		assert.Contains(t, recGet.Body.String(), "NOT_FOUND") // Исправили rec -> recGet
	})
}

func TestRouter_Integration_Users(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	mux, cleanup := setupTestEnv(t)
	defer cleanup()

	t.Run("Set User Active Status", func(t *testing.T) {
		// Подготовка: создаем команду и юзера
		teamBody := dto.Team{
			TeamName: "devs",
			Members: []dto.TeamMember{
				{UserId: "u1", Username: "pavel", IsActive: true},
			},
		}
		b, _ := json.Marshal(teamBody)
		reqAdd := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(b))
		mux.ServeHTTP(httptest.NewRecorder(), reqAdd)

		// 1. Деактивируем юзера
		reqBody := dto.PostUsersSetIsActiveJSONBody{
			UserId:   "u1",
			IsActive: false,
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewReader(bodyBytes))
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp handlers.UserResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.False(t, resp.User.IsActive)
	})

	t.Run("User Not Found Status Update", func(t *testing.T) {
		reqBody := dto.PostUsersSetIsActiveJSONBody{UserId: "missing", IsActive: true}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewReader(bodyBytes))
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestRouter_Integration_PullRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	mux, cleanup := setupTestEnv(t)
	defer cleanup()

	// Подготовка: Создаем команду с автором и ревьювером
	teamBody := dto.Team{
		TeamName: "qa",
		Members: []dto.TeamMember{
			{UserId: "author1", Username: "Alice", IsActive: true},
			{UserId: "rev1", Username: "Bob", IsActive: true},
			{UserId: "rev2", Username: "Charlie", IsActive: true},
		},
	}
	b, _ := json.Marshal(teamBody)
	reqAdd := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(b))
	mux.ServeHTTP(httptest.NewRecorder(), reqAdd)

	t.Run("Create PR and List for Reviewer", func(t *testing.T) {
		// 1. Создаем PR
		prBody := dto.PostPullRequestCreateJSONBody{
			PullRequestId:   "pr-1",
			PullRequestName: "Fix bug",
			AuthorId:        "author1",
		}
		pb, _ := json.Marshal(prBody)
		reqCreate := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(pb))
		recCreate := httptest.NewRecorder()
		mux.ServeHTTP(recCreate, reqCreate)

		assert.Equal(t, http.StatusCreated, recCreate.Code)

		var createResp handlers.PullRequestCreateResponse
		require.NoError(t, json.Unmarshal(recCreate.Body.Bytes(), &createResp))
		reviewerID := createResp.PR.AssignedReviewers[0]

		// 2. Проверяем список PR для ревьювера
		reqList := httptest.NewRequest(http.MethodGet, "/users/getReview?user_id="+reviewerID, nil)
		recList := httptest.NewRecorder()
		mux.ServeHTTP(recList, reqList)

		assert.Equal(t, http.StatusOK, recList.Code)
		var listResp handlers.PullRequestsShortResponse
		require.NoError(t, json.Unmarshal(recList.Body.Bytes(), &listResp))
		assert.NotEmpty(t, listResp.PullRequestsShort)
		assert.Equal(t, "pr-1", listResp.PullRequestsShort[0].PullRequestId)
	})

	t.Run("Merge PR", func(t *testing.T) {
		mergeBody := dto.PostPullRequestMergeJSONBody{PullRequestId: "pr-1"}
		mb, _ := json.Marshal(mergeBody)
		reqMerge := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewReader(mb))
		recMerge := httptest.NewRecorder()
		mux.ServeHTTP(recMerge, reqMerge)

		assert.Equal(t, http.StatusOK, recMerge.Code)
		var mergeResp handlers.PullRequestMergeResponse
		require.NoError(t, json.Unmarshal(recMerge.Body.Bytes(), &mergeResp))
		assert.Equal(t, "MERGED", string(mergeResp.PR.Status))
	})
}
