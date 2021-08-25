package service

import (
	ran "crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"

	"time"

	"github.com/shuaiqidechuan/SMS/smservice/model"
)

var numbers = []byte("012345678998765431234567890987654321")

type SMSVerify interface {
	OnVerifySucceed(targetID, mobile string)
	OnVerifyFailed(targetID, mobile string)
}

type Config struct {
	Host           string
	Appcode        string
	Digits         int
	ResendInterval int
	OnCheck        SMSVerify
}

type Controller struct {
	DB   *sql.DB
	Conf Config
}

func NewController(db *sql.DB, Conf *Config) *Controller {
	sm := &Controller{
		DB: db,
		Conf: Config{
			Host:           Conf.Host,
			Appcode:        Conf.Appcode,
			Digits:         Conf.Digits,
			ResendInterval: Conf.ResendInterval,
			OnCheck:        Conf.OnCheck,
		},
	}

	return sm
}

type SendSmsReply struct {
	Msg     string `json:"msg"`
	Success bool   `json:"success"`
	Code    int    `json:"code"`
	Data    struct {
		TaskID  string `json:"taskId"`
		OrderNo string `json:"orderNo"`
	} `json:"data"`
}

type SMS struct {
	Mobile string
	Date   int64
	Code   string
	Sign   string
}

func newSms() *SMS {
	sms := &SMS{}
	return sms
}

func (sms *SMS) prepare(mobile, sign string, digits int) {
	sms.Mobile = mobile
	sms.Date = time.Now().Unix()
	sms.Code = Code(digits)
	sms.Sign = sign
}

func Code(size int) string {
	data := make([]byte, size)
	out := make([]byte, size)

	buffer := len(numbers)
	_, err := ran.Read(data)
	if err != nil {
		panic(err)
	}

	for id, key := range data {
		x := byte(int(key) % buffer)
		out[id] = numbers[x]
	}

	return string(out)
}

func (sms *SMS) checkvalid(db *sql.DB, conf *Config) error {
	unixtime := sms.getDate(db)

	if unixtime > 0 && sms.Date-unixtime < int64(conf.ResendInterval) {
		return errors.New("not allowed to send twice in a short time")
	}

	if err := VailMobile(sms.Mobile); err != nil {
		return errors.New("phone number is not in line with the rules")
	}

	return nil
}

func (sms *SMS) getDate(db *sql.DB) int64 {
	unixtime, _ := model.GetDate(db, sms.Sign)
	return unixtime
}

func VailMobile(mobile string) error {
	if len(mobile) < 11 {
		return errors.New("[mobile] param wrong")
	}

	reg, err := regexp.Compile("^1[3-8][0-9]{9}$")
	if err != nil {
		panic("regexp error")
	}

	if !reg.MatchString(mobile) {
		return errors.New("phone number [mobile] incorrect format")
	}

	return nil
}

func (sms *SMS) save(db *sql.DB) error {
	err := model.Insert(db, sms.Mobile, sms.Date, sms.Code, sms.Sign)

	return err
}

func (sms *SMS) sendmsg(conf *Config) error {
	host := conf.Host

	url := fmt.Sprintf(host, sms.Mobile, sms.Code, "MD8B7116BC")

	client := &http.Client{}

	request, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	request.Header.Add("Authorization", "APPCODE "+conf.Appcode)
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return err
	}

	ssr := &SendSmsReply{}
	if err := json.Unmarshal(body, ssr); err != nil {
		return err
	}

	if ssr.Code != 200 {
		err := strconv.Itoa(ssr.Code)
		return errors.New(err)
	}

	return nil
}

func Send(mobile, sign string, conf *Config, db *sql.DB) error {
	sms := newSms()
	sms.prepare(mobile, sign, conf.Digits)

	if err := sms.checkvalid(db, conf); err != nil {
		return err
	}

	if err := sms.save(db); err != nil {
		return err
	}

	err := sms.sendmsg(conf)

	return err
}

func Check(code, sign string, conf *Config, db *sql.DB) error {
	sms := newSms()
	sms.Date = time.Now().Unix()
	sms.Code = code
	sms.Sign = sign

	getcode, err := sms.getCode(db)
	if err != nil {
		return errors.New("sign error")
	}

	if sms.Code == getcode {
		sms.delete(sms.Sign, db)
		return nil
	}

	return errors.New("code error")
}

func (sms *SMS) getCode(db *sql.DB) (string, error) {
	code, err := model.GetCode(db, sms.Sign)
	return code, err
}

func (sms *SMS) delete(sign string, db *sql.DB) { model.Delete(db, sign) }

func UID() string {
	data := make([]byte, 16)

	_, err := ran.Read(data)
	if err != nil {
		panic(err)
	}

	uuid := fmt.Sprintf("%X-%X-%X-%X-%X", data[0:4], data[4:6], data[6:8], data[8:10], data[10:])
	return uuid
}
