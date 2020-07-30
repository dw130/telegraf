package http_server

import (
	"sync"
	"fmt"
	"time"
	"net"
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
	times, err := s.ps.CPUTimes(s.PerCPU, s.TotalCPU)
	if err != nil {
		return fmt.Errorf("error getting CPU info: %s", err)
	}
	now := time.Now()

	for _, cts := range times {
		tags := map[string]string{
			"cpu": cts.CPU,
		}

		total := totalCpuTime(cts)
		active := activeCpuTime(cts)

		if s.CollectCPUTime {
			// Add cpu time metrics
			fieldsC := map[string]interface{}{
				"time_user":       cts.User,
				"time_system":     cts.System,
				"time_idle":       cts.Idle,
				"time_nice":       cts.Nice,
				"time_iowait":     cts.Iowait,
				"time_irq":        cts.Irq,
				"time_softirq":    cts.Softirq,
				"time_steal":      cts.Steal,
				"time_guest":      cts.Guest,
				"time_guest_nice": cts.GuestNice,
			}
			if s.ReportActive {
				fieldsC["time_active"] = activeCpuTime(cts)
			}
			acc.AddCounter("cpu", fieldsC, tags, now)
		}

		// Add in percentage
		if len(s.lastStats) == 0 {
			// If it's the 1st gather, can't get CPU Usage stats yet
			continue
		}

		lastCts, ok := s.lastStats[cts.CPU]
		if !ok {
			continue
		}
		lastTotal := totalCpuTime(lastCts)
		lastActive := activeCpuTime(lastCts)
		totalDelta := total - lastTotal

		if totalDelta < 0 {
			err = fmt.Errorf("Error: current total CPU time is less than previous total CPU time")
			break
		}

		if totalDelta == 0 {
			continue
		}

		fieldsG := map[string]interface{}{
			"usage_user":       100 * (cts.User - lastCts.User - (cts.Guest - lastCts.Guest)) / totalDelta,
			"usage_system":     100 * (cts.System - lastCts.System) / totalDelta,
			"usage_idle":       100 * (cts.Idle - lastCts.Idle) / totalDelta,
			"usage_nice":       100 * (cts.Nice - lastCts.Nice - (cts.GuestNice - lastCts.GuestNice)) / totalDelta,
			"usage_iowait":     100 * (cts.Iowait - lastCts.Iowait) / totalDelta,
			"usage_irq":        100 * (cts.Irq - lastCts.Irq) / totalDelta,
			"usage_softirq":    100 * (cts.Softirq - lastCts.Softirq) / totalDelta,
			"usage_steal":      100 * (cts.Steal - lastCts.Steal) / totalDelta,
			"usage_guest":      100 * (cts.Guest - lastCts.Guest) / totalDelta,
			"usage_guest_nice": 100 * (cts.GuestNice - lastCts.GuestNice) / totalDelta,
		}
		if s.ReportActive {
			fieldsG["usage_active"] = 100 * (active - lastActive) / totalDelta
		}
		acc.AddGauge("cpu", fieldsG, tags, now)
	}


	return err
}


func (s *HttpServer) Goreplay(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	fmt.Printf("params**%+v\n",params)
}


func (s *HttpServer) Metric(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	fmt.Printf("params**%+v\n",params)
}


func init() {
	inputs.Add("http_server", func() telegraf.Input {
		t := &HttpServer{
			Port: 9777
		}
		router := httprouter.New()
		router.POST("/goreplay", t.Goreplay)
		router.POST("/metrics/job/monitor/region_name/:region/app_name/:app/app_id/:appid/sevice_name/:ss/thread_index/:ti", t.Metric)
		http.ListenAndServe(":8080", router)
		return t
	})
}
