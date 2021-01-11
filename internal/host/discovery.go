package host

type discovery struct {
	h Host
}

func NewDiscovery(h Host) *discovery {
	return &discovery{h: h}
}
