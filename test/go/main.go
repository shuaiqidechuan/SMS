package main

import (
	"log"

	"github.com/shuaiqidechuan/SMS/test/pkg"
)

func main() {
	client := pkg.NewClient("02cc504d218d4ad4a61c6b68ac283a49", "MD8B7116BC")
	//
	err := client.Send("15633785802", 6) //电话， 验证码个数
	if err != nil {
		log.Panicln(err)
	}
}
