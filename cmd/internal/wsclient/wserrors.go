package wsclient

import "errors"

func ErrNotConnected() error {
	return errors.New("websocket client is not connected")
}
