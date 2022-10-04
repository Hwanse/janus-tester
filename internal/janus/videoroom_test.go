package janus

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

const RoomID = uint64(123456789)

// VideoRoom Room API Test

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
	assert.NoError(t, err)

	err = cleanRoom(handle, RoomID)
	assert.NoError(t, err)
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
	assert.NoError(t, err)

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
	assert.NoError(t, err)
	assert.True(t, exists)

	err = cleanRoom(handle, roomID)
	assert.NoError(t, err)
}

func Test_RoomList(t *testing.T) {
	handle, err := attachVideoRoomHandle()
	defer handle.Detach()
	assert.NoError(t, err)

	// insert room list for test
	id := uint64(10000000)
	idList := make([]uint64, 0)
	for i := 1; i <= 5; i++ {
		idList = append(idList, id+uint64(i))
		insertTestRoom(handle, id+uint64(i))
	}

	roomList, err := handle.RoomList()
	assert.NoError(t, err)

	assert.NotNil(t, roomList)
	assert.GreaterOrEqual(t, len(roomList), len(idList))

	for _, id := range idList {
		cleanRoom(handle, id)
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
	assert.NoError(t, err)
}

// VideoRoom Participant API Test

func Test_JoinPublisher(t *testing.T) {
	handle, err := attachVideoRoomHandle()
	defer handle.Detach()
	assert.NoError(t, err)

	handle2, err := attachVideoRoomHandle()
	defer handle2.Detach()

	roomID := uint64(123333333)
	insertTestRoom(handle, roomID)
	defer cleanRoom(handle, roomID)

	req := &JoinPublisherRequest{
		Request:  TypeJoin,
		RoomID:   roomID,
		PeerType: TypePublisher,
	}

	response, err := handle.JoinPublisher(req)
	response2, err2 := handle2.JoinPublisher(req)

	assert.NoError(t, err)
	assert.NoError(t, err2)
	assert.NotNil(t, response)
	assert.NotNil(t, response2)
	assert.NotZero(t, response.FeedID)
	assert.NotZero(t, response2.FeedID)
	assert.Greater(t, len(response2.Attendees), 0)
}

func Test_LeavePublisher(t *testing.T) {
	handle, err := attachVideoRoomHandle()
	defer handle.Detach()
	assert.NoError(t, err)

	roomID := uint64(8343348921)
	insertTestRoom(handle, roomID)
	defer cleanRoom(handle, roomID)

	joinReq := &JoinPublisherRequest{
		Request:  TypeJoin,
		RoomID:   roomID,
		PeerType: TypePublisher,
	}
	response, err := handle.JoinPublisher(joinReq)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotZero(t, response.FeedID)

	leaveReq := &LeaveRequest{TypeLeave}
	err = handle.LeavePublisher(leaveReq)

	assert.NoError(t, err)
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
			RoomID:        id,
			NotifyJoining: true,
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
