/*
Copyright 2018 Iguazio Systems Ltd.

Licensed under the Apache License, Version 2.0 (the "License") with
an addition restriction as set forth herein. You may not use this
file except in compliance with the License. You may obtain a copy of
the License at http://www.apache.org/licenses/LICENSE-2.0.

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
implied. See the License for the specific language governing
permissions and limitations under the License.

In addition, you may not use the software for any purposes that are
illegal under applicable law, and the grant of the foregoing license
under the Apache 2.0 license is conditioned upon your compliance with
such restriction.
*/

package frames

import (
	"fmt"
	"github.com/pkg/errors"
	"reflect"
	"time"
)

const (
	sliceColumnTag = "sliceCol"
)

var (
	ifaceSliceType = reflect.TypeOf([]interface{}{})
)

// SliceColumn is a column with slice data
type SliceColumn struct {
	name string
	data interface{}
	size int
}

// NewSliceColumn return a new SliceColumn
func NewSliceColumn(name string, data interface{}) (*SliceColumn, error) {
	var size int

	switch reflect.TypeOf(data) {
	case IntType:
		size = len(data.([]int))
	case FloatType:
		size = len(data.([]float64))
	case StringType:
		size = len(data.([]string))
	case TimeType:
		size = len(data.([]time.Time))
	// TODO: Why does this happen?
	case ifaceSliceType:
		data, err := convertIfaceSlice(data)
		if err != nil {
			return nil, errors.Wrap(err, "can't convert []interface{}")
		}

		return NewSliceColumn(name, data)
	default:
		return nil, fmt.Errorf("unspported data type - %T", data)
	}

	sc := &SliceColumn{
		data: data,
		name: name,
		size: size,
	}

	return sc, nil
}

// Name returns the column name
func (sc *SliceColumn) Name() string {
	return sc.name
}

// Len returns the number of elements
func (sc *SliceColumn) Len() int {
	return sc.size
}

// DType returns the data type
func (sc *SliceColumn) DType() DType {
	return reflect.TypeOf(sc.data)
}

// Ints returns data as []int
func (sc *SliceColumn) Ints() ([]int, error) {
	typedCol, ok := sc.data.([]int)
	if !ok {
		return nil, fmt.Errorf("wrong type (type is %s)", sc.DType())
	}

	return typedCol, nil
}

// Floats returns data as []float64
func (sc *SliceColumn) Floats() ([]float64, error) {
	typedCol, ok := sc.data.([]float64)
	if !ok {
		return nil, fmt.Errorf("wrong type (type is %s)", sc.DType())
	}

	return typedCol, nil
}

// Strings returns data as []string
func (sc *SliceColumn) Strings() ([]string, error) {
	typedCol, ok := sc.data.([]string)
	if !ok {
		return nil, fmt.Errorf("wrong type (type is %s)", sc.DType())
	}

	return typedCol, nil
}

// Times returns data as []time.Time
func (sc *SliceColumn) Times() ([]time.Time, error) {
	typedCol, ok := sc.data.([]time.Time)
	if !ok {
		return nil, fmt.Errorf("wrong type (type is %s)", sc.DType())
	}

	return typedCol, nil
}

// Slice returns a Column with is slice of data
func (sc *SliceColumn) Slice(start int, end int) (Column, error) {
	if err := validateSlice(start, end, sc.Len()); err != nil {
		return nil, err
	}

	var slice interface{}
	switch sc.DType() {
	case IntType:
		typedCol, _ := sc.Ints()
		slice = typedCol[start:end]
	case FloatType:
		typedCol, _ := sc.Floats()
		slice = typedCol[start:end]
	case StringType:
		typedCol, _ := sc.Strings()
		slice = typedCol[start:end]
	case TimeType:
		typedCol, _ := sc.Times()
		slice = typedCol[start:end]
	}

	return NewSliceColumn(sc.Name(), slice)
}

// Append appends a value
func (sc *SliceColumn) Append(value interface{}) error {
	switch sc.DType() {
	case IntType:
		typedVal, ok := value.(int)
		if !ok {
			return fmt.Errorf("wront type for %s - %T", sc.DType(), value)
		}

		typedCol, _ := sc.Ints()
		sc.data = append(typedCol, typedVal)
	case FloatType:
		typedVal, ok := value.(float64)
		if !ok {
			return fmt.Errorf("wront type for %s - %T", sc.DType(), value)
		}

		typedCol, _ := sc.Floats()
		sc.data = append(typedCol, typedVal)
	case StringType:
		typedVal, ok := value.(string)
		if !ok {
			return fmt.Errorf("wront type for %s - %T", sc.DType(), value)
		}
		typedCol, _ := sc.Strings()
		sc.data = append(typedCol, typedVal)
	case TimeType:
		typedVal, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("wront type for %s - %T", sc.DType(), value)
		}

		typedCol, _ := sc.Times()
		sc.data = append(typedCol, typedVal)
	default:
		return fmt.Errorf("unknown column type - %s", sc.DType())
	}

	sc.size++
	return nil
}

// SliceColumnMessage is SliceColum over-the-wirte message
type SliceColumnMessage struct {
	Name       string      `msgpack:"name"`
	IntData    []int       `msgpack:"ints, omitempty"`
	FloatData  []float64   `msgpack:"floats,omitempty"`
	StringData []string    `msgpack:"strings,omitempty"`
	TimeData   []time.Time `msgpack:"times,omitempty"`
}

// Marshal marshals to native type
func (sc *SliceColumn) Marshal() (interface{}, error) {
	msg := &SliceColumnMessage{
		Name: sc.Name(),
	}
	data := map[string]interface{}{
		"tag":  sliceColumnTag,
		"data": sc.data,
		"name": sc.Name(),
	}

	return data, nil
}

func convertIfaceSlice(data interface{}) (interface{}, error) {
	islice, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("data is not %s (it's %T)", ifaceSliceType, data)
	}

	if len(islice) == 0 {
		// TODO: Should we pick a type?
		return nil, fmt.Errorf("empty slice")
	}

	switch islice[0].(type) {
	case int:
		typedData := make([]int, len(islice))
		for i, ival := range islice {
			val, ok := ival.(int)
			if !ok {
				return nil, fmt.Errorf("%d: not int (%T)", i, ival)
			}

			typedData[i] = val
			return typedData, nil
		}
	case float64:
		typedData := make([]float64, len(islice))
		for i, ival := range islice {
			val, ok := ival.(float64)
			if !ok {
				return nil, fmt.Errorf("%d: not float64 (%T)", i, ival)
			}

			typedData[i] = val
			return typedData, nil
		}
	case string:
		typedData := make([]string, len(islice))
		for i, ival := range islice {
			val, ok := ival.(string)
			if !ok {
				return nil, fmt.Errorf("%d: not string (%T)", i, ival)
			}

			typedData[i] = val
			return typedData, nil
		}
	case time.Time:
		typedData := make([]time.Time, len(islice))
		for i, ival := range islice {
			val, ok := ival.(time.Time)
			if !ok {
				return nil, fmt.Errorf("%d: not time.Time (%T)", i, ival)
			}

			typedData[i] = val
			return typedData, nil
		}
	// TODO: Why?
	case *time.Time:
		typedData := make([]time.Time, len(islice))
		for i, ival := range islice {
			val, ok := ival.(*time.Time)
			if !ok {
				return nil, fmt.Errorf("%d: not time.Time (%T)", i, ival)
			}

			typedData[i] = *val
			return typedData, nil
		}
	}

	return nil, fmt.Errorf("unsupported type - %T", islice[0])
}
