package models

import (
	"encoding/json"
	"io/ioutil"
	// -- imports --
	// -- end --
)

type FriendRequestListResponse struct {
	Code   int              `json:"code" bson:"code"`
	Error  string           `json:"error,omitempty" bson:"error,omitempty"`
	Result []*FriendRequest `json:"result,omitempty" bson:"result,omitempty"`
	Start  int              `json:"start" bson:"start"`
	Total  int              `json:"total" bson:"total"`

	// -- extensions --
	// -- end --
}

func (t *FriendRequestListResponse) Valid() bool {
	// -- validation --
	// -- end --
	return true
}

func (v *Validator) FriendRequestListResponseFromBody() *FriendRequestListResponse {
	b, err := ioutil.ReadAll(v.r.Body)
	if err != nil {
		v.Error("body", err.Error())
		return nil
	}

	ret := &FriendRequestListResponse{}
	err = json.Unmarshal(b, ret)
	if err != nil {
		v.Error("body", err.Error())
		return nil
	}

	if !ret.Valid() {
		v.Error("body", "Invalid FriendRequestListResponse")
		return nil
	}

	return ret
}

// -- code --
// -- end --
