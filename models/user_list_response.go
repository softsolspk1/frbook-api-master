package models

import (
	"encoding/json"
	"io/ioutil"
	// -- imports --
	// -- end --
)

type UserListResponse struct {
	Code   int     `json:"code" bson:"code"`
	Error  string  `json:"error,omitempty" bson:"error,omitempty"`
	Result []*User `json:"result,omitempty" bson:"result,omitempty"`
	Start  int     `json:"start" bson:"start"`
	Total  int     `json:"total" bson:"total"`

	// -- extensions --
	// -- end --
}

func (t *UserListResponse) Valid() bool {
	// -- validation --
	// -- end --
	return true
}

func (v *Validator) UserListResponseFromBody() *UserListResponse {
	b, err := ioutil.ReadAll(v.r.Body)
	if err != nil {
		v.Error("body", err.Error())
		return nil
	}

	ret := &UserListResponse{}
	err = json.Unmarshal(b, ret)
	if err != nil {
		v.Error("body", err.Error())
		return nil
	}

	if !ret.Valid() {
		v.Error("body", "Invalid UserListResponse")
		return nil
	}

	return ret
}

// -- code --
// -- end --
