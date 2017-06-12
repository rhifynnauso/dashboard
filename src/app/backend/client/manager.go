// Copyright 2015 Google Inc. All Rights Reserved.
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

package client

import (
	"crypto/rand"
	"errors"
	"log"
	"strings"

	"github.com/emicklei/go-restful"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// Dashboard UI default values for client configs.
const (
	// High enough QPS to fit all expected use cases. QPS=0 is not set here, because
	// client code is overriding it.
	DefaultQPS = 1e6
	// High enough Burst to fit all expected use cases. Burst=0 is not set here, because
	// client code is overriding it.
	DefaultBurst = 1e6
	// Use kubernetes protobuf as content type by default
	DefaultContentType = "application/vnd.kubernetes.protobuf"
)

// ClientManager is responsible for initializing and creating clients to communicate with
// kubernetes apiserver on demand
type ClientManager interface {
	Client(req *restful.Request) (*kubernetes.Clientset, error)
	Config(req *restful.Request) (*rest.Config, error)
	CSRFKey() string
	VerberClient(req *restful.Request) (ResourceVerber, error)
}

// clientManager implements ClientManager interface
type clientManager struct {
	// Autogenerated key on backend start used to secure requests from csrf attacks
	csrfKey string
	// Path to kubeconfig file. If both kubeConfigPath and apiserverHost are empty
	// inClusterConfig will be used
	kubeConfigPath string
	// Address of apiserver host in format 'protocol://address:port'
	apiserverHost string
	// Initialized on clientManager creation and used if kubeconfigPath and apiserverHost are
	// empty
	inClusterConfig *rest.Config
}

// Client returns kubernetes client that is created based on authentication information extracted
// from request. If request is nil then authentication will be skipped.
func (self *clientManager) Client(req *restful.Request) (*kubernetes.Clientset, error) {
	cfg, err := self.Config(req)
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// Config creates rest Config based on authentication information extracted from request.
// Currently request header is only checked for existence of 'Authentication: BearerToken'
func (self *clientManager) Config(req *restful.Request) (*rest.Config, error) {
	authInfo := self.extractAuthInfo(req)

	cfg, err := self.buildConfigFromFlags(self.apiserverHost, self.kubeConfigPath)
	if err != nil {
		return nil, err
	}

	// Override auth header token. For now only bearer token is supported
	if len(authInfo.Token) > 0 {
		cfg.BearerToken = authInfo.Token
	}

	self.initConfig(cfg)
	return cfg, nil
}

// CSRFKey returns key that is generated upon client manager creation
func (self *clientManager) CSRFKey() string {
	return self.csrfKey
}

// VerberClient returns new verber client based on authentication information extracted from
// request
func (self *clientManager) VerberClient(req *restful.Request) (ResourceVerber, error) {
	client, err := self.Client(req)
	if err != nil {
		return ResourceVerber{}, err
	}

	return NewResourceVerber(client.CoreV1().RESTClient(),
		client.ExtensionsV1beta1().RESTClient(), client.AppsV1beta1().RESTClient(),
		client.BatchV1().RESTClient(), client.AutoscalingV1().RESTClient(),
		client.StorageV1beta1().RESTClient()), nil
}

// Initializes config with default values
func (self *clientManager) initConfig(cfg *rest.Config) {
	cfg.QPS = DefaultQPS
	cfg.Burst = DefaultBurst
	cfg.ContentType = DefaultContentType
}

// Returns rest Config based on provided apiserverHost and kubeConfigPath flags. If both are
// empty then in-cluster config will be used and if it is nil the error is returned.
func (self *clientManager) buildConfigFromFlags(apiserverHost, kubeConfigPath string) (
	*rest.Config, error) {
	if len(kubeConfigPath) > 0 || len(apiserverHost) > 0 {
		return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeConfigPath},
			&clientcmd.ConfigOverrides{ClusterInfo: api.Cluster{Server: apiserverHost}}).ClientConfig()
	}

	if self.isRunningInCluster() {
		return self.inClusterConfig, nil
	}

	return nil, errors.New("Could not create client config. Check logs for more information")
}

// Extracts authentication information from request header
func (self *clientManager) extractAuthInfo(req *restful.Request) api.AuthInfo {
	if req == nil {
		log.Print("No request provided. Skipping authorization header")
		return api.AuthInfo{}
	}

	authHeader := req.HeaderParameter("Authorization")
	token := ""
	if strings.HasPrefix(authHeader, "Bearer ") {
		token = strings.TrimPrefix(authHeader, "Bearer ")
	}

	return api.AuthInfo{Token: token}
}

// Initializes client manager
func (self *clientManager) init() {
	self.initInClusterConfig()
	self.initCSRFKey()
}

// Initializes csrfKey. If in-cluster config is detected then csrf key is initialized with
// service account token, otherwise it is generated
func (self *clientManager) initCSRFKey() {
	if self.inClusterConfig == nil {
		// Most likely running for a dev, so no replica issues, just generate a random key
		log.Println("Using random key for csrf signing")
		self.generateCSRFKey()
		return
	}

	// We run in a cluster, so we should use a signing key that is the same for potential replications
	log.Println("Using service account token for csrf signing")
	self.csrfKey = self.inClusterConfig.BearerToken
}

// Initializes in-cluster config if apiserverHost and kubeConfigPath were not provided.
func (self *clientManager) initInClusterConfig() {
	if len(self.apiserverHost) > 0 || len(self.kubeConfigPath) > 0 {
		log.Print("Skipping in-cluster config")
		return
	}

	log.Print("Using in-cluster config to connect to apiserver")
	cfg, err := rest.InClusterConfig()
	if err != nil {
		log.Printf("Could not init in cluster config: %s", err.Error())
		return
	}

	self.inClusterConfig = cfg
}

// Generates random csrf key
func (self *clientManager) generateCSRFKey() {
	bytes := make([]byte, 256)
	_, err := rand.Read(bytes)
	if err != nil {
		panic("Fatal error. Could not generate csrf key")
	}

	self.csrfKey = string(bytes)
}

// Returns true if in-cluster config is used
func (self *clientManager) isRunningInCluster() bool {
	return self.inClusterConfig != nil
}

// NewClientManager creates client manager based on kubeConfigPath and apiserverHost parameters.
// If both are empty then in-cluster config is used.
func NewClientManager(kubeConfigPath, apiserverHost string) ClientManager {
	result := &clientManager{
		kubeConfigPath: kubeConfigPath,
		apiserverHost:  apiserverHost,
	}

	result.init()

	return result
}
