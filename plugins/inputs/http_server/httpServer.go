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
	tagFields     map[string] map[string]interface{}
	tagTags       map[string] map[string]string
}

type Point struct {
	tags map[string]string
	fields float64
	mm string
	times int64
	tagS string
}

func (_ *HttpServer) Description() string {
	return "http server"
}

var sConfig = `
`

func (_ *HttpServer) SampleConfig() string {
	return sConfig
}

func (t *HttpServer) Gather(acc telegraf.Accumulator) error {

	//fmt.Printf("**begin gather*%v**\n", len(s.bufCh)  )

L:
	for {
		select {
			case data := <-t.bufCh:

				tagS := data.tagS
				_,ok := t.tagFields[tagS]
				if ok == false {
					t.tagFields[tagS] = map[string]interface{}{}
				}

				_,ok = t.tagTags[tagS]
				if ok == false {
					t.tagTags[tagS] = data.tags
				}


				t.tagFields[tagS][data.mm] = data.fields

    		default:
    			mm := "receive_metrics"
    			for k,_ := range t.tagFields {
    				acc.AddGauge( mm, t.tagFields[k], t.tagTags[k], time.Now() )
    				fmt.Printf("*********tt****%v***%v\n",t.tagFields[k], t.tagTags[k])
				}

				t.tagFields =  map[string] map[string]interface{}{}

    			break L
		}
	}
	//fmt.Printf("**end gather***\n")
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

    tags := map[string]string{
    	"app_name": ps.ByName("app"),
    	"app_id": ps.ByName("appid"),
    }

    tagS := fmt.Sprintf("%s%s",ps.ByName("app"),ps.ByName("appid"))

    th := ps.ByName("ti")
    if th != "" {
    	tags["thread"] = th
		tagS = fmt.Sprintf("%s%s",tagS,th)
    }

    name := ps.ByName("name")
    if name != "" {
    	tags["midd_instance"] = name
  		tagS = fmt.Sprintf("%s%s",tagS,name)
    }

    redis := ps.ByName("redis")
    var redisTag = false
    if redis != "" {
    	tags["redis_function"] = redis
  		tagS = fmt.Sprintf("%s%s",tagS,redis)
  		redisTag = true
    }

    mysql := ps.ByName("mysql")
    var mysqlTag = false
    if mysql != "" {
    	tags["mysql_function"] = mysql
  		tagS = fmt.Sprintf("%s%s",tagS,mysql)
  		mysqlTag = true
    }

    body, _ := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))

    if err := r.Body.Close(); err != nil {
        panic(err)
    }

    fmt.Fprint(w, "")
    
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
	fmt.Printf("metric:%s   val:%v  time:%v\n",metric,val,tt)

	if metrics == "connection_num" && val == 0.0 {
		return
	}
	if redisTag == true {
		metrics = metrics + "_redis"
	}
	if mysqlTag == true {
		metrics = metrics + "_mysql"
	}
	fmt.Printf("check******%v**%v****%v***%v***%v\n",redisTag,mysqlTag,metrics,tags,val)
	s.bufCh <- &Point{mm:metrics, tags:tags, fields: val,times:tt,tagS:tagS}
}


func init() {

	inputs.Add("http_server", func() telegraf.Input {
		t := &HttpServer{
			Port: 9777,
		}

		t.bufCh  =   make(chan *Point,5000)
		t.tagFields   =  map[string] map[string]interface{}{}
		t.tagTags  =  map[string] map[string]string{}

		router := httprouter.New()

		router.POST("/goreplay", t.Goreplay)
		router.POST("/metrics/job/monitor/region_name/:region/app_name/:app/app_id/:appid/MYSQL/:mysql/sevice_name/:ss/thread_index/:ti", t.Metric)
		router.POST("/metrics/job/monitor/region_name/:region/app_name/:app/app_id/:appid/REDIS/:redis/sevice_name/:ss/thread_index/:ti", t.Metric)
		router.POST("/metrics/job/monitor/region_name/:region/app_name/:app/app_id/:appid", t.Metric)
		router.POST("/metrics/job/monitor/region_name/:region/app_name/:app/app_id/:appid/Name/:name", t.Metric)
		router.POST("/metrics/job/monitor/region_name/:region/app_name/:app/app_id/:appid/sevice_name/:ss/thread_index/:ti", t.Metric)
		go http.ListenAndServe(":9777", router)

		return t
	})
}
