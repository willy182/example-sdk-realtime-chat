package chat

import "errors"

var ErrClientClosed = errors.New("client closed")
var ErrTransportNotConnected = errors.New("transport not connected")
