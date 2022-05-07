package action

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/containerd/containerd/content"
	reference "github.com/containerd/containerd/reference/docker"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	orascontent "oras.land/oras-go/pkg/content"
	"oras.land/oras-go/pkg/oras"

	"ocibundle/internal/version"
	"ocibundle/oci"
)

type PushOpts struct {
	Paths       []string
	Type        string
	Author      string
	Annotations []string
	Content     []byte
	FileName    string
}

// Push pushes an artifacts to the oci registry with origin file path
func Push(ctx context.Context, ref string, opts *PushOpts) error {
	//named, err := registry.NormalizeRef(ref, opts.Category)
	named, err := reference.ParseNormalizedNamed(ref)
	if err != nil {
		return err
	}

	layerDescs, store, cleaner, err := parseOpts(opts)
	defer func() {
		if cleaner != nil {
			cleaner()
		}
	}()

	if err != nil {
		return err
	}
	if opts.Type == oci.TypeBundle {
		return fmt.Errorf("can't set type as reserved type %s", opts.Type)
	}
	annotations := map[string]string{
		oci.AnnotationOCIBundleVersion: version.GetVersion(),
		oci.AnnotationOCIBundleType:    opts.Type,
	}
	if opts.Author != "" {
		annotations[oci.AnnotationOCIBundleAuthor] = opts.Author
	}
	for _, annotation := range opts.Annotations {
		kv := strings.Split(annotation, "=")
		if len(kv) != 2 {
			return fmt.Errorf("the format of input annotations must be the format of `a=b,c=d'")
		}
		k, v := strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])
		if _, ok := annotations[k]; ok {
			return fmt.Errorf("cat set annotation with reserved key %s", k)
		}
		annotations[k] = v
	}

	_, err = oras.Push(ctx, oci.NewRegistry(), named.String(), store, layerDescs, oras.WithManifestAnnotations(annotations), oras.WithConfigMediaType(v1.MediaTypeImageConfig))
	if err != nil {
		return err
	}
	return nil
}

func parseOpts(opts *PushOpts) (descs []v1.Descriptor, provider content.Provider, cleaner func(), err error) {
	if len(opts.Paths) > 0 {
		store := orascontent.NewFileStore("")
		cleaner = func() {
			store.Close()
		}
		for _, path := range opts.Paths {
			name := filepath.Clean(path)
			desc, err := store.Add(name, "", path)
			if err != nil {
				return nil, nil, nil, err
			}
			descs = append(descs, desc)
		}
		return descs, store, cleaner, nil
	}
	if len(opts.Content) > 0 {
		store := orascontent.NewMemoryStore()
		desc := store.Add(opts.FileName, "", opts.Content)
		return []v1.Descriptor{desc}, store, cleaner, nil
	}
	return nil, nil, nil, fmt.Errorf("path or content must be specified at lease one")
}
