package models

import (
	"encoding/json"
	"io/ioutil"
	// -- imports --
	// -- end --
)

type User struct {
	Email      string    `json:"email" bson:"email"`
	Id         int       `json:"id" bson:"_id"`
	Name       string    `json:"name" bson:"name"`
	Password   string    `json:"password" bson:"password"`
	Phone      string    `json:"phone,omitempty" bson:"phone,omitempty"`
	ProfilePic string    `json:"profile_pic,omitempty" bson:"profile_pic,omitempty"`
	ReqId      int       `json:"req_id,omitempty" bson:"req_id,omitempty"`
	Status     ReqStatus `json:"status,omitempty" bson:"status,omitempty"`
	Verified   bool      `json:"verified" bson:"verified"`

	// -- extensions --
	// -- end --
}

func (t *User) Valid() bool {
	// -- validation --
	// -- end --
	return true
}

func (v *Validator) UserFromBody() *User {
	b, err := ioutil.ReadAll(v.r.Body)
	if err != nil {
		v.Error("body", err.Error())
		return nil
	}

	ret := &User{}
	err = json.Unmarshal(b, ret)
	if err != nil {
		v.Error("body", err.Error())
		return nil
	}

	if !ret.Valid() {
		v.Error("body", "Invalid User")
		return nil
	}

	return ret
}

// -- code --
func (u *User) Serialize() ([]byte, error) {
	ret, _ := json.Marshal(u)
	return ret, nil
}

func (u *User) Parse(b []byte) error {
	return json.Unmarshal(b, u)
}

// -- end --
