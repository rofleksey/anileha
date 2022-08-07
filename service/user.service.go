package service

import (
	"anileha/config"
	"anileha/db"
	"anileha/util"
	"bytes"
	"github.com/gofrs/uuid"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
	"gorm.io/gorm"
	"io/ioutil"
	"sync"
	"text/template"
	"time"
)

type registerTemplateVars struct {
	User string
	Link string
}

type UserService struct {
	db               *gorm.DB
	log              *zap.Logger
	dialer           *gomail.Dialer
	registerTemplate *template.Template
	registerMap      sync.Map
	baseUrl          string
	from             string
	subject          string
	salt             string
}

func NewUserService(
	lifecycle fx.Lifecycle,
	database *gorm.DB,
	log *zap.Logger,
	config *config.Config,
) (*UserService, error) {
	registerTemplateBytes, err := ioutil.ReadFile(config.Mail.RegisterTemplatePath)
	if err != nil {
		return nil, err
	}
	registerTemplateStr := string(registerTemplateBytes)
	registerTemplate, err := template.New("register").Parse(registerTemplateStr)
	if err != nil {
		return nil, err
	}
	dialer := gomail.NewDialer(config.Mail.Server, int(config.Mail.Port), config.Mail.Username, config.Mail.Password)
	return &UserService{
		db:               database,
		log:              log,
		dialer:           dialer,
		registerTemplate: registerTemplate,
		from:             config.Mail.From,
		subject:          config.Mail.Subject,
		salt:             config.User.Salt,
		baseUrl:          config.Rest.BaseUrl,
	}, nil
}

func (s *UserService) SendConfirmEmail(user, email, link string) error {
	var bodyBuffer bytes.Buffer
	err := s.registerTemplate.Execute(&bodyBuffer, registerTemplateVars{
		User: user,
		Link: link,
	})
	if err != nil {
		return err
	}
	body := bodyBuffer.String()

	msg := gomail.NewMessage()
	msg.SetHeader("From", s.from)
	msg.SetHeader("To", email)
	msg.SetHeader("Subject", s.subject)
	msg.SetBody("text/html", body)

	err = s.dialer.DialAndSend(msg)
	if err != nil {
		return err
	}
	return nil
}

func (s *UserService) confirmWorker(confirmId string, user db.User, channel chan struct{}) {
	defer s.registerMap.Delete(confirmId)
	timer := time.NewTimer(1 * time.Hour)
	select {
	case <-timer.C:
		return
	case <-channel:
		if err := s.db.Create(&user).Error; err != nil {
			s.log.Error("Failed to create user", zap.Error(err))
		}
	}
}

func (s *UserService) newConfirmWorker(name, hash, email string) (string, error) {
	confirmId, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	confirmIdStr := confirmId.String()
	channel := make(chan struct{}, 1)
	user := db.NewUser(name, hash, email)
	go s.confirmWorker(confirmIdStr, user, channel)
	s.registerMap.Store(confirmIdStr, channel)
	return confirmIdStr, nil
}

func (s *UserService) RequestRegistration(name, pass, email string) (string, error) {
	hash, err := util.HashPassword(pass, s.salt)
	if err != nil {
		return "", err
	}
	confirmId, err := s.newConfirmWorker(name, hash, email)
	if err != nil {
		return "", err
	}
	return confirmId, nil
}

func (s *UserService) ConfirmRegistration(confirmId string) error {
	entry, exists := s.registerMap.LoadAndDelete(confirmId)
	if !exists {
		return util.ErrLinkExpired
	}
	channel := entry.(chan struct{})
	// this will trigger confirmWorker
	close(channel)
	return nil
}

func (s *UserService) GetUserById(id uint) (*db.User, error) {
	var user db.User
	queryResult := s.db.First(&user, "id = ?", id)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return nil, util.ErrNotFound
	}
	return &user, nil
}

func (s *UserService) GetUserByLogin(login string) (*db.User, error) {
	var user db.User
	queryResult := s.db.First(&user, "login = ?", login)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return nil, util.ErrNotFound
	}
	return &user, nil
}

func (s *UserService) GetUserByEmail(email string) (*db.User, error) {
	var user db.User
	queryResult := s.db.First(&user, "email = ?", email)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}
	if queryResult.RowsAffected == 0 {
		return nil, util.ErrNotFound
	}
	return &user, nil
}

func (s *UserService) CheckUserExists(login string, email string) error {
	_, err := s.GetUserByLogin(login)
	if err != nil {
		return util.ErrUserWithThisLoginAlreadyExists
	}
	_, err = s.GetUserByEmail(email)
	if err != nil {
		return util.ErrUserWithThisEmailAlreadyExists
	}
	return nil
}

func createAdminUser(database *gorm.DB, config *config.Config, service *UserService) {
	_, err := service.GetUserByLogin(config.Admin.Username)
	if err == nil {
		return
	}
	username := config.Admin.Username
	hash, _ := util.HashPassword(config.Admin.Password, config.User.Salt)
	user := db.NewUser(username, hash, "")
	user.Admin = true
	database.Create(&user)
}

var UserServiceExport = fx.Options(fx.Provide(NewUserService), fx.Invoke(createAdminUser))
