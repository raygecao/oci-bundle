package action

import (
	"context"

	orascontent "oras.land/oras-go/pkg/content"

	"ocibundle/oci"
)

// Load loads artifacts/bundles from the OCI-image layout to registry
func Load(ctx context.Context, path string) error {
	store, err := orascontent.NewOCIStore(path)
	if err != nil {
		return err
	}
	registry := oci.NewRegistry()
	refs := store.ListReferences()
	for ref, sDesc := range refs {
		if err := registry.PushRef(ctx, store, ref, sDesc, oci.WithVerify); err != nil {
			return err
		}
		// resolver can't be reuse here, since dockerResolver track status doesn't has a namespace.
		// and when two ref has a same layer, only write once which make's other ref lack of the reference.
		registry.Reset()
	}

	return nil
}
