package models

import (
	"fmt"
	"strings"
	"github.com/influxdata/telegraf"
)

// Makemetric applies new metric plugin and agent measurement and tag
// settings.
func makemetric(
	metric telegraf.Metric,
	nameOverride string,
	namePrefix string,
	nameSuffix string,
	tags map[string]string,
	globalTags map[string]string,
) telegraf.Metric {
	if len(nameOverride) != 0 {
		metric.SetName(nameOverride)
	}

	if len(namePrefix) != 0 {
		metric.AddPrefix(namePrefix)
	}
	if len(nameSuffix) != 0 {
		metric.AddSuffix(nameSuffix)
	}

	// Apply plugin-wide tags
	for k, v := range tags {
		_, ok := metric.GetTag(k) 
		fmt.Printf("****1*****%v**%v\n",k,ok,strings.HasPrefix(v,"mul"))
		if  k == "app_name" && ok == true && strings.HasPrefix(v,"mul") {
			metric.AddTag(k, v)
		}
		if ok == false {
			metric.AddTag(k, v)
		}
	}
	// Apply global tags
	for k, v := range globalTags {
		_, ok := metric.GetTag(k) 
		fmt.Printf("****2*****%v**%v\n",k,ok,strings.HasPrefix(v,"mul"))
		if  k == "app_name" && ok == true && strings.HasPrefix(v,"mul") {
			metric.AddTag(k, v)
		}
		if ok == false {
			metric.AddTag(k, v)
		}
	}

	return metric
}
