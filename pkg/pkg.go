package pkg

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type client struct {
	AppCode      string
	TemplateCode string
}

func VerificationCode(length int) string {
	var code string
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < length; i++ {
		code += strconv.Itoa(rand.Intn(10))
	}
	fmt.Println(code)
	return code
}
func NewClient(appCode, templateCode string) *client {
	return &client{
		AppCode:      appCode,
		TemplateCode: templateCode,
	}
}
func (c *client) Send(mobile string, length int) error {
	code := VerificationCode(length)
	url := fmt.Sprintf("http://smssend.shumaidata.com/sms/send?receive=%s&tag=%s&templateId=%s",
		mobile, code, c.TemplateCode)
	fmt.Println(url)
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "APPCODE"+" "+c.AppCode)
	resp, _ := client.Do(req)
	fmt.Println(resp.Status, "------------------")
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("respose is %v\n", string(body))
	return nil
}
