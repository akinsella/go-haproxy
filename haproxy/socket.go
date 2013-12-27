package haproxy

import (
  "io"
  "net"
)

type Socket string

// Convenience methods for common Haproxy socket queries.
func (s Socket) ShowInfo() ([]byte, error) {
  return s.Do("show info\n")
}

func (s Socket) ShowMap() ([]byte, error) {
  return s.Do("show map\n")
}

func (s Socket) ShowStat() ([]byte, error) {
  return s.Do("show stat\n")
}

// Genetic write and read socket method.
func (s Socket) Do(cmd string) ([]byte, error) {
  return s.writeAndRead(cmd)
}

// A wrapper for writing and reading from a socket (in this case a unix socket)
// Used to preserve the net.Conn used during the write to read the response.

func (s Socket) writeAndRead(cmd string) (resp []byte, err error) {
  c, err := net.Dial("unix", string(s))
  if err != nil {
    return nil, err
  }
  defer c.Close()

  if _, err = write(c, cmd); err != nil {
    return nil, err
  }

  resp, err = read(c)
  return resp, err
}

// A simple wrapper for socket writes
func write(c net.Conn, cmd string) (int, error) {
  return c.Write([]byte(cmd))
}

// A buffered read method for a net.Conn (always a unix conn in this case)
func read(c net.Conn) (resp []byte, err error) {
  buf := make([]byte, 512)
  for {
    n, err := c.Read(buf)
    switch err {
    case io.EOF:
      // EOF signals the finished message. Append the last bit of information and return.
      resp = append(resp, buf[:n]...)
      return resp, err
    case nil:
      // No EOF means there is more to come append and read again.
      resp = append(resp, buf[:n]...)
    default:
      // An unexpected error occured. Return it.
      return nil, err
    }
  }
  // Code should never get to here, but if it were to, return what we have.
  return resp, err
}
