package repositories

import (
	"context"
	"errors"
	errWrap "user-service/common/error"
	errConstant "user-service/constants/error"
	"user-service/domain/dto"
	"user-service/domain/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

type IUserRepository interface {
	Register(context.Context, *dto.RegisterRequest) (*models.User, error)
	Update(context.Context, *dto.UpdateRequest, string) (*models.User, error)
	FindByUsername(context.Context, string) (*models.User, error)
	FindByEmail(context.Context, string) (*models.User, error)
	FindByUUID(context.Context, string) (*models.User, error)
}

func NewUserRepository(db *gorm.DB) IUserRepository {
	return &UserRepository{db: db}
}

func (u UserRepository) Register(ctx context.Context, request *dto.RegisterRequest) (*models.User, error) {
	user := models.User{
		UUID:        uuid.New(),
		Name:        request.Name,
		Username:    request.Username,
		Password:    request.Password,
		PhoneNumber: request.PhoneNumber,
		Email:       request.Email,
		RoleID:      request.RoleID,
	}

	err := u.db.WithContext(ctx).Create(&user).Error
	if err != nil {
		return nil, errWrap.WrapError(errConstant.ErrSQLError)
	}

	return &user, nil
}

func (u UserRepository) Update(ctx context.Context, request *dto.UpdateRequest, uuid string) (*models.User, error) {
	user := models.User{
		Name:        request.Name,
		Username:    request.Username,
		Password:    *request.Password,
		PhoneNumber: request.PhoneNumber,
		Email:       request.Email,
	}

	err := u.db.WithContext(ctx).
		Where("uuid = ?", uuid).
		Updates(&user).Error
	if err != nil {
		return nil, errWrap.WrapError(errConstant.ErrSQLError)
	}
	return &user, nil
}

func (u UserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := u.db.WithContext(ctx).
		Preload("Role").
		Where("username = ?", username).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errConstant.ErrUserNotFound
		}
		return nil, errWrap.WrapError(errConstant.ErrSQLError)
	}
	return &user, nil
}

func (u UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := u.db.WithContext(ctx).
		Preload("Role").
		Where("email = ?", email).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errConstant.ErrUserNotFound
		}
		return nil, errWrap.WrapError(errConstant.ErrSQLError)
	}
	return &user, nil
}

func (u UserRepository) FindByUUID(ctx context.Context, uuid string) (*models.User, error) {
	var user models.User
	err := u.db.WithContext(ctx).
		Preload("Role").
		Where("uuid = ?", uuid).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errConstant.ErrUserNotFound
		}
		return nil, errWrap.WrapError(errConstant.ErrSQLError)
	}
	return &user, nil
}
