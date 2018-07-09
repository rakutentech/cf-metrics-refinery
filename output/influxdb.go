package output

import (
	"time"

	influxdb "github.com/influxdata/influxdb/client/v2"
	"github.com/pkg/errors"
	"github.com/rakutentech/cf-metrics-refinery/transformer"
)

type InfluxDB struct { // implements AsyncWriter
	c   influxdb.Client
	bpc influxdb.BatchPointsConfig
	mbe int
}

type ConfigInfluxDB struct {
	Username          string        `desc:"Username to connect to InfluxDB"`                                             // CFMR_INFLUXDB_USERNAME
	Password          string        `desc:"Password to connect to InfluxDB"`                                             // CFMR_INFLUXDB_PASSWORD
	SkipSSLValidation bool          `default:"false" desc:"Skip SSL certificate validation when connecting to InfluxDB"` // CFMR_INFLUXDB_SKIPSSLVALIDATION
	Addr              string        `required:"true" desc:"URL of InfluxDB"`                                             // CFMR_INFLUXDB_ADDR
	Timeout           time.Duration `default:"1m" desc:"Timeout for requests to InfluxDB"`                               // CFMR_INFLUXDB_TIMEOUT
	UserAgent         string        `ignored:"true"`

	Database          string        `required:"true" desc:"Name of InfluxDB database to write to"`            // CFMR_INFLUXDB_DATABASE
	RetentionPolicy   string        `desc:"Name of the retention policy to use (instead of the default one)"` // CFMR_INFLUXDB_RETENTIONPOLICY
	InfluxPingTimeout time.Duration `default:"5s" desc:"Default timeout of checking Influxdb is up or not"`   // CFMR_INFLUXDB_INFLUXPINGTIMEOUT
}

func NewInfluxDB(cfg ConfigInfluxDB) (*InfluxDB, error) {
	c, err := influxdb.NewHTTPClient(influxdb.HTTPConfig{
		Username:           cfg.Username,
		Password:           cfg.Password,
		InsecureSkipVerify: cfg.SkipSSLValidation,
		Addr:               cfg.Addr,
		Timeout:            cfg.Timeout,
		UserAgent:          cfg.UserAgent,
	})
	if err != nil {
		return nil, errors.Wrap(err, "creating InfluxDB client")
	}

	bpc := influxdb.BatchPointsConfig{
		Database:        cfg.Database,
		RetentionPolicy: cfg.RetentionPolicy,
	}

	return &InfluxDB{c: c, bpc: bpc}, nil
}

// Check if the server is up
func (o *InfluxDB) Ping(timeout time.Duration) error {
	_, _, err := o.c.Ping(timeout)
	return err
}

func (o *InfluxDB) Write(envs ...*transformer.Envelope) error {
	ps := make([]*influxdb.Point, 0, len(envs))
	for _, e := range envs {
		p, err := transformer.ToInfluxDBPoint(e)
		if err == nil {
			ps = append(ps, p)
		} else if errors.Cause(err) == transformer.ErrEventDiscarded {
			continue
		} else {
			return errors.Wrap(err, "transforming event to InfluxDB data point")
		}
	}

	b, err := influxdb.NewBatchPoints(o.bpc)
	if err != nil {
		return errors.Wrap(err, "creating InfluxDB batch container")
	}

	b.AddPoints(ps)

	err = o.c.Write(b)
	if err != nil {
		return errors.Wrap(err, "writing to InfluxDB")
	}

	return nil
}
