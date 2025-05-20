package constant

type ManagerStatus int32

const (
	ManagerStatusConnected ManagerStatus = iota
	ManagerStatusDisconnected
	ManagerStatusReconnecting
	ManagerStatusClosed
)

type ClientStatus int32

const (
	ClientInit ClientStatus = iota
	ClientRunning
	ClientReset
	ClientStop
)
