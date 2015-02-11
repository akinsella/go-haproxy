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

// A wrapper for getting the load for a given backend.
func (h Haproxy) GetLoadAsMap(backendName string) ([]*map[string]interface{}, error) {
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
    return parseLoadAsMap(headers, body, backendName)
}

type Load struct {
  Pxname        string `csv:"pxname"`
  Svname        string `csv:"svname"`
  Qcur          int    `csv:"qcur"`
  Qmax          int    `csv:"qmax"`
  Scur          int    `csv:"scur"`
  Smax          int    `csv:"smax"`
  Slim          int    `csv:"slim"`
  Stot          int    `csv:"stot"`
  Bin           int    `csv:"bin"`
  Bout          int    `csv:"bout"`
  Dreq          int    `csv:"dreq"`
  Dresp         int    `csv:"dresp"`
  Ereq          int    `csv:"ereq"`
  Econ          int    `csv:"econ"`
  Eresp         int    `csv:"eresp"`
  Wretr         int    `csv:"wretr"`
  Wredis        int    `csv:"wredis"`
  Status        string `csv:"status"`
  Checkfail     int    `csv:"chkfail"`
  Checkdown     int    `csv:"chkdown"`
  Lastchg       int    `csv:"lastchg"`
  Downtime      int    `csv:"downtime"`
  Qlimit        int    `csv:"qlimit"`
  Pid           int    `csv:"pid"`
  Iid           int    `csv:"iid"`
  Sid           int    `csv:"sid"`
  Throttle      int    `csv:"throttle"`
  Lbtot         int    `csv:"lbtot"`
  Tracked       int    `csv:"tracked"`
  Type          int    `csv:"type"`
  Rate          int    `csv:"rate"`
  Ratelim       int    `csv:"rate_lim"`
  Ratemax       int    `csv:"rate_max"`
  Checkstatus   string `csv:"check_status"`
  CheckCode     int    `csv:"check_code"`
  CheckDuration int    `csv:"check_duration"`
  Hrsp1xx       int    `csv:"hrsp_1xx"`
  Hrsp2xx       int    `csv:"hrsp_2xx"`
  Hrsp3xx       int    `csv:"hrsp_3xx"`
  Hrsp4xx       int    `csv:"hrsp_4xx"`
  Hrsp5xx       int    `csv:"hrsp_5xx"`
  HrspOther     int    `csv:"hrsp_other"`
  Hanafail      int    `csv:"hanafail"`
  Reqrate       int    `csv:"req_rate"`
  Reqratemax    int    `csv:"req_rate_max"`
  Reqtot        int    `csv:"req_tot"`
  Cliabrt       int    `csv:"cli_abrt"`
  Srvabrt       int    `csv:"srv_abrt"`
  Compin        int    `csv:"comp_in"`
  Compout       int    `csv:"comp_out"`
  Compbyp       int    `csv:"comp_byp"`
  Comprsp       int    `csv:"comp_rsp"`
  Lastsess      int    `csv:"lastsess"`
  Lastchk       string `csv:"last_chk"`
  Lastagt       int    `csv:"last_agt"`
  Qtime         int    `csv:"qtime"`
  Ctime         int    `csv:"ctime"`
  Ttime         int    `csv:"ttime"`
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

func parseLoadAsMap(headers []string, body [][]string, backendName string) (load []*map[string]interface{}, err error) {
    pxNameIndex, _ := findHeader(headers, "pxname")
    headerIndices = buildIndices(headers)

    // Used for the parsed numbers within the loop.
    var n int
    for _, fields := range body {
        // Skip any other servers than the requested.
        if fields[pxNameIndex] != backendName {
            continue
        }

        l := new(map[string]interface{})
        r := reflect.ValueOf(l).Elem()
        for name, index := range headerIndices {
            f := r.FieldByName(name)
            switch f.Kind() {
                case reflect.String:
                l[name] = fields[index]
                case reflect.Int:
                val := fields[index]
                switch val {
                    case "":
                    l[name] = 0
                    default:
                    n, err = strconv.Atoi(fields[index])
                    if err != nil {
                        return nil, err
                    }
                    l[name] = int64(n)
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
