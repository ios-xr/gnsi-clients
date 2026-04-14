// Copyright 2024 Cisco Systems, Inc.
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

package connection

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Config holds connection configuration
type Config struct {
	TargetAddr string
	TargetName string
	Timeout    time.Duration

	// TLS configuration
	InsecureSkipVerify bool

	// Certificate-based authentication (auto-detects mTLS if both provided)
	CACert     string
	ClientCert string
	ClientKey  string

	// Username/password authentication
	Username string
	Password string
}

// New creates a new gRPC connection based on the configuration
// Auto-detects authentication mode:
//   - If clientCert and clientKey are provided → mTLS
//   - Otherwise → TLS (server-side only)
//
// Server will error if it requires mTLS and client doesn't provide certificates
func New(cfg *Config) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption

	// Set timeout
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	// Auto-detect mTLS vs TLS based on client certificate presence
	if cfg.ClientCert != "" && cfg.ClientKey != "" {
		// Mutual TLS - client certificate provided
		tlsConfig, err := loadMTLSConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to load mTLS config: %w", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		// TLS only - server-side TLS
		tlsConfig, err := loadTLSConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS config: %w", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	}

	// Add username/password for AAA authentication if provided
	// If not provided, username will be extracted from certificate CN
	if cfg.Username != "" && cfg.Password != "" {
		opts = append(opts, grpc.WithPerRPCCredentials(&loginCreds{
			Username: cfg.Username,
			Password: cfg.Password,
			Insecure: cfg.InsecureSkipVerify,
		}))
	}

	// Create connection
	conn, err := grpc.Dial(cfg.TargetAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %w", cfg.TargetAddr, err)
	}

	return conn, nil
}

// loadMTLSConfig loads mutual TLS configuration
func loadMTLSConfig(cfg *Config) (*tls.Config, error) {
	// Load client certificate and key
	clientCert, err := tls.LoadX509KeyPair(cfg.ClientCert, cfg.ClientKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{clientCert},
		InsecureSkipVerify: cfg.InsecureSkipVerify,
	}

	// Load CA certificate if provided
	if cfg.CACert != "" {
		caCert, err := ioutil.ReadFile(cfg.CACert)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to append CA certificate")
		}

		tlsConfig.RootCAs = caCertPool
	}

	// Set server name if provided
	if cfg.TargetName != "" {
		tlsConfig.ServerName = cfg.TargetName
	}

	return tlsConfig, nil
}

// loadTLSConfig loads TLS configuration (server-side TLS only)
func loadTLSConfig(cfg *Config) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: cfg.InsecureSkipVerify,
	}

	// Load CA certificate if provided (for server verification)
	if cfg.CACert != "" {
		caCert, err := ioutil.ReadFile(cfg.CACert)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to append CA certificate")
		}

		tlsConfig.RootCAs = caCertPool
	}

	// Set server name if provided
	if cfg.TargetName != "" {
		tlsConfig.ServerName = cfg.TargetName
	}

	return tlsConfig, nil
}

// loginCreds implements credentials.PerRPCCredentials for username/password auth
type loginCreds struct {
	Username string
	Password string
	Insecure bool
}

func (c *loginCreds) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"username": c.Username,
		"password": c.Password,
	}, nil
}

func (c *loginCreds) RequireTransportSecurity() bool {
	return !c.Insecure
}
