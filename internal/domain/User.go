package domain

type UserId string
type User struct {
	UserId   UserId
	Username string
	TeamName string
	IsActive bool
}

func NewUser(userId UserId, username, teamName string, active bool) (*User, error) {
	if err := ValidateUserFields(userId, username, teamName); err != nil {
		return nil, err
	}

	return &User{
		UserId:   userId,
		Username: username,
		TeamName: teamName,
		IsActive: active,
	}, nil
}

func (u *User) UpdateUser(username, teamName string, active bool) bool {
	if err := ValidateUserFields(u.UserId, username, teamName); err != nil {
		return false
	}
	u.Username = username
	u.TeamName = teamName
	u.IsActive = active
	return true
}

func ValidateUserFields(userId UserId, username, teamName string) error {
	if userId == "" {
		return ErrEmptyUserID
	}
	if username == "" {
		return ErrEmptyUsername
	}
	if teamName == "" {
		return ErrEmptyTeamName
	}
	return nil
}
