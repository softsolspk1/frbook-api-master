package models

import (
	"encoding/json"
	"io/ioutil"
	// -- imports --
	// -- end --
)

type CommentListResponse struct {
	Code   int        `json:"code" bson:"code"`
	Error  string     `json:"error,omitempty" bson:"error,omitempty"`
	Result []*Comment `json:"result,omitempty" bson:"result,omitempty"`
	Start  int        `json:"start" bson:"start"`
	Total  int        `json:"total" bson:"total"`

	// -- extensions --
	// -- end --
}

func (t *CommentListResponse) Valid() bool {
	// -- validation --
	// -- end --
	return true
}

func (v *Validator) CommentListResponseFromBody() *CommentListResponse {
	b, err := ioutil.ReadAll(v.r.Body)
	if err != nil {
		v.Error("body", err.Error())
		return nil
	}

	ret := &CommentListResponse{}
	err = json.Unmarshal(b, ret)
	if err != nil {
		v.Error("body", err.Error())
		return nil
	}

	if !ret.Valid() {
		v.Error("body", "Invalid CommentListResponse")
		return nil
	}

	return ret
}

// -- code --
// -- end --
