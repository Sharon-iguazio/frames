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

package v3ioutils

import (
	"encoding/binary"
	"net/http"
	"strings"

	"github.com/nuclio/logger"
	"github.com/pkg/errors"
	"github.com/v3io/frames"
	v3io "github.com/v3io/v3io-go/pkg/dataplane"
	v3ioerrors "github.com/v3io/v3io-go/pkg/errors"
	"github.com/v3io/v3io-tsdb/pkg/utils"
)

const v3ioUsersContainer = "users"
const v3ioHomeVar = "$V3IO_HOME"

func NewContainer(v3ioContext v3io.Context,
	session *frames.Session,
	password string,
	token string,
	logger logger.Logger) (v3io.Container, error) {

	var pass string
	if password == "" {
		pass = session.Password
	} else {
		pass = password
	}

	var tok string
	if token == "" {
		tok = session.Token
	} else {
		tok = token
	}

	addr := session.Url
	// Backward compatibility for non-URL addr parameter.
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = "http://" + addr
	}

	newSessionInput := v3io.NewSessionInput{
		URL:       addr,
		Username:  session.User,
		Password:  pass,
		AccessKey: tok,
	}
	container, err := createContainer(logger, v3ioContext, session.Container, &newSessionInput)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create data container")
	}

	return container, nil
}

// CreateContainer creates a new container
func createContainer(logger logger.Logger,
	v3ioContext v3io.Context,
	containerName string,
	newSessionInput *v3io.NewSessionInput) (v3io.Container, error) {

	session, err := v3ioContext.NewSession(newSessionInput)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create session")
	}

	container, err := session.NewContainer(&v3io.NewContainerInput{ContainerName: containerName})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create container")
	}

	return container, nil
}

// AsInt64Array convert v3io blob to Int array
func AsInt64Array(val []byte) []uint64 {
	var array []uint64
	bytes := val
	for i := 16; i+8 <= len(bytes); i += 8 {
		val := binary.LittleEndian.Uint64(bytes[i : i+8])
		array = append(array, val)
	}
	return array
}

// DeleteTable deletes a table
func DeleteTable(logger logger.Logger, container v3io.Container, path, filter string, getItemsWorkers int, deleteWorkers int, ignoreMissing bool) error {

	overallNumWorkers := getItemsWorkers + deleteWorkers

	fileNameChan := make(chan string, 1024)
	getItemsTerminationChan := make(chan error, getItemsWorkers)
	deleteTerminationChan := make(chan error, deleteWorkers)
	onErrorTerminationChannel := make(chan struct{}, overallNumWorkers)

	for i := 0; i < getItemsWorkers; i++ {
		input := &v3io.GetItemsInput{
			Path:           path,
			AttributeNames: []string{"__name"},
			Filter:         filter,
			Segment:        i,
			TotalSegments:  getItemsWorkers,
		}
		go getItemsWorker(container, input, fileNameChan, getItemsTerminationChan, onErrorTerminationChannel)
	}

	for i := 0; i < deleteWorkers; i++ {
		go deleteObjectWorker(path, container, fileNameChan, deleteTerminationChan, onErrorTerminationChannel)
	}

	var getItemsTerminated, deletesTerminated int
	for deletesTerminated < deleteWorkers {
		select {
		case err := <-getItemsTerminationChan:
			if err != nil {
				// If requested to ignore non-existing tables do not return error.
				if errorWithStatusCode, ok := err.(v3ioerrors.ErrorWithStatusCode); !ok || !ignoreMissing || errorWithStatusCode.StatusCode() != http.StatusNotFound {
					for i := 0; i < overallNumWorkers; i++ {
						onErrorTerminationChannel <- struct{}{}
					}
					return errors.Wrapf(err, "GetItems failed during recursive delete of '%s'.", path)
				}
			}
			getItemsTerminated++
			if getItemsTerminated == getItemsWorkers {
				close(fileNameChan)
			}
		case err := <-deleteTerminationChan:
			if err != nil {
				for i := 0; i < overallNumWorkers; i++ {
					onErrorTerminationChannel <- struct{}{}
				}
				return errors.Wrapf(err, "Delete failed during recursive delete of '%s'.", path)
			}
			deletesTerminated++
		}
	}

	if filter == "" {
		err := container.DeleteObjectSync(&v3io.DeleteObjectInput{Path: path})
		if err != nil {
			if !utils.IsNotExistsOrConflictError(err) {
				return errors.Wrapf(err, "Failed to delete table object '%s'.", path)
			}
		}
	}

	return nil
}

func getItemsWorker(container v3io.Container, input *v3io.GetItemsInput, fileNameChan chan<- string, terminationChan chan<- error, onErrorTerminationChannel <-chan struct{}) {
	for {
		select {
		case <-onErrorTerminationChannel:
			terminationChan <- nil
			return
		default:
		}
		if input.Filter != "" {
			input.Filter += " and __name != '.#schema'"
		}
		resp, err := container.GetItemsSync(input)
		if err != nil {
			terminationChan <- err
			return
		}
		resp.Release()
		output := resp.Output.(*v3io.GetItemsOutput)
		for _, item := range output.Items {
			fileNameChan <- item.GetField("__name").(string)
		}
		if output.Last {
			terminationChan <- nil
			return
		}
		input.Marker = output.NextMarker
	}
}

func deleteObjectWorker(tablePath string, container v3io.Container, fileNameChan <-chan string, terminationChan chan<- error, onErrorTerminationChannel <-chan struct{}) {
	for {
		select {
		case fileName, ok := <-fileNameChan:
			if !ok {
				terminationChan <- nil
				return
			}
			input := &v3io.DeleteObjectInput{Path: tablePath + "/" + fileName}
			err := container.DeleteObjectSync(input)
			if err != nil {
				if !utils.IsNotExistsOrConflictError(err) {
					terminationChan <- err
					return
				}
			}
		case <-onErrorTerminationChannel:
			return
		}
	}
}

func ProcessPaths(session *frames.Session, path string, addSlash bool) (string, string, error) {

	container := session.Container

	if container == "" {
		sp := strings.SplitN(path, "/", 2)
		if len(sp) < 2 {
			return "", "", errors.New("Please specify a data container name via the container parameters or the path prefix e.g. bigdata/mytable")
		}
		container = sp[0]
		path = sp[1]
	}

	if container == v3ioHomeVar {
		container = v3ioUsersContainer
		path = session.User + "/" + path
	}

	if strings.HasPrefix(path, v3ioHomeVar+"/") {
		container = v3ioUsersContainer
		path = session.User + "/" + path[len(v3ioHomeVar)+1:]
	}

	if addSlash && !strings.HasSuffix(path, "/") {
		path += "/"
	}

	return container, path, nil
}
