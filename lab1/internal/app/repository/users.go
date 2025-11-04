package repository

import (
	"lab1/internal/app/ds"
)

func (r *Repository) Register(user *ds.User) error {
	return r.db.Create(user).Error
}

func (r *Repository) GetUserByLogin(login string) (ds.User, error) {
	var user ds.User
	err := r.db.Where("name = ?", login).First(&user).Error
	return user, err
}

func (r *Repository) GetUserByID(id uint) (*ds.User, error) {
	user := &ds.User{}
	err := r.db.First(user, id).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) UpdateUser(id uint, req ds.UserUpdateRequest) error {
	updates := make(map[string]interface{})

	if req.Login != nil {
		updates["name"] = *req.Login
	}
	if req.Password != nil {
		updates["pass"] = *req.Password
	}

	return r.db.Model(&ds.User{}).Where("id = ?", id).Updates(updates).Error
}
