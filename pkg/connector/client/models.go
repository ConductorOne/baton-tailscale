package client

import "time"

type Resource struct {
	Id          string `json:"id"`
	DisplayName string `json:"name"`
}

type UsersAPIData struct {
	Users []User `json:"users"`
}

type User struct {
	ID            string    `json:"id,omitempty"`
	DisplayName   string    `json:"displayName,omitempty"`
	LoginName     string    `json:"loginName,omitempty"`
	ProfilePicURL string    `json:"profilePicUrl"`
	TailnetID     string    `json:"tailnetId,omitempty"`
	Created       time.Time `json:"created,omitempty"`
	Type          string    `json:"type,omitempty"`
	Role          string    `json:"role,omitempty"`
	Status        string    `json:"status,omitempty"`
	DeviceCount   int       `json:"deviceCount,omitempty"`
	LastSeen      time.Time `json:"lastSeen,omitempty"`
}

type Role struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}
