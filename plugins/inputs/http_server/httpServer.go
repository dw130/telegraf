package http_server

import (
	"sync"
	"fmt"
	"io/ioutil"
	"io"
	"strings"
	//"time"
	//"net"
	"net/http"
	"github.com/julienschmidt/httprouter"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type HttpServer struct {

	mux      sync.Mutex
	Port     int `toml:"port"`
}


func (_ *HttpServer) Description() string {
	return "http server"
}

var sConfig = `
`

func (_ *HttpServer) SampleConfig() string {
	return sConfig
}

func (s *HttpServer) Gather(acc telegraf.Accumulator) error {

	return nil
}


func (s *HttpServer) Goreplay(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

    body, _ := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))

    if err := r.Body.Close(); err != nil {
        panic(err)
    }

	fmt.Printf("params**%+v\n",body)
}


func (s *HttpServer) Metric(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

    w.WriteHeader(http.StatusOk)

    region := ps.ByName("region")
    ti := ps.ByName("ti")
    ss := ps.ByName("ss")
    appid := ps.ByName("appid")
    app := ps.ByName("app")

    fmt.Printf("region：%v***ti:%v**ss:%v***appid:%v***app:%v\n",region,ti,ss,appid,app)

    body, _ := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))

    if err := r.Body.Close(); err != nil {
        panic(err)
    }

    strList := strings.Split(string(body)," ",-1)
    if len(strList) < 3 {
    	fmt.Printf( "*catch wrong metric:%v\n", string(body) )
    	return
    }
    metric := strings.TrimRight(strList[0])

	val, _ := strconv.ParseFloat(strList[1], 64)

	tt, _ := strconv.ParseInt(strList[2], 10, 64)

	fmt.Printf("metric:%+v   val:%v time:\n",metric,val,tt)
}


func init() {

	inputs.Add("http_server", func() telegraf.Input {
		t := &HttpServer{
			Port: 9777,
		}
		router := httprouter.New()

		router.POST("/goreplay", t.Goreplay)
		router.POST("/metrics/job/monitor/region_name/:region/app_name/:app/app_id/:appid/sevice_name/:ss/thread_index/:ti", t.Metric)
		go http.ListenAndServe(":9777", router)

		return t
	})
}
