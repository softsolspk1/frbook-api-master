package models

import (
	"encoding/json"
	"io/ioutil"
	// -- imports --
	// -- end --
)

type SmsEvent struct {
	Content string `json:"content" bson:"content"`
	Email   string `json:"email" bson:"email"`
	Subject string `json:"subject" bson:"subject"`
	Success bool   `json:"success" bson:"success"`

	// -- extensions --
	// -- end --
}

func (t *SmsEvent) Valid() bool {
	// -- validation --
	// -- end --
	return true
}

func (v *Validator) SmsEventFromBody() *SmsEvent {
	b, err := ioutil.ReadAll(v.r.Body)
	if err != nil {
		v.Error("body", err.Error())
		return nil
	}

	ret := &SmsEvent{}
	err = json.Unmarshal(b, ret)
	if err != nil {
		v.Error("body", err.Error())
		return nil
	}

	if !ret.Valid() {
		v.Error("body", "Invalid SmsEvent")
		return nil
	}

	return ret
}

// -- code --

func (t *SmsEvent) Serialize() ([]byte, error) {
	return json.Marshal(t)
}

func (t *SmsEvent) Parse(b []byte) error {
	return json.Unmarshal(b, t)
}

// -- end --
