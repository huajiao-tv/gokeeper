package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huajiao-tv/gokeeper/dashboard/api/models"
	km "github.com/huajiao-tv/gokeeper/model"
)

func getDomainConfig(c *gin.Context) {
	domain := c.Param("domain")
	cs, err := models.KeeperAdminClient.QueryConfig(domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Error(err))
		return
	}
	c.JSON(http.StatusOK, NewResponse(cs))
}

func updateDomainConfig(c *gin.Context) {
	op := &km.Operate{}
	if err := c.BindJSON(&op); err != nil {
		c.JSON(http.StatusBadRequest, Error(err))
		return
	}
	if err := models.KeeperAdminClient.ManageConfig(op.Domain, []*km.Operate{op}, op.Note); err != nil {
		c.JSON(http.StatusInternalServerError, Error(err))
		return
	}
	c.JSON(http.StatusOK, NewResponse(""))
}
