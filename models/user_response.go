package models

import (
	"encoding/json"
	"io/ioutil"
	// -- imports --
	// -- end --
)

type UserResponse struct {
	Code   int    `json:"code" bson:"code"`
	Error  string `json:"error,omitempty" bson:"error,omitempty"`
	Result *User  `json:"result,omitempty" bson:"result,omitempty"`

	// -- extensions --
	// -- end --
}

func (t *UserResponse) Valid() bool {
	// -- validation --
	// -- end --
	return true
}

func (v *Validator) UserResponseFromBody() *UserResponse {
	b, err := ioutil.ReadAll(v.r.Body)
	if err != nil {
		v.Error("body", err.Error())
		return nil
	}

	ret := &UserResponse{}
	err = json.Unmarshal(b, ret)
	if err != nil {
		v.Error("body", err.Error())
		return nil
	}

	if !ret.Valid() {
		v.Error("body", "Invalid UserResponse")
		return nil
	}

	return ret
}

// -- code --
// -- end --
