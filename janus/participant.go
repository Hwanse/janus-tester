package janus

import (
	"github.com/mitchellh/mapstructure"
)

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

func (handle *Handle) JoinSubscriber(req *JoinSubscriberRequest) (*JoinSubscriberResponse, error) {
	msg, err := handle.Message(req, nil)
	if err != nil {
		return nil, WrapError("failed to join the subscriber", err.Error())
	}

	response := JoinSubscriberResponse{}
	err = mapstructure.Decode(msg.Plugindata.Data, &response)
	if isUnexpectedResponse(response.VideoRoomResponseType.Type, SuccessAttached) {
		return nil, WrapError("failed to join the publisher", response.Error())
	}
	response.Jsep = msg.Jsep

	return &response, nil
}

func (handle *Handle) Publish(req *PublishRequest, jsep interface{}) (interface{}, error) {
	msg, err := handle.Message(req, jsep)
	if err != nil {
		return nil, WrapError("failed to publish", err.Error())
	}

	response := PublishResponse{}
	err = mapstructure.Decode(msg.Plugindata.Data, &response)
	if isUnexpectedResponse(response.VideoRoomResponseType.Type, TypeEvent) ||
		isUnexpectedResponse(response.Configured, OK) {
		return nil, WrapError("failed to publish", response.Error())
	}

	return msg.Jsep, nil
}

func (handle *Handle) SubscribeStart(req *SubscribeStartRequest, jsep interface{}) error {
	msg, err := handle.Message(req, jsep)
	if err != nil {
		return WrapError("failed to subscribe", err.Error())
	}

	response := SubscribeStartResponse{}
	err = mapstructure.Decode(msg.Plugindata.Data, &response)
	if isUnexpectedResponse(response.VideoRoomResponseType.Type, TypeEvent) ||
		isUnexpectedResponse(response.Started, OK) {
		return WrapError("failed to subscribe", response.Error())
	}

	return nil
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
		return WrapError("failed to leave the room", response.Error())
	}

	return nil
}

func (handle *Handle) LeaveSubscriber(req *LeaveRequest) error {
	msg, err := handle.Message(req, nil)
	if err != nil {
		return WrapError("failed to leave the room", err.Error())
	}

	response := LeaveSubscriberResponse{}
	mapstructure.Decode(msg.Plugindata.Data, &response)
	if isUnexpectedResponse(response.VideoRoomResponseType.Type, TypeEvent) ||
		isUnexpectedResponse(response.Left, OK) {
		return WrapError("failed to leave the room", response.Error())
	}

	return nil
}

// TODO: participant advanced api
