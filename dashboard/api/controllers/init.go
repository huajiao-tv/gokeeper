package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Init(r *gin.Engine) error {
	keeperGroup := r.Group("/keeper")
	{
		keeperGroup.GET("/domains", getDomains)
		keeperGroup.PUT("/:domain", addDomain)
	}

	configGroup := r.Group("/config")
	{
		configGroup.GET("/:domain", getDomainConfig)
		configGroup.POST("/:domain", updateDomainConfig)
	}

	serviceGroup := r.Group("/discovery")
	{
		serviceGroup.GET("/services", getServices)
		serviceGroup.GET("/get/:service", getService)
	}

	return nil
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"error"`
	Data    interface{} `json:"data"`
}

func (resp *Response) String() string {
	v, _ := json.Marshal(resp)
	return string(v)
}

func NewResponse(data interface{}) *Response {
	return &Response{
		Data: data,
	}
}

func Error(err error) *Response {
	return &Response{
		Code:    http.StatusInternalServerError,
		Message: err.Error(),
	}
}

func BadRequest(msg string) *Response {
	return &Response{
		Code:    http.StatusBadRequest,
		Message: msg,
	}
}
