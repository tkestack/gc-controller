/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	componentbaseconfig "k8s.io/component-base/config"
	gcconfig "k8s.io/kubernetes/pkg/controller/garbagecollector/config"
)

// Config is the main context object for the controller manager.
type Config struct {
	ClientConnection componentbaseconfig.ClientConnectionConfiguration
	LeaderElection   componentbaseconfig.LeaderElectionConfiguration

	EnableGarbageCollector bool
	ConcurrentGCSyncs      int32
	GCIgnoredResources     []gcconfig.GroupResource
	GCGroup                string

	Client        *clientset.Clientset
	Kubeconfig    *restclient.Config
	EventRecorder record.EventRecorder
	KubeConfPath  string
}

type completedConfig struct {
	*Config
}

// CompletedConfig same as Config, just to swap private object.
type CompletedConfig struct {
	// Embed a private pointer that cannot be instantiated outside of this package.
	*completedConfig
}
