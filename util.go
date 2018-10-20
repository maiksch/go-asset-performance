package performance

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

func readBody(body io.ReadCloser, object interface{}) error {
	read, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	defer body.Close()

	err = json.Unmarshal(read, &object)
	if err != nil {
		return err
	}

	return nil
}
