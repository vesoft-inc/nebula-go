// Copyright 2025 vesoft inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// 
//     http://www.apache.org/licenses/LICENSE-2.0
// 
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package grpcutil

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"math"
	"time"

	"github.com/vesoft-inc/nebula-go/v5/internal/internal_error"
	"google.golang.org/grpc"
	grpccodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	grpcstatus "google.golang.org/grpc/status"
)

var defaultMsgSize = math.MaxInt64

func NewGrpcClient(address string, timeout time.Duration, tlsCfg *tls.Config) (*grpc.ClientConn, error) {
	var (
		err      error
		grpcConn *grpc.ClientConn
		cred     grpc.DialOption
	)

	if tlsCfg != nil {
		cred = grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg))
	} else {
		cred = grpc.WithInsecure()
	}
	duration := time.Duration(timeout)
	grpcConn, err = grpc.NewClient(address, cred, grpc.WithBlock(), grpc.WithTimeout(duration),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(defaultMsgSize), grpc.MaxCallRecvMsgSize(defaultMsgSize)))
	if err != nil {
		return nil, internal_error.ErrConnCannotOpen(address, err.Error())
	}
	return grpcConn, nil
}

func NewTLSConfig(host string, ca, cert, key string, insecureSkipVerify bool) (*tls.Config, error) {
	if insecureSkipVerify {
		tlsCfg := &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS12,
		}
		if cert != "" || key != "" {
			if cert, err := tls.LoadX509KeyPair(cert, key); err != nil {
				return nil, internal_error.ErrTLS(err.Error())
			} else {
				tlsCfg.Certificates = []tls.Certificate{cert}
			}
		}
		return tlsCfg, nil
	}

	tlsCfg := &tls.Config{
		ServerName: host,
		MinVersion: tls.VersionTLS12,
	}
	if ca == "" {
		return nil, internal_error.ErrTLS("No CA certificate provide")
	}

	CAs := x509.NewCertPool()
	if ca, err := ioutil.ReadFile(ca); err == nil {
		if !CAs.AppendCertsFromPEM(ca) {
			return nil, internal_error.ErrTLS("AppendCertsFromPEM failed")
		}
		tlsCfg.RootCAs = CAs
	} else {
		return nil, internal_error.ErrTLS(err.Error())
	}

	if cert != "" || key != "" {
		if cert, err := tls.LoadX509KeyPair(cert, key); err != nil {
			return nil, internal_error.ErrTLS(err.Error())
		} else {
			tlsCfg.Certificates = []tls.Certificate{cert}
		}
	}

	tlsCfg.VerifyPeerCertificate = func(certificates [][]byte, _ [][]*x509.Certificate) error {
		certs := make([]*x509.Certificate, len(certificates))
		for i, data := range certificates {
			cert, err := x509.ParseCertificate(data)
			if err != nil {
				return internal_error.ErrTLS(err.Error())
			}
			certs[i] = cert
		}

		opts := x509.VerifyOptions{
			Roots:         tlsCfg.RootCAs,
			DNSName:       tlsCfg.ServerName,
			Intermediates: x509.NewCertPool(),
		}

		for _, cert := range certs[1:] {
			opts.Intermediates.AddCert(cert)
		}

		_, err := certs[0].Verify(opts)
		if err != nil {
			return internal_error.ErrTLS(err.Error())
		}
		return nil
	}

	return tlsCfg, nil
}

func GetCtxDuration(ctx context.Context) time.Duration {
	deadline, ok := ctx.Deadline()
	if !ok {
		return 0
	}
	return time.Until(deadline)
}

func GetGrpcError(address string, err error, duration time.Duration) error {
	rpcErr, ok := grpcstatus.FromError(err)
	if !ok {
		return err
	}
	switch rpcErr.Code() {
	case grpccodes.DeadlineExceeded, grpccodes.Canceled:
		return internal_error.ErrConnRequestTimeout(address, err.Error(), duration)
	case grpccodes.Unavailable:
		return internal_error.ErrConnUnavailable(address, err.Error())
	}
	return err
}

var defaultTimeout = 3 * time.Second
