package service

import (
	"anileha/config"
	"anileha/db"
	"anileha/db/repo"
	"anileha/rest"
	"anileha/util"
	"bytes"
	"fmt"
	"github.com/gofrs/uuid"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
	"os"
	"sync"
	"text/template"
	"time"
)

type registerTemplateVars struct {
	User string
	Link string
}

type UserService struct {
	userRepo         *repo.UserRepo
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
	userRepo *repo.UserRepo,
	log *zap.Logger,
	config *config.Config,
) (*UserService, error) {
	registerTemplateBytes, err := os.ReadFile(config.Mail.RegisterTemplatePath)
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
		userRepo:         userRepo,
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
	defer timer.Stop()
	select {
	case <-timer.C:
		return
	case <-channel:
		if _, err := s.userRepo.Create(&user); err != nil {
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
	user := db.User{
		Login: name,
		Hash:  hash,
		Email: email,
	}
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
		return rest.ErrLinkExpired
	}
	channel := entry.(chan struct{})
	// this will trigger confirmWorker
	close(channel)
	return nil
}

func (s *UserService) GetById(id uint) (*db.User, error) {
	user, err := s.userRepo.GetById(id)
	if err != nil {
		return nil, rest.ErrInternal(err.Error())
	}
	if user == nil {
		return nil, rest.ErrNotFoundInst
	}
	return user, nil
}

func (s *UserService) GetByLogin(login string) (*db.User, error) {
	user, err := s.userRepo.GetByLogin(login)
	if err != nil {
		return nil, rest.ErrInternal(err.Error())
	}
	if user == nil {
		return nil, rest.ErrNotFoundInst
	}
	return user, nil
}

func (s *UserService) CheckExists(login string, email string) error {
	user, err := s.userRepo.GetByLogin(login)
	if err != nil {
		return rest.ErrInternal(err.Error())
	}
	if user != nil {
		return rest.ErrUserWithThisLoginAlreadyExists
	}

	user, err = s.userRepo.GetByEmail(email)
	if err != nil {
		return rest.ErrInternal(err.Error())
	}
	if user != nil {
		return rest.ErrUserWithThisEmailAlreadyExists
	}

	return nil
}

func createAdminUser(userRepo *repo.UserRepo, config *config.Config) error {
	existingUser, err := userRepo.GetByLogin(config.Admin.Username)
	if err == nil && existingUser != nil {
		return nil
	}
	username := config.Admin.Username
	hash, _ := util.HashPassword(config.Admin.Password, config.User.Salt)
	user := db.User{
		Login: username,
		Hash:  hash,
		Admin: true,
	}
	_, err = userRepo.Create(&user)
	if err != nil {
		return fmt.Errorf("failed to automatically create admin user: %w", err)
	}
	return nil
}

var UserServiceExport = fx.Options(fx.Provide(NewUserService), fx.Invoke(createAdminUser))
