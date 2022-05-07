package bundle

import (
	"context"
	"encoding/json"

	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"ocibundle/oci"
)

func prepareIndex(config ocispec.Descriptor, manifests []ocispec.Descriptor, annotations map[string]string) (ocispec.Descriptor, []byte, error) {

	all := make([]ocispec.Descriptor, len(manifests)+1)
	all[0] = config
	copy(all[1:], manifests)
	index := ocispec.Index{
		Versioned: specs.Versioned{
			SchemaVersion: 2,
		},
		Manifests:   all,
		Annotations: annotations,
	}
	indexPayload, err := json.Marshal(index)
	if err != nil {
		return ocispec.Descriptor{}, nil, err
	}
	indexDesc := ocispec.Descriptor{
		Digest:      digest.FromBytes(indexPayload),
		MediaType:   ocispec.MediaTypeImageIndex,
		Size:        int64(len(indexPayload)),
		Annotations: annotations,
	}
	return indexDesc, indexPayload, nil
}

func pushConfig(ctx context.Context, registry *oci.Registry, ref string, configPayload []byte) (ocispec.Descriptor, error) {
	// push config payload
	configDesc := ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageConfig,
		Digest:    digest.FromBytes(configPayload),
		Size:      int64(len(configPayload)),
	}
	err := registry.PushPayload(ctx, ref, configDesc, configPayload)
	if err != nil {
		return ocispec.Descriptor{}, nil
	}

	manifest := ocispec.Manifest{
		Versioned: specs.Versioned{
			SchemaVersion: 2,
		},
		Config: configDesc,
		Annotations: map[string]string{
			oci.AnnotationOCIBundleType: "bundle-config",
		},
	}

	// push config manifest payload
	manifestPayload, err := json.Marshal(manifest)
	if err != nil {
		return ocispec.Descriptor{}, err
	}

	manifestDesc := ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageManifest,
		Digest:    digest.FromBytes(manifestPayload),
		Size:      int64(len(manifestPayload)),
	}
	err = registry.PushPayload(ctx, ref, manifestDesc, manifestPayload)

	return manifestDesc, err
}
