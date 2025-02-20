package models

import (
	"encoding/json"
	"io/ioutil"
	// -- imports --
	// -- end --
)

type ArticleResponse struct {
	Code   int      `json:"code" bson:"code"`
	Error  string   `json:"error,omitempty" bson:"error,omitempty"`
	Result *Article `json:"result,omitempty" bson:"result,omitempty"`

	// -- extensions --
	// -- end --
}

func (t *ArticleResponse) Valid() bool {
	// -- validation --
	// -- end --
	return true
}

func (v *Validator) ArticleResponseFromBody() *ArticleResponse {
	b, err := ioutil.ReadAll(v.r.Body)
	if err != nil {
		v.Error("body", err.Error())
		return nil
	}

	ret := &ArticleResponse{}
	err = json.Unmarshal(b, ret)
	if err != nil {
		v.Error("body", err.Error())
		return nil
	}

	if !ret.Valid() {
		v.Error("body", "Invalid ArticleResponse")
		return nil
	}

	return ret
}

// -- code --
// -- end --
