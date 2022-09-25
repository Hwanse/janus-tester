package janus

import "fmt"

/*

room-<unique room ID>: {
        description = This is my awesome room
        is_private = true|false (private rooms don't appear when you do a 'list' request, default=false)
        secret = <optional password needed for manipulating (e.g. destroying) the room>
        pin = <optional password needed for joining the room>
        require_pvtid = true|false (whether subscriptions are required to provide a valid private_id
                                 to associate with a publisher, default=false)
        signed_tokens = true|false (whether access to the room requires signed tokens; default=false,
                                 only works if signed tokens are used in the core as well)
        publishers = <max number of concurrent senders> (e.g., 6 for a video
                                 conference or 1 for a webinar, default=3)
        bitrate = <max video bitrate for senders> (e.g., 128000)
        bitrate_cap = <true|false, whether the above cap should act as a limit to dynamic bitrate changes by publishers, default=false>,
        fir_freq = <send a FIR to publishers every fir_freq seconds> (0=disable)
        audiocodec = opus|g722|pcmu|pcma|isac32|isac16 (audio codec to force on publishers, default=opus
                                can be a comma separated list in order of preference, e.g., opus,pcmu)
        videocodec = vp8|vp9|h264|av1|h265 (video codec to force on publishers, default=vp8
                                can be a comma separated list in order of preference, e.g., vp9,vp8,h264)
        vp9_profile = VP9-specific profile to prefer (e.g., "2" for "profile-id=2")
        h264_profile = H.264-specific profile to prefer (e.g., "42e01f" for "profile-level-id=42e01f")
        opus_fec = true|false (whether inband FEC must be negotiated; only works for Opus, default=true)
        opus_dtx = true|false (whether DTX must be negotiated; only works for Opus, default=false)
        video_svc = true|false (whether SVC support must be enabled; only works for VP9, default=false)
        audiolevel_ext = true|false (whether the ssrc-audio-level RTP extension must be
                negotiated/used or not for new publishers, default=true)
        audiolevel_event = true|false (whether to emit event to other users or not, default=false)
        audio_active_packets = 100 (number of packets with audio level, default=100, 2 seconds)
        audio_level_average = 25 (average value of audio level, 127=muted, 0='too loud', default=25)
        videoorient_ext = true|false (whether the video-orientation RTP extension must be
                negotiated/used or not for new publishers, default=true)
        playoutdelay_ext = true|false (whether the playout-delay RTP extension must be
                negotiated/used or not for new publishers, default=true)
        transport_wide_cc_ext = true|false (whether the transport wide CC RTP extension must be
                negotiated/used or not for new publishers, default=true)
        record = true|false (whether this room should be recorded, default=false)
        rec_dir = <folder where recordings should be stored, when enabled>
        lock_record = true|false (whether recording can only be started/stopped if the secret
                                is provided, or using the global enable_recording request, default=false)
        notify_joining = true|false (optional, whether to notify all participants when a new
                                participant joins the room. The Videoroom plugin by design only notifies
                                new feeds (publishers), and enabling this may result extra notification
                                traffic. This flag is particularly useful when enabled with require_pvtid
                                for admin to manage listening only participants. default=false)
        require_e2ee = true|false (whether all participants are required to publish and subscribe
                                using end-to-end media encryption, e.g., via Insertable Streams; default=false)
        dummy_publisher = true|false (whether a dummy publisher should be created in this room,
                                with one separate m-line for each codec supported in the room; this is
                                useful when there's a need to create subscriptions with placeholders
                                for some or all m-lines, even when they aren't used yet; default=false)
        dummy_streams = in case dummy_publisher is set to true, array of codecs to offer,
                                optionally with a fmtp attribute to match (codec/fmtp properties).
                                If not provided, all codecs enabled in the room are offered, with no fmtp.
                                Notice that the fmtp is parsed, and only a few codecs are supported.
}

*/

const (
	VideoRoomPluginName = "janus.plugin.videoroom"

	// Request
	TypeCreate  = "create"
	TypeDestroy = "destroy"

	// Response
	TypeEvent          = "event"
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

type DestroyRoomResponse struct {
	VideoRoomResponseType `mapstructure:",squash"`
	RoomID                uint64 `mapstructure:"room"`
	Permanent             bool
	ErrorResponse         `mapstructure:",squash"`
}
