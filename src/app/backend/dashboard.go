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

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/kubernetes/dashboard/src/app/backend/auth"
	authApi "github.com/kubernetes/dashboard/src/app/backend/auth/api"
	"github.com/kubernetes/dashboard/src/app/backend/auth/jwe"
	"github.com/kubernetes/dashboard/src/app/backend/client"
	"github.com/kubernetes/dashboard/src/app/backend/handler"
	"github.com/kubernetes/dashboard/src/app/backend/integration"
	integrationapi "github.com/kubernetes/dashboard/src/app/backend/integration/api"
	"github.com/kubernetes/dashboard/src/app/backend/sync"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/pflag"
)

var (
	argInsecurePort        = pflag.Int("insecure-port", 9090, "The port to listen to for incoming HTTP requests.")
	argPort                = pflag.Int("port", 8443, "The secure port to listen to for incoming HTTPS requests.")
	argInsecureBindAddress = pflag.IP("insecure-bind-address", net.IPv4(127, 0, 0, 1), "The IP address on which to serve the --port (set to 0.0.0.0 for all interfaces).")
	argBindAddress         = pflag.IP("bind-address", net.IPv4(0, 0, 0, 0), "The IP address on which to serve the --secure-port (set to 0.0.0.0 for all interfaces).")
	argCertFile            = pflag.String("tls-cert-file", "", "File containing the default x509 Certificate for HTTPS.")
	argKeyFile             = pflag.String("tls-key-file", "", "File containing the default x509 private key matching --tls-cert-file.")
	argApiserverHost       = pflag.String("apiserver-host", "", "The address of the Kubernetes Apiserver "+
		"to connect to in the format of protocol://address:port, e.g., "+
		"http://localhost:8080. If not specified, the assumption is that the binary runs inside a "+
		"Kubernetes cluster and local discovery is attempted.")
	argHeapsterHost = pflag.String("heapster-host", "", "The address of the Heapster Apiserver "+
		"to connect to in the format of protocol://address:port, e.g., "+
		"http://localhost:8082. If not specified, the assumption is that the binary runs inside a "+
		"Kubernetes cluster and service proxy will be used.")
	argKubeConfigFile = pflag.String("kubeconfig", "", "Path to kubeconfig file with authorization and master location information.")
)

func main() {
	// Set logging output to standard console out
	log.SetOutput(os.Stdout)

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	flag.CommandLine.Parse(make([]string, 0)) // Init for glog calls in kubernetes packages

	if *argApiserverHost != "" {
		log.Printf("Using apiserver-host location: %s", *argApiserverHost)
	}
	if *argKubeConfigFile != "" {
		log.Printf("Using kubeconfig file: %s", *argKubeConfigFile)
	}

	clientManager := client.NewClientManager(*argKubeConfigFile, *argApiserverHost)
	versionInfo, err := clientManager.InsecureClient().Discovery().ServerVersion()
	if err != nil {
		handleFatalInitError(err)
	}

	log.Printf("Successful initial request to the apiserver, version: %s", versionInfo.String())

	// Init auth manager
	authManager := initAuthManager(clientManager)

	// Init integrations
	integrationManager := integration.NewIntegrationManager(clientManager)
	err = integrationManager.Metric().
		ConfigureHeapster(*argHeapsterHost).
		Enable(integrationapi.HeapsterIntegrationID)
	if err != nil {
		log.Printf("Could not enable metric client: %s. Continuing.", err)
	}

	apiHandler, err := handler.CreateHTTPAPIHandler(
		integrationManager,
		clientManager,
		authManager)
	if err != nil {
		handleFatalInitError(err)
	}

	// Run a HTTP server that serves static public files from './public' and handles API calls.
	// TODO(bryk): Disable directory listing.
	http.Handle("/", handler.MakeGzipHandler(handler.CreateLocaleHandler()))
	http.Handle("/api/", apiHandler)
	// TODO(maciaszczykm): Move to /appConfig.json as it was discussed in #640.
	http.Handle("/api/appConfig.json", handler.AppHandler(handler.ConfigHandler))
	http.Handle("/api/sockjs/", handler.CreateAttachHandler("/api/sockjs"))
	http.Handle("/metrics", prometheus.Handler())

	// Listen for http and https
	addr := fmt.Sprintf("%s:%d", *argInsecureBindAddress, *argInsecurePort)
	log.Printf("Serving insecurely on HTTP port: %d", *argInsecurePort)
	go func() { log.Fatal(http.ListenAndServe(addr, nil)) }()
	secureAddr := fmt.Sprintf("%s:%d", *argBindAddress, *argPort)
	if *argCertFile != "" && *argKeyFile != "" {
		log.Printf("Serving securely on HTTPS port: %d", *argPort)
		go func() { log.Fatal(http.ListenAndServeTLS(secureAddr, *argCertFile, *argKeyFile, nil)) }()
	}
	select {}
}

func initAuthManager(clientManager client.ClientManager) authApi.AuthManager {
	insecureClient := clientManager.InsecureClient()

	// Init default encryption key synchronizer
	synchronizerManager := sync.NewSynchronizerManager(insecureClient)
	keySynchronizer := synchronizerManager.Secret(authApi.EncryptionKeyHolderNamespace, authApi.EncryptionKeyHolderName)

	// Register synchronizer. Overwatch will be responsible for restarting it in case of error.
	sync.Overwatch.RegisterSynchronizer(keySynchronizer, sync.AlwaysRestart)

	// Init encryption key holder and token manager
	keyHolder := jwe.NewRSAKeyHolder(keySynchronizer)
	tokenManager := jwe.NewJWETokenManager(keyHolder)

	// Set token manager for client manager.
	clientManager.SetTokenManager(tokenManager)
	return auth.NewAuthManager(clientManager, tokenManager)
}

/**
 * Handles fatal init error that prevents server from doing any work. Prints verbose error
 * message and quits the server.
 */
func handleFatalInitError(err error) {
	log.Fatalf("Error while initializing connection to Kubernetes apiserver. "+
		"This most likely means that the cluster is misconfigured (e.g., it has "+
		"invalid apiserver certificates or service accounts configuration) or the "+
		"--apiserver-host param points to a server that does not exist. Reason: %s\n"+
		"Refer to the troubleshooting guide for more information: "+
		"https://github.com/kubernetes/dashboard/blob/master/docs/user-guide/troubleshooting.md", err)
}
