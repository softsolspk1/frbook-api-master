package models

import (
	"encoding/json"
	"io/ioutil"
	// -- imports --
	// -- end --
)

type CallEvent struct {
	Channel string        `json:"channel,omitempty" bson:"channel,omitempty"`
	FromPic string        `json:"from_pic,omitempty" bson:"from_pic,omitempty"`
	Kind    CallEventType `json:"kind" bson:"kind"`
	ToId    int           `json:"to_id,omitempty" bson:"_id,omitempty"`
	ToPic   string        `json:"to_pic,omitempty" bson:"to_pic,omitempty"`

	// -- extensions --
	// -- end --
}

func (t *CallEvent) Valid() bool {
	// -- validation --
	// -- end --
	return true
}

func (v *Validator) CallEventFromBody() *CallEvent {
	b, err := ioutil.ReadAll(v.r.Body)
	if err != nil {
		v.Error("body", err.Error())
		return nil
	}

	ret := &CallEvent{}
	err = json.Unmarshal(b, ret)
	if err != nil {
		v.Error("body", err.Error())
		return nil
	}

	if !ret.Valid() {
		v.Error("body", "Invalid CallEvent")
		return nil
	}

	return ret
}

// -- code --
func (u *CallEvent) Serialize() ([]byte, error) {
	ret, _ := json.Marshal(u)
	return ret, nil
}

func (u *CallEvent) Parse(b []byte) error {
	return json.Unmarshal(b, u)
}

// -- end --
