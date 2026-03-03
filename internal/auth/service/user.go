package service

import (
	"VPSBenchmarkBackend/internal/auth/model"
	"VPSBenchmarkBackend/internal/auth/store"
)

func GetUser(id int64) (*model.User, error) {
	user, err := store.GetUserByID(id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func ListUsers() ([]model.User, error) {
	users, err := store.ListUsers()
	if err != nil {
		return nil, err
	}
	return users, nil
}

func UpdateUser(user *model.User) error {
	_, err := store.UpdateUser(*user)
	return err
}

func DeleteUser(id int64) error {
	_, err := store.DeleteUser(id)
	return err
}

func CreateUserGroup(group model.UserGroup) error {
	return store.CreateUserGroup(&group)
}

func ListUserGroups() ([]model.UserGroup, error) {
	groups, err := store.ListUserGroups()
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func UpdateUserGroup(group *model.UserGroup) error {
	err := store.UpdateUserGroup(*group)
	return err
}

func DeleteUserGroup(id uint32) error {
	_, err := store.DeleteUserGroup(id)
	return err
}
