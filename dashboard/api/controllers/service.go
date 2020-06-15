package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huajiao-tv/gokeeper/dashboard/api/models"
)

func getServices(c *gin.Context) {
	services, err := models.KeeperAdminClient.GetServices()
	if err != nil {
		log.Println("get services error:", err)
		c.JSON(http.StatusBadRequest, Error(err))
		return
	}
	c.JSON(http.StatusOK, services)
	return
}

func getService(c *gin.Context) {
	serviceName := c.Param("service")
	service, err := models.KeeperAdminClient.GetService(serviceName)
	if err != nil {
		log.Println("get service error:", err)
		c.JSON(http.StatusBadRequest, Error(err))
		return
	}
	c.JSON(http.StatusOK, service)
	return
}
