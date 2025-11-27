package roblox

import "time"

type UniverseIdResponse struct {
	UniverseId int `json:"universeId"`
}

type Universe struct {
	Path                    string    `json:"path"`
	CreateTime              time.Time `json:"createTime"`
	UpdateTime              time.Time `json:"updateTime"`
	DisplayName             string    `json:"displayName"`
	Description             string    `json:"description"`
	User                    string    `json:"user"`
	Visibility              string    `json:"visibility"`
	VoiceChatEnabled        bool      `json:"voiceChatEnabled"`
	AgeRating               string    `json:"ageRating"`
	PrivateServerPriceRobux string    `json:"privateServerPriceRobux"`
	DesktopEnabled          bool      `json:"desktopEnabled"`
	MobileEnabled           bool      `json:"mobileEnabled"`
	TabletEnabled           bool      `json:"tabletEnabled"`
	ConsoleEnabled          bool      `json:"consoleEnabled"`
	VrEnabled               bool      `json:"vrEnabled"`
	RootPlace               string    `json:"rootPlace"`
	TemplateRootPlace       string    `json:"templateRootPlace"`
}

type User struct {
	Path        string    `json:"path"`
	CreateTime  time.Time `json:"createTime"`
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"displayName"`
	About       string    `json:"about"`
	Locale      string    `json:"locale"`
	Premium     bool      `json:"premium"`
}
