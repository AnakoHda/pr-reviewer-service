package storage

import (
	"pr-reviewer-service/internal/domain"
	"context"
)

func (pgPepo *Repository) GetUserByID(ctx context.Context, userID domain.UserId) (*domain.User, error) {
	const selectUserByIDQ = `
SELECT u.id,
	u.username,
	t.name AS team_name,
	u.is_active
FROM users u
JOIN teams t ON u.team_id = t.id
WHERE u.id = $1
    `
	//достаём юзера
	var tmpUser domain.User
	err := pgPepo.db.QueryRowContext(ctx, selectUserByIDQ, userID).Scan(
		&tmpUser.UserId,
		&tmpUser.Username,
		&tmpUser.TeamName,
		&tmpUser.IsActive,
	)
	if err != nil {
		if IsErrorPGNotExist(err) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return domain.NewUser(
		tmpUser.UserId, tmpUser.Username,
		tmpUser.TeamName, tmpUser.IsActive,
	)
}

func (pgPepo *Repository) UpdateUser(ctx context.Context, user domain.User) error {
	const selectTeamIDByTeamNameQ = `
SELECT id
FROM teams
WHERE name = $1
`
	var teamId int
	err := pgPepo.db.QueryRowContext(ctx, selectTeamIDByTeamNameQ, user.TeamName).Scan(
		&teamId,
	)
	if err != nil {
		if IsErrorPGNotExist(err) {
			return domain.ErrTeamNotFound
		}
		return err
	}

	const updateUserQ = `
UPDATE users
SET
    username = $2,
    team_id  = $3,
    is_active = $4
WHERE id = $1
`

	res, err := pgPepo.db.ExecContext(ctx, updateUserQ,
		user.UserId,
		user.Username,
		teamId,
		user.IsActive,
	)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (pgPepo *Repository) ListUsersByTeamName(ctx context.Context, teamName string) ([]domain.User, error) {
	const selectTeamIDByTeamNameQ = `
SELECT id
FROM teams
WHERE name = $1
`
	var teamID int
	err := pgPepo.db.QueryRowContext(ctx, selectTeamIDByTeamNameQ, teamName).Scan(&teamID)
	if err != nil {
		if IsErrorPGNotExist(err) {
			return nil, domain.ErrTeamNotFound
		}
		return nil, err
	}

	//Достаём всех пользователей по teamID
	const selectUsersByTeamIdQ = `
SELECT id, username, is_active
FROM users
WHERE team_id = $1
`
	rows, err := pgPepo.db.QueryContext(ctx, selectUsersByTeamIdQ, teamID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	users := make([]domain.User, 0)

	for rows.Next() {
		var tmpUser domain.User
		if err := rows.Scan(&tmpUser.UserId, &tmpUser.Username, &tmpUser.IsActive); err != nil {
			return nil, err
		}
		//если юзер битый, пропускаем
		newUser, err := domain.NewUser(tmpUser.UserId, tmpUser.Username, teamName, tmpUser.IsActive)
		if err != nil {
			continue
		}
		users = append(users, *newUser)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}
