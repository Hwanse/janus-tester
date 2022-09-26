package janus

import "fmt"

const (
	VideoRoomPluginName = "janus.plugin.videoroom"
	TypePublisher       = "publisher"
	TypeSubscriber      = "subscriber"

	// Request
	TypeCreate  = "create"
	TypeDestroy = "destroy"
	TypeExists  = "exists"
	TypeList    = "list"
	TypeJoin    = "join"
	TypeLeave   = "leave"

	// Response
	TypeEvent          = "event"
	Success            = "success"
	OK                 = "ok"
	SuccessCreateRoom  = "created"
	SuccessDestroyRoom = "destroyed"
	SuccessJoin        = "joined"
)

type VideoRoomRequestType string

type VideoRoomResponseType struct {
	Type string `mapstructure:"videoroom"`
}

type ErrorResponse struct {
	ErrorDescription string `mapstructure:"error"`
	ErrorCode        uint8  `mapstructure:"error_code"`
}

func (err *ErrorResponse) Error() string {
	return fmt.Sprintf("error_code: %d, error_desc: %s", err.ErrorCode, err.ErrorDescription)
}

type Room struct {
	RoomID              uint64 `json:"room" mapstructure:"room"`
	Description         string `json:"description,omitempty"`
	IsPrivate           bool   `json:"is_private" mapstructure:"is_private"`
	Secret              string `json:"secret,omitempty"`
	Pin                 string `json:"pin,omitempty"`
	PublisherLimitCount int    `json:"publishers,omitempty" mapstructure:"publishers"`
	UseRecord           bool   `json:"record,omitempty" mapstructure:"record"`
	RecordDirectory     string `json:"rec_dir,omitempty" mapstructure:"rec_dir"`
	LockRecord          bool   `json:"lock_record,omitempty" mapstructure:"lock_record"`
	NotifyJoining       bool   `json:"notify_joining,omitempty" mapstructure:"notify_joining"`
	Bitrate             int    `json:"bitrate,omitempty"`
	BitrateCap          bool   `json:"bitrate_cap,omitempty" mapstructure:"bitrate_cap"`
	AudioCodec          string `json:"audiocodec,omitempty" mapstructure:"audiocodec"`
	VideoCodec          string `json:"videocodec,omitempty" mapstructure:"videocodec"`
	RequirePvtID        bool   `json:"require_pvtid,omitempty" mapstructure:"require_pvtid"`
	SignedTokens        bool   `json:"signed_tokens,omitempty" mapstructure:"signed_tokens"`
	FirFreq             int    `json:"fir_freq,omitempty" mapstructure:"fir_freq"`
	VP9Profile          string `json:"vp9_profile,omitempty" mapstructure:"vp9_profile"`
	H264Profile         string `json:"h264_profile,omitempty" mapstructure:"h264_profile"`
	OpusFec             bool   `json:"opus_fec,omitempty" mapstructure:"opus_fec"`
	OpusDtx             bool   `json:"opus_dtx,omitempty" mapstructure:"opus_dtx"`
	VideoSvc            bool   `json:"video_svc,omitempty" mapstructure:"video_svc"`
	AudioLevelExt       bool   `json:"audiolevel_ext,omitempty" mapstructure:"audiolevel_ext"`
	AudioLevelEvent     bool   `json:"audiolevel_event,omitempty" mapstructure:"audiolevel_event"`
	AudioActivePackets  int    `json:"audio_active_packets,omitempty" mapstructure:"audio_active_packets"`
	AudioLevelAverage   int    `json:"audio_level_average,omitempty" mapstructure:"audio_level_average"`
	VideoOrientExt      bool   `json:"videoorient_ext,omitempty" mapstructure:"videoorient_ext"`
	PlayOutDelayExt     bool   `json:"playoutdelay_ext,omitempty" mapstructure:"playoutdelay_ext"`
	TransportWideCCExt  bool   `json:"transport_wide_cc_ext,omitempty" mapstructure:"transport_wide_cc_ext"`
	RequireE2ee         bool   `json:"require_e2ee,omitempty" mapstructure:"require_e2ee"`
	DummyPublisher      bool   `json:"dummy_publisher,omitempty" mapstructure:"dummy_publisher"`
	DummyStreams        bool   `json:"dummy_streams,omitempty" mapstructure:"dummy_streams"`
}

type Publisher struct {
	FeedID           uint64 `mapstructure:"id"`
	DisplayName      string `mapstructure:"display"`
	IsDummyPublisher bool   `mapstructure:"dummy"`
	Streams          []Stream
	IsTalking        bool `mapstructure:"talking"`
}

// Attendee non-activate publisher
type Attendee struct {
	ID          uint64 `mapstructure:"id"`
	DisplayName string `mapstructure:"display"`
}

// Stream Participant's stream info
type Stream struct {
	MediaType   string `mapstructure:"type"`
	MIndex      int    `mapstructure:"mindex"`
	MID         int    `mapstructure:"mid"`
	IsDisabled  bool   `mapstructure:"disabled"`
	Codec       string `mapstructure:"codec"`
	Description string `mapstructure:"description"`
	Moderated   bool   `mapstructure:"moderated"`
	Simulcast   bool   `mapstructure:"simulcast"`
	SVC         bool   `mapstructure:"svc"`
	IsTalking   bool   `mapstructure:"talking"`
}

// Room API Request

type CreateRoomRequest struct {
	Request VideoRoomRequestType `json:"request"`
	Room
}

type ExistsRoomRequest struct {
	Request VideoRoomRequestType `json:"request"`
	RoomID  uint64               `json:"room"`
}

type RoomListRequest struct {
	Request VideoRoomRequestType `json:"request"`
}

type DestroyRoomRequest struct {
	Request   VideoRoomRequestType `json:"request"`
	RoomID    uint64               `json:"room"`
	Permanent bool                 `json:"permanent"`
}

// Publisher API Request

type JoinPublisherRequest struct {
	Request     VideoRoomRequestType `json:"request"`
	RoomID      uint64               `json:"room"`
	PeerType    string               `json:"ptype"`
	FeedID      uint64               `json:"id,omitempty"`
	DisplayName string               `json:"display,omitempty"`
	Token       string               `json:"token,omitempty"`
}

type LeaveRequest struct {
	Request VideoRoomRequestType `json:"request"`
}

// Room API Response

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

type RoomListResponse struct {
	VideoRoomResponseType `mapstructure:",squash"`
	List                  []Room
	ErrorResponse         `mapstructure:",squash"`
}

type DestroyRoomResponse struct {
	VideoRoomResponseType `mapstructure:",squash"`
	RoomID                uint64 `mapstructure:"room"`
	Permanent             bool
	ErrorResponse         `mapstructure:",squash"`
}

// Publisher API Response

type JoinPublisherResponse struct {
	VideoRoomResponseType `mapstructure:",squash"`
	RoomID                uint64      `mapstructure:"room"`
	RoomDescription       string      `mapstructure:"description,omitempty"`
	FeedID                uint64      `mapstructure:"id"`
	PrivateID             uint64      `mapstructure:"private_id"`
	Publishers            []Publisher // Type에 맞게 변환됨
	Attendees             []Attendee  // Type에 맞게 변환됨
	ErrorResponse         `mapstructure:",squash"`
}

type LeavePublisherResponse struct {
	VideoRoomResponseType `mapstructure:",squash"`
	Leaving               string `mapstructure:"leaving"`
	ErrorResponse         `mapstructure:",squash"`
}
