package repo

import (
	"anileha/db"
	"anileha/rest"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserRepo struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewUserRepo(db *gorm.DB, log *zap.Logger) *UserRepo {
	return &UserRepo{
		db:  db,
		log: log,
	}
}

func (r *UserRepo) GetById(id uint) (*db.User, error) {
	var user db.User
	queryResult := r.db.First(&user, "id = ?", id)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return nil, nil
	}
	return &user, nil
}

func (r *UserRepo) GetByLogin(login string) (*db.User, error) {
	var user db.User
	queryResult := r.db.First(&user, "login = ?", login)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return nil, nil
	}
	return &user, nil
}

func (r *UserRepo) GetByEmail(email string) (*db.User, error) {
	var user db.User
	queryResult := r.db.First(&user, "email = ?", email)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return nil, nil
	}
	return &user, nil
}

func (r *UserRepo) Create(user *db.User) (*uint, error) {
	queryResult := r.db.Create(user)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return nil, rest.ErrCreationFailed
	}
	return &user.ID, nil
}

var UserRepoExport = fx.Options(fx.Provide(NewUserRepo))
