package models

import (
	"encoding/json"
	"io/ioutil"
	// -- imports --
	// -- end --
)

type StringResponse struct {
	Code   int    `json:"code" bson:"code"`
	Error  string `json:"error,omitempty" bson:"error,omitempty"`
	Result string `json:"result,omitempty" bson:"result,omitempty"`

	// -- extensions --
	// -- end --
}

func (t *StringResponse) Valid() bool {
	// -- validation --
	// -- end --
	return true
}

func (v *Validator) StringResponseFromBody() *StringResponse {
	b, err := ioutil.ReadAll(v.r.Body)
	if err != nil {
		v.Error("body", err.Error())
		return nil
	}

	ret := &StringResponse{}
	err = json.Unmarshal(b, ret)
	if err != nil {
		v.Error("body", err.Error())
		return nil
	}

	if !ret.Valid() {
		v.Error("body", "Invalid StringResponse")
		return nil
	}

	return ret
}

// -- code --
// -- end --
