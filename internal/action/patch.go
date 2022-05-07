package action

import (
	"context"
	"fmt"
	"os"

	orascontent "oras.land/oras-go/pkg/content"

	"ocibundle/oci"
)

type PatchOpts struct {
	Origin string
}

func Patch(ctx context.Context, patchPath string, opts *PatchOpts) error {
	if _, err := os.Stat(patchPath); err != nil {
		return fmt.Errorf("patch path `%s' doesn't exist", patchPath)
	}
	info, err := getPatchInfo(patchPath)
	if err != nil {
		return err
	}
	source := opts.Origin
	if info != nil && info.Origin != "" {
		source = info.Origin
	}
	if source == "" {
		return fmt.Errorf("source reference must be specified")
	}

	registry := oci.NewRegistry()

	coreStore, err := orascontent.NewOCIStore(patchPath)
	if err != nil {
		return err
	}

	fetcher, err := registry.Fetcher(ctx, source)
	if err != nil {
		return err
	}
	bkStore := oci.NewFetcherStore(fetcher)
	store, err := oci.NewHybridStore(coreStore, bkStore)
	if err != nil {
		return err
	}

	refs := coreStore.ListReferences()
	for ref, sDesc := range refs {
		// resolver can't be reuse here, since dockerResolver track status doesn't has a namespace.
		// and when two ref has a same layer, only write once which make's other ref lack of the reference.
		registry.Reset()
		if err := registry.PushRef(ctx, store, ref, sDesc, oci.WithVerify); err != nil {
			return err
		}
	}

	return nil
}
