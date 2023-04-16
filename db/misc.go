package db

import "os"

type AuthUser struct {
	ID    uint     `json:"id"`
	Roles []string `json:"roles"`
}

func NewAuthUser(user User) AuthUser {
	return AuthUser{
		ID:    user.ID,
		Roles: user.Roles,
	}
}

type Thumb struct {
	Path string `gorm:"column:thumb_path"`
	Url  string `gorm:"column:thumb_url"`
}

func (t *Thumb) Delete() {
	_ = os.Remove(t.Path)
}
