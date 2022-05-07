package action

import (
	"context"

	"github.com/containerd/containerd/log"
	orascontent "oras.land/oras-go/pkg/content"
	"oras.land/oras-go/pkg/oras"

	"ocibundle/oci"
)

type PullOpts struct {
	Output         string
	AllowOverwrite bool
}

// Pull pulls an artifact from OCI registry.
func Pull(ctx context.Context, ref string, opts *PullOpts) error {
	store := orascontent.NewFileStore(opts.Output)
	store.DisableOverwrite = !opts.AllowOverwrite
	defer store.Close()
	registry := oci.NewRegistry()
	desc, artifacts, err := oras.Pull(ctx, registry, ref, store,
		oras.WithAllowedMediaTypes(nil),
		oras.WithPullEmptyNameAllowed(),
	)
	if err != nil {
		return err
	}
	if len(artifacts) == 0 {
		log.G(ctx).Warnln("Pack empty artifact")
	}
	log.G(ctx).Infof("pulled %s with desc %+v", ref, desc)
	return nil
}
