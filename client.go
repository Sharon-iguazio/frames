package frames

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/nuclio/logger"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack"
)

// Client is v3io streaming client
type Client struct {
	URL    string
	apiKey string
	logger logger.Logger
}

// NewClient returns a new client
func NewClient(url string, apiKey string, logger logger.Logger) (*Client, error) {
	var err error
	if logger == nil {
		logger, err = NewLogger("info")
		if err != nil {
			return nil, errors.Wrap(err, "Can't create logger")
		}
	}

	client := &Client{
		URL:    url,
		apiKey: apiKey,
		logger: logger,
	}

	return client, nil
}

// Query runs a query on the client
func (c *Client) Query(query string) (chan *Message, error) {
	queryObj := map[string]interface{}{
		"query":   query,
		"limit":   100,
		"columns": []string{"first", "last"},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(queryObj); err != nil {
		return nil, errors.Wrap(err, "can't encode query")
	}

	req, err := http.NewRequest("POST", c.URL, &buf)
	if err != nil {
		return nil, errors.Wrap(err, "can't create HTTP request")
	}
	req.Header.Set("Autohrization", c.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "can't call API")
	}

	ch := make(chan *Message) // TODO: Buffered channel?

	go func() {
		defer resp.Body.Close()
		dec := msgpack.NewDecoder(resp.Body)
		for {
			msg := &Message{}
			if err := dec.Decode(msg); err != nil {
				// TODO: log
				return
			}
			if err != nil {
				// TODO: log
			}
			ch <- msg
		}
	}()

	return ch, nil
}
