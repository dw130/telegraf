package http_server

import (
	"sync"
	"fmt"
	"io/ioutil"
	"io"
	"strings"
	"strconv"
	"time"
	//"net"
	"net/http"
	"github.com/julienschmidt/httprouter"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type HttpServer struct {
	bufCh    chan *Point
	mux      sync.Mutex
	Port     int `toml:"port"`
}

type Point struct {
	tags map[string]string
	fields map[string]interface{}
	mm string
	times int64
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

	fmt.Printf("**begin gather*%v**\n", len(s.bufCh)  )
	for {
		select {
			case data := <-s.bufCh:
				nn := time.Unix(data.times / 1000 , 0)
				fmt.Printf("data:%+v  nn:%v\n",data,nn)
				acc.AddGauge(data.mm, data.fields, data.tags, nn)

    		default:
    			fmt.Printf("********default**************\n")
    			break
		}
	}
	fmt.Printf("**end gather***\n")
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

    w.WriteHeader(http.StatusOK)

    ss := ps.ByName("ss")

    tags := map[string]string{
    	"app_name": ps.ByName("app"),
    	"app_id": ps.ByName("appid"),
    	//"thread": ps.ByName("ti"),
    }

    th := ps.ByName("ti")
    if th != "" {
    	tags["thread"] = th
    }

    if ss != ps.ByName("app") && ss != "" {
    	tags["sevice_name"] = ss
    }

    name := ps.ByName("name")
    if name != "" {
    	tags["name"] = name
    }

    body, _ := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))

    if err := r.Body.Close(); err != nil {
        panic(err)
    }

    strList := strings.Split(string(body)," ")
    if len(strList) < 3 {
    	fmt.Printf( "*catch wrong metric:%v\n", string(body) )
    	return
    }

    cc := strings.Index(strList[0], "{")
    if cc <= 0 {
    	return
    }

    metric := strings.TrimPrefix(strList[0][0:cc],"MONITOR_")
    metrics := strings.ToLower(metric)
	val, _ := strconv.ParseFloat(strList[1], 64)
	tt, _ := strconv.ParseInt(strings.Trim(strList[2],"\n"), 10, 64)
	//fmt.Printf("metric:%s   val:%v  time:%v\n",metric,val,tt)

	s.bufCh <- &Point{mm:metrics, tags:tags, fields: map[string]interface{}{"val":val},times:tt}
}


func init() {

	inputs.Add("http_server", func() telegraf.Input {
		t := &HttpServer{
			Port: 9777,
		}

		t.bufCh  =   make(chan *Point,5000)

		router := httprouter.New()

		router.POST("/goreplay", t.Goreplay)
		router.POST("/metrics/job/monitor/region_name/:region/app_name/:app/app_id/:appid/Name/:name", t.Metric)
		router.POST("/metrics/job/monitor/region_name/:region/app_name/:app/app_id/:appid/sevice_name/:ss/thread_index/:ti", t.Metric)
		go http.ListenAndServe(":9777", router)

		return t
	})
}
