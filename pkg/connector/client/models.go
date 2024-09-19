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

type DevicesAPIData struct {
	Devices []Device `json:"devices,omitempty"`
}

type Device struct {
	Addresses                 []string  `json:"addresses,omitempty"`
	Authorized                bool      `json:"authorized,omitempty"`
	BlocksIncomingConnections bool      `json:"blocksIncomingConnections,omitempty"`
	ClientVersion             string    `json:"clientVersion,omitempty"`
	Created                   time.Time `json:"created,omitempty"`
	Expires                   time.Time `json:"expires,omitempty"`
	Hostname                  string    `json:"hostname,omitempty"`
	ID                        string    `json:"id,omitempty"`
	IsExternal                bool      `json:"isExternal,omitempty"`
	KeyExpiryDisabled         bool      `json:"keyExpiryDisabled,omitempty"`
	LastSeen                  time.Time `json:"lastSeen,omitempty"`
	MachineKey                string    `json:"machineKey,omitempty"`
	Name                      string    `json:"name,omitempty"`
	NodeID                    string    `json:"nodeId,omitempty"`
	NodeKey                   string    `json:"nodeKey,omitempty"`
	Os                        string    `json:"os,omitempty"`
	TailnetLockError          string    `json:"tailnetLockError,omitempty"`
	TailnetLockKey            string    `json:"tailnetLockKey,omitempty"`
	UpdateAvailable           bool      `json:"updateAvailable,omitempty"`
	User                      string    `json:"user,omitempty"`
}

type UserInvitesAPIData []struct {
	ID              string    `json:"id"`
	Role            string    `json:"role"`
	TailnetID       int64     `json:"tailnetId"`
	InviterID       int64     `json:"inviterId"`
	Email           string    `json:"email"`
	LastEmailSentAt time.Time `json:"lastEmailSentAt"`
	InviteURL       string    `json:"inviteUrl"`
}
