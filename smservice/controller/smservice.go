package controller

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shuaiqidechuan/SMS/smservice/model"
	"github.com/shuaiqidechuan/SMS/smservice/service"
)

type SMController struct {
	ser *service.Controller
}

func New(db *sql.DB, conf *service.Config) *SMController {
	return &SMController{
		ser: service.NewController(db, conf),
	}
}

func (s *SMController) RegisterRouter(r gin.IRouter) {
	if r == nil {
		log.Fatal("[InitRouter]: server is nil")
	}

	err := model.CreateTable(s.ser.DB)
	if err != nil {
		log.Fatal(err)
	}

	r.POST("/send", s.Send)
	r.POST("/check", s.Check)
}

func (s *SMController) Send(c *gin.Context) {
	var (
		req struct {
			Mobile string `json:"mobile"`
			Sign   string `json:"sign"`
		}
	)

	err := c.ShouldBind(&req)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest})
		return
	}

	err = service.Send(req.Mobile, req.Sign, &s.ser.Conf, s.ser.DB)
	fmt.Println(err)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusBadGateway, gin.H{"status": http.StatusBadGateway})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK})
}

func (s *SMController) Check(c *gin.Context) {
	var (
		req struct {
			Code string `json:"code"`
			Sign string `json:"sign"`
		}

		resp struct {
			sign   string
			mobile string
		}
	)

	err := c.ShouldBind(&req)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest})
		return
	}

	resp.sign = req.Sign
	resp.mobile, _ = model.GetMobile(s.ser.DB, resp.sign)

	err = service.Check(req.Code, req.Sign, &s.ser.Conf, s.ser.DB)
	if err != nil {
		s.ser.Conf.OnCheck.OnVerifyFailed(resp.sign, resp.mobile)

		c.Error(err)
		c.JSON(http.StatusBadGateway, gin.H{"status": http.StatusBadGateway})
		return
	}

	s.ser.Conf.OnCheck.OnVerifySucceed(resp.sign, resp.mobile)

	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK})
}
