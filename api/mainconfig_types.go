package api

import (
	cl "warp-server/pkg/controlloop"
)

const (
	VPNConnectedCondition     = "VPNConnected"
	PFDisabledCondition       = "PFDisabled"
	SSHKeysInstalledCondition = "SSHKeysInstalled"
	TunnelEnabledCondition    = "TunnelEnabled"

	VPNReasonReady                   = "VPNDisconnected"
	VPNConnectionFailedReason        = "VPNConnectionFailed"
	VPNConnectionStateFailedReason   = "VPNConnectionStateFailed"
	VPNDisconnectFailedReason        = "VPNDisconnectFailed"
	VPNConnectionStopFiled           = "VPNStopFailed"
	PFDisabledFailedReason           = "PFDisabledFailed"
	SSHKeysFailedReason              = "SSHKeysFailed"
	TunnelInitializationFailedReason = "TunnelInitializationFailed"
	TunnelDisablingFailedReason      = "TunnelDisablingFailed"
)

type MainConfig struct {
	cl.Resource
	Spec MainConfigSpec
}

type MainConfigSpec struct{}

// todo generate
func (c *MainConfig) DeepCopy() *MainConfig {
	return cl.DeepCopyStruct(c).(*MainConfig)
}
