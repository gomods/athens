package proxy

// State is state of the proxy
type State struct {
	OlympusEndpoint string `json:"olympus_enpdpoint"`
	SequenceID      string `json:"sequence_id"`
}
