package models

import (
	"encoding/json"
	"io/ioutil"
	"time"
	// -- imports --
	// -- end --
)

type Article struct {
	AuthorName  string     `json:"author_name,omitempty" bson:"author_name,omitempty"`
	Content     string     `json:"content,omitempty" bson:"content,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty" bson:"created_at,omitempty"`
	Description string     `json:"description,omitempty" bson:"description,omitempty"`
	Id          int        `json:"id,omitempty" bson:"_id,omitempty"`
	Pdf         string     `json:"pdf,omitempty" bson:"pdf,omitempty"`
	Photo       string     `json:"photo,omitempty" bson:"photo,omitempty"`
	ProfilePic  string     `json:"profile_pic,omitempty" bson:"profile_pic,omitempty"`
	Tags        []string   `json:"tags,omitempty" bson:"tags,omitempty"`
	Title       string     `json:"title,omitempty" bson:"title,omitempty"`
	UserId      int        `json:"user_id,omitempty" bson:"user_id,omitempty"`

	// -- extensions --
	// -- end --
}

func (t *Article) Valid() bool {
	// -- validation --
	// -- end --
	return true
}

func (v *Validator) ArticleFromBody() *Article {
	b, err := ioutil.ReadAll(v.r.Body)
	if err != nil {
		v.Error("body", err.Error())
		return nil
	}

	ret := &Article{}
	err = json.Unmarshal(b, ret)
	if err != nil {
		v.Error("body", err.Error())
		return nil
	}

	if !ret.Valid() {
		v.Error("body", "Invalid Article")
		return nil
	}

	return ret
}

// -- code --
// -- end --
