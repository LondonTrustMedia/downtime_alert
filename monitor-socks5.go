package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"strconv"

	"github.com/DanielOaks/proxyclient"
)

// CheckSocks5 checks the given SOCKS5 proxy and returns an error if it doesn't work.
func CheckSocks5(config Socks5Config) error {
	log.Println("Checking SOCKS5 proxy", config.Host)

	// assemble socks5 url
	userpass := url.UserPassword(config.Username, config.Password)
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

	c, err := p.Dial("tcp", net.JoinHostPort(config.TestDomain, "80"))
	if err != nil {
		return err
	}

	io.WriteString(c, fmt.Sprintf("GET / HTTP/1.0\r\nHOST:%s\r\n\r\n", config.TestDomain))
	_, err = ioutil.ReadAll(c)
	if err != nil {
		return err
	}

	// we don't care about whatever http errors we get, just that we actually get a response
	return nil
}
