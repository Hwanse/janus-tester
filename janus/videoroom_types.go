package janus

import "fmt"

const (
	VideoRoomPluginName = "janus.plugin.videoroom"

	// Request
	TypeCreate  = "create"
	TypeDestroy = "destroy"
	TypeExists  = "exists"

	// Response
	TypeEvent          = "event"
	Success            = "success"
	SuccessCreateRoom  = "created"
	SuccessDestroyRoom = "destroyed"
)

type VideoRoomRequestType string
type VideoRoomResponseType struct {
	Type string `mapstructure:"videoroom"`
}

type Room struct {
	RoomID              uint64 `json:"room"`
	Description         string `json:"description,omitempty"`
	IsPrivate           bool   `json:"is_private"`
	PublisherLimitCount int    `json:"publishers,omitempty"`
	UseRecord           bool   `json:"record,omitempty"`
	RecordDirectory     string `json:"rec_dir,omitempty"`
	NotifyJoining       bool   `json:"notify_joining,omitempty"`
	Bitrate             int    `json:"bitrate,omitempty"`
	BitrateCap          bool   `json:"bitrate_cap,omitempty"`
	AudioCodec          string `json:"audiocodec,omitempty"`
	VideoCodec          string `json:"videocodec,omitempty"`
}

type CreateRoomRequest struct {
	Request VideoRoomRequestType `json:"request"`
	Room
}

type ExistsRoomRequest struct {
	Request VideoRoomRequestType `json:"request"`
	RoomID  uint64               `json:"room"`
}

type DestroyRoomRequest struct {
	Request   VideoRoomRequestType `json:"request"`
	RoomID    uint64               `json:"room"`
	Permanent bool                 `json:"permanent"`
}

type ErrorResponse struct {
	ErrorDescription string `mapstructure:"error"`
	ErrorCode        uint8  `mapstructure:"error_code"`
}

func (err *ErrorResponse) Error() string {
	return fmt.Sprintf("error_code: %d, error_desc: %s", err.ErrorCode, err.ErrorDescription)
}

type CreateRoomResponse struct {
	VideoRoomResponseType `mapstructure:",squash"`
	RoomID                uint64 `mapstructure:"room"`
	Permanent             bool
	ErrorResponse         `mapstructure:",squash"`
}

type ExistsRoomResponse struct {
	VideoRoomResponseType `mapstructure:",squash"`
	RoomID                uint64 `mapstructure:"room"`
	IsExists              bool   `mapstructure:"exists"`
	ErrorResponse         `mapstructure:",squash"`
}

type DestroyRoomResponse struct {
	VideoRoomResponseType `mapstructure:",squash"`
	RoomID                uint64 `mapstructure:"room"`
	Permanent             bool
	ErrorResponse         `mapstructure:",squash"`
}
