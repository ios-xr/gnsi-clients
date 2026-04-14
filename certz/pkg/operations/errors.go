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

import "errors"

// Common errors
var (
	ErrNoCertInResponse   = errors.New("no certificate in response")
	ErrNoCSRInResponse    = errors.New("no CSR in response")
	ErrInvalidPEM         = errors.New("failed to decode PEM block")
	ErrEmptyCertFile      = errors.New("certificate file is empty")
	ErrEmptyKeyFile       = errors.New("key file is empty")
	ErrEmptyCAFile        = errors.New("CA bundle file is empty")
	ErrUnsupportedKeyType = errors.New("unsupported key type")
	ErrValidationFailed   = errors.New("certificate validation failed")
)
