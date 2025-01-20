/*
 *
 * Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License.
 *
 */

package nebula_go

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/vesoft-inc/nebula-go/v3/nebula"
	"github.com/vesoft-inc/nebula-go/v3/nebula/graph"
	"golang.org/x/net/http2"
)

type connection struct {
	severAddress HostAddress
	timeout      time.Duration
	returnedAt   time.Time // the connection was created or returned.
	sslConfig    *tls.Config
	useHTTP2     bool
	httpHeader   http.Header
	handshakeKey string
	graph        *graph.GraphServiceClient
	transport    thrift.TTransport
}

func newConnection(severAddress HostAddress) *connection {
	return &connection{
		severAddress: severAddress,
		timeout:      0 * time.Millisecond,
		returnedAt:   time.Now(),
		sslConfig:    nil,
		handshakeKey: "",
		graph:        nil,
		transport:    nil,
	}
}

// create socket based transport
func getTransportAndProtocolFactory(hostAddress HostAddress, timeout time.Duration, sslConfig *tls.Config) (thrift.TTransport, thrift.TProtocolFactory, error) {
	newAdd := net.JoinHostPort(hostAddress.Host, strconv.Itoa(hostAddress.Port))

	var transport thrift.TTransport
	var pf thrift.TProtocolFactory
	var sock thrift.TTransport
	if sslConfig != nil {
		sock = thrift.NewTSSLSocketConf(newAdd, &thrift.TConfiguration{
			ConnectTimeout: timeout, // Use 0 for no timeout
			SocketTimeout:  timeout, // Use 0 for no timeout

			TLSConfig: sslConfig,
		})

		//sock, err = thrift.NewTSSLSocketTimeout(newAdd, sslConfig, timeout, timeout)
	} else {
		sock = thrift.NewTSocketConf(newAdd, &thrift.TConfiguration{
			ConnectTimeout: timeout, // Use 0 for no timeout
			SocketTimeout:  timeout, // Use 0 for no timeout
		})
		//sock, err = thrift.NewTSocketTimeout(newAdd, timeout, timeout)
	}

	// Set transport
	bufferSize := 128 << 10
	bufferedTransFactory := thrift.NewTBufferedTransportFactory(bufferSize)
	buffTransport, err := bufferedTransFactory.GetTransport(sock)
	if err != nil {
		return nil, nil, err
	}

	transport = thrift.NewTHeaderTransport(buffTransport)

	//pf = thrift.NewTHeaderProtocolFactory()
	pf = thrift.NewTHeaderProtocolFactoryConf(
		&thrift.TConfiguration{})

	return transport, pf, nil
}

func getTransportAndProtocolFactoryForHttp2(hostAddress HostAddress, sslConfig *tls.Config, httpHeader http.Header) (thrift.TTransport, thrift.TProtocolFactory, error) {

	newAdd := net.JoinHostPort(hostAddress.Host, strconv.Itoa(hostAddress.Port))
	var (
		err       error
		transport thrift.TTransport
		pf        thrift.TProtocolFactory
	)

	if sslConfig != nil {
		transport, err = thrift.NewTHttpClientWithOptions("https://"+newAdd,
			thrift.THttpClientOptions{
				Client: &http.Client{
					Transport: &http2.Transport{
						TLSClientConfig: sslConfig,
					},
				},
			})
	} else {
		transport, err = thrift.NewTHttpClientWithOptions("https://"+newAdd, thrift.THttpClientOptions{
			Client: &http.Client{
				Transport: &http2.Transport{
					// So http2.Transport doesn't complain the URL scheme isn't 'https'
					AllowHTTP: true,
					// Pretend we are dialing a TLS endpoint. (Note, we ignore the passed tls.Config)
					DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
						_ = cfg
						var d net.Dialer
						return d.DialContext(ctx, network, addr)
					},
				},
			},
		})

	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create a net.Conn-backed Transport,: %s", err.Error())
	}

	//pf = thrift.NewTBinaryProtocolFactoryDefault()
	pf = thrift.NewTBinaryProtocolFactoryConf(&thrift.TConfiguration{})

	if httpHeader != nil {
		client, ok := transport.(*thrift.THttpClient)
		if !ok {
			return nil, nil, fmt.Errorf("failed to get thrift http client")
		}
		for k, vv := range httpHeader {
			if k == "Content-Type" {
				// fbthrift will add "Content-Type" header, so we need to skip it
				continue
			}
			for _, v := range vv {
				// fbthrift set header with http.Header.Add, so we need to set header one by one
				client.SetHeader(k, v)
			}
		}
	}

	return transport, pf, nil
}

// open opens transport for the connection
// if sslConfig is not nil, an SSL transport will be created
func (cn *connection) open(ctx context.Context, hostAddress HostAddress, timeout time.Duration, sslConfig *tls.Config,
	useHTTP2 bool, httpHeader http.Header, handshakeKey string) error {
	cn.timeout = timeout
	cn.useHTTP2 = useHTTP2
	cn.handshakeKey = handshakeKey

	var (
		err       error
		transport thrift.TTransport
		pf        thrift.TProtocolFactory
	)

	if useHTTP2 {
		transport, pf, err = getTransportAndProtocolFactoryForHttp2(hostAddress, sslConfig, httpHeader)
	} else {
		transport, pf, err = getTransportAndProtocolFactory(hostAddress, timeout, sslConfig)
	}

	cn.transport = transport
	cn.graph = graph.NewGraphServiceClientFactory(transport, pf)

	if err = cn.transport.Open(); err != nil {
		return fmt.Errorf("failed to open transport, error: %s", err.Error())
	}

	if !cn.transport.IsOpen() {
		return fmt.Errorf("transport is off")
	}

	return cn.verifyClientVersion(ctx)
}

func (cn *connection) verifyClientVersion(ctx context.Context) error {
	req := graph.NewVerifyClientVersionReq()
	if cn.handshakeKey != "" {
		req.Version = []byte(cn.handshakeKey)
	}
	resp, err := cn.graph.VerifyClientVersion(ctx, req)
	if err != nil {
		cn.close()
		return fmt.Errorf("failed to verify client handshakeKey: %s", err.Error())
	}
	if resp.GetErrorCode() != nebula.ErrorCode_SUCCEEDED {
		return fmt.Errorf("incompatible handshakeKey between client and server: %s", string(resp.GetErrorMsg()))
	}
	return nil
}

// reopen reopens the current connection.
// Because the code generated by Fbthrift does not handle the seqID,
// the message will be dislocated when the timeout occurs, resulting in unexpected response.
// When the timeout occurs, the connection will be reopened to avoid the impact of the message.
func (cn *connection) reopen(ctx context.Context) error {
	cn.close()
	return cn.open(ctx, cn.severAddress, cn.timeout, cn.sslConfig, cn.useHTTP2, cn.httpHeader, cn.handshakeKey)
}

// Authenticate
func (cn *connection) authenticate(ctx context.Context, username, password string) (*graph.AuthResponse, error) {
	resp, err := cn.graph.Authenticate(ctx, []byte(username), []byte(password))
	if err != nil {
		err = fmt.Errorf("authentication fails, %s", err.Error())

		if e := cn.transport.Close(); e != nil {
			err = fmt.Errorf("fail to close transport, error: %s", e.Error())
		}
		return nil, err
	}

	return resp, nil
}

func (cn *connection) execute(ctx context.Context, sessionID int64, stmt string) (*graph.ExecutionResponse, error) {
	return cn.executeWithParameter(ctx, sessionID, stmt, map[string]*nebula.Value{})
}

func (cn *connection) executeWithParameter(ctx context.Context, sessionID int64, stmt string,
	params map[string]*nebula.Value) (*graph.ExecutionResponse, error) {
	resp, err := cn.graph.ExecuteWithParameter(ctx, sessionID, []byte(stmt), params)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (cn *connection) executeWithParameterTimeout(ctx context.Context, sessionID int64, stmt string, params map[string]*nebula.Value, timeoutMs int64) (*graph.ExecutionResponse, error) {
	return cn.executeWithParameterTimeoutDuration(ctx, sessionID, stmt, params, time.Duration(timeoutMs)*time.Millisecond)
}

func (cn *connection) executeWithParameterTimeoutDuration(ctx context.Context, sessionID int64, stmt string, params map[string]*nebula.Value, timeout time.Duration) (*graph.ExecutionResponse, error) {
	ctxWithTimeout, _ := context.WithTimeout(ctx, timeout)
	return cn.graph.ExecuteWithParameter(ctxWithTimeout, sessionID, []byte(stmt), params)
}

func (cn *connection) executeJson(ctx context.Context, sessionID int64, stmt string) ([]byte, error) {
	return cn.ExecuteJsonWithParameter(ctx, sessionID, stmt, map[string]*nebula.Value{})
}

func (cn *connection) ExecuteJsonWithParameter(ctx context.Context, sessionID int64, stmt string, params map[string]*nebula.Value) ([]byte, error) {
	jsonResp, err := cn.graph.ExecuteJsonWithParameter(ctx, sessionID, []byte(stmt), params)
	if err != nil {
		// reopen the connection if timeout
		var TTransportException thrift.TTransportException
		if errors.As(err, &TTransportException) {
			if err.(thrift.TTransportException).TypeId() == thrift.TIMED_OUT {
				reopenErr := cn.reopen(ctx)
				if reopenErr != nil {
					return nil, reopenErr
				}
				return cn.graph.ExecuteJsonWithParameter(ctx, sessionID, []byte(stmt), params)
			}
		}
	}

	return jsonResp, err
}

// Check connection to host address
func (cn *connection) ping(ctx context.Context) bool {
	_, err := cn.execute(ctx, 0, "YIELD 1")
	return err == nil
}

// Sign out and release session ID
func (cn *connection) signOut(ctx context.Context, sessionID int64) error {
	// Release session ID to graphd
	return cn.graph.Signout(ctx, sessionID)
}

// Update returnedAt for cleaner
func (cn *connection) release() {
	cn.returnedAt = time.Now()
}

// Close transport
func (cn *connection) close() error {
	if e := cn.transport.Close(); e != nil {
		err := fmt.Errorf("fail to close transport, error: %s", e.Error())

		return err
	}
	return nil
}
