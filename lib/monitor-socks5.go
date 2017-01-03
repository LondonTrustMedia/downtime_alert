package lib

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/url"
	"strconv"
	"time"

	"net/http"

	"strings"

	"code.cloudfoundry.org/bytefmt"
	"github.com/DanielOaks/proxyclient"
	"github.com/LondonTrustMedia/downtime_alert/lib/slo"
)

// CheckSocks5 checks the given SOCKS5 proxy and returns an error if it doesn't work.
func CheckSocks5(tracker *slo.Tracker, config Socks5Config, credsToUse int) error {
	log.Println("Checking SOCKS5 proxy", config.Host)

	// assemble socks5 url
	userpass := url.UserPassword(config.Credentials[credsToUse].Username, config.Credentials[credsToUse].Password)
	socksURL := url.URL{
		Scheme: "socks5",
		User:   userpass,
		Host:   net.JoinHostPort(config.Host, strconv.Itoa(config.Port)),
	}

	// test proxy and connectivity through the proxy
	p, err := proxyclient.NewProxyClient(socksURL.String())
	if err != nil {
		return err
	}

	urlString := config.TestDownload.URL
	if strings.Contains(urlString, "{{random-int}}") {
		urlString = strings.Replace(urlString, "{{random-int}}", strconv.Itoa(rand.Int()), -1)
	}
	req, err := http.NewRequest("GET", urlString, nil)
	if err != nil {
		return err
	}
	if config.TestDownload.MaxBytesToDL > 0 {
		req.Header.Add("Range", fmt.Sprintf("bytes=0-%d", config.TestDownload.MaxBytesToDL))
	}

	u, err := url.Parse(config.TestDownload.URL)
	if err != nil {
		return err
	}

	var host string
	if strings.Contains(u.Host, ":") {
		// bleh
		host = u.Host
	} else {
		host = net.JoinHostPort(u.Host, "80")
	}

	conn, err := p.Dial("tcp", host)
	if err != nil {
		return err
	}

	// write our actual HTTP request
	req.Write(conn)

	// download
	downloadStartedTime := time.Now()
	defer conn.Close()
	body, err := ioutil.ReadAll(conn)
	if err != nil {
		return err
	}
	downloadEndedTime := time.Now()

	downloadElapsedSeconds := downloadEndedTime.Unix() - downloadStartedTime.Unix()
	downloadSizeBytes := uint64(len(body))
	downloadSpeedBytesPerSecond := downloadSizeBytes / uint64(downloadElapsedSeconds)

	log.Println("SOCKS5", config.Host, "- Downloaded", bytefmt.ByteSize(downloadSizeBytes), "in", downloadElapsedSeconds, "seconds --", fmt.Sprintf("%s/s", bytefmt.ByteSize(downloadSpeedBytesPerSecond)))

	tracker.AddDownload(downloadStartedTime, downloadSpeedBytesPerSecond)

	// no errors!
	return nil
}
