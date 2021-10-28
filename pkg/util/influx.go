package util

import (
	"context"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	log "github.com/sirupsen/logrus"
	"time"
)

type InfluxDataPoint interface {
	Fields() map[string]interface{}
	Tags() map[string]string
	Measurement() string
	Timestamp() time.Time
}

type AtmoData struct {
	T         float64
	H         float64
	PM25      float64
	PM10      float64
	Ts time.Time
}

func (d AtmoData) Fields() map[string]interface{} {
	return map[string]interface{}{
		"t":    d.T,
		"h":    d.H,
		"pm25": d.PM25,
		"pm10": d.PM10,
		// TODO: compute estimated AQI
	}
}
func (d AtmoData) Tags() map[string]string {
	return map[string]string{
		"node": "mbed",
	}
}
func (d AtmoData) Measurement() string {
	return "atmo"
}

func (d AtmoData) Timestamp() time.Time {
	return d.Ts
}

func FlushInfluxBuffer(data []InfluxDataPoint, client api.WriteAPIBlocking) error{
	for _, d := range data {
		p := influxdb2.NewPoint(d.Measurement(), d.Tags(), d.Fields(), d.Timestamp())
		err := client.WritePoint(context.Background(), p)
		if err != nil {
			log.WithField("point", p).WithError(err).Error("Error writing to influx")
			return err
		}
	}
	return nil

}
