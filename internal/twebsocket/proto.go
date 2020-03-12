package twebsocket

import "github.com/mitchellh/mapstructure"

type Payload struct {
	Data interface{} `json:"data,omitempty"`
}

func (p *Payload) DecodeData(data interface{}) error {
	return mapstructure.Decode(p.Data, data)
}

func (p *Payload) EncodeData(data interface{}) error {
	p.Data = data
	return nil
}
type Request interface {
	Command() string
	Sequence() int64
	DecodeData(data interface{}) error
	Client() Client
}

type Response interface {
	EncodeData(data interface{}, code int32, msg string)
}

type RequestData struct {
	Cmd   string      `json:"cmd"`
	Seq   int64       `json:"seq"`
	Immed bool        `json:"immed,omitempty"`
	Payload
}

type ResponseData struct {
	Cmd  string      `json:"cmd"`
	Seq  int64       `json:"seq"`
	Code int32       `json:"code"`
	Msg  string      `json:"msg,omitempty"`
	Payload
}

type request struct {
	data *RequestData
	cli  *client
}

func (req *request) Client() Client {
	return req.cli
}

func (req *request) Command() string {
	return req.data.Cmd
}

func (req *request) Sequence() int64 {
	return req.data.Seq
}

func (req *request) DecodeData(data interface{}) error {
	return mapstructure.Decode(req.data.Data, data)
}

type response struct {
	data *ResponseData
}

func (rsp *response) EncodeData(data interface{}, code int32, msg string) {
	rsp.data.Code = code
	rsp.data.Msg = msg
	rsp.data.Data = data
}
