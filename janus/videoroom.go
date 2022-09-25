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

func (handle *Handle) ExistsRoom(req *ExistsRoomRequest) (bool, error) {
	msg, err := handle.Request(req)
	if err != nil {
		return false, WrapError("failed to exists room", err.Error())
	}

	response := ExistsRoomResponse{}
	err = mapstructure.Decode(msg.PluginData.Data, &response)
	if err != nil {
		return false, err
	}

	if isUnexpectedResponse(response.VideoRoomResponseType.Type, Success) {
		return false, WrapError("failed to exists room", response.Error())
	}

	return response.IsExists, nil
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

func (handle *Handle) RoomList() ([]Room, error) {
	req := &RoomListRequest{Request: TypeList}
	msg, err := handle.Request(req)
	if err != nil {
		return nil, WrapError("failed to get all room list", err.Error())
	}

	response := RoomListResponse{}
	err = mapstructure.Decode(msg.PluginData.Data, &response)
	if err != nil {
		return nil, err
	}

	if isUnexpectedResponse(response.VideoRoomResponseType.Type, Success) {
		return nil, WrapError("failed to get all room list", response.Error())
	}

	roomList, err := mapSliceConvertToRoomSlice(&response.List)
	if err != nil {
		return nil, err
	}

	return roomList, nil
}

func mapSliceConvertToRoomSlice(list *[]map[string]interface{}) ([]Room, error) {
	roomList := make([]Room, 0)
	for _, data := range *list {
		room := Room{}
		err := mapstructure.Decode(data, &room)
		if err != nil {
			return nil, err
		}
		roomList = append(roomList, room)
	}
	return roomList, nil
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
