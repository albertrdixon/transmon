package transmission

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"time"

	"github.com/albertrdixon/gearbox/logger"
	"github.com/albertrdixon/gearbox/util"
	"github.com/cenkalti/backoff"
	"github.com/tubbebubbe/transmission"
)

func (r *RawClient) UpdatePort(port int) error {
	req := newRequest("session-set", "peer-port", port, "port-forwarding-enabled", true)
	tag := req.Tag
	logger.Debugf("Marshalling %v", req)
	body, er := json.Marshal(req)
	if er != nil {
		return er
	}
	out, er := r.Post(string(body))
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

func UpdateSettings(path, ip string, port int) error {
	data, er := ioutil.ReadFile(path)
	if er != nil {
		return er
	}

	s := new(settings)
	if er := json.Unmarshal(data, s); er != nil {
		return er
	}

	s.Addr = ip
	s.Port = port
	s.Forward = true

	data, er = json.Marshal(s)
	if er != nil {
		return er
	}

	info, _ := os.Stat(path)
	return ioutil.WriteFile(path, data, info.Mode().Perm())
}

func (c *Client) CleanTorrents() error {
	torrents, er := c.GetTorrents()
	if er != nil {
		return er
	}

	torrents.SortByID(false)
	for _, t := range torrents {
		id := util.Hashf(md5.New(), t.ID, t.Name)
		status := &torrentStatus{Torrent: t, id: id, failures: 0}
		status.setFailures()

		if st, ok := seen[id]; ok {
			status.failures = status.failures + st.failures
			if !updated(st.Torrent, status.Torrent) {
				status.failures++
			}
		} else {
			seen[id] = status
		}
	}

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 15 * time.Second
	remove := make([]*torrentStatus, 0, 1)
	for _, t := range seen {
		if t.failed() {
			b.Reset()
			logger.Infof("[Torrent %d: %q] Removing", t.ID, t.Name)
			er := backoff.RetryNotify(delTorrent(c, t.Torrent), b, func(e error, w time.Duration) {
				logger.Errorf("[Torrent %d: %q] Failed to remove (retry in %v): %v", t.ID, t.Name, w, e)
			})
			if er == nil {
				remove = append(remove, t)
			} else {
				logger.Errorf("[Torrent %d: %q] Failed to remove, will retry next cycle", t.ID, t.Name)
			}
		}
	}

	for i := range remove {
		delete(seen, remove[i].id)
	}
	return nil
}

func (s *torrentStatus) setFailures() {
	switch {
	case s.Error != 0:
		logger.Warnf("[Torrent %d: %q] Error: %s", s.ID, s.Name, s.ErrorString)
		s.failures++
	case s.IsFinished:
		logger.Infof("[Torrent %d: %q] Finished", s.ID, s.Name)
		s.failures = 3
	}
}

func (s *torrentStatus) failed() bool {
	return s.failures >= 3
}

func delTorrent(c *Client, t transmission.Torrent) backoff.Operation {
	return func() (e error) {
		del, er := transmission.NewDelCmd(t.ID, true)
		if er != nil {
			return er
		}
		_, e = c.ExecuteCommand(del)
		return
	}
}

func updated(a, b transmission.Torrent) bool {
	return a.PercentDone != b.PercentDone || a.UploadRatio != b.UploadRatio
}
