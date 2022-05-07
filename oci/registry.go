package oci

import (
	"context"
	"fmt"
	"net/http"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/log"
	reference "github.com/containerd/containerd/reference/docker"
	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
)

// Registry is a stub for push artifacts to registry
type Registry struct {
	remotes.Resolver
}

func NewRegistry() *Registry {
	return &Registry{Resolver: newResolver()}
}

type PushRefOpts struct {
	verify bool
}

type PushRefOpt func(opts *PushRefOpts)

func WithVerify(opts *PushRefOpts) {
	opts.verify = true
}

// LinkRepository links blob reference from repository `sRef' to repository `dRef' within a registry
func (r *Registry) LinkRepository(ctx context.Context, sRef, dRef reference.Named) (ocispec.Descriptor, error) {
	log.G(ctx).Debugf("begin to link artifacts from %s to %s", sRef.String(), dRef.String())
	_, desc, err := r.Resolve(ctx, sRef.String())
	if err != nil {
		return desc, err
	}
	fetcher, err := r.Fetcher(ctx, sRef.Name())
	if err != nil {
		return desc, err
	}
	store := NewFetcherStore(fetcher)

	pusher, err := r.Pusher(ctx, dRef.Name())
	if err != nil {
		return desc, err
	}
	err = remotes.PushContent(ctx, pusher, desc, store, nil, nil, nil)
	return desc, err
}

// PushRef pushes descriptor `sDesc' with all children iteratively to a ref with content store
func (r *Registry) PushRef(ctx context.Context, store content.Store, ref string, sDesc ocispec.Descriptor, opts ...PushRefOpt) error {
	pushRefOpts := new(PushRefOpts)
	for _, opt := range opts {
		opt(pushRefOpts)
	}

	_, dDesc, err := r.Resolve(ctx, ref)
	// the resolve ref err but not `NotFound'
	if err != nil && !errdefs.IsNotFound(err) {
		return err
	}

	// resolved ref exists
	if err == nil {
		log.G(ctx).Infof("ref %s has been exist with digest matches, skip push", ref)
		if sDesc.Digest.String() == dDesc.Digest.String() {
			// digest matches, don't need push again
			return nil
		}
		if pushRefOpts.verify {
			return fmt.Errorf("ref %s has been exist in registry and digest mismatch", ref)
		}
		log.G(ctx).Warnf("ref %s digest mismatch, origin digest: %s, overwrote with new digest %s", ref, sDesc.Digest, dDesc.Digest)
	}

	// resolved ref doesn't exist
	pusher, err := r.Pusher(ctx, ref)
	if err != nil {
		return err
	}
	if err := remotes.PushContent(ctx, pusher, sDesc, store, nil, nil, nil); err != nil {
		return err
	}
	log.G(ctx).Infof("successfully push ref %s", ref)
	return nil
}

// PushPayload pushes the payload to the descriptor
func (r *Registry) PushPayload(ctx context.Context, ref string, descriptor ocispec.Descriptor, payload []byte) error {
	log.G(ctx).Debugf("begin push ref %s with payload:\n%s ", ref, string(payload))

	pusher, err := r.Pusher(ctx, ref)
	if err != nil {
		return err
	}
	writer, err := pusher.Push(ctx, descriptor)
	if err != nil {
		if errors.Cause(err) == errdefs.ErrAlreadyExists {
			return nil
		}
		return err
	}
	defer writer.Close()
	if _, err := writer.Write(payload); err != nil {
		if errors.Cause(err) == errdefs.ErrAlreadyExists {
			return nil
		}
		return err
	}
	err = writer.Commit(ctx, descriptor.Size, descriptor.Digest)
	if errors.Cause(err) == errdefs.ErrAlreadyExists {
		return nil
	}
	return err
}

// Reset rest resolver for clear the track cache.
// Because trackState doesn't support namespace, a layer pushed to two repository with one resolver will be unexpected.
func (r *Registry) Reset() {
	r.Resolver = newResolver()
}

func newResolver() remotes.Resolver {
	// TODO: add auth
	opts := docker.ResolverOptions{
		PlainHTTP: true,
		Client:    http.DefaultClient,
	}
	return docker.NewResolver(opts)
}
