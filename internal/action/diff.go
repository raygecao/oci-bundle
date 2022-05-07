package action

import (
	"context"
	"os"
	"path"

	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/log"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"gopkg.in/yaml.v3"
	orascontent "oras.land/oras-go/pkg/content"
	"oras.land/oras-go/pkg/oras"

	"ocibundle/oci"
)

const patchFile = ".ocibundle-patch"

type PatchInfo struct {
	Origin string
	Target string
}

type DiffOpts struct {
	Output string
}

func Diff(ctx context.Context, target, source string, opts *DiffOpts) error {
	// store origin descriptors
	registry := oci.NewRegistry()
	originFetcher, err := registry.Fetcher(ctx, source)
	if err != nil {
		return err
	}
	_, sourceDesc, err := registry.Resolve(ctx, source)
	if err != nil {
		return err
	}
	tree, err := oci.GenerateDescTree(ctx, oci.NewFetcherStore(originFetcher), sourceDesc)
	if err != nil {
		return err
	}

	// whether target is subset of origin
	_, targetDesc, err := registry.Resolve(ctx, target)
	if err != nil {
		return err
	}
	if tree.Has(targetDesc.Digest) {
		log.G(ctx).Warnf("%s has been contained by %s, diff is empty", target, source)
		return nil
	}

	targetFetcher, err := registry.Fetcher(ctx, target)
	if err != nil {
		return err
	}

	// diff target descriptors with origin descriptors
	ociStore, err := orascontent.NewOCIStore(opts.Output)
	if err != nil {
		return err
	}

	diffHandler := func(ctx context.Context, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
		if tree.Has(desc.Digest) {
			return nil, images.ErrStopHandler
		}
		return nil, nil
	}

	desc, artifacts, err := oras.Pull(ctx, registry, target, ociStore,
		oras.WithAllowedMediaTypes(nil),
		oras.WithPullEmptyNameAllowed(),
		oras.WithContentProvideIngester(
			oci.NewFetcherStore(targetFetcher).WithIngester(ociStore),
		),
		oras.WithPullBaseHandler(images.HandlerFunc(diffHandler)),
	)

	if err != nil {
		return err
	}
	ociStore.AddReference(target, desc)
	if err := saveIndex(ctx, desc, ociStore, true); err != nil {
		return err
	}
	_ = desc
	if len(artifacts) == 0 {
		log.G(ctx).Warnln("No diff generated")
	}
	log.G(ctx).Infof("get %d diffs", len(artifacts))
	return writePatchInfo(&PatchInfo{Origin: source, Target: target}, opts.Output)
}

func writePatchInfo(info *PatchInfo, output string) error {
	file, err := os.Create(path.Join(output, patchFile))
	if err != nil {
		return err
	}
	return yaml.NewEncoder(file).Encode(info)
}

func getPatchInfo(output string) (*PatchInfo, error) {
	file, err := os.Open(path.Join(output, patchFile))
	if err != nil {
		log.G(context.TODO()).Warnf("a non-standard patch is specified, a source reference must be specified")
		return nil, nil
	}
	info := new(PatchInfo)
	err = yaml.NewDecoder(file).Decode(info)
	if err != nil {
		return nil, err
	}
	return info, err
}
