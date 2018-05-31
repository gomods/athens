package eventlog

// PointerRegistry is a key/value store that stores an event log pointer for one
// or more Olympus deployments. It is used in proxies (Athens) and Olympus
// deployments as part of the event log sync process
type PointerRegistry interface {
	LookupPointer(deploymentID string) (string, error)
}
