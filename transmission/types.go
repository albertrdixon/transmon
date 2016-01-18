package transmission

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tubbebubbe/transmission"
)

var seen map[string]*torrentStatus

type settings struct {
	Stuff   map[string]*json.RawMessage `json:",inline"`
	Addr    string                      `json:"bind-address-ipv4"`
	Port    int                         `json:"peer-port"`
	Forward bool                        `json:"port-forwarding-enabled"`
}

type RawClient struct {
	transmission.ApiClient
}

type Client struct {
	transmission.TransmissionClient
}

type torrentStatus struct {
	transmission.Torrent
	id       string
	failures int
}

type request struct {
	Method string  `json:"method"`
	Tag    int     `json:"tag,omitempty"`
	Args   argList `json:"arguments,omitempty"`
}

type response struct {
	Result string  `json:"result"`
	Tag    int     `json:"tag,omitempty"`
	Args   argList `json:"arguments,omitempty"`
}

type argList map[string]*json.RawMessage

func (r *request) String() string {
	return fmt.Sprintf("Request(method=%q, tag=%d, args=%v)", r.Method, r.Tag, r.Args)
}

func (a argList) String() string {
	list := make([]string, 0, len(a))
	for k, v := range a {
		list = append(list, fmt.Sprintf("%q: %v", k, string([]byte(*v))))
	}
	return fmt.Sprintf("{%s}", strings.Join(list, ", "))
}

func NewRawClient(url, user, pass string) *RawClient {
	return &RawClient{ApiClient: transmission.NewClient(url, user, pass)}
}

func NewClient(url, user, pass string) *Client {
	return &Client{TransmissionClient: transmission.New(url, user, pass)}
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

func init() {
	seen = make(map[string]*torrentStatus)
}
