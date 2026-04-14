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
	"fmt"
	"os"
)

// ValidateRotateRequest validates a RotateRequest before processing
func ValidateRotateRequest(req *RotateRequest) error {
	if req.ProfileID == "" {
		return fmt.Errorf("profile_id is required")
	}

	// Validate certificate rotation
	if req.CertFile != "" {
		if err := validateFileExists(req.CertFile, "certificate"); err != nil {
			return err
		}
	}

	if req.KeyFile != "" {
		if err := validateFileExists(req.KeyFile, "private key"); err != nil {
			return err
		}
	}

	// Certificate and key must be provided together
	if (req.CertFile != "" && req.KeyFile == "") || (req.CertFile == "" && req.KeyFile != "") {
		return fmt.Errorf("certificate and private key must be provided together")
	}

	// Validate CA bundle
	if req.CABundleFile != "" {
		if err := validateFileExists(req.CABundleFile, "CA bundle"); err != nil {
			return err
		}
	}

	// Validate CRL bundle
	if req.CRLBundleFile != "" {
		if err := validateFileExists(req.CRLBundleFile, "CRL bundle"); err != nil {
			return err
		}
	}

	// Validate CSR generation parameters
	if req.GenerateCSR {
		if req.CSRParams == nil {
			return fmt.Errorf("CSR parameters are required when generate_csr is true")
		}
		if req.CSRParams.CommonName == "" {
			return fmt.Errorf("common_name is required for CSR generation")
		}
		if req.CASignCertFile == "" || req.CASignKeyFile == "" {
			return fmt.Errorf("ca_sign_cert and ca_sign_key are required for CSR signing")
		}
		if err := validateFileExists(req.CASignCertFile, "CA signing certificate"); err != nil {
			return err
		}
		if err := validateFileExists(req.CASignKeyFile, "CA signing key"); err != nil {
			return err
		}
	}

	// Validate existing entity copy
	if req.CopyFrom != "" {
		if req.EntityType == 0 {
			return fmt.Errorf("entity_type is required when copying from another profile")
		}
	}

	// Validation requires ConnConfig
	if req.Validate && req.ConnConfig == nil {
		return fmt.Errorf("connection config is required when validation is enabled")
	}

	return nil
}

// validateFileExists checks if a file exists and is readable
func validateFileExists(filePath, fileType string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s file not found: %s", fileType, filePath)
		}
		return fmt.Errorf("cannot access %s file %s: %w", fileType, filePath, err)
	}

	if info.IsDir() {
		return fmt.Errorf("%s path is a directory, not a file: %s", fileType, filePath)
	}

	return nil
}
