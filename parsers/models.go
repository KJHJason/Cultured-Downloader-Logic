package parsers

import (
	"net/http"
)

type cookieInfo struct {
	Domain   string
	Name     string
	SameSite http.SameSite
}

// For the exported cookies in JSON instead of Netscape format
type ExportedCookies []struct {
	Domain   string  `json:"domain"`
	Expire   float64 `json:"expirationDate"`
	HttpOnly bool    `json:"httpOnly"`
	Name     string  `json:"name"`
	Path     string  `json:"path"`
	Secure   bool    `json:"secure"`
	Value    string  `json:"value"`
	Session  bool    `json:"session"`
}

type cookieInfoArgs struct {
	name     string
	sameSite http.SameSite
}

func NewCookieInfoArgs(name string, sameSite http.SameSite) *cookieInfoArgs {
	return &cookieInfoArgs{
		name:     name,
		sameSite: sameSite,
	}
}
