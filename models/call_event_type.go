package models

import (
	"errors"
	// -- imports --
	// -- end --
)

type CallEventType int

const (
	CallEventTypeIncoming CallEventType = iota

	CallEventTypeStartCall

	CallEventTypeAcceptCall

	CallEventTypeEndCall

	CallEventTypeInit
)

func (c CallEventType) String() string {
	return [...]string{"CallEventTypeIncoming", "CallEventTypeStartCall", "CallEventTypeAcceptCall", "CallEventTypeEndCall", "CallEventTypeInit"}[c]
}

func CallEventTypeValues() []CallEventType {
	return []CallEventType{CallEventTypeIncoming, CallEventTypeStartCall, CallEventTypeAcceptCall, CallEventTypeEndCall, CallEventTypeInit}
}

func CallEventTypeFromString(s string) (CallEventType, error) {
	switch s {

	case "CallEventTypeIncoming":
		return CallEventTypeIncoming, nil

	case "CallEventTypeStartCall":
		return CallEventTypeStartCall, nil

	case "CallEventTypeAcceptCall":
		return CallEventTypeAcceptCall, nil

	case "CallEventTypeEndCall":
		return CallEventTypeEndCall, nil

	case "CallEventTypeInit":
		return CallEventTypeInit, nil

	}

	return CallEventTypeIncoming, errors.New("Can't parse enum")
}

func CallEventTypeFromInt(i int) (CallEventType, error) {
	switch CallEventType(i) {

	case 0:
		return CallEventTypeIncoming, nil

	case 1:
		return CallEventTypeStartCall, nil

	case 2:
		return CallEventTypeAcceptCall, nil

	case 3:
		return CallEventTypeEndCall, nil

	case 4:
		return CallEventTypeInit, nil

	}

	return CallEventTypeIncoming, errors.New("Can't parse enum")
}

// -- code --
// -- end --
