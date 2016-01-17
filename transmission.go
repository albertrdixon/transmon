package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/albertrdixon/gearbox/logger"
	"github.com/longnguyen11288/go-transmission/client"
)

type transmission struct {
	client.ApiClient
}

type request struct {
	Method string  `json:"method"`
	Tag    int     `json:"tag,omitempty"`
	Args   argList `json:"arguments,omitempty"`
}

func (r *request) String() string {
	return fmt.Sprintf("Request(method=%q, tag=%d, args=%v)", r.Method, r.Tag, r.Args)
}

type response struct {
	Result string  `json:"result"`
	Tag    int     `json:"tag,omitempty"`
	Args   argList `json:"arguments,omitempty"`
}

type argList map[string]*json.RawMessage

func (a argList) String() string {
	list := make([]string, 0, len(a))
	for k, v := range a {
		list = append(list, fmt.Sprintf("%q: %v", k, v))
	}
	return fmt.Sprintf("{%s}", strings.Join(list, ", "))
}

func newTransmissionClient(url, user, pass string) *transmission {
	return &transmission{ApiClient: client.NewClient(url, user, pass)}
}

func newRequest(method string, args ...interface{}) *request {
	c := &request{
		Method: method,
		Tag:    tag(),
		Args:   make(map[string]*json.RawMessage),
	}
	if !even(len(args)) {
		return c
	}
	for i := range args {
		s, ok := args[i].(string)
		if ok && even(i) {
			if out, er := json.Marshal(args[i+1]); er == nil {
				jr := json.RawMessage(out)
				c.Args[s] = &jr
			}
		}
	}
	return c
}

func (t *transmission) updatePort(port int) error {
	req := newRequest("session-set", "peer-port", port, "port-forwarding-enabled", true)
	tag := req.Tag
	logger.Debugf("Marshalling %v", req)
	body, er := json.Marshal(req)
	if er != nil {
		return er
	}
	logger.Debugf("Post body: %v", string(body))
	out, er := t.Post(string(body))
	if er != nil {
		return er
	}

	response := new(response)
	if er := json.Unmarshal(out, response); er != nil {
		return er
	}
	if response.Tag != tag {
		return errors.New("Request and response tags do not match")
	}
	if response.Result != "success" {
		return errors.New(response.Result)
	}
	return nil
}
