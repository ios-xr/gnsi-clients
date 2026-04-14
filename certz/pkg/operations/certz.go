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

package operations

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/cisco/gnsi-certz-client/pkg/connection"
	"github.com/golang/protobuf/proto"
	certzpb "github.com/openconfig/gnsi/certz"
	"google.golang.org/grpc"
)

// CertzClient wraps the gRPC certz client
type CertzClient struct {
	client    certzpb.CertzClient
	conn      *grpc.ClientConn
	dumpProto bool
	verbose   bool
}

// NewCertzClient creates a new Certz client
func NewCertzClient(conn *grpc.ClientConn) *CertzClient {
	return &CertzClient{
		client: certzpb.NewCertzClient(conn),
		conn:   conn,
	}
}

// Close closes the gRPC connection
func (c *CertzClient) Close() error {
	return c.conn.Close()
}

// SetDumpProto enables prototext dumping and verbose output
func (c *CertzClient) SetDumpProto(dumpProto, verbose bool) {
	c.dumpProto = dumpProto
	c.verbose = verbose
}

// logRequest logs a request in prototext format if dump-proto is enabled
func (c *CertzClient) logRequest(name string, req proto.Message) {
	if c.dumpProto {
		fmt.Printf("[PROTO] %s Request:\n%s\n", name, proto.MarshalTextString(req))
	}
}

// logResponse logs a response in prototext format if dump-proto is enabled
func (c *CertzClient) logResponse(name string, resp proto.Message) {
	if c.dumpProto && resp != nil {
		fmt.Printf("[PROTO] %s Response:\n%s\n", name, proto.MarshalTextString(resp))
	}
}

// log logs a message if verbose is enabled
func (c *CertzClient) log(format string, args ...interface{}) {
	if c.verbose {
		fmt.Printf(format+"\n", args...)
	}
}

// AddProfile creates a new SSL profile
func (c *CertzClient) AddProfile(ctx context.Context, profileID string) error {
	req := &certzpb.AddProfileRequest{
		SslProfileId: profileID,
	}

	c.logRequest("AddProfile", req)

	resp, err := c.client.AddProfile(ctx, req)
	if err != nil {
		return fmt.Errorf("AddProfile failed: %w", err)
	}

	c.logResponse("AddProfile", resp)

	return nil
}

// DeleteProfile deletes an existing SSL profile
func (c *CertzClient) DeleteProfile(ctx context.Context, profileID string) error {
	req := &certzpb.DeleteProfileRequest{
		SslProfileId: profileID,
	}

	c.logRequest("DeleteProfile", req)

	resp, err := c.client.DeleteProfile(ctx, req)
	if err != nil {
		return fmt.Errorf("DeleteProfile failed: %w", err)
	}

	c.logResponse("DeleteProfile", resp)

	return nil
}

// GetProfileList retrieves all SSL profiles
func (c *CertzClient) GetProfileList(ctx context.Context) ([]string, error) {
	req := &certzpb.GetProfileListRequest{}

	c.logRequest("GetProfileList", req)

	resp, err := c.client.GetProfileList(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("GetProfileList failed: %w", err)
	}

	c.logResponse("GetProfileList", resp)

	return resp.SslProfileIds, nil
}

// CanGenerateCSR checks if device can generate CSR
func (c *CertzClient) CanGenerateCSR(ctx context.Context, params *certzpb.CSRParams) (bool, error) {
	req := &certzpb.CanGenerateCSRRequest{
		Params: params,
	}

	c.logRequest("CanGenerateCSR", req)

	resp, err := c.client.CanGenerateCSR(ctx, req)
	if err != nil {
		return false, fmt.Errorf("CanGenerateCSR failed: %w", err)
	}

	c.logResponse("CanGenerateCSR", resp)

	return resp.CanGenerate, nil
}

// RotateRequest holds parameters for certificate rotation
type RotateRequest struct {
	ProfileID      string
	ForceOverwrite bool
	Version        string
	CreatedOn      uint64

	// Certificate rotation
	CertFile   string
	KeyFile    string
	CertSource certzpb.Certificate_CertSource
	KeySource  certzpb.Certificate_KeySource

	// CA bundle
	CABundleFile string

	// CRL bundle
	CRLBundleFile string

	// Device-generated CSR
	GenerateCSR    bool
	CSRParams      *certzpb.CSRParams
	CASignCertFile string
	CASignKeyFile  string

	// Existing entity
	CopyFrom   string
	EntityType certzpb.ExistingEntity_EntityType

	// Validation before finalize
	Validate   bool
	ConnConfig *connection.Config
}

// Rotate performs certificate rotation
func (c *CertzClient) Rotate(ctx context.Context, req *RotateRequest) error {
	// Validate request parameters
	if err := ValidateRotateRequest(req); err != nil {
		return fmt.Errorf("invalid rotation request: %w", err)
	}

	stream, err := c.client.Rotate(ctx)
	if err != nil {
		return fmt.Errorf("failed to create rotate stream: %w", err)
	}

	// Handle device-generated CSR if requested
	if req.GenerateCSR {
		if err := c.rotateWithDeviceCSR(stream, req); err != nil {
			return err
		}
	} else {
		// Upload certificate if provided
		if req.CertFile != "" || req.CertSource != certzpb.Certificate_CERT_SOURCE_UNSPECIFIED {
			if err := c.uploadCertificate(stream, req); err != nil {
				return err
			}
		}
	}

	// Upload CA bundle if provided
	if req.CABundleFile != "" {
		if err := c.uploadCABundle(stream, req); err != nil {
			return err
		}
	}

	// Upload CRL bundle if provided
	if req.CRLBundleFile != "" {
		if err := c.uploadCRLBundle(stream, req); err != nil {
			return err
		}
	}

	// Copy existing entity if requested
	if req.CopyFrom != "" {
		if err := c.uploadExistingEntity(stream, req); err != nil {
			return err
		}
	}

	// Validate rotated certificates before finalizing (if requested)
	if req.Validate {
		if err := c.validateRotation(ctx, req); err != nil {
			// Validation failed - cancel stream to trigger rollback
			return fmt.Errorf("validation failed, rotation cancelled: %w", err)
		}
	}

	// Finalize rotation
	finalizeReq := &certzpb.RotateCertificateRequest{
		ForceOverwrite: req.ForceOverwrite,
		SslProfileId:   req.ProfileID,
		RotateRequest: &certzpb.RotateCertificateRequest_FinalizeRotation{
			FinalizeRotation: &certzpb.FinalizeRequest{},
		},
	}

	c.logRequest("RotateCertificate-Finalize", finalizeReq)

	if err := stream.Send(finalizeReq); err != nil {
		return fmt.Errorf("failed to send finalize request: %w", err)
	}

	// Wait for stream to close
	finalizeResp, err := stream.Recv()
	if err != nil && err != io.EOF {
		return fmt.Errorf("finalize failed: %w", err)
	}

	if finalizeResp != nil {
		c.logResponse("RotateCertificate-Finalize", finalizeResp)
	}

	return nil
}

// rotateWithDeviceCSR handles device-generated CSR workflow
func (c *CertzClient) rotateWithDeviceCSR(stream certzpb.Certz_RotateClient, req *RotateRequest) error {
	// Request CSR generation
	csrReq := &certzpb.RotateCertificateRequest{
		ForceOverwrite: req.ForceOverwrite,
		SslProfileId:   req.ProfileID,
		RotateRequest: &certzpb.RotateCertificateRequest_GenerateCsr{
			GenerateCsr: &certzpb.GenerateCSRRequest{
				Params: req.CSRParams,
			},
		},
	}

	c.logRequest("RotateCertificate-GenerateCSR", csrReq)

	if err := stream.Send(csrReq); err != nil {
		return fmt.Errorf("failed to send CSR request: %w", err)
	}

	// Receive CSR response
	resp, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("failed to receive CSR: %w", err)
	}

	c.logResponse("RotateCertificate-GenerateCSR", resp)

	generatedCSR := resp.GetGeneratedCsr()
	if generatedCSR == nil {
		return fmt.Errorf("no CSR in response")
	}

	// Sign the CSR
	signedCert, err := signCSR(generatedCSR.CertificateSigningRequest, req.CASignCertFile, req.CASignKeyFile, req.CSRParams.CsrSuite)
	if err != nil {
		return fmt.Errorf("failed to sign CSR: %w", err)
	}

	// Upload signed certificate
	certChain := &certzpb.CertificateChain{
		Certificate: &certzpb.Certificate{
			Type:     certzpb.CertificateType_CERTIFICATE_TYPE_X509,
			Encoding: certzpb.CertificateEncoding_CERTIFICATE_ENCODING_PEM,
			CertificateType: &certzpb.Certificate_RawCertificate{
				RawCertificate: signedCert,
			},
			PrivateKeyType: &certzpb.Certificate_KeySource_{
				KeySource: certzpb.Certificate_KEY_SOURCE_GENERATED,
			},
		},
	}

	uploadReq := &certzpb.RotateCertificateRequest{
		ForceOverwrite: req.ForceOverwrite,
		SslProfileId:   req.ProfileID,
		RotateRequest: &certzpb.RotateCertificateRequest_Certificates{
			Certificates: &certzpb.UploadRequest{
				Entities: []*certzpb.Entity{
					{
						Version:   req.Version,
						CreatedOn: req.CreatedOn,
						Entity: &certzpb.Entity_CertificateChain{
							CertificateChain: certChain,
						},
					},
				},
			},
		},
	}

	c.logRequest("RotateCertificate-UploadSignedCert", uploadReq)

	if err := stream.Send(uploadReq); err != nil {
		return fmt.Errorf("failed to upload signed certificate: %w", err)
	}

	uploadResp, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("failed to receive upload response: %w", err)
	}

	c.logResponse("RotateCertificate-UploadSignedCert", uploadResp)

	return nil
}

// uploadCertificate uploads a certificate to the device
func (c *CertzClient) uploadCertificate(stream certzpb.Certz_RotateClient, req *RotateRequest) error {
	var certChain *certzpb.CertificateChain

	if req.CertSource != certzpb.Certificate_CERT_SOURCE_UNSPECIFIED {
		// Use device-native certificate
		certChain = &certzpb.CertificateChain{
			Certificate: &certzpb.Certificate{
				Type:     certzpb.CertificateType_CERTIFICATE_TYPE_X509,
				Encoding: certzpb.CertificateEncoding_CERTIFICATE_ENCODING_PEM,
				CertificateType: &certzpb.Certificate_CertSource_{
					CertSource: req.CertSource,
				},
				PrivateKeyType: &certzpb.Certificate_KeySource_{
					KeySource: req.KeySource,
				},
			},
		}
	} else {
		// Load certificate chain and key from files
		certPEM, err := ioutil.ReadFile(req.CertFile)
		if err != nil {
			return fmt.Errorf("failed to read certificate: %w", err)
		}

		keyPEM, err := ioutil.ReadFile(req.KeyFile)
		if err != nil {
			return fmt.Errorf("failed to read key: %w", err)
		}

		// Parse all certificates from the file (can be a chain)
		var certs []*x509.Certificate
		remaining := certPEM
		for {
			block, rest := pem.Decode(remaining)
			if block == nil {
				break
			}

			if block.Type == "CERTIFICATE" {
				cert, err := x509.ParseCertificate(block.Bytes)
				if err != nil {
					return fmt.Errorf("failed to parse certificate: %w", err)
				}
				certs = append(certs, cert)
			}
			remaining = rest
		}

		if len(certs) == 0 {
			return fmt.Errorf("no certificates found in certificate file")
		}

		// Build certificate chain
		// First certificate (leaf) gets the private key
		// Subsequent certificates (intermediates) are linked as parents
		var chainMessage *certzpb.CertificateChain
		for i, cert := range certs {
			certPEMBytes := pem.EncodeToMemory(&pem.Block{
				Type:  "CERTIFICATE",
				Bytes: cert.Raw,
			})

			certMsg := &certzpb.Certificate{
				Type:     certzpb.CertificateType_CERTIFICATE_TYPE_X509,
				Encoding: certzpb.CertificateEncoding_CERTIFICATE_ENCODING_PEM,
				CertificateType: &certzpb.Certificate_RawCertificate{
					RawCertificate: certPEMBytes,
				},
			}

			// Only the first (leaf) certificate gets the private key
			if i == 0 {
				certMsg.PrivateKeyType = &certzpb.Certificate_RawPrivateKey{
					RawPrivateKey: keyPEM,
				}
			}

			// Build chain structure
			if i == 0 {
				// First certificate - create the chain
				chainMessage = &certzpb.CertificateChain{
					Certificate: certMsg,
				}
			} else {
				// Subsequent certificates - add as parent
				chainMessage.Parent = &certzpb.CertificateChain{
					Certificate: certMsg,
					Parent:      chainMessage.Parent,
				}
			}
		}

		certChain = chainMessage
	}

	uploadReq := &certzpb.RotateCertificateRequest{
		ForceOverwrite: req.ForceOverwrite,
		SslProfileId:   req.ProfileID,
		RotateRequest: &certzpb.RotateCertificateRequest_Certificates{
			Certificates: &certzpb.UploadRequest{
				Entities: []*certzpb.Entity{
					{
						Version:   req.Version,
						CreatedOn: req.CreatedOn,
						Entity: &certzpb.Entity_CertificateChain{
							CertificateChain: certChain,
						},
					},
				},
			},
		},
	}

	c.logRequest("Rotate Certificate-UploadCertificate", uploadReq)

	if err := stream.Send(uploadReq); err != nil {
		return fmt.Errorf("failed to upload certificate: %w", err)
	}

	uploadResp, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("failed to receive upload response: %w", err)
	}

	c.logResponse("RotateCertificate-UploadCertificate", uploadResp)

	return nil
}

// uploadCABundle uploads a CA bundle
func (c *CertzClient) uploadCABundle(stream certzpb.Certz_RotateClient, req *RotateRequest) error {
	caPEM, err := ioutil.ReadFile(req.CABundleFile)
	if err != nil {
		return fmt.Errorf("failed to read CA bundle: %w", err)
	}

	// Parse certificates from PEM
	var certs []*x509.Certificate
	remaining := caPEM
	for {
		block, rest := pem.Decode(remaining)
		if block == nil {
			break
		}

		if block.Type == "CERTIFICATE" {
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return fmt.Errorf("failed to parse certificate: %w", err)
			}
			certs = append(certs, cert)
		}
		remaining = rest
	}

	if len(certs) == 0 {
		return fmt.Errorf("no certificates found in CA bundle")
	}

	// Build certificate chain
	var certChain *certzpb.CertificateChain
	for i := len(certs) - 1; i >= 0; i-- {
		certPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: certs[i].Raw,
		})

		newChain := &certzpb.CertificateChain{
			Certificate: &certzpb.Certificate{
				Type:     certzpb.CertificateType_CERTIFICATE_TYPE_X509,
				Encoding: certzpb.CertificateEncoding_CERTIFICATE_ENCODING_PEM,
				CertificateType: &certzpb.Certificate_RawCertificate{
					RawCertificate: certPEM,
				},
			},
		}

		if certChain != nil {
			newChain.Parent = certChain
		}
		certChain = newChain
	}

	uploadReq := &certzpb.RotateCertificateRequest{
		ForceOverwrite: req.ForceOverwrite,
		SslProfileId:   req.ProfileID,
		RotateRequest: &certzpb.RotateCertificateRequest_Certificates{
			Certificates: &certzpb.UploadRequest{
				Entities: []*certzpb.Entity{
					{
						Version:   req.Version,
						CreatedOn: req.CreatedOn,
						Entity: &certzpb.Entity_TrustBundle{
							TrustBundle: certChain,
						},
					},
				},
			},
		},
	}

	c.logRequest("RotateCertificate-UploadCABundle", uploadReq)

	if err := stream.Send(uploadReq); err != nil {
		return fmt.Errorf("failed to upload CA bundle: %w", err)
	}

	caResp, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("failed to receive CA bundle upload response: %w", err)
	}

	c.logResponse("RotateCertificate-UploadCABundle", caResp)

	return nil
}

// uploadCRLBundle uploads a CRL bundle
func (c *CertzClient) uploadCRLBundle(stream certzpb.Certz_RotateClient, req *RotateRequest) error {
	crlPEM, err := ioutil.ReadFile(req.CRLBundleFile)
	if err != nil {
		return fmt.Errorf("failed to read CRL bundle: %w", err)
	}

	var crlBundle certzpb.CertificateRevocationListBundle
	remaining := crlPEM
	for {
		block, rest := pem.Decode(remaining)
		if block == nil {
			break
		}

		if block.Type == "X509 CRL" {
			crlPEMBytes := pem.EncodeToMemory(&pem.Block{
				Type:  "X509 CRL",
				Bytes: block.Bytes,
			})

			crl := &certzpb.CertificateRevocationList{
				Type:                      certzpb.CertificateType_CERTIFICATE_TYPE_X509,
				Encoding:                  certzpb.CertificateEncoding_CERTIFICATE_ENCODING_PEM,
				CertificateRevocationList: crlPEMBytes,
				Id:                        req.Version,
			}
			crlBundle.CertificateRevocationLists = append(crlBundle.CertificateRevocationLists, crl)
		}
		remaining = rest
	}

	if len(crlBundle.CertificateRevocationLists) == 0 {
		return fmt.Errorf("no CRLs found in bundle")
	}

	uploadReq := &certzpb.RotateCertificateRequest{
		ForceOverwrite: req.ForceOverwrite,
		SslProfileId:   req.ProfileID,
		RotateRequest: &certzpb.RotateCertificateRequest_Certificates{
			Certificates: &certzpb.UploadRequest{
				Entities: []*certzpb.Entity{
					{
						Version:   req.Version,
						CreatedOn: req.CreatedOn,
						Entity: &certzpb.Entity_CertificateRevocationListBundle{
							CertificateRevocationListBundle: &crlBundle,
						},
					},
				},
			},
		},
	}

	c.logRequest("RotateCertificate-UploadCRLBundle", uploadReq)

	if err := stream.Send(uploadReq); err != nil {
		return fmt.Errorf("failed to upload CRL bundle: %w", err)
	}

	crlResp, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("failed to receive CRL bundle upload response: %w", err)
	}

	c.logResponse("RotateCertificate-UploadCRLBundle", crlResp)

	return nil
}

// uploadExistingEntity copies an entity from another profile
func (c *CertzClient) uploadExistingEntity(stream certzpb.Certz_RotateClient, req *RotateRequest) error {
	uploadReq := &certzpb.RotateCertificateRequest{
		ForceOverwrite: req.ForceOverwrite,
		SslProfileId:   req.ProfileID,
		RotateRequest: &certzpb.RotateCertificateRequest_Certificates{
			Certificates: &certzpb.UploadRequest{
				Entities: []*certzpb.Entity{
					{
						Version:   req.Version,
						CreatedOn: req.CreatedOn,
						Entity: &certzpb.Entity_ExistingEntity{
							ExistingEntity: &certzpb.ExistingEntity{
								SslProfileId: req.CopyFrom,
								EntityType:   req.EntityType,
							},
						},
					},
				},
			},
		},
	}

	c.logRequest("RotateCertificate-CopyExistingEntity", uploadReq)

	if err := stream.Send(uploadReq); err != nil {
		return fmt.Errorf("failed to upload existing entity: %w", err)
	}

	existingResp, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("failed to receive existing entity upload response: %w", err)
	}

	c.logResponse("RotateCertificate-CopyExistingEntity", existingResp)

	return nil
}

// validateRotation tests the rotated certificates by creating a new connection
// and verifying connectivity before finalizing the rotation.
//
// This function:
// 1. Creates a new gRPC connection using the rotated certificate/key/CA files
// 2. Calls CanGenerateCSR RPC to verify the connection works
// 3. Returns nil if successful, error if validation fails
//
// If validation fails, the caller should cancel the rotation stream,
// which triggers automatic rollback on the device.
func (c *CertzClient) validateRotation(ctx context.Context, req *RotateRequest) error {
	c.log("INFO: Starting validation of rotated credentials...")

	// Create a new connection config using the rotated certificate/key/CA
	testConfig := &connection.Config{
		TargetAddr:         req.ConnConfig.TargetAddr,
		TargetName:         req.ConnConfig.TargetName,
		Timeout:            req.ConnConfig.Timeout,
		InsecureSkipVerify: req.ConnConfig.InsecureSkipVerify,
		CACert:             req.CABundleFile, // Use the rotated CA bundle (if provided)
		ClientCert:         req.CertFile,     // Use the rotated client certificate (if provided)
		ClientKey:          req.KeyFile,      // Use the rotated client key (if provided)
		Username:           req.ConnConfig.Username,
		Password:           req.ConnConfig.Password,
	}

	// If no client cert provided in rotation, use original connection certs
	if testConfig.ClientCert == "" {
		testConfig.ClientCert = req.ConnConfig.ClientCert
		testConfig.ClientKey = req.ConnConfig.ClientKey
	}

	// If no CA provided in rotation, use original CA
	if testConfig.CACert == "" {
		testConfig.CACert = req.ConnConfig.CACert
	}

	c.log("INFO: Attempting test connection with rotated credentials...")

	// Establish test connection with rotated credentials
	testConn, err := connection.New(testConfig)
	if err != nil {
		return fmt.Errorf("test connection failed with rotated credentials (check certificate/key match and CA trust): %w", err)
	}
	defer testConn.Close()

	// Create a Certz client with the test connection
	testClient := certzpb.NewCertzClient(testConn)

	// Call CanGenerateCSR to verify connectivity works with rotated credentials
	csrParams := &certzpb.CSRParams{
		CsrSuite:   certzpb.CSRSuite_CSRSUITE_X509_KEY_TYPE_ECDSA_PRIME256V1_SIGNATURE_ALGORITHM_SHA_2_256,
		CommonName: "validation-test",
		Country:    "US",
	}

	canGenReq := &certzpb.CanGenerateCSRRequest{
		Params: csrParams,
	}

	c.log("INFO: Verifying device connectivity after rotation...")

	if _, err := testClient.CanGenerateCSR(ctx, canGenReq); err != nil {
		return fmt.Errorf("RPC call failed with rotated certificates (device may not trust new cert or CA): %w", err)
	}

	c.log("INFO: Validation successful - device is accessible with rotated credentials")

	return nil
}
