package haproxy_test

import (
  "fmt"
  "math/rand"
  "net"
  "os"
  "testing"
  "time"
)

func init() {
  rand.Seed(time.Now().UnixNano())
}

type testListener struct {
  Command  string
  Response string
  T        *testing.T
  Path     string
  net.Listener
}

func NewTestListener(t *testing.T) testListener {
  path := fmt.Sprintf("/tmp/test-socket-%d", rand.Int63())
  l, e := net.Listen("unix", path)
  if e != nil {
    panic(e)
  }
  return testListener{T: t, Path: path, Listener: l}
}

// Opens a
func ListenForAndRespondWith(t *testing.T, command string, response string) testListener {
  tl := NewTestListener(t)
  tl.Command = command
  tl.Response = response
  go tl.ListenAndRespond(command, response)
  return tl
}

func (t testListener) ListenAndRespond(command string, expected string) {
  c, err := t.Accept()
  if err != nil {
    t.T.Errorf("Accept error: %s", err)
  }
  b := make([]byte, len(command))
  if n, err := c.Read(b); n != len(command) || err != nil || string(b) != command {
    t.T.Errorf("Expected Command: %s - Got: %s", string(command), string(b))
  }

  c.Write([]byte(expected))
  c.Close()
  t.Close()
}

func (t testListener) Close() {
  go os.Remove(t.Path)
  t.Listener.Close()
}

func ListenForeverAndRespondWith(resp string) (net.Listener, string) {
  path := fmt.Sprintf("/tmp/test-socket-%d", rand.Int63())
  l, e := net.Listen("unix", path)
  if e != nil {
    panic(e)
  }
  go func() {
    for {
      c, _ := l.Accept()
      if c == nil {
        break
      }
      c.Read([]byte{})
      c.Write([]byte(resp))
      c.Close()
    }
  }()
  return l, path
}
