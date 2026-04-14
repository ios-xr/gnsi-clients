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

package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/cisco/gnsi-certz-client/pkg/connection"
	"github.com/cisco/gnsi-certz-client/pkg/operations"
	certzpb "github.com/openconfig/gnsi/certz"
)

var (
	// Connection flags
	targetAddr = flag.String("target_addr", "localhost:9339", "Target address (host:port)")
	targetName = flag.String("target_name", "", "Target server name for TLS verification")
	timeout    = flag.Duration("timeout", 30*time.Second, "Operation timeout")

	// TLS configuration
	insecureSkipVerify = flag.Bool("insecure_skip_verify", false, "Skip TLS certificate verification (testing only)")

	// Certificate-based authentication (auto-detects mTLS if provided)
	caCert     = flag.String("ca_cert", "", "CA certificate file")
	clientCert = flag.String("client_cert", "", "Client certificate file (or chain: leaf + intermediates)")
	clientKey  = flag.String("client_key", "", "Client private key file (for leaf certificate)")

	// Username/password authentication
	username = flag.String("username", "", "Username for authentication")
	password = flag.String("password", "", "Password for authentication")

	// Operation flags
	operation = flag.String("op", "get-profile-list", "Operation: add-profile, delete-profile, get-profile-list, can-generate-csr, rotate")

	// Profile flags
	profileID     = flag.String("profile_id", "", "SSL profile ID")
	fromProfileID = flag.String("from_profile_id", "", "Source profile ID (for copying)")

	// Rotation flags
	forceOverwrite = flag.Bool("force_overwrite", true, "Force overwrite existing version")
	version        = flag.String("version", "1.0", "Entity version")
	createdOn      = flag.Uint64("created_on", uint64(time.Now().Unix()), "Creation timestamp")

	// Certificate rotation flags
	certFile   = flag.String("cert_file", "", "Certificate file (PEM) - can be a chain (leaf + intermediates)")
	keyFile    = flag.String("key_file", "", "Private key file (PEM) - for leaf certificate")
	certSource = flag.String("cert_source", "", "Device certificate source: IDevID, oIDevID, self-signed")

	// CA bundle flags
	caBundleFile = flag.String("ca_bundle_file", "", "CA bundle file (PEM)")

	// CRL bundle flags
	crlBundleFile = flag.String("crl_bundle_file", "", "CRL bundle file (PEM)")

	// CSR generation flags
	generateCSR  = flag.Bool("generate_csr", false, "Request device to generate CSR")
	csrSuite     = flag.String("csr_suite", "rsa-2048-sha256", "CSR suite. Options: rsa-2048-sha256, rsa-2048-sha384, rsa-2048-sha512, rsa-3072-sha256, rsa-3072-sha384, rsa-3072-sha512, rsa-4096-sha256, rsa-4096-sha384, rsa-4096-sha512, ecdsa-p256-sha256, ecdsa-p256-sha384, ecdsa-p256-sha512, ecdsa-p384-sha256, ecdsa-p384-sha384, ecdsa-p384-sha512, ecdsa-p521-sha256, ecdsa-p521-sha384, ecdsa-p521-sha512, eddsa-ed25519")
	commonName   = flag.String("common_name", "", "Common Name (CN)")
	country      = flag.String("country", "", "Country (C)")
	state        = flag.String("state", "", "State (ST)")
	city         = flag.String("city", "", "City (L)")
	organization = flag.String("organization", "", "Organization (O)")
	orgUnit      = flag.String("organizational_unit", "", "Organizational Unit (OU)")
	ipAddress    = flag.String("ip_address", "", "IP address")
	emailID      = flag.String("email_id", "", "Email address")
	sanDNS       = flag.String("san_dns", "", "SAN DNS names (comma-separated)")
	sanEmails    = flag.String("san_emails", "", "SAN email addresses (comma-separated)")
	sanIPs       = flag.String("san_ips", "", "SAN IP addresses (comma-separated)")
	sanURIs      = flag.String("san_uris", "", "SAN URIs (comma-separated)")

	// CSR signing flags (for device-generated CSR)
	caSignCert = flag.String("ca_sign_cert", "", "CA certificate for signing CSR")
	caSignKey  = flag.String("ca_sign_key", "", "CA key for signing CSR")

	// Existing entity flags
	copyEntityFrom = flag.String("copy_entity_from", "", "Copy entity from this profile")
	entityType     = flag.String("entity_type", "", "Entity type to copy: cert, ca_bundle, crl_bundle")

	// Validation flag
	validate = flag.Bool("validate", false, "Validate rotated certificates before finalizing (test new connection)")

	// Verbose flag
	verbose = flag.Bool("v", false, "Verbose output")

	// Dump proto flag for prototext output
	dumpProto = flag.Bool("dump-proto", false, "Dump requests/responses in prototext format")
)

func main() {
	flag.Parse()

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Validate required flags
	if *targetName == "" && !*insecureSkipVerify {
		return fmt.Errorf("target_name is required (unless using insecure_skip_verify)")
	}

	// Validate CA certificate requirement for secure connections
	if !*insecureSkipVerify && *caCert == "" {
		return fmt.Errorf("ca_cert is required for secure connections (unless using insecure_skip_verify)")
	}

	// Validate client certificate and key pairing
	if (*clientCert != "" && *clientKey == "") || (*clientCert == "" && *clientKey != "") {
		return fmt.Errorf("client_cert and client_key must be provided together")
	}

	// Create connection configuration
	connCfg := &connection.Config{
		TargetAddr:         *targetAddr,
		TargetName:         *targetName,
		Timeout:            *timeout,
		InsecureSkipVerify: *insecureSkipVerify,
		CACert:             *caCert,
		ClientCert:         *clientCert,
		ClientKey:          *clientKey,
		Username:           *username,
		Password:           *password,
	}

	// Create connection
	if *verbose {
		fmt.Printf("Connecting to %s...\n", *targetAddr)
	}

	conn, err := connection.New(connCfg)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	if *verbose {
		fmt.Println("Connected successfully")
	}

	// Create Certz client
	client := operations.NewCertzClient(conn)
	client.SetDumpProto(*dumpProto, *verbose)
	defer client.Close()

	// Create context
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	// Execute operation
	switch *operation {
	case "add-profile":
		return addProfile(ctx, client)

	case "delete-profile":
		return deleteProfile(ctx, client)

	case "get-profile-list":
		return getProfileList(ctx, client)

	case "can-generate-csr":
		return canGenerateCSR(ctx, client)

	case "rotate":
		return rotate(ctx, client, connCfg)

	default:
		return fmt.Errorf("unknown operation: %s", *operation)
	}
}

func addProfile(ctx context.Context, client *operations.CertzClient) error {
	if *profileID == "" {
		return fmt.Errorf("profile_id is required")
	}

	if *verbose {
		fmt.Printf("Adding profile: %s\n", *profileID)
	}

	if err := client.AddProfile(ctx, *profileID); err != nil {
		return err
	}

	fmt.Printf("✓ Profile '%s' added successfully\n", *profileID)
	return nil
}

func deleteProfile(ctx context.Context, client *operations.CertzClient) error {
	if *profileID == "" {
		return fmt.Errorf("profile_id is required")
	}

	if *verbose {
		fmt.Printf("Deleting profile: %s\n", *profileID)
	}

	if err := client.DeleteProfile(ctx, *profileID); err != nil {
		return err
	}

	fmt.Printf("✓ Profile '%s' deleted successfully\n", *profileID)
	return nil
}

func getProfileList(ctx context.Context, client *operations.CertzClient) error {
	if *verbose {
		fmt.Println("Retrieving profile list...")
	}

	profiles, err := client.GetProfileList(ctx)
	if err != nil {
		return err
	}

	fmt.Println("SSL Profiles:")
	for _, profile := range profiles {
		fmt.Printf("  - %s\n", profile)
	}

	return nil
}

func canGenerateCSR(ctx context.Context, client *operations.CertzClient) error {
	if *commonName == "" {
		return fmt.Errorf("common_name is required")
	}

	params := buildCSRParams()

	if *verbose {
		fmt.Printf("Checking CSR generation capability for CN=%s, Suite=%s\n", *commonName, *csrSuite)
	}

	canGenerate, err := client.CanGenerateCSR(ctx, params)
	if err != nil {
		return err
	}

	if canGenerate {
		fmt.Println("✓ Device can generate CSR with specified parameters")
	} else {
		fmt.Println("✗ Device cannot generate CSR with specified parameters")
	}

	return nil
}

func rotate(ctx context.Context, client *operations.CertzClient, connCfg *connection.Config) error {
	if *profileID == "" {
		return fmt.Errorf("profile_id is required")
	}

	req := &operations.RotateRequest{
		ProfileID:      *profileID,
		ForceOverwrite: *forceOverwrite,
		Version:        *version,
		CreatedOn:      *createdOn,
		CertFile:       *certFile,
		KeyFile:        *keyFile,
		CABundleFile:   *caBundleFile,
		CRLBundleFile:  *crlBundleFile,
		GenerateCSR:    *generateCSR,
		CASignCertFile: *caSignCert,
		CASignKeyFile:  *caSignKey,
		CopyFrom:       *copyEntityFrom,
		Validate:       *validate,
		ConnConfig:     connCfg,
	}

	// Handle certificate source
	if *certSource != "" {
		switch strings.ToLower(*certSource) {
		case "idevid":
			req.CertSource = certzpb.Certificate_CERT_SOURCE_IDEVID
			req.KeySource = certzpb.Certificate_KEY_SOURCE_IDEVID_TPM
		case "oidevid":
			req.CertSource = certzpb.Certificate_CERT_SOURCE_OIDEVID
			req.KeySource = certzpb.Certificate_KEY_SOURCE_IDEVID_TPM
		case "self-signed":
			req.CertSource = certzpb.Certificate_CERT_SOURCE_SELFSIGNED
			req.KeySource = certzpb.Certificate_KEY_SOURCE_SELFSIGNED
		default:
			return fmt.Errorf("invalid cert_source: %s", *certSource)
		}
	}

	// Handle entity type for copying
	if *copyEntityFrom != "" {
		switch strings.ToLower(*entityType) {
		case "cert":
			req.EntityType = certzpb.ExistingEntity_ENTITY_TYPE_CERTIFICATE_CHAIN
		case "ca_bundle", "cabundle":
			req.EntityType = certzpb.ExistingEntity_ENTITY_TYPE_TRUST_BUNDLE
		case "crl_bundle", "crlbundle":
			req.EntityType = certzpb.ExistingEntity_ENTITY_TYPE_CERTIFICATE_REVOCATION_LIST_BUNDLE
		default:
			return fmt.Errorf("invalid entity_type: %s", *entityType)
		}
	}

	// Handle CSR generation
	if *generateCSR {
		if *commonName == "" {
			return fmt.Errorf("common_name is required for CSR generation")
		}
		if *caSignCert == "" || *caSignKey == "" {
			return fmt.Errorf("ca_sign_cert and ca_sign_key are required for CSR signing")
		}

		// Validate CA key type matches CSR suite before starting rotation
		if err := validateCAKeyMatchesCSRSuite(*caSignKey, *csrSuite); err != nil {
			return err
		}

		req.CSRParams = buildCSRParams()
	}

	if *verbose {
		fmt.Printf("Starting certificate rotation for profile: %s\n", *profileID)
	}

	if err := client.Rotate(ctx, req); err != nil {
		return err
	}

	fmt.Printf("✓ Certificate rotation completed successfully for profile '%s'\n", *profileID)
	return nil
}

func buildCSRParams() *certzpb.CSRParams {
	suiteValue := csrSuiteFromString(*csrSuite)
	params := &certzpb.CSRParams{
		CsrSuite:           suiteValue,
		CommonName:         *commonName,
		Country:            *country,
		State:              *state,
		City:               *city,
		Organization:       *organization,
		OrganizationalUnit: *orgUnit,
		IpAddress:          *ipAddress,
		EmailId:            *emailID,
	}

	// Handle SAN
	if *sanDNS != "" || *sanEmails != "" || *sanIPs != "" || *sanURIs != "" {
		san := &certzpb.V3ExtensionSAN{}

		if *sanDNS != "" {
			san.Dns = strings.Split(*sanDNS, ",")
		}
		if *sanEmails != "" {
			san.Emails = strings.Split(*sanEmails, ",")
		}
		if *sanIPs != "" {
			san.Ips = strings.Split(*sanIPs, ",")
		}
		if *sanURIs != "" {
			san.Uris = strings.Split(*sanURIs, ",")
		}

		params.San = san
	}

	return params
}

func csrSuiteFromString(suite string) certzpb.CSRSuite {
	switch suite {
	case "rsa-2048-sha256":
		return certzpb.CSRSuite_CSRSUITE_X509_KEY_TYPE_RSA_2048_SIGNATURE_ALGORITHM_SHA_2_256
	case "rsa-2048-sha384":
		return certzpb.CSRSuite_CSRSUITE_X509_KEY_TYPE_RSA_2048_SIGNATURE_ALGORITHM_SHA_2_384
	case "rsa-2048-sha512":
		return certzpb.CSRSuite_CSRSUITE_X509_KEY_TYPE_RSA_2048_SIGNATURE_ALGORITHM_SHA_2_512
	case "rsa-3072-sha256":
		return certzpb.CSRSuite_CSRSUITE_X509_KEY_TYPE_RSA_3072_SIGNATURE_ALGORITHM_SHA_2_256
	case "rsa-3072-sha384":
		return certzpb.CSRSuite_CSRSUITE_X509_KEY_TYPE_RSA_3072_SIGNATURE_ALGORITHM_SHA_2_384
	case "rsa-3072-sha512":
		return certzpb.CSRSuite_CSRSUITE_X509_KEY_TYPE_RSA_3072_SIGNATURE_ALGORITHM_SHA_2_512
	case "rsa-4096-sha256":
		return certzpb.CSRSuite_CSRSUITE_X509_KEY_TYPE_RSA_4096_SIGNATURE_ALGORITHM_SHA_2_256
	case "rsa-4096-sha384":
		return certzpb.CSRSuite_CSRSUITE_X509_KEY_TYPE_RSA_4096_SIGNATURE_ALGORITHM_SHA_2_384
	case "rsa-4096-sha512":
		return certzpb.CSRSuite_CSRSUITE_X509_KEY_TYPE_RSA_4096_SIGNATURE_ALGORITHM_SHA_2_512
	case "ecdsa-p256-sha256":
		return certzpb.CSRSuite_CSRSUITE_X509_KEY_TYPE_ECDSA_PRIME256V1_SIGNATURE_ALGORITHM_SHA_2_256
	case "ecdsa-p256-sha384":
		return certzpb.CSRSuite_CSRSUITE_X509_KEY_TYPE_ECDSA_PRIME256V1_SIGNATURE_ALGORITHM_SHA_2_384
	case "ecdsa-p256-sha512":
		return certzpb.CSRSuite_CSRSUITE_X509_KEY_TYPE_ECDSA_PRIME256V1_SIGNATURE_ALGORITHM_SHA_2_512
	case "ecdsa-p384-sha256":
		return certzpb.CSRSuite_CSRSUITE_X509_KEY_TYPE_ECDSA_SECP384R1_SIGNATURE_ALGORITHM_SHA_2_256
	case "ecdsa-p384-sha384":
		return certzpb.CSRSuite_CSRSUITE_X509_KEY_TYPE_ECDSA_SECP384R1_SIGNATURE_ALGORITHM_SHA_2_384
	case "ecdsa-p384-sha512":
		return certzpb.CSRSuite_CSRSUITE_X509_KEY_TYPE_ECDSA_SECP384R1_SIGNATURE_ALGORITHM_SHA_2_512
	case "ecdsa-p521-sha256":
		return certzpb.CSRSuite_CSRSUITE_X509_KEY_TYPE_ECDSA_SECP521R1_SIGNATURE_ALGORITHM_SHA_2_256
	case "ecdsa-p521-sha384":
		return certzpb.CSRSuite_CSRSUITE_X509_KEY_TYPE_ECDSA_SECP521R1_SIGNATURE_ALGORITHM_SHA_2_384
	case "ecdsa-p521-sha512":
		return certzpb.CSRSuite_CSRSUITE_X509_KEY_TYPE_ECDSA_SECP521R1_SIGNATURE_ALGORITHM_SHA_2_512
	case "eddsa-ed25519":
		return certzpb.CSRSuite_CSRSUITE_X509_KEY_TYPE_EDDSA_ED25519
	default:
		// Default to RSA 2048 SHA256 - warn user about invalid value
		fmt.Fprintf(os.Stderr, "WARNING: Unknown CSR suite '%s', defaulting to rsa-2048-sha256\n", suite)
		return certzpb.CSRSuite_CSRSUITE_X509_KEY_TYPE_RSA_2048_SIGNATURE_ALGORITHM_SHA_2_256
	}
}

// validateCAKeyMatchesCSRSuite validates that the CA signing key type matches the CSR suite
func validateCAKeyMatchesCSRSuite(caKeyFile, csrSuite string) error {
	// Read CA key file
	keyPEM, err := ioutil.ReadFile(caKeyFile)
	if err != nil {
		return fmt.Errorf("failed to read CA key file: %w", err)
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return fmt.Errorf("failed to decode CA key PEM")
	}

	// Determine CA key type
	var caKeyType string
	if key, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes); err == nil {
		switch key.(type) {
		case *rsa.PrivateKey:
			caKeyType = "RSA"
		case *ecdsa.PrivateKey:
			caKeyType = "ECDSA"
		case ed25519.PrivateKey:
			caKeyType = "Ed25519"
		default:
			return fmt.Errorf("unsupported CA key type")
		}
	} else if _, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes); err == nil {
		caKeyType = "RSA"
	} else if _, err := x509.ParseECPrivateKey(keyBlock.Bytes); err == nil {
		caKeyType = "ECDSA"
	} else {
		return fmt.Errorf("unsupported CA key format")
	}

	// Determine required key type based on CSR suite
	var requiredKeyType string
	if strings.HasPrefix(csrSuite, "rsa-") {
		requiredKeyType = "RSA"
	} else if strings.HasPrefix(csrSuite, "ecdsa-") {
		requiredKeyType = "ECDSA"
	} else if strings.HasPrefix(csrSuite, "eddsa-") {
		requiredKeyType = "Ed25519"
	} else {
		requiredKeyType = "RSA" // Default
	}

	// Check if they match
	if caKeyType != requiredKeyType {
		return fmt.Errorf("CA key type mismatch: CSR suite '%s' requires %s key, but CA key is %s", csrSuite, requiredKeyType, caKeyType)
	}

	return nil
}
