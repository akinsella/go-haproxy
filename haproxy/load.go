package haproxy

import (
  "bytes"
  "encoding/csv"
  "errors"
  "io"
  "strconv"
)

// A wrapper for getting the load for a given backend.
func (h Haproxy) GetLoad(backendName string) ([]*Load, error) {
  resp, err := h.Socket.ShowStat()
  if err != nil && err != io.EOF {
    return nil, err
  }

  r := csv.NewReader(bytes.NewReader(resp))
  stats, err := r.ReadAll()
  if err != nil {
    return nil, err
  }

  headers, body := seperateHeaders(stats)
  return parseLoad(headers, body, backendName)
}

type Load struct {
  Name        string `csv:"svname"`
  Current     int    `csv:"scur"`
  Max         int    `csv:"smax"`
  Health      string `csv:"status"`
  FailedCheck string `csv:"chkdown"`
}

// Haproxy CSV first entry is like: # pvname, svname,......
// Golang csv parses this as ["# pvname", "svname",...]
// Edit the first entry to remove the extra "# "
// Returns the headers and remaining rows of data.
func seperateHeaders(resp [][]string) ([]string, [][]string) {
  headers := resp[0]
  headers[0] = headers[0][2:]
  return headers[:len(headers)-1], resp[1:]
}

// TODO: Use reflect with the Load struct to scan the data into []*Load automatically.
func parseLoad(headers []string, body [][]string, backendName string) (load []*Load, err error) {
  var nameIndex, pxNameIndex, currentLoadIndex, maxLoadIndex int

  // A bunch of manual parsing. See above TODO.
  if pxNameIndex, err = findHeader(headers, "pxname"); err != nil {
    return nil, err
  }
  if nameIndex, err = findHeader(headers, "svname"); err != nil {
    return nil, err
  }
  if currentLoadIndex, err = findHeader(headers, "scur"); err != nil {
    return nil, err
  }
  if maxLoadIndex, err = findHeader(headers, "smax"); err != nil {
    return nil, err
  }

  // Used for the parsed numbers within the loop.
  var n int
  for _, fields := range body {
    // Skip any other servers than the requested.
    if fields[pxNameIndex] != backendName {
      continue
    }
    l := new(Load)
    l.Name = fields[nameIndex]
    if n, err = strconv.Atoi(fields[currentLoadIndex]); err != nil {
      return nil, err
    }
    l.Current = n
    if n, err = strconv.Atoi(fields[maxLoadIndex]); err != nil {
      return nil, err
    }
    l.Max = n
    load = append(load, l)
  }

  return load, nil
}

// A helper func to enable the manual parsing above. Hopefully obsoleted by a true scanning.
func findHeader(headers []string, header string) (int, error) {
  for i, h := range headers {
    if h == header {
      return i, nil
    }
  }
  return -1, errors.New("Header not found.")
}
