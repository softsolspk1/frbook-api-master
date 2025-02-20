package models

import (
	"encoding/json"
	"io/ioutil"
	"time"
	// -- imports --
	// -- end --
)

type FriendEntry struct {
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	FromId    int       `json:"from_id" bson:"from_id"`
	Id        int       `json:"id" bson:"_id"`
	ToId      int       `json:"to_id" bson:"to_id"`

	// -- extensions --
	// -- end --
}

func (t *FriendEntry) Valid() bool {
	// -- validation --
	// -- end --
	return true
}

func (v *Validator) FriendEntryFromBody() *FriendEntry {
	b, err := ioutil.ReadAll(v.r.Body)
	if err != nil {
		v.Error("body", err.Error())
		return nil
	}

	ret := &FriendEntry{}
	err = json.Unmarshal(b, ret)
	if err != nil {
		v.Error("body", err.Error())
		return nil
	}

	if !ret.Valid() {
		v.Error("body", "Invalid FriendEntry")
		return nil
	}

	return ret
}

// -- code --
// -- end --
