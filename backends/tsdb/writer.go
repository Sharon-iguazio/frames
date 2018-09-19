package tsdb

import (
	"fmt"
	"github.com/nuclio/logger"
	"github.com/pkg/errors"
	"github.com/v3io/frames"
	"github.com/v3io/v3io-tsdb/pkg/tsdb"
	"github.com/v3io/v3io-tsdb/pkg/utils"
	"sort"
	"time"
)

func (b *Backend) Write(request *frames.WriteRequest) (frames.FrameAppender, error) {

	appender, err := b.adapter.Appender()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Appender")
	}

	newTsdbAppender := tsdbAppender{
		request:  request,
		appender: appender,
		logger:   b.logger,
	}

	if request.ImmidiateData != nil {
		err := newTsdbAppender.Add(request.ImmidiateData)
		if err != nil {
			return &newTsdbAppender, err
		}
	}

	return &newTsdbAppender, nil
}

// Appender is key/value appender
type tsdbAppender struct {
	request  *frames.WriteRequest
	appender tsdb.Appender
	logger   logger.Logger
}

type metricCtx struct {
	lset utils.Labels
	ref  uint64
	data []float64
}

func (a *tsdbAppender) Add(frame frames.Frame) error {

	if frame.Len() == 0 {
		return nil
	}

	names := frame.Names()
	if len(names) == 0 {
		return fmt.Errorf("empty frame")
	}

	lset := make(utils.Labels, 0, len(frame.Labels()))
	for name, val := range frame.Labels() {
		lset = append(lset, utils.Label{Name: name, Value: fmt.Sprintf("%s", val)})
	}
	sort.Sort(lset)

	values := map[string][]float64{}
	tarray := make([]int64, frame.Len())
	var lastTime int64

	indexCol, err := frame.Column(frame.IndexName())
	if err != nil {
		return err
	}

	times, err := indexCol.Times()
	if err != nil {
		return errors.Wrap(err, "TimeSeries index is not of type Time")
	}

	for i := 0; i < frame.Len(); i++ {
		t := times[i].Unix()*1000 + times[i].UnixNano()/1000
		if t < lastTime {
			return errors.Wrap(err, "time column is out of order (need to be sorted by time)")
		}
		tarray[i] = t
	}

	for _, name := range names {
		if name != frame.IndexName() {
			col, err := frame.Column(name)
			if err != nil {
				return err
			}

			switch col.DType() {
			case frames.FloatType:
				asFloat, _ := col.Floats()
				values[name] = asFloat
			case frames.IntType:
				asInt, _ := col.Ints()
				data := make([]float64, frame.Len())
				for i := 0; i < frame.Len(); i++ {
					data[i] = float64(asInt[i])
				}
				values[name] = data
			default:
				return fmt.Errorf("cannot write type %v as time series value", col.DType())
			}
		}
	}

	metrics := make([]metricCtx, 0, len(values))
	for name, metric := range values {
		lset, err := newLset(frame.Labels(), name, len(values) == 1)
		if err != nil {
			return err
		}
		metrics = append(metrics, metricCtx{lset: lset, data: metric})
	}

	for i := 0; i < frame.Len(); i++ {

		for _, metric := range metrics {
			if i == 0 {
				metric.ref, err = a.appender.Add(metric.lset, tarray[0], metric.data[0])
				if err != nil {
					return errors.Wrap(err, "failed to Add")
				}
			} else {
				err := a.appender.AddFast(nil, metric.ref, tarray[i], metric.data[i])
				if err != nil {
					return errors.Wrap(err, "failed to AddFast")
				}
			}
		}
	}

	return nil
}

func newLset(labels map[string]interface{}, name string, singleCol bool) (utils.Labels, error) {
	lset := make(utils.Labels, 0, len(labels))
	var hadName bool
	for name, val := range labels {
		if name == "__name__" {
			if !singleCol {
				return nil, fmt.Errorf("label __name__ cannot be set with multi column TSDB frames")
			}
			val = name
			hadName = true
		}
		lset = append(lset, utils.Label{Name: name, Value: fmt.Sprintf("%s", val)})
	}
	if !hadName {
		lset = append(lset, utils.Label{Name: "__name__", Value: name})
	}
	sort.Sort(lset)
	return lset, nil
}

func (a *tsdbAppender) WaitForComplete(timeout time.Duration) error {
	_, err := a.appender.WaitForCompletion(timeout)
	return err
}
