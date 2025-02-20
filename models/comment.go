package models

import (
	"encoding/json"
	"io/ioutil"
	"time"
	// -- imports --
	// -- end --
)

type Comment struct {
	Content    string    `json:"content" bson:"content"`
	CreatedAt  time.Time `json:"created_at" bson:"created_at"`
	Id         int       `json:"id" bson:"_id"`
	Name       string    `json:"name,omitempty" bson:"name,omitempty"`
	PostId     int       `json:"post_id" bson:"post_id"`
	ProfilePic string    `json:"profile_pic,omitempty" bson:"profile_pic,omitempty"`
	UserId     int       `json:"user_id" bson:"user_id"`

	// -- extensions --
	// -- end --
}

func (t *Comment) Valid() bool {
	// -- validation --
	// -- end --
	return true
}

func (v *Validator) CommentFromBody() *Comment {
	b, err := ioutil.ReadAll(v.r.Body)
	if err != nil {
		v.Error("body", err.Error())
		return nil
	}

	ret := &Comment{}
	err = json.Unmarshal(b, ret)
	if err != nil {
		v.Error("body", err.Error())
		return nil
	}

	if !ret.Valid() {
		v.Error("body", "Invalid Comment")
		return nil
	}

	return ret
}

// -- code --
// -- end --
