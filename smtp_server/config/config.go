package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"gopkg.in/ini.v1"
)

var GC Cfg

type Cfg struct {
	preName string
	path    string

	Canonicalization  string
	SigExpireIn       uint64
	MaxLenBytes       uint64
	MaxRcpt           uint64
	ConnextPort       string
	AllowInsecureAuth bool

	RelayServer              []*RelayServer
	DkimSigns                []*DKIMSign
	DkimSigns_onlySameDomain []*DKIMSign
}

type RelayServer struct {
	Addr       string
	Port       string
	Relay_addr string
	Relay_port string
	Username   string
	Password   string
	HeloDomain string
	Weighed    string
}

type DKIMSign struct {
	IsUse          bool
	Domain         string
	Selector       string
	PrivKey_base64 string
}

func (t *Cfg) Load() error {
	// 섹션 지정
	t.preName = "smtp_server__"

	path := "config.ini"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		path = "config__dev.ini"
	}
	t.path = path
	cfg, err := ini.InsensitiveLoad(t.path)
	if err != nil {
		cfg = ini.Empty()
	}

	// base
	t.Canonicalization = cfg.Section(t.preName + "base").Key("canonicalization").MustString("relaxed/relaxed")
	sigExpireIn := cfg.Section(t.preName + "base").Key("sig_expire_in").MustString("3600")
	t.SigExpireIn, err = strconv.ParseUint(sigExpireIn, 10, 64)
	if err != nil {
		t.SigExpireIn = 3600
	}
	maxlen := cfg.Section(t.preName + "base").Key("max_msg_len_bytes").MustString("1048576")
	t.MaxLenBytes, err = strconv.ParseUint(maxlen, 10, 64)
	if err != nil {
		t.MaxLenBytes = 1048576
	}
	maxRept := cfg.Section(t.preName + "base").Key("max_rcpt").MustString("100")
	t.MaxRcpt, err = strconv.ParseUint(maxRept, 10, 64)
	if err != nil {
		t.MaxRcpt = 100
	}
	AllowInsecureAuth := cfg.Section(t.preName + "base").Key("allow_insecure_auth").MustString("true")
	if AllowInsecureAuth == "true" {
		t.AllowInsecureAuth = true
	} else {
		t.AllowInsecureAuth = false
	}
	t.ConnextPort = cfg.Section(t.preName + "base").Key("connect_port").MustString("25")
	if err != nil {
		t.ConnextPort = "25"
	}

	// relay server
	{
		t.RelayServer = make([]*RelayServer, 0, 10)
		section := cfg.Section(t.preName + fmt.Sprintf("relay_server"))
		for i := 0; ; i++ {
			if section.HasKey(fmt.Sprintf("%d__addr", i)) == false {
				break
			}
			relayServer := &RelayServer{}
			{
				relayServer.Addr = section.Key(fmt.Sprintf("%d__addr", i)).MustString("")
				relayServer.Port = section.Key(fmt.Sprintf("%d__port", i)).MustString("")
				relayServer.Relay_addr = section.Key(fmt.Sprintf("%d__relay_addr", i)).MustString("")
				relayServer.Relay_port = section.Key(fmt.Sprintf("%d__relay_port", i)).MustString("")
				relayServer.Username = section.Key(fmt.Sprintf("%d__username", i)).MustString("")
				relayServer.Password = section.Key(fmt.Sprintf("%d__password", i)).MustString("")
				relayServer.HeloDomain = section.Key(fmt.Sprintf("%d__helo_domain", i)).MustString("")
				relayServer.Weighed = section.Key(fmt.Sprintf("%d__weighted", i)).MustString("")
			}
			t.RelayServer = append(t.RelayServer, relayServer)
		}
	}
	// dkim sign - always
	{
		t.DkimSigns = make([]*DKIMSign, 0, 10)
		section := cfg.Section(t.preName + "dkim_signing__always")
		for i := 0; ; i++ {
			if section.HasKey(fmt.Sprintf("%d__is_use", i)) == false {
				break
			}
			dkimSign := &DKIMSign{}
			{
				isUse := section.Key(fmt.Sprintf("%d__is_use", i)).MustString("false")
				if isUse == "true" {
					dkimSign.IsUse = true
				} else {
					dkimSign.IsUse = false
				}
				dkimSign.Domain = section.Key(fmt.Sprintf("%d__domain", i)).MustString("")
				dkimSign.Selector = section.Key(fmt.Sprintf("%d__selector", i)).MustString("")
				dkimSign.PrivKey_base64 = section.Key(fmt.Sprintf("%d__privkey_base64", i)).MustString("")
			}
			t.DkimSigns = append(t.DkimSigns, dkimSign)
		}
	}
	// dkim sign - only same domain
	{
		t.DkimSigns_onlySameDomain = make([]*DKIMSign, 0, 10)
		pt_section := cfg.Section(t.preName + "dkim_signing__only_same_domain")
		for i := 0; ; i++ {
			if pt_section.HasKey(fmt.Sprintf("%d__is_use", i)) == false {
				break
			}
			dkimSigning := &DKIMSign{}
			{
				isUse := pt_section.Key(fmt.Sprintf("%d__is_use", i)).MustString("false")
				if isUse == "true" {
					dkimSigning.IsUse = true
				} else {
					dkimSigning.IsUse = false
				}
				dkimSigning.Domain = pt_section.Key(fmt.Sprintf("%d__domain", i)).MustString("")
				dkimSigning.Selector = pt_section.Key(fmt.Sprintf("%d__selector", i)).MustString("")
				dkimSigning.PrivKey_base64 = pt_section.Key(fmt.Sprintf("%d__privkey_base64", i)).MustString("")
			}
			t.DkimSigns_onlySameDomain = append(t.DkimSigns_onlySameDomain, dkimSigning)
		}
	}

	// save config
	err = cfg.SaveTo(t.path)
	if err != nil {
		return errors.New("[ " + t.path + " ] is not found / Create default setting file.")
	}

	return nil
}
