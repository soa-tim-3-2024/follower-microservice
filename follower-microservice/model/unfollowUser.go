package model

import (
	"encoding/json"
	"io"
)

type UnfollowUser struct {
	UserId           string `json:"userId,omitempty"`
	UserToUnfollowId string `json:"userToUnfollowId,omitempty"`
}

func (o *UnfollowUser) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(o)
}
