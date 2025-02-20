package models

import (
	"encoding/json"
	"io/ioutil"
	"time"
	// -- imports --
	// -- end --
)

type Post struct {
	CommentsCount int       `json:"comments_count,omitempty" bson:"comments_count,omitempty"`
	Content       string    `json:"content,omitempty" bson:"content,omitempty"`
	CreatedAt     time.Time `json:"created_at" bson:"created_at"`
	Id            int       `json:"id" bson:"_id"`
	Image         string    `json:"image,omitempty" bson:"image,omitempty"`
	Liked         bool      `json:"liked,omitempty" bson:"liked,omitempty"`
	Likes         []int     `json:"likes,omitempty" bson:"likes,omitempty"`
	LikesCount    int       `json:"likes_count,omitempty" bson:"likes_count,omitempty"`
	Name          string    `json:"name,omitempty" bson:"name,omitempty"`
	ProfilePic    string    `json:"profile_pic,omitempty" bson:"profile_pic,omitempty"`
	Title         string    `json:"title,omitempty" bson:"title,omitempty"`
	UserId        int       `json:"user_id" bson:"user_id"`
	Video         string    `json:"video,omitempty" bson:"video,omitempty"`

	// -- extensions --
	// -- end --
}

func (t *Post) Valid() bool {
	// -- validation --
	// -- end --
	return true
}

func (v *Validator) PostFromBody() *Post {
	b, err := ioutil.ReadAll(v.r.Body)
	if err != nil {
		v.Error("body", err.Error())
		return nil
	}

	ret := &Post{}
	err = json.Unmarshal(b, ret)
	if err != nil {
		v.Error("body", err.Error())
		return nil
	}

	if !ret.Valid() {
		v.Error("body", "Invalid Post")
		return nil
	}

	return ret
}

// -- code --
// -- end --
