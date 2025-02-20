package models

import (
	"encoding/json"
	"io/ioutil"
	// -- imports --
	// -- end --
)

type IntResponse struct {
	Code   int    `json:"code" bson:"code"`
	Error  string `json:"error,omitempty" bson:"error,omitempty"`
	Result int    `json:"result,omitempty" bson:"result,omitempty"`

	// -- extensions --
	// -- end --
}

func (t *IntResponse) Valid() bool {
	// -- validation --
	// -- end --
	return true
}

func (v *Validator) IntResponseFromBody() *IntResponse {
	b, err := ioutil.ReadAll(v.r.Body)
	if err != nil {
		v.Error("body", err.Error())
		return nil
	}

	ret := &IntResponse{}
	err = json.Unmarshal(b, ret)
	if err != nil {
		v.Error("body", err.Error())
		return nil
	}

	if !ret.Valid() {
		v.Error("body", "Invalid IntResponse")
		return nil
	}

	return ret
}

// -- code --
// -- end --
