// Copyright 2017 The Kubernetes Dashboard Authors.
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

package jwe

import (
	"reflect"
	"testing"

	authApi "github.com/kubernetes/dashboard/src/app/backend/auth/api"
	"github.com/kubernetes/dashboard/src/app/backend/sync"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/clientcmd/api"
)

func getTokenManager() authApi.TokenManager {
	c := fake.NewSimpleClientset()
	syncManager := sync.NewSynchronizerManager(c)
	holder := NewRSAKeyHolder(syncManager.Secret("", ""))
	return NewJWETokenManager(holder)
}

func areErrorsEqual(err1, err2 error) bool {
	return (err1 != nil && err2 != nil && err1.Error() == err2.Error()) ||
		(err1 == nil && err2 == nil)
}

func TestJweTokenManager_Generate(t *testing.T) {
	cases := []struct {
		info        string
		authInfo    api.AuthInfo
		expectedErr error
	}{
		{
			"Should generate encrypted token",
			api.AuthInfo{Token: "test-token"},
			nil,
		},
	}

	for _, c := range cases {
		tokenManager := getTokenManager()
		token, err := tokenManager.Generate(c.authInfo)

		if !areErrorsEqual(err, c.expectedErr) {
			t.Errorf("Test Case: %s. Expected error to be: %v, but got %v.",
				c.info, c.expectedErr, err)
		}

		if len(token) == 0 {
			t.Errorf("Test Case: %s. Expected token not to be empty.", c.info)
		}
	}
}

func TestJweTokenManager_Decrypt(t *testing.T) {
	cases := []struct {
		info        string
		authInfo    api.AuthInfo
		expected    *api.AuthInfo
		expectedErr error
	}{
		{
			"Should decrypt encrypted token",
			api.AuthInfo{Token: "test-token"},
			&api.AuthInfo{Token: "test-token"},
			nil,
		},
	}

	for _, c := range cases {
		tokenManager := getTokenManager()
		token, _ := tokenManager.Generate(c.authInfo)
		authInfo, err := tokenManager.Decrypt(token)

		if !areErrorsEqual(err, c.expectedErr) {
			t.Errorf("Test Case: %s. Expected error to be: %v, but got %v.",
				c.info, c.expectedErr, err)
		}

		if !reflect.DeepEqual(authInfo, c.expected) {
			t.Errorf("Test Case: %s. Expected: %v, but got %v.", c.info, c.expected, authInfo)
		}
	}
}
