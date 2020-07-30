package http_server

import (
	"sync"
	"fmt"
	"io/ioutil"
	"io"
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
    region := ps.ByName("region")
    ti := ps.ByName("ti")

    fmt.Printf("region：%v***ti:%v\n",region,ti)
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
