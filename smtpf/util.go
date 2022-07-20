package smtpf

import (
	"fmt"
	"strings"
)

func ParseDomain(_emailAddr string) (domain string, err error) {
	emailAddrs := strings.Split(_emailAddr, "@")
	if len(emailAddrs) != 2 {
		return "", fmt.Errorf("invalid address | %s", _emailAddr)
	}
	domain = emailAddrs[1]
	return domain, nil
}
