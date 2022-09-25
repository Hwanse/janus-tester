package janus

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

const RoomID = uint64(123456789)

func Test_CreateRoom(t *testing.T) {
	handle, err := attachVideoRoomHandle()
	defer handle.Detach()
	assert.NoError(t, err)

	request := &CreateRoomRequest{
		Request: TypeCreate,
		Room: Room{
			RoomID:              RoomID,
			IsPrivate:           false,
			PublisherLimitCount: 5,
			UseRecord:           false,
			NotifyJoining:       false,
			Bitrate:             128000,
			BitrateCap:          false,
		},
	}
	err = handle.CreateRoom(request)
	if err != nil {
		assert.NoError(t, err)
		return
	}

	err = cleanRoom(handle, RoomID)
	if err != nil {
		assert.NoError(t, err)
		return
	}
}

func Test_Exists_False(t *testing.T) {
	handle, err := attachVideoRoomHandle()
	defer handle.Detach()
	assert.NoError(t, err)

	roomID := uint64(1111111)

	req := &ExistsRoomRequest{
		Request: TypeExists,
		RoomID:  roomID,
	}

	exists, err := handle.ExistsRoom(req)
	if err != nil {
		assert.NoError(t, err)
		return
	}

	assert.False(t, exists)
}

func Test_Exists_True(t *testing.T) {
	handle, err := attachVideoRoomHandle()
	defer handle.Detach()
	assert.NoError(t, err)

	roomID := uint64(2222222)
	insertTestRoom(handle, roomID)

	req := &ExistsRoomRequest{
		Request: TypeExists,
		RoomID:  roomID,
	}

	exists, err := handle.ExistsRoom(req)
	if err != nil {
		assert.NoError(t, err)
		return
	}

	assert.True(t, exists)

	err = cleanRoom(handle, roomID)
	if err != nil {
		assert.NoError(t, err)
		return
	}
}

func Test_DestroyRoom(t *testing.T) {
	handle, err := attachVideoRoomHandle()
	defer handle.Detach()
	assert.NoError(t, err)

	roomID := uint64(12341234)
	insertTestRoom(handle, roomID)

	req := &DestroyRoomRequest{
		Request: TypeDestroy,
		RoomID:  roomID,
	}
	err = handle.DestroyRoom(req)
	if err != nil {
		assert.NoError(t, err)
		return
	}
}

func attachVideoRoomHandle() (*Handle, error) {
	url := fmt.Sprintf("ws://%s:%s/", JanusLocalHost, JanusWebsocketPort)
	client, err := WsConnect(url)
	if err != nil {
		return nil, err
	}

	session, err := client.Create()
	if err != nil {
		return nil, err
	}

	handle, err := session.Attach(VideoRoomPluginName)
	if err != nil {
		return nil, err
	}

	return handle, nil
}

func insertTestRoom(handle *Handle, id uint64) error {
	req := &CreateRoomRequest{
		Request: TypeCreate,
		Room: Room{
			RoomID: id,
		},
	}

	err := handle.CreateRoom(req)
	if err != nil {
		return err
	}

	return nil
}

func cleanRoom(handle *Handle, id uint64) error {
	req := &DestroyRoomRequest{
		Request:   TypeDestroy,
		RoomID:    id,
		Permanent: false,
	}

	err := handle.DestroyRoom(req)
	if err != nil {
		return err
	}

	return nil
}
