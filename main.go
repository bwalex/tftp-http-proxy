package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
	"github.com/pin/tftp"
)

const httpBaseUrlDefault = "http://127.0.0.1/tftp"
const tftpTimeoutDefault = 5 * time.Second

var globalState = struct {
	httpBaseUrl	string
	httpClient	*http.Client
}{
	httpBaseUrl:	httpBaseUrlDefault,
	httpClient:	nil,
}

func tftpReadHandler(filename string, rf io.ReaderFrom) error {
	raddr := rf.(tftp.OutgoingTransfer).RemoteAddr() // net.UDPAddr

	log.Printf("INFO: New TFTP request (%s) from %s", filename, raddr.IP.String())

	uri := globalState.httpBaseUrl

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

func main() {
	httpBaseUrlPtr := flag.String("http-base-url", httpBaseUrlDefault, "HTTP base URL")
	tftpTimeoutPtr := flag.Duration("tftp-timeout", tftpTimeoutDefault, "TFTP timeout")

	flag.Parse()

	globalState.httpBaseUrl = *httpBaseUrlPtr
	globalState.httpClient = &http.Client{}

	s := tftp.NewServer(tftpReadHandler, nil)
	s.SetTimeout(*tftpTimeoutPtr)
	err := s.ListenAndServe(":69")
	if err != nil {
		log.Fatalf("FATAL: tftp server: %v\n", err)
	}
}
