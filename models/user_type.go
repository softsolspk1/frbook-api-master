package models

import (
	"errors"
	// -- imports --
	// -- end --
)

type UserType int

const (
	UserTypeUser UserType = iota

	UserTypeAdmin
)

func (u UserType) String() string {
	return [...]string{"UserTypeUser", "UserTypeAdmin"}[u]
}

func UserTypeValues() []UserType {
	return []UserType{UserTypeUser, UserTypeAdmin}
}

func UserTypeFromString(s string) (UserType, error) {
	switch s {

	case "UserTypeUser":
		return UserTypeUser, nil

	case "UserTypeAdmin":
		return UserTypeAdmin, nil

	}

	return UserTypeUser, errors.New("Can't parse enum")
}

func UserTypeFromInt(i int) (UserType, error) {
	switch UserType(i) {

	case 0:
		return UserTypeUser, nil

	case 1:
		return UserTypeAdmin, nil

	}

	return UserTypeUser, errors.New("Can't parse enum")
}

// -- code --
// -- end --
