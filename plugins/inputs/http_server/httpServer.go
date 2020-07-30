package http_server

import (
	"sync"
	"fmt"
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
	params := httprouter.ParamsFromContext(r.Context())
	fmt.Printf("params**%+v\n",params)
}


func (s *HttpServer) Metric(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	params := httprouter.ParamsFromContext(r.Context())
	fmt.Printf("params**%+v\n",params)
}


func init() {
	inputs.Add("http_server", func() telegraf.Input {
		t := &HttpServer{
			Port: 9777,
		}
		router := httprouter.New()
		router.POST("/goreplay", t.Goreplay)
		router.POST("/metrics/job/monitor/region_name/:region/app_name/:app/app_id/:appid/sevice_name/:ss/thread_index/:ti", t.Metric)
		http.ListenAndServe(":8080", router)
		return t
	})
}
