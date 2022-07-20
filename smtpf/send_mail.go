package smtpf

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
)

//---------------------------------------------------------------------------------------------------//
// Send mail

func SendMail(_addr string, _port string, _heloDomain string, _username string, _password string, _emailJob *EmailJob) error {
	mail, err := _emailJob.Conv__bt()
	if err != nil {
		return err
	}

	// auth 처리
	var auth smtp.Auth
	if _username != "" && _password != "" {
		auth = smtp.PlainAuth("", _username, _password, _addr)
	} else {
		auth = nil
	}

	return SendMail_raw(_addr, _port, _heloDomain, auth, _emailJob.from, _emailJob.to, &mail)
}

//---------------------------------------------------------------------------------------------------//
// Send mail ( by bt )

func SendMail_raw(_addr string, _port string, _heloDomain string, _auth smtp.Auth, _from string, _to []string, _message *[]byte) error {
	// 1. 전처리
	if len(*_message) == 0 {
		return fmt.Errorf("msg is nil")
	}
	if len(_port) != 0 && _port[0:1] != ":" {
		_port = ":" + _port
	}

	// 2. dial
	client, err := smtp.Dial(_addr + _port)
	if err != nil {
		return err
	}
	defer client.Close()

	// 3. cert
	{
		// hello
		if _heloDomain != "" {
			if err = client.Hello(_heloDomain); err != nil {
				return err
			}
		}

		// start tls
		isExistTLS, _ := client.Extension("STARTTLS")
		if isExistTLS == true {
			config := &tls.Config{
				ServerName:         _addr,
				InsecureSkipVerify: true,
			}
			if client.StartTLS(config); err != nil {
				return err
			}
		}

		// auth
		if _auth != nil {
			isExistExtAuth, _ := client.Extension("AUTH")
			if !isExistExtAuth {
				return fmt.Errorf("smtp server doesn't support AUTH")
			}
			if err = client.Auth(_auth); err != nil {
				return err
			}
		}

	}

	// 4. Send
	{
		if err := client.Mail(_from); err != nil {
			return err
		}
		for _, s_to := range _to {
			if err = client.Rcpt(s_to); err != nil {
				return err
			}
		}
		w, err := client.Data()
		if err != nil {
			return err
		}
		_, err = w.Write(*_message)
		if err != nil {
			return err
		}
		if err = w.Close(); err != nil {
			return err
		}
	}
	return nil
}
