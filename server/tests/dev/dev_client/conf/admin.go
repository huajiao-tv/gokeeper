package main

import (
	"encoding/json"
	"errors"
	"github.com/huajiao-tv/gokeeper/model"
	"io/ioutil"
	"net/http"
	"net/url"
)

const filePath = "data/config/testDomain/test.conf"
const fileName = "test.conf"
const domain = "testDomain"

type BaseResponse struct {
	Code    int    `json:"error_code"`
	Message string `json:"error"`
}

func initConfig() error{
	file,err :=ioutil.ReadFile(filePath)
	if err != nil{
		return errors.New("read config file "+filePath+" failed:"+err.Error())
	}
	args := url.Values{
		"domain": []string{domain},
		"file":   []string{fileName},
		"conf":   []string{string(file)},
		"note":   []string{"test add file"},
	}
	res,err :=http.PostForm("http://"+*adminAddr+"/add/file",args)
	if err != nil{
		return err
	}
	body,_ := ioutil.ReadAll(res.Body)
	r := &BaseResponse{}
	if err = json.Unmarshal(body,r);err !=nil{
		return err
	}
	if r.Code!=0{
		return errors.New("add file failed,message:"+r.Message)
	}
	return nil
}

func changeConfig(key string,typ string,value string) error{
	op:=&model.Operate{
		Opcode:  model.OpcodeUpdate,
		Domain:  domain,
		File:    fileName,
		Section: "DEFAULT",
		Key:     key,
		Type:    typ,
		Value:   value,
		Note:    "test change config",
	}
	data, err := json.Marshal([]*model.Operate{op})
	if err != nil {
		return err
	}
	args := url.Values{
		"domain":   []string{domain},
		"operates": []string{string(data)},
		"note":     []string{"test change config"},
	}
	res,err := http.PostForm("http://"+*adminAddr+"/conf/manage",args)
	if err != nil{
		return err
	}
	body,_ := ioutil.ReadAll(res.Body)
	r := &BaseResponse{}
	if err = json.Unmarshal(body,r);err !=nil{
		return err
	}
	if r.Code!=0{
		return errors.New("add file failed,message:"+r.Message)
	}
	return nil
}
