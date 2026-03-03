package store

import (
	"context"

	"VPSBenchmarkBackend/internal/auth/model"
	"VPSBenchmarkBackend/internal/common"
	"gorm.io/gorm"
)

var ctx = context.Background()

var (
	db         *gorm.DB
	users      gorm.Interface[model.User]
	userGroups gorm.Interface[model.UserGroup]
)

func init() {
	// Register the initializer
	common.RegisterDBInitializer(InitUserStore)
}

func InitUserStore(dbPath string) error {
	db = common.GetDB()
	users = gorm.G[model.User](db)
	userGroups = gorm.G[model.UserGroup](db)
	if err := db.AutoMigrate(&model.User{}, &model.UserGroup{}); err != nil {
		return err
	}
	return nil
}

// User CRUD

func CreateUser(user *model.User) error {
	return users.Create(ctx, user)
}

func GetUserByID(id int64) (model.User, error) {
	return users.Where("id = ?", id).First(ctx)
}

func UpdateUser(user model.User) (int, error) {
	return users.Where("id = ?", user.ID).Updates(ctx, user)
}

func DeleteUser(id int64) (int, error) {
	return users.Where("id = ?", id).Delete(ctx)
}

func ListUsers() ([]model.User, error) {
	return users.Find(ctx)
}

// UserGroup CRUD

func CreateUserGroup(group *model.UserGroup) error {
	return userGroups.Create(ctx, group)
}

func GetUserGroupByID(id uint32) (model.UserGroup, error) {
	return userGroups.Where("id = ?", id).First(ctx)
}

func UpdateUserGroup(group model.UserGroup) (int, error) {
	return userGroups.Where("id = ?", group.ID).Updates(ctx, group)
}

func DeleteUserGroup(id uint32) (int, error) {
	return userGroups.Where("id = ?", id).Delete(ctx)
}

func ListUserGroups() ([]model.UserGroup, error) {
	return userGroups.Find(ctx)
}
