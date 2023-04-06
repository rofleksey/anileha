package db

type AuthUser struct {
	ID    uint
	Login string
	Admin bool
}

func NewAuthUser(user User) AuthUser {
	return AuthUser{
		ID:    user.ID,
		Login: user.Login,
		Admin: user.Admin,
	}
}
