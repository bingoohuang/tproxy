package protocol

import (
	"bufio"
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
)

type (
	httpInterop struct {
	}
)

func newHttpInterop() *httpInterop {
	return &httpInterop{}
}

func (i *httpInterop) Dump(r io.Reader, source string, id int, quiet bool) {
	buf := bufio.NewReader(r)
	switch source {
	case ClientSide:
		for {
			req, err := http.ReadRequest(buf)
			if err != nil {
				log.Printf("[%d] read request error: %v", id, err)
				return
			}

			dump, err := httputil.DumpRequest(req, true)
			if err != nil {
				log.Printf("[%d] dump request error: %v", id, err)
				continue
			}

			if !quiet {
				log.Printf("[%d] Request >>>", id)
				log.Printf("Request 【%s】", dump)
			}
		}

	case ServerSide:
		for {
			rsp, err := http.ReadResponse(buf, nil)
			if err != nil {
				log.Printf("[%d] read response error: %v", id, err)
				return
			}

			dump, err := httputil.DumpResponse(rsp, false)
			if err != nil {
				log.Printf("[%d] dump response error: %v", id, err)
				continue
			}

			if quiet {
				io.Copy(io.Discard, rsp.Body)
			} else {
				log.Printf("[%d] Response <<<", id)
				log.Printf("Response Header【%s】", dump)

				if rsp.Header.Get("Content-Encoding") == "gzip" {
					log.Printf("gzip gzip")
					rsp.Body, err = gzip.NewReader(rsp.Body)
					if err != nil {
						log.Printf("[%d] gzip read error: %v\n", id, err)
					}
				}

				plainBody, _ := io.ReadAll(rsp.Body)
				log.Printf("Response PlainBody【%s】", plainBody)
			}
		}
	}
}
