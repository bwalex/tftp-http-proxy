package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
	"github.com/pin/tftp"
)

const httpBaseUrlDefault = "http://127.0.0.1/tftp"
const tftpTimeoutDefault = 5 * time.Second
const tftpBindAddrDefault = ":69"
const appendPathDefault = false

var globalState = struct {
	httpBaseUrl	string
	httpClient	*http.Client
	appendPath	bool
}{
	httpBaseUrl:	httpBaseUrlDefault,
	httpClient:	nil,
	appendPath:	appendPathDefault,
}

func tftpReadHandler(filename string, rf io.ReaderFrom) error {
	raddr := rf.(tftp.OutgoingTransfer).RemoteAddr() // net.UDPAddr

	log.Printf("INFO: New TFTP request (%s) from %s", filename, raddr.IP.String())

	uri := globalState.httpBaseUrl
	if globalState.appendPath {
		// No need to validate url any further, http.NewRequest does
		// this for us using url.Parse().  We already checked that base
		// contains scheme and host and ends with a slash.  We assume
		// that appending filename does not change scheme, host and initial
		// part of path of URL.
		uri = uri + strings.TrimLeft(filename, "/")
	}

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		log.Printf("ERR: http request setup failed: %v", err)
		return err
	}
	req.Header.Add("X-TFTP-IP", raddr.IP.String())
	req.Header.Add("X-TFTP-Port", fmt.Sprintf("%d", raddr.Port))
	req.Header.Add("X-TFTP-File", filename)

	resp, err := globalState.httpClient.Do(req)
	if err != nil {
		log.Printf("ERR: http request failed: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("INFO: http FileNotFound response: %s", resp.Status)
		return fmt.Errorf("File not found")
	} else if resp.StatusCode != http.StatusOK {
		log.Printf("ERR: http request returned status %s", resp.Status)
		return fmt.Errorf("HTTP request error: %s", resp.Status)
	}

	// Use ContentLength, if provided, to set TSize option
	if resp.ContentLength >= 0 {
		rf.(tftp.OutgoingTransfer).SetSize(resp.ContentLength)
	}

	_, err = rf.ReadFrom(resp.Body)
	if err != nil {
		log.Printf("ERR: ReadFrom failed: %v", err)
		return err
	}

	return nil
}

func parseBaseURL(baseUrl string, appendPath bool) string {
	u, err := url.ParseRequestURI(baseUrl)
	if err != nil {
		log.Panicf("FATAL: invalid base URL: %v\n", err)
	}
	if (u.Scheme == "") {
		log.Panicf("FATAL: invalid base URL: No scheme found.\n")
	}
	if (u.Host == "") {
		log.Panicf("FATAL: invalid base URL: No host found.\n")
	}
	base := u.String()
	if appendPath && !strings.HasSuffix(base, "/") {
		return base + "/"
	} else {
		return base
	}
}

func main() {
	httpBaseUrlPtr := flag.String("http-base-url", httpBaseUrlDefault, "HTTP base URL")
	appendPathPtr := flag.Bool("http-append-path", appendPathDefault, "append TFTP filename to URL")
	tftpTimeoutPtr := flag.Duration("tftp-timeout", tftpTimeoutDefault, "TFTP timeout")
	bindAddrPtr := flag.String("tftp-bind-address", tftpBindAddrDefault, "TFTP addr to bind to")

	flag.Parse()

	globalState.httpBaseUrl = parseBaseURL(*httpBaseUrlPtr, *appendPathPtr)
	globalState.httpClient = &http.Client{}
	globalState.appendPath = *appendPathPtr

	s := tftp.NewServer(tftpReadHandler, nil)
	s.SetTimeout(*tftpTimeoutPtr)
	log.Printf("Listening TFTP requests on: %s", *bindAddrPtr)
	err := s.ListenAndServe(*bindAddrPtr)
	if err != nil {
		log.Panicf("FATAL: tftp server: %v\n", err)
	}
}
