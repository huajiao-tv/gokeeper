package controllers

import (
	"bytes"
	"io"
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
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxSize)
	r, err := c.Request.MultipartReader()
	if err != nil {
		c.JSON(http.StatusBadRequest, Error(err))
		return
	}

	for {
		part, err := r.NextPart()
		if err == io.EOF {
			break
		}
		buf := bytes.Buffer{}
		if _, err = buf.ReadFrom(part); err != nil {
			c.JSON(http.StatusInternalServerError, Error(err))
			return
		}
		ops, err := models.ParseConfigFile(domain, part.FileName(), buf.String())
		if err != nil {
			c.JSON(http.StatusInternalServerError, Error(err))
			return
		}
		if err := models.KeeperAdminClient.ManageConfig(domain, ops, "create domain with files"); err != nil {
			c.JSON(http.StatusInternalServerError, Error(err))
			return
		}
	}

	c.JSON(http.StatusOK, NewResponse(""))
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
