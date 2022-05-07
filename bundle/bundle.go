package bundle

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/containerd/containerd/reference/docker"
	reference "github.com/containerd/containerd/reference/docker"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	"ocibundle/oci"
)

type Bundle struct {
	Name      string `yaml:"name" json:"name"`
	Tag       string `yaml:"tag" json:"tag"`
	Artifacts []struct {
		Name    string `yaml:"name" json:"name"`
		Type    string `yaml:"type" json:"type"`
		Runtime string `yaml:"runtime,omitempty" json:"runtime,omitempty"`
	} `yaml:"artifacts" json:"artifacts"`
	Author      string            `yaml:"author,omitempty" json:"author,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty" json:"annotations,omitempty"`
}

func (b *Bundle) Upload(ctx context.Context) error {
	dRef, err := reference.ParseNormalizedNamed(b.Name)
	if err != nil {
		return err
	}
	registry := oci.NewRegistry()
	var manifests []v1.Descriptor

	for _, art := range b.Artifacts {
		sRef, err := reference.ParseNormalizedNamed(art.Name)
		if err != nil {
			return err
		}
		manifest, err := registry.LinkRepository(ctx, sRef, dRef)
		if err != nil {
			return err
		}
		if manifest.Annotations == nil {
			manifest.Annotations = make(map[string]string)
		}
		manifest.Annotations[v1.AnnotationRefName] = sRef.String()
		manifests = append(manifests, manifest)
	}

	indexRef, err := docker.WithTag(dRef, b.Tag)
	if err != nil {
		return err
	}
	content, err := json.Marshal(b)
	if err != nil {
		return err
	}
	configDesc, err := pushConfig(ctx, registry, indexRef.Name(), content)
	if err != nil {
		return err
	}
	annotations := map[string]string{
		oci.AnnotationOCIBundleType: oci.TypeBundle,
	}
	if b.Author != "" {
		annotations[oci.AnnotationOCIBundleAuthor] = b.Author
	}
	for k, v := range b.Annotations {
		if _, ok := annotations[k]; ok {
			return fmt.Errorf("bundle with reserved annotation key %s", k)
		}
		annotations[k] = v
	}
	indexDesc, indexPayload, err := prepareIndex(configDesc, manifests, annotations)
	if err != nil {
		return err
	}
	return registry.PushPayload(ctx, indexRef.String(), indexDesc, indexPayload)
}
