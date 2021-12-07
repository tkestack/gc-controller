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
	"fmt"
	"net"

	v1 "k8s.io/api/core/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	apiserveroptions "k8s.io/apiserver/pkg/server/options"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	clientset "k8s.io/client-go/kubernetes"
	clientgokubescheme "k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"
	"k8s.io/component-base/metrics"
	cmoptions "k8s.io/controller-manager/options"
	kubectrlmgrconfigv1alpha1 "k8s.io/kube-controller-manager/config/v1alpha1"
	kubecontrollerconfig "k8s.io/kubernetes/cmd/kube-controller-manager/app/config"
	"k8s.io/kubernetes/pkg/cluster/ports"
	kubectrlmgrconfig "k8s.io/kubernetes/pkg/controller/apis/config"
	kubectrlmgrconfigscheme "k8s.io/kubernetes/pkg/controller/apis/config/scheme"
	"k8s.io/kubernetes/pkg/controller/garbagecollector"
	garbagecollectorconfig "k8s.io/kubernetes/pkg/controller/garbagecollector/config"

	// add the kubernetes feature gates
	_ "k8s.io/kubernetes/pkg/features"
)

const (
	// GCManagerUserAgent is the userAgent name when starting kube-controller managers.
	GCManagerUserAgent = "kube-controller-manager" //todo to check,what if change to platform-gc
)

// KubeControllerManagerOptions is the main context object for the kube-controller manager.
type GCManagerOptions struct {
	Generic                    *cmoptions.GenericControllerManagerConfigurationOptions
	GarbageCollectorController *GarbageCollectorControllerOptions
	SecureServing              *apiserveroptions.SecureServingOptionsWithLoopback
	// TODO: remove insecure serving mode
	InsecureServing *apiserveroptions.DeprecatedInsecureServingOptionsWithLoopback
	Authentication  *apiserveroptions.DelegatingAuthenticationOptions
	Authorization   *apiserveroptions.DelegatingAuthorizationOptions
	Metrics         *metrics.Options
	Logs            *logs.Options

	Master                      string
	Kubeconfig                  string
	ShowHiddenMetricsForVersion string
}

// NewKubeControllerManagerOptions creates a new KubeControllerManagerOptions with a default config.
func NewGCManagerOptions() (*GCManagerOptions, error) {
	componentConfig, err := NewDefaultComponentConfig(ports.InsecureKubeControllerManagerPort)
	if err != nil {
		return nil, err
	}

	s := GCManagerOptions{
		Generic: cmoptions.NewGenericControllerManagerConfigurationOptions(&componentConfig.Generic),

		GarbageCollectorController: &GarbageCollectorControllerOptions{
			&componentConfig.GarbageCollectorController,
		},

		SecureServing: apiserveroptions.NewSecureServingOptions().WithLoopback(),
		InsecureServing: (&apiserveroptions.DeprecatedInsecureServingOptions{
			BindAddress: net.ParseIP(componentConfig.Generic.Address),
			BindPort:    int(componentConfig.Generic.Port),
			BindNetwork: "tcp",
		}).WithLoopback(),
		Authentication: apiserveroptions.NewDelegatingAuthenticationOptions(),
		Authorization:  apiserveroptions.NewDelegatingAuthorizationOptions(),
		Metrics:        metrics.NewOptions(),
		Logs:           logs.NewOptions(),
	}

	s.Authentication.RemoteKubeConfigFileOptional = true
	s.Authorization.RemoteKubeConfigFileOptional = true

	// Set the PairName but leave certificate directory blank to generate in-memory by default
	s.SecureServing.ServerCert.CertDirectory = ""
	s.SecureServing.ServerCert.PairName = "kube-controller-manager"
	s.SecureServing.BindPort = ports.KubeControllerManagerPort

	gcIgnoredResources := make([]garbagecollectorconfig.GroupResource, 0, len(garbagecollector.DefaultIgnoredResources()))
	for r := range garbagecollector.DefaultIgnoredResources() {
		gcIgnoredResources = append(gcIgnoredResources, garbagecollectorconfig.GroupResource{Group: r.Group, Resource: r.Resource})
	}

	s.GarbageCollectorController.GCIgnoredResources = gcIgnoredResources
	s.Generic.LeaderElection.ResourceName = "kube-controller-manager"
	s.Generic.LeaderElection.ResourceNamespace = "kube-system"

	return &s, nil
}

// NewDefaultComponentConfig returns kube-controller manager configuration object.
func NewDefaultComponentConfig(insecurePort int32) (kubectrlmgrconfig.KubeControllerManagerConfiguration, error) {
	versioned := kubectrlmgrconfigv1alpha1.KubeControllerManagerConfiguration{}
	kubectrlmgrconfigscheme.Scheme.Default(&versioned)

	internal := kubectrlmgrconfig.KubeControllerManagerConfiguration{}
	if err := kubectrlmgrconfigscheme.Scheme.Convert(&versioned, &internal, nil); err != nil {
		return internal, err
	}
	internal.Generic.Port = insecurePort
	return internal, nil
}

// Flags returns flags for a specific APIServer by section name
func (s *GCManagerOptions) Flags() cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}
	s.Generic.AddFlags(&fss, []string{"gabagecollector"}, []string{})

	s.SecureServing.AddFlags(fss.FlagSet("secure serving"))
	s.InsecureServing.AddUnqualifiedFlags(fss.FlagSet("insecure serving"))
	s.Authentication.AddFlags(fss.FlagSet("authentication"))
	s.Authorization.AddFlags(fss.FlagSet("authorization"))

	s.GarbageCollectorController.AddFlags(fss.FlagSet("garbagecollector controller"))
	s.Metrics.AddFlags(fss.FlagSet("metrics"))
	s.Logs.AddFlags(fss.FlagSet("logs"))

	fs := fss.FlagSet("misc")
	fs.StringVar(&s.Master, "master", s.Master, "The address of the Kubernetes API server (overrides any value in kubeconfig).")
	fs.StringVar(&s.Kubeconfig, "kubeconfig", s.Kubeconfig, "Path to kubeconfig file with authorization and master location information.")
	utilfeature.DefaultMutableFeatureGate.AddFlag(fss.FlagSet("generic"))

	return fss
}

// ApplyTo fills up controller manager config with options.
func (s *GCManagerOptions) ApplyTo(c *kubecontrollerconfig.Config) error {
	if err := s.Generic.ApplyTo(&c.ComponentConfig.Generic); err != nil {
		return err
	}

	if err := s.GarbageCollectorController.ApplyTo(&c.ComponentConfig.GarbageCollectorController); err != nil {
		return err
	}

	if err := s.InsecureServing.ApplyTo(&c.InsecureServing, &c.LoopbackClientConfig); err != nil {
		return err
	}
	if err := s.SecureServing.ApplyTo(&c.SecureServing, &c.LoopbackClientConfig); err != nil {
		return err
	}
	if s.SecureServing.BindPort != 0 || s.SecureServing.Listener != nil {
		if err := s.Authentication.ApplyTo(&c.Authentication, c.SecureServing, nil); err != nil {
			return err
		}
		if err := s.Authorization.ApplyTo(&c.Authorization); err != nil {
			return err
		}
	}

	// sync back to component config
	// TODO: find more elegant way than syncing back the values.
	c.ComponentConfig.Generic.Port = int32(s.InsecureServing.BindPort)
	c.ComponentConfig.Generic.Address = s.InsecureServing.BindAddress.String()

	return nil
}

// Validate is used to validate the options and config before launching the controller manager
func (s *GCManagerOptions) Validate(allControllers []string, disabledByDefaultControllers []string) error {
	var errs []error

	errs = append(errs, s.Generic.Validate(allControllers, disabledByDefaultControllers)...)
	errs = append(errs, s.GarbageCollectorController.Validate()...)
	errs = append(errs, s.SecureServing.Validate()...)
	errs = append(errs, s.InsecureServing.Validate()...)
	errs = append(errs, s.Authentication.Validate()...)
	errs = append(errs, s.Authorization.Validate()...)
	errs = append(errs, s.Metrics.Validate()...)
	errs = append(errs, s.Logs.Validate()...)

	// TODO: validate component config, master and kubeconfig

	return utilerrors.NewAggregate(errs)
}

// Config return a controller manager config objective
func (s GCManagerOptions) Config() (*kubecontrollerconfig.Config, error) {
	if err := s.SecureServing.MaybeDefaultWithSelfSignedCerts("localhost", nil, []net.IP{net.ParseIP("127.0.0.1")}); err != nil {
		return nil, fmt.Errorf("error creating self-signed certificates: %v", err)
	}
	kubeconfig, err := clientcmd.BuildConfigFromFlags(s.Master, s.Kubeconfig)
	if err != nil {
		return nil, err
	}
	kubeconfig.DisableCompression = true
	kubeconfig.ContentConfig.AcceptContentTypes = s.Generic.ClientConnection.AcceptContentTypes
	kubeconfig.ContentConfig.ContentType = s.Generic.ClientConnection.ContentType
	kubeconfig.QPS = s.Generic.ClientConnection.QPS
	kubeconfig.Burst = int(s.Generic.ClientConnection.Burst)

	client, err := clientset.NewForConfig(restclient.AddUserAgent(kubeconfig, GCManagerUserAgent))
	if err != nil {
		return nil, err
	}

	eventRecorder := createRecorder(client, GCManagerUserAgent)
	c := &kubecontrollerconfig.Config{
		Client:        client,
		Kubeconfig:    kubeconfig,
		EventRecorder: eventRecorder,
	}
	if err := s.ApplyTo(c); err != nil {
		return nil, err
	}
	s.Metrics.Apply()

	s.Logs.Apply()
	return c, nil
}

func createRecorder(kubeClient clientset.Interface, userAgent string) record.EventRecorder {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	return eventBroadcaster.NewRecorder(clientgokubescheme.Scheme, v1.EventSource{Component: userAgent})
}
