package url

type URL struct {
	Hostname  string `json:"hostname"`
	Malicious bool   `json:"malicious"`
}

func NewURL() *URL {
	return &URL{}
}

func (u *URL) SetHostname(hostname string) {
	u.Hostname = hostname
}

func (u *URL) SetMalicious(malicious bool) {
	u.Malicious = malicious
}

func (u *URL) IsMalicious() bool {
	return u.Malicious == true
}
