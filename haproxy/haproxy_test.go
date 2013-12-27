package haproxy_test

import (
  "bytes"
  "encoding/csv"
  "github.com/macb/go-haproxy/haproxy"
  "io"
  "testing"
)

// TODO: Create stub socket to mock Haproxy responses.
// For now though, it assumes you have Haproxy up with a socket at /tmp/haproxy
var socketPath = "/tmp/haproxy"

func TestShowInfo(t *testing.T) {
  _, err := haproxy.Socket(socketPath).ShowInfo()
  if err != nil && err != io.EOF {
    t.Error(err)
  }
}

func TestShowInto(t *testing.T) {
  _, err := haproxy.Socket(socketPath).ShowMap()
  if err != nil && err != io.EOF {
    t.Error(err)
  }
}

func TestShowStat(t *testing.T) {
  resp, err := haproxy.Socket(socketPath).ShowStat()
  if err != nil && err != io.EOF {
    t.Error(err)
  }

  r := csv.NewReader(bytes.NewReader(resp))
  s, err := r.ReadAll()
  t.Logf("%#v", s[0][0][2:])
}

func TestGetLoad(t *testing.T) {
  load, err := haproxy.Haproxy{Socket: "/tmp/haproxy"}.GetLoad("elastic-ocean")
  if err != nil {
    t.Error(err)
  }
  for _, thing := range load {
    t.Log(*thing)
  }
}
