package http

import "github.com/xtls/xray-core/common/protocol"

// Authenticator validates HTTP proxy credentials against the mixed inbound's
// dynamic user registry.
type Authenticator interface {
	Authenticate(username, password string) (*protocol.MemoryUser, bool)
}
