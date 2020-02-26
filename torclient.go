package torclient

import (
	"bytes"
	"github.com/shellrausch/virgo/pkg/options"
	"github.com/shellrausch/virgo/pkg/virgo"
	"golang.org/x/net/proxy"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

type TorClient struct {
	Tor *virgo.Client
}

func New() *TorClient {
	o := options.New()
	o.UserAgent = "Mozilla/5.0 (Windows NT 10.0; rv:68.0) Gecko/20100101 Firefox/68.0" // Default Tor user agent

	client := virgo.New()
	client.SetOptions(o)
	client.SetClient(initTORClient())

	torClient := &TorClient{}
	torClient.Tor = client

	return torClient
}

func (torClient *TorClient) CheckTorConnectivity() string {
	resultCh := make(chan *virgo.Result)
	go torClient.Tor.Start([]string{"https://check.torproject.org"}, resultCh)
	resp := <-resultCh

	if resp.Err != nil {
		log.Fatalln("Something went wrong on the Tor status site: ", resp.Err)
	}

	successMsg := "Congratulations. This browser is configured to use Tor."
	if !bytes.Contains(resp.Body, []byte(successMsg)) {
		log.Fatalln("The response body of the Tor status site does not contain a success message")
	}

	userAgentCheck := "However, it does not appear to be Tor Browser."
	if bytes.Contains(resp.Body, []byte(userAgentCheck)) {
		log.Fatalln("You are using the Tor network but the user agent is not Tor Browser")
	}

	rgx := regexp.MustCompile("[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}")
	ip := rgx.FindString(string(resp.Body))

	return ip
}

// stolen and extended from https://gist.github.com/Yawning/bac58e08a05fc378a8cc
func initTORClient() http.Client {
	// Create a transport that uses Tor Browser's SocksPort.  If
	// talking to a system tor, this may be an AF_UNIX socket, or
	// 127.0.0.1:9050 instead.
	tbProxyURL, err := url.Parse("socks5://127.0.0.1:9150")
	if err != nil {
		log.Fatalf("Failed to parse proxy URL: %v\n", err)
	}

	// Get a proxy Dialer that will create the connection on our
	// behalf via the SOCKS5 proxy.  Specify the authentication
	// and re-create the dialer/transport/client if tor's
	// IsolateSOCKSAuth is needed.
	tbDialer, err := proxy.FromURL(tbProxyURL, proxy.Direct)
	if err != nil {
		log.Fatalf("Failed to obtain proxy dialer: %v\n", err)
	}

	// Make a http.Transport that uses the proxy dialer, and a
	// http.Client that uses the transport.
	tbTransport := &http.Transport{Dial: tbDialer.Dial}
	client := http.Client{
		Transport: tbTransport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
		Timeout: time.Duration(10000) * time.Millisecond,
	}

	return client
}
