package storage

import (
	"pr-reviewer-service/internal/domain"
	"context"
	"errors"
)

func (pgPepo *Repository) ListPullRequestsByAuthorID(ctx context.Context, authorID domain.UserId) ([]domain.PullRequest, error) {
	const selectPRsByAuthorQ = `
SELECT id, pull_requests_name, author_id, status, created_at, merged_at
FROM pull_requests
WHERE author_id = $1
`
	rows, err := pgPepo.db.QueryContext(ctx, selectPRsByAuthorQ, authorID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	result := make([]domain.PullRequest, 0)
	for rows.Next() {
		var tmpPR domain.PullRequest
		err := rows.Scan(
			&tmpPR.PullRequestId,
			&tmpPR.PullRequestName,
			&tmpPR.AuthorId,
			&tmpPR.Status,
			&tmpPR.CreatedAt,
			&tmpPR.MergedAt,
		)
		if err != nil {
			return nil, err
		}

		// Для каждого PR нужно достать ревьюверов (N+1, но сохраняем текущий паттерн репозитория)
		const reviewersSelectQ = `
SELECT reviewer_id
FROM pull_request_reviewers
WHERE pull_request_id = $1
`
		revRows, err := pgPepo.db.QueryContext(ctx, reviewersSelectQ, tmpPR.PullRequestId)
		if err != nil {
			return nil, err
		}
		
		tmpPR.AssignedReviewers = make([]domain.UserId, 0)
		for revRows.Next() {
			var revID domain.UserId
			if err := revRows.Scan(&revID); err != nil {
				_ = revRows.Close()
				return nil, err
			}
			tmpPR.AssignedReviewers = append(tmpPR.AssignedReviewers, revID)
		}
		_ = revRows.Close()

		newPR, err := domain.NewPullRequest(
			tmpPR.PullRequestId, tmpPR.PullRequestName,
			tmpPR.AuthorId, tmpPR.AssignedReviewers,
		)
		if err == nil {
			result = append(result, *newPR)
		}
	}

	return result, nil
}

func (pgPepo *Repository) ListPullRequestsByReviewerID(ctx context.Context, reviewerID domain.UserId) ([]domain.PullRequest, error) {
	//достаём все строки PRID где reviewer_id=reviewerID
	const selectPullRequestIDsQ = `
SELECT pull_request_id
FROM pull_request_reviewers
WHERE reviewer_id = $1
`
	rows, err := pgPepo.db.QueryContext(ctx, selectPullRequestIDsQ, reviewerID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	prIDs := make([]domain.PullRequestId, 0)
	//собираем в список всех PullRequestId
	for rows.Next() {
		var prID domain.PullRequestId
		if err := rows.Scan(&prID); err != nil {
			return nil, err
		}
		prIDs = append(prIDs, prID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	//Если не нашли ни одного то и возвращаем пустой список
	if len(prIDs) == 0 {
		return []domain.PullRequest{}, nil
	}

	result := make([]domain.PullRequest, 0, len(prIDs))

	//Имея PullRequestId достаём все PullRequests
	for _, prID := range prIDs {
		pr, err := pgPepo.GetPullRequestByID(ctx, prID)
		if err != nil {
			//если не оказалось PR с ID, то пропускаем
			if errors.Is(err, domain.ErrPullRequestNotFound) {
				continue
			}
			return nil, err
		}
		result = append(result, *pr)
	}

	return result, nil
}
func (pgPepo *Repository) GetPullRequestByID(ctx context.Context, pullRequestId domain.PullRequestId) (*domain.PullRequest, error) {
	//достаём PullRequesr из таблицы pull_requests
	const pullRequestSelectQ = `
SELECT id, pull_requests_name, author_id, status, created_at, merged_at
FROM pull_requests
WHERE id = $1
`
	var tmpPR domain.PullRequest

	err := pgPepo.db.QueryRowContext(ctx, pullRequestSelectQ, pullRequestId).Scan(
		&tmpPR.PullRequestId,
		&tmpPR.PullRequestName,
		&tmpPR.AuthorId,
		&tmpPR.Status,
		&tmpPR.CreatedAt,
		&tmpPR.MergedAt,
	)
	if err != nil {
		//если PR не найден
		if IsErrorPGNotExist(err) {
			return nil, domain.ErrPullRequestNotFound
		}
		return nil, err
	}

	//Достаём reviewers из таблицы pull_request_reviewers для этого RP
	const reviewersSelectQ = `
SELECT reviewer_id
FROM pull_request_reviewers
WHERE pull_request_id = $1
`
	rows, err := pgPepo.db.QueryContext(ctx, reviewersSelectQ, pullRequestId)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	tmpPR.AssignedReviewers = make([]domain.UserId, 0, 2)
	//собирвем PR и reviewers
	for rows.Next() {
		var reviewerID domain.UserId
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, err
		}
		tmpPR.AssignedReviewers = append(tmpPR.AssignedReviewers, reviewerID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	//Отправляем валидные данные
	return domain.NewPullRequest(
		tmpPR.PullRequestId, tmpPR.PullRequestName,
		tmpPR.AuthorId, tmpPR.AssignedReviewers,
	)
}
func (pgPepo *Repository) CreatePullRequest(ctx context.Context, pullRequest domain.PullRequest) error {
	tx, err := pgPepo.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	const pullRequestInsertQ = `
INSERT INTO pull_requests (id, pull_requests_name, author_id, status, created_at, merged_at)
VALUES ($1, $2, $3, $4, $5, $6)
`
	_, err = tx.ExecContext(ctx, pullRequestInsertQ,
		pullRequest.PullRequestId, pullRequest.PullRequestName,
		pullRequest.AuthorId, pullRequest.Status,
		pullRequest.CreatedAt, pullRequest.MergedAt,
	)
	if err != nil {
		if IsErrorPGAlreadyExist(err) {
			return domain.ErrPullRequestAlreadyExists
		}
		return err
	}

	const insertReviewerQ = `
INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id)
VALUES ($1, $2)
`
	for _, reviewer := range pullRequest.AssignedReviewers {
		if _, err := tx.ExecContext(ctx, insertReviewerQ, pullRequest.PullRequestId, reviewer); err != nil {
			//если уже существовала такая пара пропустить и добавить следующие
			if IsErrorPGAlreadyExist(err) {
				continue
			}
			return err
		}
	}

	return tx.Commit()
}

func (pgPepo *Repository) UpdatePullRequest(ctx context.Context, pullRequest domain.PullRequest) error {
	tx, err := pgPepo.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	//обновляем PR
	const pullRequestUpdateQ = `
UPDATE pull_requests
SET
    pull_requests_name = $2,
    author_id          = $3,
    status             = $4,
    created_at         = $5,
    merged_at          = $6
WHERE id = $1
`
	res, err := tx.ExecContext(ctx, pullRequestUpdateQ,
		pullRequest.PullRequestId, pullRequest.PullRequestName,
		pullRequest.AuthorId, pullRequest.Status,
		pullRequest.CreatedAt, pullRequest.MergedAt,
	)
	if err != nil {
		return err
	}
	// если нет изменений, значит не существует
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrPullRequestNotFound
	}
	//удаляем старых reviewer
	const deleteReviewersQ = `
DELETE FROM pull_request_reviewers
WHERE pull_request_id = $1
`
	if _, err := tx.ExecContext(ctx, deleteReviewersQ, pullRequest.PullRequestId); err != nil {
		return err
	}
	//создаём новых из списка AssignedReviewers
	const insertReviewerQ = `
INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id)
VALUES ($1, $2)
`
	for _, reviewer := range pullRequest.AssignedReviewers {
		if _, err := tx.ExecContext(ctx, insertReviewerQ, pullRequest.PullRequestId, reviewer); err != nil {
			//если уже существовала такая пара пропустить и добавить следующие
			if IsErrorPGAlreadyExist(err) {
				continue
			}
			return err
		}
	}

	return tx.Commit()
}
