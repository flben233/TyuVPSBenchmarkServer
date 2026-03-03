package util

import "VPSBenchmarkBackend/internal/auth/store"

func IsAdmin(id int64) bool {
	user, err := store.GetUserByID(id)
	if err != nil {
		return false
	}
	group, err := store.GetUserGroupByID(user.GroupID)
	if err != nil {
		return false
	}
	return group.IsAdmin
}

func CheckHostQuota(id int64, current int64) bool {
	user, err := store.GetUserByID(id)
	if err != nil {
		return false
	}
	group, err := store.GetUserGroupByID(user.GroupID)
	if err != nil {
		return false
	}
	return current < int64(group.MaxHostNum)
}

func CheckInspectorQuota(id int64, current int64) bool {
	user, err := store.GetUserByID(id)
	if err != nil {
		return false
	}
	group, err := store.GetUserGroupByID(user.GroupID)
	if err != nil {
		return false
	}
	return current < int64(group.InspectorNum)
}
