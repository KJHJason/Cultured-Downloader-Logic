package cf

type Cookie struct {
	Name         string `json:"name"`
	Value        string `json:"value"`
	Domain       string `json:"domain"`
	Path         string `json:"path"`
	Expires      int    `json:"expires"`
	Size         int    `json:"size"`
	HttpOnly     bool   `json:"httpOnly"`
	Secure       bool   `json:"secure"`
	Session      bool   `json:"session"`
	Priority     string `json:"priority"`
	SameParty    bool   `json:"sameParty"`
	SourceScheme string `json:"sourceScheme"`
	SourcePort   string `json:"sourcePort"`
}

type Cookies []*Cookie
