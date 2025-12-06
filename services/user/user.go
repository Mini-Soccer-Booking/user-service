package services

import (
	"context"
	"errors"
	"strings"
	"time"
	"user-service/config"
	"user-service/constants"
	errConstant "user-service/constants/error"
	"user-service/domain/dto"
	"user-service/domain/models"
	"user-service/repositories"

	"github.com/golang-jwt/jwt/v5"
	googleUUID "github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repository repositories.IRepositoryRegistry
}

type IUserService interface {
	Login(context.Context, *dto.LoginRequest) (*dto.LoginResponse, error)
	Register(context.Context, *dto.RegisterRequest) (*dto.RegisterResponse, error)
	Update(context.Context, *dto.UpdateRequest, string) (*dto.UserResponse, error)
	GetUserLogin(context.Context) (*dto.UserResponse, error)
	GetUserByUUID(context.Context, string) (*dto.UserResponse, error)
}

func NewUserService(repository repositories.IRepositoryRegistry) IUserService {
	return &UserService{repository: repository}
}

type Claims struct {
	User *dto.UserResponse
	jwt.RegisteredClaims
}

func (u *UserService) Login(ctx context.Context, request *dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := u.repository.GetUser().FindByUsername(ctx, request.Username)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password))
	if err != nil {
		return nil, err
	}

	expirationTime := time.Now().Add(time.Duration(config.Config.JwtExpirationTime) * time.Minute).Unix()
	data := &dto.UserResponse{
		UUID:        user.UUID,
		Name:        user.Name,
		Username:    user.Username,
		PhoneNumber: user.PhoneNumber,
		Email:       user.Email,
		Role:        strings.ToLower(user.Role.Code),
	}

	claims := &Claims{
		User: data,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Unix(expirationTime, 0)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.Config.JwtSecretKey))
	if err != nil {
		return nil, err
	}

	response := &dto.LoginResponse{
		User:  *data,
		Token: tokenString,
	}

	return response, nil
}

func (u *UserService) Register(ctx context.Context, request *dto.RegisterRequest) (*dto.RegisterResponse, error) {
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	checkUsername, err := u.repository.GetUser().FindByUsername(ctx, request.Username)
	if err == nil && checkUsername != nil {
		return nil, errConstant.ErrUsernameAlreadyExists
	}

	if err != nil && !errors.Is(err, errConstant.ErrUserNotFound) {
		return nil, err
	}

	checkEmail, err := u.repository.GetUser().FindByEmail(ctx, request.Email)
	if err == nil && checkEmail != nil {
		return nil, errConstant.ErrEmailAlreadyExists
	}
	if err != nil && !errors.Is(err, errConstant.ErrUserNotFound) {
		return nil, err
	}

	if request.Password != request.ConfirmPassword {
		return nil, errConstant.ErrPasswordDoesNotMatch
	}

	user, err := u.repository.GetUser().Register(ctx, &dto.RegisterRequest{
		Name:        request.Name,
		Username:    request.Username,
		Password:    string(hashPassword),
		Email:       request.Email,
		PhoneNumber: request.PhoneNumber,
		RoleID:      constants.Customer,
	})

	if err != nil {
		return nil, err
	}

	response := &dto.RegisterResponse{
		User: dto.UserResponse{
			UUID:        user.UUID,
			Name:        user.Name,
			Username:    user.Username,
			PhoneNumber: user.PhoneNumber,
			Email:       user.Email,
		},
	}

	return response, nil

}

func (u *UserService) Update(ctx context.Context, request *dto.UpdateRequest, uuid string) (*dto.UserResponse, error) {
	var (
		password         string
		hashedPassword   []byte
		user, userResult *models.User
		err              error
		data             dto.UserResponse
	)

	user, err = u.repository.GetUser().FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	if request.Username != "" && request.Username != user.Username {
		existUser, err := u.repository.GetUser().FindByUsername(ctx, request.Username)
		if err == nil || existUser != nil {
			return nil, errConstant.ErrUsernameAlreadyExists
		}

		if err != nil && !errors.Is(err, errConstant.ErrUserNotFound) {
			return nil, err
		}
	}

	if request.Email != "" && request.Email != user.Email {
		existUserEmail, err := u.repository.GetUser().FindByEmail(ctx, request.Email)
		if err == nil || existUserEmail != nil {
			return nil, errConstant.ErrEmailAlreadyExists
		}

		if err != nil && !errors.Is(err, errConstant.ErrUserNotFound) {
			return nil, err
		}
	}

	if request.Password != nil {
		if request.ConfirmPassword == nil || *request.Password != *request.ConfirmPassword {
			return nil, errConstant.ErrPasswordDoesNotMatch
		}
		hashedPassword, err = bcrypt.GenerateFromPassword([]byte(*request.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		password = string(hashedPassword)
	}

	userResult, err = u.repository.GetUser().Update(ctx, &dto.UpdateRequest{
		Name:        request.Name,
		Username:    request.Username,
		Password:    &password,
		Email:       request.Email,
		PhoneNumber: request.PhoneNumber,
	}, uuid)

	if err != nil {
		return nil, err
	}

	parsedUUID, err := googleUUID.Parse(uuid)
	if err != nil {
		return nil, err // Return error jika string UUID tidak valid
	}

	data = dto.UserResponse{
		UUID:        parsedUUID,
		Name:        userResult.Name,
		Username:    userResult.Username,
		Email:       userResult.Email,
		PhoneNumber: userResult.PhoneNumber,
	}
	return &data, nil
}

func (u *UserService) GetUserLogin(ctx context.Context) (*dto.UserResponse, error) {
	var (
		userLogin = ctx.Value(constants.UserLogin).(*dto.UserResponse)
		data      dto.UserResponse
	)

	data = dto.UserResponse{
		UUID:        userLogin.UUID,
		Name:        userLogin.Name,
		Username:    userLogin.Username,
		PhoneNumber: userLogin.PhoneNumber,
		Email:       userLogin.Email,
		Role:        userLogin.Role,
	}

	return &data, nil
}

func (u *UserService) GetUserByUUID(ctx context.Context, uuid string) (*dto.UserResponse, error) {
	user, err := u.repository.GetUser().FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	data := dto.UserResponse{
		UUID:        user.UUID,
		Name:        user.Name,
		Username:    user.Username,
		PhoneNumber: user.PhoneNumber,
		Email:       user.Email,
	}

	return &data, nil
}
