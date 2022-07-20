package smtpf

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"math/big"
	"os"

	timef "github.com/gokch/time"

	"gopkg.in/gomail.v2"
)

func NewEmailjob(
	_from string, _to []string,
	_title string, _contents string,
	_filePath_contents string,
	_filePath_attach []string,
	_unsubscribeURL string,
) *EmailJob {
	emailJob := &EmailJob{}
	emailJob.Init(_from, _to, _title, _contents, _filePath_contents, _filePath_attach, _unsubscribeURL)
	return emailJob
}

//---------------------------------------------------------------------------------------------------//
// email job

type EmailJob struct {
	from              string
	to                []string
	title             string
	contents          string
	filePath_contents string
	filePath_attach   []string

	unsubscribe_url string
}

func (t *EmailJob) MarshalJSON() ([]byte, error) {
	return json.Marshal(t)
}

func (t *EmailJob) UnMarshalJSON(_bt []byte) error {
	return json.Unmarshal(_bt, t)
}

func (t *EmailJob) Init(
	_from string,
	_to []string,
	_title string,
	_contents string,
	_filePath_contents string,
	_filePath_attach []string,
	_unsubscribeURL string,
) error {
	t.from = _from
	t.to = _to
	t.title = _title
	t.contents = _contents
	t.filePath_contents = _filePath_contents
	t.filePath_attach = _filePath_attach
	t.unsubscribe_url = _unsubscribeURL
	return nil
}

func (t *EmailJob) conv__gomail() (msgGomail *gomail.Message, err error) {
	msgId, err := t.GenerateMessageID(t.from)
	if err != nil {
		return nil, err
	}
	if len(t.to) == 0 {
		return nil, fmt.Errorf("invalid address to | %v", t.to)
	}

	msgGomail = gomail.NewMessage()
	{
		msgGomail.SetHeader("From", t.from)
		msgGomail.SetHeader("To", t.to...)
		msgGomail.SetHeader("Subject", t.title)
		msgGomail.SetHeader("Message-ID", msgId)
		msgGomail.SetHeader("List-Unsubscribe", t.unsubscribe_url)
		msgGomail.SetDateHeader("Date", timef.UTC_now())
		msgGomail.SetBody("text/html", t.contents)
		if len(t.filePath_contents) > 0 {
			msgGomail.AddAlternative("text/html", t.filePath_contents)
		}
		for _, filePath_attach := range t.filePath_attach {
			msgGomail.Attach(filePath_attach)
		}
	}
	return msgGomail, nil
}

func (t *EmailJob) Conv__bt() (msg []byte, err error) {
	gomail, err := t.conv__gomail()
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	_, err = gomail.WriteTo(buf)
	if err != nil {
		return nil, err
	}
	msg = buf.Bytes()
	return msg, nil
}

func (t *EmailJob) GenerateMessageID(_from string) (string, error) {
	dtnNow := timef.UTC_now().UnixNano()
	pid := os.Getpid()
	rand, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return "", err
	}

	domain, err := ParseDomain(_from)
	if err != nil {
		return "", err
	}
	msgid := fmt.Sprintf("<%d.%d.%d@%s>", dtnNow, pid, rand, domain)
	return msgid, nil
}

func (t *EmailJob) ParseTemplate(_htmlPath string, _parse interface{}) error {
	template, err := template.ParseFiles(_htmlPath)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	if err = template.Execute(buf, _parse); err != nil {
		return err
	}
	t.filePath_contents = buf.String()
	return nil
}
