package db

import "os"

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

type Thumb struct {
	Path string `gorm:"column:thumb_path"`
	Url  string `gorm:"column:thumb_url"`
}

func (t *Thumb) Delete() {
	_ = os.Remove(t.Path)
}
