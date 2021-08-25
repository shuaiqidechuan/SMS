package main

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/shuaiqidechuan/SMS/smservice/controller"
	"github.com/shuaiqidechuan/SMS/smservice/service"
)

type funcv struct{}

func (v funcv) OnVerifySucceed(targetID, mobile string) {}
func (v funcv) OnVerifyFailed(targetID, mobile string)  {}

func main() {
	var v funcv

	router := gin.Default()

	dbConn, err := sql.Open("mysql", "root:123456@tcp(localhost:3306)/project?parseTime=true")
	if err != nil {
		panic(err)
	}
	con := &service.Config{
		Host:           "http://smssend.shumaidata.com/sms/send?receive=%s&tag=%s&templateId=%s",
		Appcode:        "02cc504d218d4ad4a61c6b68ac283a49",
		Digits:         6,
		ResendInterval: 60,
		OnCheck:        v,
	}
	smserviceCon := controller.New(dbConn, con)
	smserviceCon.RegisterRouter(router.Group("/api/v1/message"))

	log.Fatal(router.Run(":8000"))
}
