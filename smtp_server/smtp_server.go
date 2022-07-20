package smtp_server

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/smtp"
	"smtp/smtp_server/config"
	"smtp/smtpf"
	"time"

	smtpd "github.com/emersion/go-smtp"
)

func RunServe() error {
	// config load
	err := config.GC.Load()
	if err != nil {
		return err
	}

	for _, relayServer := range config.GC.RelayServer {
		// relay server 별 goroutine 실행
		go func(_pt_config *config.RelayServer) {
			// 1. init
			smtpServer := runServe(_pt_config)
			// 2. ListenAndServe
			log.Println("Starting server at", smtpServer.Addr)
			if err := smtpServer.ListenAndServe(); err != nil {
				log.Fatal(err)
			}
		}(relayServer)
	}
	return nil
}

func runServe(_cfg *config.RelayServer) *smtpd.Server {
	pt_server := smtpd.NewServer(&Backend{
		addr:       _cfg.Addr,
		port:       _cfg.Port,
		username:   _cfg.Username,
		password:   _cfg.Password,
		relayAddr:  _cfg.Relay_addr,
		relayPort:  _cfg.Relay_port,
		heloDomain: _cfg.HeloDomain,
		weighted:   _cfg.Weighed,
	})
	pt_server.Addr = ":" + _cfg.Port
	pt_server.Domain = _cfg.Addr
	pt_server.ReadTimeout = 10 * time.Second
	pt_server.WriteTimeout = 10 * time.Second
	pt_server.MaxMessageBytes = int(config.GC.MaxLenBytes)
	pt_server.MaxRecipients = int(config.GC.MaxRcpt)
	pt_server.AllowInsecureAuth = config.GC.AllowInsecureAuth

	return pt_server
}

type Backend struct {
	addr       string
	port       string
	username   string
	password   string
	relayAddr  string
	relayPort  string
	heloDomain string
	weighted   string
}

func (t *Backend) Login(state *smtpd.ConnectionState, username, password string) (smtpd.Session, error) {
	if username != t.username || password != t.password {
		return nil, errors.New("invalid username or password")
	}

	return &Session{
		addr:       t.addr,
		port:       t.port,
		username:   username,
		password:   password,
		relayAddr:  t.relayAddr,
		relayPort:  t.relayPort,
		heloDomain: t.heloDomain,
	}, nil
}

// 에러 처리
func (t *Backend) AnonymousLogin(state *smtpd.ConnectionState) (smtpd.Session, error) {
	return nil, smtpd.ErrAuthRequired
}

type Session struct {
	addr       string
	port       string
	username   string
	password   string
	relayAddr  string
	relayPort  string
	heloDomain string

	from string
	to   string
}

func (t *Session) Mail(from string, opts smtpd.MailOptions) error {
	t.from = from
	return nil
}

func (t *Session) Rcpt(to string) error {
	t.to = to
	return nil
}

func (t *Session) Reset() {}

func (t *Session) Logout() error {
	t.username = ""
	t.password = ""
	return nil
}

func (t *Session) Data(_r io.Reader) error {
	// 1. read email contents
	bt_email, err := ioutil.ReadAll(_r)
	if err != nil {
		return err
	}
	// 2. sign DKIM
	err = t.SignDKIM(&bt_email)
	if err != nil {
		return err
	}
	// 3. dial and send
	err = t.DialAndSend(bt_email)
	if err != nil {
		return err
	}

	return nil
}

func (t *Session) SignDKIM(_email *[]byte) error {
	// 전처리
	s_canonicalization := config.GC.Canonicalization
	u8_sig_expire_in := config.GC.SigExpireIn
	arrs_header := []string{"from", "to", "subject", "date"}

	// DKIM 구조체 제작
	pt_dkim_sign := smtpf.New_DKIM()
	// 1. domain - always
	for _, pt_master := range config.GC.DkimSigns {
		pt_dkim_sign.Add_Option(
			false,
			pt_master.Domain,
			pt_master.Selector,
			pt_master.PrivKey_base64,
			arrs_header,
			s_canonicalization,
			u8_sig_expire_in,
		)
	}
	// 2. domain - only same domain
	for _, pt_sub := range config.GC.DkimSigns_onlySameDomain {
		pt_dkim_sign.Add_Option(
			false,
			pt_sub.Domain,
			pt_sub.Selector,
			pt_sub.PrivKey_base64,
			arrs_header,
			s_canonicalization,
			u8_sig_expire_in,
		)
	}
	return pt_dkim_sign.Sign(t.from, _email)
}

func (t *Session) DialAndSend(_email []byte) error {
	var err error
	if t.relayAddr == "" {
		// direct server - addr to domain 의 mx record 로 전송
		err = smtpf.MXRecord_SendToMostPriortyRecord(t.to, func(_s_host_name string) (is_send bool, err error) {
			err = smtpf.SendMail_raw(_s_host_name, config.GC.ConnextPort, t.heloDomain, nil, t.from, []string{t.to}, &_email)
			if err != nil {
				// 임시 - 에러 상태값에 따라 nil 반환 필요 ( nil 반환하면 다음 mx record 순회 )
				// 우리 이메일 서버 에러가 아니라면 도메인 내 모든 mx record 를 순회해야 하기 때문에
				return false, err
			}
			return true, nil
		})
	} else {
		// relay server - 지정된 relay address 로 전송
		t_auth := smtp.PlainAuth("", t.username, t.password, t.addr)
		err = smtpf.SendMail_raw(t.relayAddr, t.relayPort, t.heloDomain, t_auth, t.from, []string{t.to}, &_email)
	}
	if err != nil {
		log.Print("smtp_server,dial_and_send,fail", "err - %v", err)
		return err
	}
	log.Print("smtp_server,dial_and_send,success", "")
	return nil
}
