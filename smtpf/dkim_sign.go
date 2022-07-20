package smtpf

import (
	"encoding/base64"

	"github.com/toorop/go-dkim"
)

func New_DKIM() *T_DKIM_sign {
	pt_dkim_sign := &T_DKIM_sign{}
	return pt_dkim_sign
}

//---------------------------------------------------------------------------------------------------//
// dkim sign

type T_DKIM_sign struct {
	Opt_DKIMSign []*Opt_DKIMSign
}

func (t *T_DKIM_sign) Init() {
	if t.Opt_DKIMSign == nil {
		t.Opt_DKIMSign = make([]*Opt_DKIMSign, 0, 10)
	}
}

func (t *T_DKIM_sign) Add_Option(
	_isOnlySameDomain bool,
	_domain string,
	_selector string,
	_privkey_base64 string,
	_headers []string,
	_canonicalization string,
	_sigExpireIn uint64,
) {
	t.Init()
	pt_dkim_sign__one := &Opt_DKIMSign{}
	pt_dkim_sign__one.Init(_isOnlySameDomain, _domain, _selector, _privkey_base64, _headers, _canonicalization, _sigExpireIn)
	t.Opt_DKIMSign = append(t.Opt_DKIMSign, pt_dkim_sign__one)
}

func (t *T_DKIM_sign) Sign(_s_addr_from string, _bt_email *[]byte) (err error) {
	s_addr_from__domain, err := ParseDomain(_s_addr_from)
	if err != nil {
		return err
	}
	for _, pt_dkim_sign_one := range t.Opt_DKIMSign {
		_, err = pt_dkim_sign_one.sign(s_addr_from__domain, _bt_email)
		if err != nil {
			return err
		}
	}
	return nil
}

type Opt_DKIMSign struct {
	isSignOnlySameDomain bool
	dkim.SigOptions
}

func (t *Opt_DKIMSign) Init(
	_isSignOnlySameDomain bool,
	_domain string,
	_selector string,
	_privkey_base64 string,
	_headers []string,
	_canonicalization string,
	_sigExpireIn uint64,
) {
	// base64 decode
	privkey, _ := base64.RawStdEncoding.DecodeString(_privkey_base64)

	t.isSignOnlySameDomain = _isSignOnlySameDomain
	t.SigOptions = dkim.NewSigOptions()
	{
		t.PrivateKey = privkey
		t.Domain = _domain
		t.Selector = _selector
		t.AddSignatureTimestamp = true
		t.Headers = _headers
		t.Canonicalization = _canonicalization
		t.SignatureExpireIn = _sigExpireIn
	}
}

func (t *Opt_DKIMSign) sign(_domain string, _email *[]byte) (isSigned bool, err error) {
	if t.isSignOnlySameDomain == true && t.Domain != _domain {
		return false, nil
	}
	err = dkim.Sign(_email, t.SigOptions)
	if err != nil {
		return false, err
	}
	return true, nil
}
