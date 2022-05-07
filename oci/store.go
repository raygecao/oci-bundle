package oci

import (
	"context"
	"fmt"
	"io"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/remotes"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// wrapper store is a wrapper of a provider or a ingester.
// A empty Info method is offered for called during remote.PushContext
type fetcherStore struct {
	content.Store
	provider content.Provider
	ingester content.Ingester
}

type WrapperOpt func(store *fetcherStore)

func NewFetcherStore(fetcher remotes.Fetcher) *fetcherStore {
	return &fetcherStore{provider: &fetcherProvider{fetcher: fetcher}}
}

func (w *fetcherStore) WithIngester(ingester content.Ingester) *fetcherStore {
	w.ingester = ingester
	return w
}

func (w *fetcherStore) Writer(ctx context.Context, opts ...content.WriterOpt) (content.Writer, error) {
	if w.ingester != nil {
		return w.ingester.Writer(ctx, opts...)
	}
	return nil, fmt.Errorf("not implement")
}

func (w *fetcherStore) ReaderAt(ctx context.Context, desc ocispec.Descriptor) (content.ReaderAt, error) {
	if w.provider != nil {
		return w.provider.ReaderAt(ctx, desc)
	}
	return nil, fmt.Errorf("not implement")
}

func (w *fetcherStore) Info(_ context.Context, _ digest.Digest) (content.Info, error) {
	return content.Info{}, nil
}

type fetcherProvider struct {
	fetcher remotes.Fetcher
}

func (p *fetcherProvider) ReaderAt(ctx context.Context, desc ocispec.Descriptor) (content.ReaderAt, error) {
	rc, err := p.fetcher.Fetch(ctx, desc)
	if err != nil {
		return nil, err
	}
	return &fetcherReaderAt{ReadCloser: rc, currentOffset: 0, size: desc.Size}, nil
}

type fetcherReaderAt struct {
	io.ReadCloser
	currentOffset int64
	size          int64
}

func (r *fetcherReaderAt) Size() int64 {
	return r.size
}

func (r *fetcherReaderAt) ReadAt(p []byte, off int64) (int, error) {
	if off != r.currentOffset {
		return 0, fmt.Errorf("at the moment this reader only supports offset at %d, requested offset was %d", r.currentOffset, off)
	}
	n, err := r.Read(p)
	r.currentOffset += int64(n)
	if err == io.EOF && n == len(p) {
		return n, nil
	}
	if err != nil || n == len(p) {
		return n, err
	}
	n2, err := r.ReadAt(p[n:], r.currentOffset)
	n += n2
	return n, err
}

type hybridStore struct {
	content.Store
	backups []content.Store
}

func NewHybridStore(stores ...content.Store) (*hybridStore, error) {
	if len(stores) == 0 {
		return nil, fmt.Errorf("must at least 1 store specified")
	}
	return &hybridStore{
		Store:   stores[0],
		backups: stores[1:],
	}, nil
}

func (h *hybridStore) ReaderAt(ctx context.Context, desc ocispec.Descriptor) (content.ReaderAt, error) {
	for _, cdt := range append([]content.Store{h.Store}, h.backups...) {
		if ra, err := cdt.ReaderAt(ctx, desc); err == nil {
			return ra, err
		}
	}
	//return h.Store.ReaderAt(ctx, desc)
	return nil, fmt.Errorf("err")
}

func (h *hybridStore) Writer(ctx context.Context, opts ...content.WriterOpt) (content.Writer, error) {
	for _, cdt := range append([]content.Store{h.Store}, h.backups...) {
		if wr, err := cdt.Writer(ctx, opts...); err == nil {
			return wr, err
		}
	}
	//return h.Store.Writer(ctx, opts...)
	return nil, fmt.Errorf("err")
}

func (h *hybridStore) Info(ctx context.Context, dgst digest.Digest) (content.Info, error) {
	for _, cdt := range append([]content.Store{h.Store}, h.backups...) {
		if wr, err := cdt.Info(ctx, dgst); err == nil {
			return wr, err
		}
	}
	//return h.Store.Writer(ctx, opts...)
	return content.Info{}, fmt.Errorf("err")
}
