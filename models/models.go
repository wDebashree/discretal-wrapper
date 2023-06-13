package models

import (
	"time"
)

type User struct {
	ID       string                 `json:"id"`
	Email    string                 `json:"email"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type Connect struct {
	Channel_ids []string
	Thing_ids   []string
}

type RegisterUserReq struct {
	Firstname string `json:"firstName" binding:"required" example:"John Doe"`
	Lastname  string `json:"lastName" binding:"required" example:"John Doe"`
	Email     string `json:"email" binding:"required" example:"user1@example.com"`
	Password  string `json:"password" binding:"required" example:"pass@1234"`
}
type RegisterUserRes struct {
	ID string `json:"id" binding:"required"`
}

type Metadata map[string]interface{}

type LoginUserReq struct {
	Email    string `json:"email" binding:"required" example:"user1@example.com"`
	Password string `json:"password" binding:"required" example:"pass@1234"`
}

type LoginUserRes struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJkaXNjcmV0YWwuYXV0aCIsInN1YiI6InVzZXIxQGV4YW1wbGUuY29tIiwiZXhwIjoxNjcyMDkyNDYzLCJpYXQiOjE2NzIwNTY0NjMsImlzc3Vlcl9pZCI6ImY5ZGJiZjIyLTcxZWQtNGIxZC1hZTU3LTk3ZjIxYjA4YTJiOSIsInR5cGUiOjB9.-Lcm4eWaR82W_oEVIgB24-ao6kI2NE80qR-nAiwh_c8"`
}

type ThingReq struct {
	Name        string                 `json:"name" binding:"required" example:"device1"`
	Coordinates map[string]interface{} `json:"coordinates"`
	Metadata    Metadata               `json:"metadata,omitempty"`
}

type ThingRes struct {
	Name     string   `json:"name" example:"device1"`
	ID       string   `json:"id" example:"8c0c7129-8857-4e50-a7e0-698e2865b0aa"`
	Key      string   `json:"key" example:"ef751d71-fb43-423c-a2eb-8602e6232cb4"`
	Metadata Metadata `json:"metadata,omitempty"`
}

type ThingResAll struct {
	Name     string   `json:"name" example:"device1"`
	Owner    string   `json:"owner" example:"user@example.com"`
	ID       string   `json:"id" example:"8c0c7129-8857-4e50-a7e0-698e2865b0aa"`
	Key      string   `json:"key" example:"ef751d71-fb43-423c-a2eb-8602e6232cb4"`
	Groups   []string `json:"groups,omitempty"`
	Metadata Metadata `json:"metadata,omitempty"`
}

type ChannelReq struct {
	Name     string   `json:"name" binding:"required" example:"channel1"`
	Metadata Metadata `json:"metadata,omitempty"`
}

type ChannelRes struct {
	Name     string   `json:"name" example:"channel1"`
	ID       string   `json:"id" example:"880d7429-8857-4e50-a7e0-698e2865b0aa"`
	Metadata Metadata `json:"metadata,omitempty"`
}
type ChannelResAll struct {
	Name     string   `json:"name" example:"channel1"`
	Owner    string   `json:"owner" example:"user@example.com"`
	ID       string   `json:"id" example:"880d7429-8857-4e50-a7e0-698e2865b0aa"`
	Metadata Metadata `json:"metadata,omitempty"`
}

type ThingsList struct {
	Things []ThingRes `json:"things"`
}

type ChannelsList struct {
	Channels []ChannelResAll `json:"channels"`
}

type GroupsList struct {
	Groups []GroupRes `json:"groups"`
}

type ItemId struct {
	Id string `uri:"id" binding:"required"`
}

type GroupReq struct {
	Name        string   `json:"name" binding:"required" example:"group1"`
	Description string   `json:"description" example:"group1"`
	Metadata    Metadata `json:"metadata,omitempty"`
}

// type GroupRes struct {
// 	ID          string   `json:"id,omitempty"`
// 	OwnerID     string   `json:"ownerid,omitempty"`
// 	ParentID    string   `json:"parentid,omitempty"`
// 	Name        string   `json:"name,omitempty"`
// 	Description string   `json:"description,omitempty"`
// 	Metadata    Metadata `json:"metadata,omitempty"`
// 	// Indicates a level in tree hierarchy.
// 	// Root node is level 1.
// 	Level int `json:"level,omitempty"`
// 	// Path in a tree consisting of group ids
// 	// parentID1.parentID2.childID1
// 	// e.g. 01EXPM5Z8HRGFAEWTETR1X1441.01EXPKW2TVK74S5NWQ979VJ4PJ.01EXPKW2TVK74S5NWQ979VJ4PJ
// 	Path      string      `json:"path,omitempty"`
// 	Children  []*GroupRes `json:"children,omitempty"`
// 	CreatedAt time.Time   `json:"createdat,omitempty"`
// 	UpdatedAt time.Time   `json:"updatedat,omitempty"`
// }

type GroupRes struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	OwnerID     string                 `json:"owner_id"`
	ParentID    string                 `json:"parent_id,omitempty"`
	Description string                 `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	// Indicates a level in tree hierarchy from first group node - root.
	Level int `json:"level"`
	// Path in a tree consisting of group ids
	// parentID1.parentID2.childID1
	// e.g. 01EXPM5Z8HRGFAEWTETR1X1441.01EXPKW2TVK74S5NWQ979VJ4PJ.01EXPKW2TVK74S5NWQ979VJ4PJ
	Path      string      `json:"path"`
	Children  []*GroupRes `json:"children,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type GrpPageRes struct {
	Limit  uint64 `json:"limit,omitempty"`
	Offset uint64 `json:"offset,omitempty"`
	Total  uint64 `json:"total"`
	Level  uint64 `json:"level"`
	Name   string `json:"name"`
}
type GroupPageRes struct {
	GrpPageRes
	Groups []GroupRes `json:"groups"`
}
type ViewGroupRes struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type GroupId struct {
	ID string `uri:"groupID" binding:"required"`
}

type AssignReq struct {
	Type    string   `json:"type,omitempty" binding:"required"`
	Members []string `json:"members" binding:"required"`
}

type AssignGroupReq struct {
	// Type   string   `json:"type,omitempty" binding:"required"`
	Type   string
	Groups []string `json:"groups" binding:"required"`
}

type Message interface{}

type RespError struct {
	Error string `json:"error,omitempty"`
}

type AddMapReq struct {
	ThingID     string   `json:"thing_id" binding:"required"`
	Coordinates Metadata `json:"coordinates" binding:"required"`
}

type pageRes struct {
	Total  uint64 `json:"total"`
	Offset uint64 `json:"offset"`
	Limit  uint64 `json:"limit"`
	Order  string `json:"order"`
	Dir    string `json:"direction"`
}
type ViewMapRes struct {
	ThingID     string                 `json:"thing_id"`
	Coordinates map[string]interface{} `json:"coordinates"`
	Status      bool                   `json:"status"`
	LastOnline  string                 `json:"lastOnline"`
	Owner       string                 `json:"owner"`
	Name        string                 `json:"name"`
	Key         string                 `json:"key"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}
type MapsPageRes struct {
	pageRes
	Maps []ViewMapRes `json:"maps"`
}

type UpdateMapReq struct {
	Coordinates Metadata `json:"coordinates" binding:"required"`
}

type LastSentTime struct {
	LatestTimes map[string]float64 `json:"lastSentTime"`
}

type thingsPageRes struct {
	Total     uint64 `json:"total"`
	Offset    uint64 `json:"offset"`
	Limit     uint64 `json:"limit"`
	Order     string `json:"order,omitempty"`
	Direction string `json:"dir,omitempty"`
	IsAdmin   bool   `json:"isadmin,omitempty"`
}
type ThingsPageRes struct {
	thingsPageRes
	Things []ThingResAll `json:"things"`
}

type GroupIDs struct {
	IDs []string `json:"group_ids"`
}
