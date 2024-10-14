package easy_json

import (
	"encoding/json"

	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	/* i love this library */
	backoff "github.com/cenkalti/backoff"
)

const (
	REQUEST_TIMEOUT = 20 * time.Second
)

type Form map[string]string
type RawForm []byte

/*
By Rodrigo Monteiro Junior
ter 10 set 2024 14:24:16 -03

-- FROM V0.2 --

Generic reposnse from the specific API i'm consuming,
again, don't really use this if you don't need it.
*/
type RespostaAPI struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	// TODO: Handle action field
	// Action
}

func CheckForSuccessMessage(body []byte) (err error) {

	//log.Println(string(body)) // XXX: Debugging

	/*
		By Rodrigo Monteiro Junior
		ter 10 set 2024 14:30:47 -03

		( NOTE: This is specific for my personal usage,
			if your api or whatever you're using that
			for doesn't return any 'status' like fields,
			please remove this thing. (actually, why are you
			using this?)
		)
		patch for checking a `status` response.
		(this is a nasty workaround for faster debugging)
	*/
	var check RespostaAPI

	err = json.Unmarshal(body, &check)

	if err != nil {
		/* FIXME: remove excessive loggin */
		log.Printf("WARN: Can't unmarshal response JSON into type %T\n", check)

		/* Can safely ignore this, since it's simply meant for error reporting */
	}

	if check.Status == "error" {
		err = fmt.Errorf("API returned error status: %s\n", check.Message)

		return
	}

	return
}

/*
	By Rodrigo Monteiro Junior
	Last modified: ter 17 set 2024 14:43:40 -03

Execute an HTTP POST request to a JSON api,
passing a JSON form with no response.

this function retries to make the request up to 20 seconds
using a backoff algorithm.
*/
func SimpleRawRequest(url string, data RawForm, contentType string) (err error) {

	var res *http.Response

	bf := backoff.NewExponentialBackOff()
	bf.MaxElapsedTime = REQUEST_TIMEOUT

	err = backoff.Retry(
		func() (err error) {
			req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))

			if err != nil {
				err = fmt.Errorf("Error creating request: %s\n", err)

				return
			}

			req.Header.Set("Content-Type", contentType)

			res, err = http.DefaultClient.Do(req)

			/* FIXME: remove excessive loggin */
			if err != nil {
				log.Println("Error sending request:", err)
			}

			return
		},

		bf,
	)

	if err != nil {
		return
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("Error connecting to '%s': got HTTP %d\n", url, res.StatusCode)

		return
	}

	body, err := io.ReadAll(res.Body)

	if err != nil {
		err = fmt.Errorf("Error reading response body: %s\n", err)

		return
	}

	// NOTE: You can comment out this section safely
	err = CheckForSuccessMessage(body)

	return
}

/*
	By Rodrigo Monteiro Junior
	Last modified in:                   • ter 27 ago 2024 15:30:34 -03
			  ( we're here -> ) • sex 13 set 2024 08:34:42 -03

	- Version 0.1:
		| receives a 'data' parameter
		| as a map[string]string.

	- Version 0.2 * (Modifications exclusive to Envio):
		| 'data' parameter renamed to jsonData
		| and now represents Marshalled json data.

	- Version 0.3
	  • sex 13 set 2024 08:34:34 -03
	  	| Moved to its separate function.
		| renamed to RawRequest and now takes a
	        | RawForm parameter in the form of a byte
		| slice.

Execute an HTTP POST request to a JSON api,
passing a JSON form and getting a response in
a user-defined struct.

this function retries to make the request up to 20 seconds
using a backoff algorithm.
*/
func RawRequest(url string, data RawForm, jsonOutput interface{}) (err error) {

	var res *http.Response

	bf := backoff.NewExponentialBackOff()
	bf.MaxElapsedTime = REQUEST_TIMEOUT

	err = backoff.Retry(
		func() (err error) {

			req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))

			if err != nil {
				err = fmt.Errorf("Error creating request: %s\n", err)

				return
			}

			req.Header.Set("Content-Type", "application/json")

			res, err = http.DefaultClient.Do(req)

			/* FIXME: remove excessive loggin */
			if err != nil {
				log.Println("Error sending request:", err)
			}

			return
		},

		bf,
	)

	if err != nil {
		return
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("Error connecting to '%s': got HTTP %d\n", url, res.StatusCode)

		return
	}

	body, err := io.ReadAll(res.Body)

	if err != nil {
		err = fmt.Errorf("Error reading response body: %s\n", err)

		return
	}

	//log.Println(string(body))

	// NOTE: You can comment out this section safely
	err = CheckForSuccessMessage(body)

	if err != nil {
		return
	}

	err = json.Unmarshal(body, &jsonOutput)

	if err != nil {
		err = fmt.Errorf("Error unmarshaling response JSON: %s\n", err)
	}

	return
}

/*
	--TODO--: this request module is getting too big, time
	to turn it into a repo/gist or own project.
	NOTE: Done that.

	By Rodrigo Monteiro Junior

	- Version 0.1:

		receives a 'data' parameter
		as a map[string]string.

	- Version 0.1.1:
	  ter 10 set 2024 14:24:16 -03

		Minor patch to check errors
		related to the `status` field.

	- Version 0.2 * ( FIXME: CONFLICTING patch with 'github.com/mytempoesp/Envio/request.go' ):
	  ter 10 set 2024 15:07:49 -03

		Major patch to error reporting, won't affect
		much usage, but conform to proper idiomatic
		error handling and avoid redundancy.

Execute an HTTP POST request to a JSON api,
passing a JSON form and getting a response in
a user-defined struct.

this function retries to make the request up to 20 seconds
using a backoff algorithm.
*/
func JSONRequest(url string, data Form, jsonOutput interface{}) (err error) {

	var res *http.Response

	jsonData, err := json.Marshal(data)

	if err != nil {
		err = fmt.Errorf("Error marshaling JSON: %s\n", err)

		return
	}

	bf := backoff.NewExponentialBackOff()
	bf.MaxElapsedTime = REQUEST_TIMEOUT

	err = backoff.Retry(
		func() (err error) {

			req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))

			if err != nil {
				err = fmt.Errorf("Error creating request: %s\n", err)

				return
			}

			req.Header.Set("Content-Type", "application/json")

			res, err = http.DefaultClient.Do(req)

			if err != nil {
				/* FIXME: remove excessive loggin */
				log.Println("Error sending request:", err)
			}

			return
		},

		bf,
	)

	if err != nil {
		return
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("Error connecting to '%s': got HTTP %d", url, res.StatusCode)

		return
	}

	body, err := io.ReadAll(res.Body)

	if err != nil {
		err = fmt.Errorf("Error reading response body: %s\n", err)

		return
	}

	// NOTE: You can comment out this section safely
	err = CheckForSuccessMessage(body)

	if err != nil {
		return
	}

	err = json.Unmarshal(body, &jsonOutput)

	if err != nil {
		err = fmt.Errorf("Error unmarshaling response JSON: %s\n", err)
	}

	return
}

/*
By Rodrigo Monteiro Junior
sex 13 set 2024 08:36:49 -03

Do a simple POST request and ignore the response,
only treat it in case of errors etc.

Version 0.2:
  - Better error handling ( i don't care much about this function )
*/
func JSONSimpleRequest(url string, data Form) (err error) {

	var res *http.Response

	jsonData, err := json.Marshal(data)

	if err != nil {
		err = fmt.Errorf("Error marshaling JSON: %s\n", err)

		return
	}

	bf := backoff.NewExponentialBackOff()
	bf.MaxElapsedTime = REQUEST_TIMEOUT

	err = backoff.Retry(
		func() (err error) {

			req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))

			if err != nil {
				err = fmt.Errorf("Error creating request: %s\n", err)

				return
			}

			req.Header.Set("Content-Type", "application/json")

			res, err = http.DefaultClient.Do(req)

			if err != nil {
				/* FIXME: remove excessive loggin */
				log.Println("Error sending request:", err)
			}

			return
		},

		bf,
	)

	if err != nil {
		err = fmt.Errorf("Error doing request (%s elapsed): %s\n", bf.MaxElapsedTime, err)

		return
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("Error connecting to '%s': got HTTP %d", url, res.StatusCode)
	}

	return
}
