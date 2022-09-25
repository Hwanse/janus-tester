package janus

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
)

func (handle *Handle) CreateRoom(req *CreateRoomRequest) error {
	msg, err := handle.Request(req)
	if err != nil {
		return WrapError("failed to create room", err.Error())
	}

	response := CreateRoomResponse{}
	err = mapstructure.Decode(msg.PluginData.Data, &response)
	if err != nil {
		return err
	}

	if isUnexpectedResponse(response.VideoRoomResponseType.Type, SuccessCreateRoom) {
		return WrapError("failed to create room", response.Error())
	}

	return nil
}

func (handle *Handle) DestroyRoom(req *DestroyRoomRequest) error {
	msg, err := handle.Request(req)
	if err != nil {
		return WrapError("failed to destroy room", err.Error())
	}

	response := DestroyRoomResponse{}
	err = mapstructure.Decode(msg.PluginData.Data, &response)
	if err != nil {
		return err
	}

	if isUnexpectedResponse(response.VideoRoomResponseType.Type, SuccessDestroyRoom) {
		return WrapError("failed to destroy room", response.Error())
	}

	return nil
}

func isUnexpectedResponse(responseKey, successKey string) bool {
	if responseKey != successKey {
		return true
	}
	return false
}

func WrapError(description string, errText string) error {
	return fmt.Errorf("%s : %s", description, errText)
}
