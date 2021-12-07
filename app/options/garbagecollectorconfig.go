package options

import (
	"github.com/spf13/pflag"

	garbagecollectorconfig "k8s.io/kubernetes/pkg/controller/garbagecollector/config"
)

// GarbageCollectorControllerOptions holds the GarbageCollectorController options.
type GarbageCollectorControllerOptions struct {
	*garbagecollectorconfig.GarbageCollectorControllerConfiguration
}

// AddFlags adds flags related to GarbageCollectorController for controller manager to the specified FlagSet.
func (o *GarbageCollectorControllerOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}

	fs.Int32Var(&o.ConcurrentGCSyncs, "concurrent-gc-syncs", o.ConcurrentGCSyncs, "The number of garbage collector workers that are allowed to sync concurrently.")
	fs.BoolVar(&o.EnableGarbageCollector, "enable-garbage-collector", o.EnableGarbageCollector, "Enables the generic garbage collector. MUST be synced with the corresponding flag of the kube-apiserver.")
	fs.StringVar(&o.GCgroup, "gcgroup", "platform.tkestack.io", "specified resourece group to be gc")
}

// ApplyTo fills up GarbageCollectorController config with options.
func (o *GarbageCollectorControllerOptions) ApplyTo(cfg *garbagecollectorconfig.GarbageCollectorControllerConfiguration) error {
	if o == nil {
		return nil
	}

	cfg.ConcurrentGCSyncs = o.ConcurrentGCSyncs
	cfg.GCIgnoredResources = o.GCIgnoredResources
	cfg.EnableGarbageCollector = o.EnableGarbageCollector
	cfg.GCgroup = o.GCgroup

	return nil
}

// Validate checks validation of GarbageCollectorController.
func (o *GarbageCollectorControllerOptions) Validate() []error {
	if o == nil {
		return nil
	}

	errs := []error{}
	return errs
}
