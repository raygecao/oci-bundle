package action

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/log"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	orascontent "oras.land/oras-go/pkg/content"
	"oras.land/oras-go/pkg/oras"

	"ocibundle/oci"
)

type PackOpts struct {
	Output string
}

// Pack packs an artifact/bundle from OCI registry as the form of OCI-image layout
func Pack(ctx context.Context, ref string, opts *PackOpts) error {
	store, err := orascontent.NewOCIStore(opts.Output)
	if err != nil {
		return err
	}

	registry := oci.NewRegistry()
	fetcher, err := registry.Fetcher(ctx, ref)
	if err != nil {
		return err
	}
	desc, artifacts, err := oras.Pull(ctx, registry, ref, store,
		oras.WithAllowedMediaTypes(nil),
		oras.WithPullEmptyNameAllowed(),
		oras.WithContentProvideIngester(
			oci.NewFetcherStore(fetcher).WithIngester(store),
		),
	)
	if err != nil {
		return err
	}
	if _, err := store.Info(ctx, desc.Digest); err != nil {
		return fmt.Errorf("entry blob %s does not exist in ociStore", desc.Digest)
	}
	store.AddReference(ref, desc)
	if err := saveIndex(ctx, desc, store, false); err != nil {
		return err
	}
	if len(artifacts) == 0 {
		log.G(ctx).Warnln("Pull empty artifact")
	}
	log.G(ctx).Infof("pull %d artifacts with desc %+v", len(artifacts), desc)
	return nil
}

// saveIndex generates index.json for the layout.
func saveIndex(ctx context.Context, desc ocispec.Descriptor, store *orascontent.OCIStore, patch bool) error {
	if desc.MediaType == ocispec.MediaTypeImageIndex || desc.MediaType == images.MediaTypeDockerSchema2ManifestList {
		ra, err := store.ReaderAt(ctx, desc)
		if err != nil {
			return err
		}
		payload := make([]byte, desc.Size)
		_, err = ra.ReadAt(payload, 0)
		if err != nil {
			return err
		}
		var index ocispec.Index
		if err := json.Unmarshal(payload, &index); err != nil {
			return err
		}
		for _, m := range index.Manifests {
			if _, err := store.Info(ctx, m.Digest); err != nil {
				// diff for generating patch, allow some desc lost in blobs
				if patch {
					continue
				}
				return fmt.Errorf("blob %s not exist in ociStore", m.Digest)
			}
			if ref, ok := m.Annotations[ocispec.AnnotationRefName]; ok {
				store.AddReference(ref, m)
			}
			if m.MediaType == ocispec.MediaTypeImageIndex || m.MediaType == images.MediaTypeDockerSchema2ManifestList {
				if err := saveIndex(ctx, m, store, patch); err != nil {
					return err
				}
			}
		}
	}
	return store.SaveIndex()
}
