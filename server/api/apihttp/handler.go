package apihttp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	"reflect"
	"strings"
	"time"

	"github.com/huajiao-tv/gokeeper/server/logger"
	"github.com/huajiao-tv/gokeeper/server/setting"
)

const (
	API_REQUEST_PARAMS_ERROR = 400
	API_INTERNAL_ERROR       = 500
)

//
var (
	httpReadTimeout  = time.Duration(10) * time.Second
	httpWriteTimeout = time.Duration(10) * time.Second
)

func StartHttpServer() error {
	port := setting.AdminListen

	logger.Logex.Trace("admin listen", port)
	http.HandleFunc("/", ServiceHandler)
	server := &http.Server{
		Addr:         port,
		ReadTimeout:  httpReadTimeout,
		WriteTimeout: httpWriteTimeout,
	}
	if err := server.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

// HelloWorldAction say i'm alive
func (s *ServiceController) HelloWorldAction() {
	s.response.Write([]byte("hello world"))
}

// Resp define response data
type Resp struct {
	ErrorCode int         `json:"error_code"`
	Error     string      `json:"error"`
	Data      interface{} `json:"data"`
}

// ServiceController ...
type ServiceController struct {
	response http.ResponseWriter
	request  *http.Request
}

// ServiceHandler provide routing parser
func ServiceHandler(w http.ResponseWriter, r *http.Request) {
	remoteip := clientIp(r)
	if !setting.TestMode {
		params := getGuidParams(r)
		guid := query(r, "guid")
		if !CheckServerGUID(params, guid) {
			fmt.Fprintf(w, "access deny")
			return
		}
	}
	pathInfo := strings.Trim(r.URL.Path, "/")
	if pathInfo == "favicon.ico" {
		return
	}
	parts := strings.Split(pathInfo, "/")
	for k, v := range parts {
		parts[k] = strings.Title(v)
	}
	action := strings.Join(parts, "")
	service := &ServiceController{response: w, request: r}
	controller := reflect.ValueOf(service)

	logger.Logex.Trace("[access] ServerHandler", remoteip, r.URL.Path, fmt.Sprintf("%#v", r.URL.Query()))

	method := controller.MethodByName(action + "Action")
	if !method.IsValid() {
		method = controller.MethodByName("HelloWorldAction")
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")

	method.Call([]reflect.Value{})
}

func (s *ServiceController) renderJSON(result Resp) {
	r, err := json.Marshal(result)
	if err != nil {
		result = Resp{
			ErrorCode: -1,
			Error:     err.Error(),
		}
		r, _ = json.Marshal(result)
	}
	s.response.Write(r)
	return
}

func (s *ServiceController) required(args ...string) bool {
	for _, v := range args {
		if s.query(v) == "" {
			s.renderJSON(Resp{ErrorCode: 1, Error: fmt.Sprintf("%s is required", v)})
			return false
		}
	}
	return true
}

func (s *ServiceController) query(q string) string {
	return query(s.request, q)
}

func query(r *http.Request, q string) string {
	var v string
	if v = r.URL.Query().Get(q); v == "" {
		v = r.FormValue(q)
	}
	return strings.Trim(v, " ")
}

func (s *ServiceController) readBody() (string, error) {
	body, err := ioutil.ReadAll(s.request.Body)
	return string(body), err
}

func clientIp(r *http.Request) string {
	ip := r.Header.Get("X-Real-Ip")
	if ip == "" {
		s := strings.Split(r.RemoteAddr, ":")
		ip = s[0]
	}
	return ip
}

func getGuidParams(r *http.Request) *GuidParams {
	return &GuidParams{
		Partner: query(r, "partner"),
		Rand:    query(r, "rand"),
		Time:    query(r, "time"),
	}
}
