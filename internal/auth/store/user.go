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

const (
	DefaultUserGroupId  uint32 = 0
	DefaultAdminGroupId uint32 = 1
)

func init() {
	// Register the initializer
	common.RegisterDBInitializer(InitUserStore)
}

// Create a default user group if it doesn't exist
func createDefaultUserGroup() error {
	count, err := userGroups.Where("id = ?", DefaultUserGroupId).Count(ctx, "*")
	if err != nil {
		return err
	} else if count > 0 {
		return nil
	}
	defaultGroup := model.UserGroup{
		ID:           DefaultUserGroupId,
		Name:         "Default",
		MaxHostNum:   5,
		InspectorNum: 10,
		IsAdmin:      false,
	}
	return userGroups.Create(ctx, &defaultGroup)
}

func createDefaultAdminGroup() error {
	count, err := userGroups.Where("id = ?", DefaultAdminGroupId).Count(ctx, "*")
	if err != nil {
		return err
	} else if count > 0 {
		return nil
	}
	defaultAdminGroup := model.UserGroup{
		ID:           DefaultAdminGroupId,
		Name:         "Admin",
		MaxHostNum:   65535,
		InspectorNum: 65535,
		IsAdmin:      true,
	}
	return userGroups.Create(ctx, &defaultAdminGroup)
}

func InitUserStore(dbPath string) error {
	db = common.GetDB()
	users = gorm.G[model.User](db)
	userGroups = gorm.G[model.UserGroup](db)
	if err := db.AutoMigrate(&model.User{}, &model.UserGroup{}); err != nil {
		return err
	}
	err := createDefaultUserGroup()
	if err != nil {
		return err
	}
	err = createDefaultAdminGroup()
	return err
}

// User CRUD

func CreateUser(user *model.User) error {
	return users.Create(ctx, user)
}

func UserExists(id int64) (bool, error) {
	count, err := users.Where("id = ?", id).Count(ctx, "*")
	return count > 0, err
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
