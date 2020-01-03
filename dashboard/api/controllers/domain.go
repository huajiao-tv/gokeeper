package controllers

import (
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huajiao-tv/gokeeper/dashboard/api/models"
)

const (
	MaxSize = 10 << 20
)

func addDomain(c *gin.Context) {
	domain := c.Param("domain")
	domains, err := models.KeeperAdminClient.QueryClusters()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Error(err))
		return
	}
	for _, d := range domains {
		if d.Name == domain {
			c.JSON(http.StatusBadRequest, BadRequest("domain already exists"))

			return
		}
	}

	file, fileHeader, err := c.Request.FormFile("files")
	if err != nil {
		c.JSON(http.StatusBadRequest, BadRequest("read file failed:"+err.Error()))
		return
	}
	if file == nil || fileHeader == nil {
		c.JSON(http.StatusBadRequest, BadRequest("read file empty"))
		return
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, BadRequest("read file failed:"+err.Error()))
	}
	fileName := fileHeader.Filename
	if err := models.KeeperAdminClient.AddFile(domain, fileName, string(data), "add domain"); err != nil {
		c.JSON(http.StatusBadRequest, BadRequest("add file failed:"+err.Error()))
		return
	}
	c.JSON(http.StatusOK, NewResponse(""))
	return
}

func getDomains(c *gin.Context) {
	var list []map[string]string
	dup := make(map[string]string)

	domains, err := models.KeeperAdminClient.QueryClusters()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Error(err))
		return
	}
	for _, domain := range domains {
		list = append(list, map[string]string{
			"domain": domain.Name,
			"status": dup[domain.Name],
		})
		dup[domain.Name] = "DUP"
	}
	c.JSON(http.StatusOK, NewResponse(list))
}
