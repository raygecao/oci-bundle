package action

import (
	"context"
	"path/filepath"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"ocibundle/oci"
)

type PlotOpts struct {
	Output    string
	HideLayer bool
}

func Plot(ctx context.Context, ref string, opts *PlotOpts) error {
	registry := oci.NewRegistry()
	fetcher, err := registry.Fetcher(ctx, ref)
	if err != nil {
		return err
	}
	_, desc, err := registry.Resolve(ctx, ref)
	if err != nil {
		return err
	}
	if desc.Annotations == nil {
		desc.Annotations = make(map[string]string)
	}
	desc.Annotations[ocispec.AnnotationRefName] = ref
	tree, err := oci.GenerateDescTree(ctx, oci.NewFetcherStore(fetcher), desc)
	if err != nil {
		return err
	}
	return tree.Graphviz(filepath.Clean(opts.Output), opts.HideLayer)
}
