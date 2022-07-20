package smtpf

import (
	"fmt"
	"math/rand"
	"net"
)

func MXRecord_SendToMostPriortyRecord(_to string, _fn_cb_sendMail func(_hostName_mostPriorty string) (isSend bool, err error),
) error {
	pt_mx_records := &MXRecords{}
	err := pt_mx_records.Set(_to)
	if err != nil {
		return err
	}
	return pt_mx_records.SendToMostPriortyRecord(_fn_cb_sendMail)
}

//---------------------------------------------------------------------------------------------------//
// mx records

type MXRecords struct {
	Mxs []*net.MX
}

func (t *MXRecords) Init() {
	t.Mxs = make([]*net.MX, 0, 10)
}

func (t *MXRecords) Set(_to string) error {
	if t.Mxs == nil {
		t.Init()
	}

	domain, err := ParseDomain(_to)
	if err != nil {
		return err
	}
	t.Mxs, err = net.LookupMX(domain)
	if err != nil {
		return err
	}

	return nil
}

func (t *MXRecords) SendToMostPriortyRecord(_fn_cb_sendMail func(_host_mostPriorty string) (isSend bool, err error)) error {
	if len(t.Mxs) == 0 {
		return fmt.Errorf("mx records is empty")
	}

	var posStart, posEnd int
	minPref := t.Mxs[0].Pref
	lenMxs := len(t.Mxs)
	// 가중치가 가장 높은 mx record 에 발송 ( 가중치가 높다 = mx record 의 Pref 값이 낮다 )
	// 발송이 실패하면 그 다음으로 높은 가중치를 가진 mx record 에 발송
	// 가장 높은 가중치를 가진 mx record가 동시에 여러 개 있다면 그 중에서 랜덤 값을 얻어 그 위치부터 한 바퀴 순회
	for i := 0; i < lenMxs; i++ {
		if i == lenMxs-1 || t.Mxs[i+1].Pref > minPref {
			posEnd = i + 1
			// 1. rand 값 [n_pos_start, n_pos_end] 산출
			rand_betweenStartEnd := rand.Intn(posEnd - posStart)
			rand_betweenStartEnd += posStart

			// 2. [rand, end) 순회
			for j := rand_betweenStartEnd; j < posEnd; j++ {
				host := t.Mxs[j].Host
				isSend, err := _fn_cb_sendMail(host)
				if err != nil { // 발송 취소
					return err
				} else if isSend == true { // 정상 발송
					return nil
				}
			}

			// 3. [start, rand) 순회
			for j := posStart; j < rand_betweenStartEnd; j++ {
				s_host := t.Mxs[j].Host
				isSend, err := _fn_cb_sendMail(s_host)
				if err != nil { // 발송 취소
					return err
				} else if isSend == true { // 정상 발송
					return nil
				}
			}

			// 4. 다음 가중치 순회를 위한 후처리
			posStart = i + 1
			if i != lenMxs-1 {
				minPref = t.Mxs[i+1].Pref
			}
		}
	}
	// 에러 처리 - 모든 mx 레코드에서 send 가 거절당한 경우
	return fmt.Errorf("failed send_mail | every mx records refuse to send mail")
}
