package g

import "net/smtp"

type HotmailStmpAuth struct {
	userName string
	passwd   string
}

func NewHotmailStmpAuth(username, passwd string) smtp.Auth {
	return &HotmailStmpAuth{
		userName: username,
		passwd:   passwd,
	}
}

func (h *HotmailStmpAuth) Start(server *smtp.ServerInfo) (proto string, toServer []byte, err error) {
	return "LOGIN", []byte(h.userName), nil
}
func (a *HotmailStmpAuth) Next(fromServer []byte, more bool) (toServer []byte, err error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.userName), nil
		case "Password:":
			return []byte(a.passwd), nil
		}
	}
	return nil, nil
}
