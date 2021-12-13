package api

import (
	"errors"
	"fmt"
	"pdserver/pkg/api/model"

	"github.com/jinzhu/gorm"
)

type UserDB struct {
	Storage *gorm.DB
}

type UserAPIService interface {
	Post(model.User) error
	GetWithID(uint64) (*model.User, error)
	Get(*model.User) (model.User, error)
	Delete(uint64) error
	Available(string, string) (bool, error)
}

// Create user database
func NewUserDB(db *gorm.DB) *UserDB {
	return &UserDB{Storage: db}
}

// PostUser - Post user to database
func (db *UserDB) Post(user *model.User) error {
	if res := db.Storage.Save(&user); res.Error != nil {
		fmt.Println(res.Error.Error())
		return res.Error
	}
	return nil
}

func (db *UserDB) GetWithID(userID uint64) (*model.User, error) {
	var user model.User
	fmt.Println(userID)
	if res := db.Storage.First(&user, userID); res.Error != nil {
		return nil, res.Error
	}
	return &user, nil
}

// GetUser - Get user from database
func (db *UserDB) Get(user *model.User) (*model.User, error) {
	if res := db.Storage.
		Where(&model.User{Email: user.Email, Password: user.Password, Name: user.Name, Nickname: user.Nickname, Birth: user.Birth}).
		First(&user); res.Error != nil {
		return nil, res.Error
	}
	return user, nil
}

// DeleteUser - Delete user from database
func (db *UserDB) Delete(id uint64) error {
	var user model.User
	if res := db.Storage.Delete(&user, id); res.Error != nil {
		return res.Error
	}
	return nil
}

// GetAvailable - Configure val which type is key can be used
func (db *UserDB) Available(val string, key string) (bool, error) {
	query := fmt.Sprintf("%s = ?", key)
	res := db.Storage.Where(query, val).Find(&model.User{})
	if res.Error == nil {
		return false, nil
	}
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return true, nil
	}
	return false, res.Error
}
