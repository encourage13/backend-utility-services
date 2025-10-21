package repository

import (
	"lab1/internal/app/ds"

	"golang.org/x/crypto/bcrypt"
)

func (r *Repository) CreateUser(user *ds.User) error {
	return r.db.Create(user).Error
}

func (r *Repository) GetUserByID(id uint) (ds.User, error) {
	var user ds.User
	if err := r.db.First(&user, id).Error; err != nil {
		return ds.User{}, err
	}
	return user, nil
}

func (r *Repository) UpdateUser(id uint, req ds.UserUpdateRequest) error {
	updates := make(map[string]interface{})

	if req.Login != nil {
		updates["login"] = *req.Login
	}
	if req.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		updates["hashed_password"] = string(hashedPassword)
	}
	if req.IsModerator != nil {
		updates["is_moderator"] = *req.IsModerator
	}

	if len(updates) == 0 {
		return nil
	}

	return r.db.Model(&ds.User{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) GetUserByUsername(login string) (ds.User, error) {
	var user ds.User
	if err := r.db.Where("login = ?", login).First(&user).Error; err != nil {
		return ds.User{}, err
	}
	return user, nil
}
