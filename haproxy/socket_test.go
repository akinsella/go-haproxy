package haproxy_test

import (
  "github.com/macb/go-haproxy/haproxy"
  "testing"
)

func TestShowInfo(t *testing.T) {

  tl := ListenForAndRespondWith(t, "show info\n", "resp")

  resp, err := haproxy.Socket(tl.Path).ShowInfo()
  if err != nil {
    t.Error(err)
  }

  if string(resp) != tl.Response {
    t.Errorf("Expected: %s - Got: %s", string(resp), tl.Response)
  }
}

func TestShowMap(t *testing.T) {

  tl := ListenForAndRespondWith(t, "show map\n", "resp")

  resp, err := haproxy.Socket(tl.Path).ShowMap()
  if err != nil {
    t.Error(err)
  }

  if string(resp) != tl.Response {
    t.Errorf("Expected: %s - Got: %s", string(resp), tl.Response)
  }
}

func TestShowStat(t *testing.T) {

  tl := ListenForAndRespondWith(t, "show stat\n", "resp")

  resp, err := haproxy.Socket(tl.Path).ShowStat()
  if err != nil {
    t.Error(err)
  }

  if string(resp) != tl.Response {
    t.Errorf("Expected: %s - Got: %s", string(resp), tl.Response)
  }
}
