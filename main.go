// Copyright 2016 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"math"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/prometheus/common/model"
	"github.com/spf13/pflag"
	"github.com/prometheus/prometheus/prompb"
	"github.com/Microsoft/ApplicationInsights-Go/appinsights"
)

func main() {
	ikey := pflag.String("ikey", "", "IKey for target Application Insights app to receive Prometheus metrics.")

	pflag.Parse()

	if *ikey == "" {
		log.Fatalln("No instrumentation key provided.")
	}

	client := appinsights.NewTelemetryClient(*ikey)

	appinsights.NewDiagnosticsMessageListener(func(msg string) error {
		fmt.Printf("[%s] %s\n", time.Now().Format(time.UnixDate), msg)
		return nil
	})

	http.HandleFunc("/receive", func(w http.ResponseWriter, r *http.Request) {
		compressed, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		reqBuf, err := snappy.Decode(nil, compressed)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var req prompb.WriteRequest
		if err := proto.Unmarshal(reqBuf, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		samples := protoToSamples(&req)
		for _, sample := range samples {
			name := string(sample.Metric[model.MetricNameLabel])
			value := float64(sample.Value)
			if math.IsNaN(value) || math.IsInf(value, 0) {
				continue
			}
			metric := appinsights.NewMetricTelemetry(name, value)
			metric.Properties = cleanLabels(sample.Metric)
			client.Track(metric)
		}
	})

	log.Fatal(http.ListenAndServe(":1234", nil))
}


func protoToSamples(req *prompb.WriteRequest) model.Samples {
	var samples model.Samples
	for _, ts := range req.Timeseries {
		metric := make(model.Metric, len(ts.Labels))
		for _, l := range ts.Labels {
			metric[model.LabelName(l.Name)] = model.LabelValue(l.Value)
		}
		for _, s := range ts.Samples {
			samples = append(samples, &model.Sample{
				Metric:    metric,
				Value:     model.SampleValue(s.Value),
				Timestamp: model.Time(s.Timestamp),
			})
		}
	}
	return samples
}

func cleanLabels(metric map[model.LabelName]model.LabelValue) map[string]string {
	result := make(map[string]string)
	for name, val := range metric {
		if name != model.MetricNameLabel {
			result[string(name)] = string(val)
		}
	}
	return result
}