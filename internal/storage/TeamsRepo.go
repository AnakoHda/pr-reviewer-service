package storage

import (
	"pr-reviewer-service/internal/domain"
	"context"
)

func (pgPepo *Repository) CreateTeam(ctx context.Context, team domain.Team) error {
	tx, err := pgPepo.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	const insertTeamQ = `
		INSERT INTO teams (name)
		VALUES ($1)
		RETURNING id
	`
	var teamID int
	err = tx.QueryRowContext(ctx, insertTeamQ, team.TeamName).Scan(&teamID)
	if err != nil {
		//если существует такой TeamName
		if IsErrorPGAlreadyExist(err) {
			return domain.ErrTeamAlreadyExists
		}
		return err
	}

	const insertOrUpdateUserQ = `
	INSERT INTO users (id, username, team_id, is_active)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (id) DO UPDATE SET
		username  = EXCLUDED.username,
		team_id   = EXCLUDED.team_id,
		is_active = EXCLUDED.is_active
	`

	for _, u := range team.Members {
		if _, err := tx.ExecContext(ctx, insertOrUpdateUserQ,
			u.UserId, u.Username,
			teamID, u.IsActive,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}
