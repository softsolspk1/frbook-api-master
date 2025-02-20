package models

import (
	"errors"
	// -- imports --
	// -- end --
)

type ReqStatus int

const (
	ReqStatusNone ReqStatus = iota

	ReqStatusPending

	ReqStatusTakeAction
)

func (r ReqStatus) String() string {
	return [...]string{"ReqStatusNone", "ReqStatusPending", "ReqStatusTakeAction"}[r]
}

func ReqStatusValues() []ReqStatus {
	return []ReqStatus{ReqStatusNone, ReqStatusPending, ReqStatusTakeAction}
}

func ReqStatusFromString(s string) (ReqStatus, error) {
	switch s {

	case "ReqStatusNone":
		return ReqStatusNone, nil

	case "ReqStatusPending":
		return ReqStatusPending, nil

	case "ReqStatusTakeAction":
		return ReqStatusTakeAction, nil

	}

	return ReqStatusNone, errors.New("Can't parse enum")
}

func ReqStatusFromInt(i int) (ReqStatus, error) {
	switch ReqStatus(i) {

	case 0:
		return ReqStatusNone, nil

	case 1:
		return ReqStatusPending, nil

	case 2:
		return ReqStatusTakeAction, nil

	}

	return ReqStatusNone, errors.New("Can't parse enum")
}

// -- code --
// -- end --
