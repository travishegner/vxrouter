package vxrouter

import (
	"time"
)

// useful constants for the whole project
const (
	Version                    = "0.0.11"
	EnvPrefix                  = "VXR_"
	NetworkDriver              = "vxrNet"
	IpamDriver                 = "vxrIpam"
	DefaultReqAddrSleepTime    = 100 * time.Millisecond
	DefaultReqAddrRetryTimeout = 100 * time.Millisecond
	DefaultReqAddrTimeout      = 10 * time.Second
	DefaultRouteProto          = 192
)
