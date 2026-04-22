package app

import "github.com/gustavoz65/MoniMaster/internal/shared"

type Session struct {
	Mode     string
	Identity *shared.Identity
}

func (s Session) Actor() string {
	if s.Identity != nil {
		return s.Identity.Username
	}
	return "anonymous"
}
