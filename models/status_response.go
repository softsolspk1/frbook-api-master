package models

import (
	"encoding/json"
	"io/ioutil"
	// -- imports --
	// -- end --
)

type StatusResponse struct {
	Code  int    `json:"code" bson:"code"`
	Error string `json:"error,omitempty" bson:"error,omitempty"`

	// -- extensions --
	// -- end --
}

func (t *StatusResponse) Valid() bool {
	// -- validation --
	// -- end --
	return true
}

func (v *Validator) StatusResponseFromBody() *StatusResponse {
	b, err := ioutil.ReadAll(v.r.Body)
	if err != nil {
		v.Error("body", err.Error())
		return nil
	}

	ret := &StatusResponse{}
	err = json.Unmarshal(b, ret)
	if err != nil {
		v.Error("body", err.Error())
		return nil
	}

	if !ret.Valid() {
		v.Error("body", "Invalid StatusResponse")
		return nil
	}

	return ret
}

// -- code --
// -- end --
