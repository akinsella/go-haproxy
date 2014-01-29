package haproxy

import (
  "bytes"
  "encoding/csv"
  "errors"
  "reflect"
  "strconv"
)

var headerIndices map[string]int

// A wrapper for getting the load for a given backend.
func (h Haproxy) GetLoad(backendName string) ([]*Load, error) {
  resp, err := h.Socket.ShowStat()
  if err != nil {
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
  FailedCheck int    `csv:"chkdown"`
  CheckStatus string `csv:"check_status"`
  CheckCode   int    `csv:"check_code"`
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

func parseLoad(headers []string, body [][]string, backendName string) (load []*Load, err error) {
  pxNameIndex, _ := findHeader(headers, "pxname")
  headerIndices = buildIndices(headers)

  // Used for the parsed numbers within the loop.
  var n int
  for _, fields := range body {
    // Skip any other servers than the requested.
    if fields[pxNameIndex] != backendName {
      continue
    }

    l := new(Load)
    r := reflect.ValueOf(l).Elem()
    for name, index := range headerIndices {
      f := r.FieldByName(name)
      switch f.Kind() {
      case reflect.String:
        f.SetString(fields[index])
      case reflect.Int:
        val := fields[index]
        switch val {
        case "":
          f.SetInt(0)
        default:
          n, err = strconv.Atoi(fields[index])
          if err != nil {
            return nil, err
          }
          f.SetInt(int64(n))
        }
      }
    }
    load = append(load, l)
  }

  return load, nil
}

func buildIndices(headers []string) map[string]int {
  if headerIndices != nil {
    return headerIndices
  }
  h := map[string]int{}
  r := reflect.TypeOf(Load{})
  for i := 0; i < r.NumField(); i++ {
    f := r.Field(i)

    // TODO Terribly inefficient. Loops through the headers for every field. :\
    h[f.Name], _ = findHeader(headers, f.Tag.Get("csv"))
  }
  return h
}

// A helper func to find the index of a given header.
func findHeader(headers []string, header string) (int, error) {
  for i, h := range headers {
    if h == header {
      return i, nil
    }
  }
  return -1, errors.New("Header not found.")
}
