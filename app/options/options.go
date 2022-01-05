/*
Copyright 2014 The Kubernetes Authors.

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

// Package options provides the flags used for the controller manager.
//
package options

import (
	gcconfig "github.com/tkestack/gc-controller/app/config"
	v1 "k8s.io/api/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	clientgokubescheme "k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	cliflag "k8s.io/component-base/cli/flag"
	componentbaseconfig "k8s.io/component-base/config"
	"k8s.io/component-base/config/options"
	"k8s.io/kubernetes/pkg/controller/garbagecollector"
	garbagecollectorconfig "k8s.io/kubernetes/pkg/controller/garbagecollector/config"

	// add the kubernetes feature gates
	"k8s.io/component-base/logs"
	_ "k8s.io/kubernetes/pkg/features"
)

const (
	// KubeControllerManagerUserAgent is the userAgent name when starting kube-controller managers.
	KubeControllerManagerUserAgent = "kube-controller-manager"
)

type GarbageCollectorControllerOptions struct {
	LeaderElection         componentbaseconfig.LeaderElectionConfiguration
	ClientConnection       componentbaseconfig.ClientConnectionConfiguration
	EnableGarbageCollector bool
	ConcurrentGCSyncs      int32
	GCIgnoredResources     []garbagecollectorconfig.GroupResource
	GCGroup                string
	Master                 string
	Kubeconfig             string
	Logs                   *logs.Options
}

// NewKubeControllerManagerOptions creates a new KubeControllerManagerOptions with a default config.
func NewPlatformGcControllerOptions() (*GarbageCollectorControllerOptions, error) {
	g := GarbageCollectorControllerOptions{
		ClientConnection:       componentbaseconfig.ClientConnectionConfiguration{},
		EnableGarbageCollector: true,
		ConcurrentGCSyncs:      5,
		GCIgnoredResources:     []garbagecollectorconfig.GroupResource{},
		Master:                 "",
		Kubeconfig:             "",
		Logs:                   logs.NewOptions(),
	}
	gcIgnoredResources := make([]garbagecollectorconfig.GroupResource, 0, len(garbagecollector.DefaultIgnoredResources()))
	for r := range garbagecollector.DefaultIgnoredResources() {
		gcIgnoredResources = append(gcIgnoredResources, garbagecollectorconfig.GroupResource{Group: r.Group, Resource: r.Resource})
	}
	g.GCIgnoredResources = gcIgnoredResources
	g.LeaderElection.ResourceName = "platfor-gc-controller"
	g.LeaderElection.ResourceNamespace = "kube-system"

	return &g, nil
}

// Flags returns flags for a specific APIServer by section name
func (g *GarbageCollectorControllerOptions) Flags() cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}
	genericfs := fss.FlagSet("generic")
	genericfs.StringVar(&g.ClientConnection.ContentType, "kube-api-content-type", g.ClientConnection.ContentType, "Content type of requests sent to apiserver.")
	genericfs.Float32Var(&g.ClientConnection.QPS, "kube-api-qps", g.ClientConnection.QPS, "QPS to use while talking with kubernetes apiserver.")
	genericfs.Int32Var(&g.ClientConnection.Burst, "kube-api-burst", g.ClientConnection.Burst, "Burst to use while talking with kubernetes apiserver.")
	genericfs.StringVar(&g.Master, "master", g.Master, "The address of the Kubernetes API server (overrides any value in kubeconfig).")
	genericfs.StringVar(&g.Kubeconfig, "kubeconfig", "/app/conf/tdcc-platform-config.conf", "kubeconfig file path with authorization and master location info.")
	options.BindLeaderElectionFlags(&g.LeaderElection, genericfs)

	gcfs := fss.FlagSet("garbagecollector controller")
	gcfs.Int32Var(&g.ConcurrentGCSyncs, "concurrent-gc-syncs", g.ConcurrentGCSyncs, "Number of garbage collector workers that are allowed to sync concurrently.")
	gcfs.BoolVar(&g.EnableGarbageCollector, "enable-garbage-collector", g.EnableGarbageCollector, "Enables the generic garbage collector. MUST be synced with the corresponding flag of the kube-apiserver.")
	gcfs.StringVar(&g.GCGroup, "gcgroup", "platform.tkestack.io", "specify resoure group to be gc,useage:--gcgroup=platform.tkestack.io")
	g.Logs.AddFlags(fss.FlagSet("logs"))
	return fss
}

// ApplyTo fills up controller manager config with options.
func (g *GarbageCollectorControllerOptions) ApplyTo(cfg *gcconfig.Config) error {
	if g == nil {
		return nil
	}
	cfg.ClientConnection = g.ClientConnection
	cfg.LeaderElection = g.LeaderElection

	cfg.ConcurrentGCSyncs = g.ConcurrentGCSyncs
	cfg.GCIgnoredResources = g.GCIgnoredResources
	cfg.EnableGarbageCollector = g.EnableGarbageCollector
	cfg.GCGroup = g.GCGroup
	cfg.KubeConfPath = g.Kubeconfig

	return nil
}

// Config return a controller manager config objective
func (s GarbageCollectorControllerOptions) Config() (*gcconfig.Config, error) {
	kubeconfig, err := clientcmd.BuildConfigFromFlags(s.Master, s.Kubeconfig)
	if err != nil {
		return nil, err
	}
	kubeconfig.DisableCompression = true
	kubeconfig.ContentConfig.AcceptContentTypes = s.ClientConnection.AcceptContentTypes
	kubeconfig.ContentConfig.ContentType = s.ClientConnection.ContentType
	kubeconfig.QPS = s.ClientConnection.QPS
	kubeconfig.Burst = int(s.ClientConnection.Burst)

	client, err := clientset.NewForConfig(restclient.AddUserAgent(kubeconfig, KubeControllerManagerUserAgent))
	if err != nil {
		return nil, err
	}

	eventRecorder := createRecorder(client, KubeControllerManagerUserAgent)
	c := &gcconfig.Config{
		Client:        client,
		Kubeconfig:    kubeconfig,
		EventRecorder: eventRecorder,
	}
	if err := s.ApplyTo(c); err != nil {
		return nil, err
	}

	return c, nil
}

func createRecorder(kubeClient clientset.Interface, userAgent string) record.EventRecorder {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	return eventBroadcaster.NewRecorder(clientgokubescheme.Scheme, v1.EventSource{Component: userAgent})
}
