package app

import (
	"bytes"
	"context"
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"github.com/webitel/engine/model"
	"net"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/http2"
)

// Apple HTTP/2 Development & Production urls
const (
	HostDevelopment = "https://api.sandbox.push.apple.com"
	HostProduction  = "https://api.push.apple.com"
)

// DefaultHost is a mutable var for testing purposes
var DefaultHost = HostDevelopment

var (
	// HTTPClientTimeout specifies a time limit for requests made by the
	// HTTPClient. The timeout includes connection time, any redirects,
	// and reading the response body.
	HTTPClientTimeout = 60 * time.Second

	// ReadIdleTimeout is the timeout after which a health check using a ping
	// frame will be carried out if no frame is received on the connection. If
	// zero, no health check is performed.
	ReadIdleTimeout = 15 * time.Second

	// TCPKeepAlive specifies the keep-alive period for an active network
	// connection. If zero, keep-alive probes are sent with a default value
	// (currently 15 seconds)
	TCPKeepAlive = 15 * time.Second

	// TLSDialTimeout is the maximum amount of time a dial will wait for a connect
	// to complete.
	TLSDialTimeout = 20 * time.Second
)

// DialTLS is the default dial function for creating TLS connections for
// non-proxied HTTPS requests.
var DialTLS = func(network, addr string, cfg *tls.Config) (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout:   TLSDialTimeout,
		KeepAlive: TCPKeepAlive,
	}
	return tls.DialWithDialer(dialer, network, addr, cfg)
}

// Client represents a connection with the APNs
type ApnClient struct {
	Host          string
	CustomHeaders map[string]string
	Certificate   tls.Certificate
	HTTPClient    *http.Client
}

type ApnMessage struct {
	Priority int               `json:"priority"`
	Aps      map[string]string `json:"aps"`
}

func ApnCertificate(config *model.PushConfig) (tls.Certificate, model.AppError) {
	blockPEM := func(data []byte, typ string) (block *pem.Block) {
		block, data = pem.Decode(data)
		if block == nil {
			block = &pem.Block{
				Type:  typ,
				Bytes: data,
			}
		}
		return
	}

	var appErr model.AppError
	var certData []byte
	var keyData []byte

	if certData, appErr = readFile(config.ApnCertFile); appErr != nil {
		return tls.Certificate{}, appErr
	}

	if keyData, appErr = readFile(config.ApnKeyFile); appErr != nil {
		return tls.Certificate{}, appErr
	}

	certDER := blockPEM(certData, "CERTIFICATE")
	pkeyDER := blockPEM(keyData, "PRIVATE KEY")
	if len(pkeyDER.Bytes) == 0 {
		return tls.Certificate{}, model.NewInternalError("app.apn.valid.private_file", "tls pkey required")
	}

	pkey, err := parsePrivateKey(pkeyDER.Bytes)
	if err != nil {
		return tls.Certificate{}, model.NewInternalError("app.apn.config.private", err.Error())
	}
	if len(certDER.Bytes) == 0 {
		return tls.Certificate{}, model.NewInternalError("app.apn.valid.cert_file", "tls cert required")
	}
	cert := tls.Certificate{
		Certificate: [][]byte{
			certDER.Bytes,
		},
		PrivateKey: pkey,
	}
	cert.Leaf, err = x509.ParseCertificate(
		certDER.Bytes,
	)

	if err != nil {

	}

	return cert, nil
}

// NewClient returns a new Client with an underlying http.Client configured with
// the correct APNs HTTP/2 transport settings. It does not connect to the APNs
// until the first Notification is sent via the Push method.
//
// As per the Apple APNs Provider API, you should keep a handle on this client
// so that you can keep your connections with APNs open across multiple
// notifications; don’t repeatedly open and close connections. APNs treats rapid
// connection and disconnection as a denial-of-service attack.
//
// If your use case involves multiple long-lived connections, consider using
// the ClientManager, which manages clients for you.
func NewApnClient(certificate tls.Certificate, headers map[string]string) *ApnClient {
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
	}
	if len(certificate.Certificate) > 0 {
		tlsConfig.BuildNameToCertificate()
	}
	transport := &http2.Transport{
		TLSClientConfig: tlsConfig,
		DialTLS:         DialTLS,
		ReadIdleTimeout: ReadIdleTimeout,
	}
	return &ApnClient{
		HTTPClient: &http.Client{
			Transport: transport,
			Timeout:   HTTPClientTimeout,
		},
		Certificate:   certificate,
		Host:          DefaultHost,
		CustomHeaders: headers,
	}
}

func (apn *ApnClient) Push(ctx context.Context, device string, r *model.SendPush) model.AppError {
	data, err := json.Marshal(ApnMessage{
		Priority: int(r.Priority),
		Aps:      r.Data,
	})
	if err != nil {
		return model.NewInternalError("app.apn.marshal", err.Error())
	}
	req, err := http.NewRequest(http.MethodPost, apn.Host+"/3/device/"+device, bytes.NewBuffer(data))
	if err != nil {
		return model.NewInternalError("app.apn.request.prepare", err.Error())
	}
	req.WithContext(ctx)

	for k, v := range apn.CustomHeaders {
		req.Header.Add(k, v)
	}

	if r.Expiration > 0 {
		//req.Header.Add("apns-expiration", strconv.Itoa(int(model.GetMillis())+int(r.Expiration)))
	}

	res, err := apn.HTTPClient.Do(req)
	if err != nil {
		return model.NewInternalError("app.apn.request.send", err.Error())
	}

	defer res.Body.Close()

	return nil
}

// Development sets the Client to use the APNs development push endpoint.
func (c *ApnClient) Development() *ApnClient {
	c.Host = HostDevelopment
	return c
}

// Production sets the Client to use the APNs production push endpoint.
func (c *ApnClient) Production() *ApnClient {
	c.Host = HostProduction
	return c
}

func parsePrivateKey(bytes []byte) (crypto.PrivateKey, error) {
	var key crypto.PrivateKey
	var err error
	var privPemBytes = bytes

	key, err = x509.ParsePKCS1PrivateKey(privPemBytes)
	if err == nil {
		return key, nil
	}
	key, err = x509.ParsePKCS8PrivateKey(privPemBytes)
	if err == nil {
		return key, nil
	}

	return nil, model.NewInternalError("app.cert.parse", "failed to parse private key")
}

func readFile(loc string) ([]byte, model.AppError) {
	data, err := os.ReadFile(loc)
	if err != nil {
		return nil, model.NewInternalError("app.file.read", err.Error())
	}

	return data, nil
}
