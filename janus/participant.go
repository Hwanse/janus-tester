package janus

import (
	"github.com/mitchellh/mapstructure"
)

// TODO participant api
func (handle *Handle) JoinPublisher(req *JoinPublisherRequest) (*JoinPublisherResponse, error) {
	msg, err := handle.Message(req, nil)
	if err != nil {
		return nil, WrapError("failed to join the publisher", err.Error())
	}

	response := JoinPublisherResponse{}
	err = mapstructure.Decode(msg.Plugindata.Data, &response)
	if isUnexpectedResponse(response.VideoRoomResponseType.Type, SuccessJoin) {
		return nil, WrapError("failed to join the publisher", response.Error())
	}

	return &response, nil
}

func (handle *Handle) LeavePublisher(req *LeaveRequest) error {
	msg, err := handle.Message(req, nil)
	if err != nil {
		return WrapError("failed to leave the room", err.Error())
	}

	response := LeavePublisherResponse{}
	mapstructure.Decode(msg.Plugindata.Data, &response)
	if isUnexpectedResponse(response.VideoRoomResponseType.Type, TypeEvent) ||
		isUnexpectedResponse(response.Leaving, OK) {
		return WrapError("failed to join the publisher", response.Error())
	}

	return nil
}
