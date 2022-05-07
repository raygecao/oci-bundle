package oci

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/awalterschulze/gographviz"
	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/images"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type DescTree struct {
	root    *DescNode
	descMap map[digest.Digest]*DescNode
}

type DescNode struct {
	ocispec.Descriptor
	typ      string
	children []*DescNode
}

func (n *DescNode) addNode(graph *gographviz.Escape) (string, error) {
	if n.Digest.String() == "" {
		return "", fmt.Errorf("an empty descriptor")
	}
	name := n.Annotations[ocispec.AnnotationRefName]
	if name != "" {
		parts := strings.Split(name, "/")
		name = parts[len(parts)-1]
	}
	id := strings.TrimPrefix(n.Digest.String(), "sha256:")[:7]
	if name != "" {
		id = fmt.Sprintf("%s-%s", name, id)
	}
	attrs := map[string]string{
		string(gographviz.Label): id + "\n" + n.typ,
		string(gographviz.Style): "filled",
	}
	switch n.MediaType {
	case ocispec.MediaTypeImageIndex, images.MediaTypeDockerSchema2ManifestList:
		attrs[string(gographviz.FillColor)] = "cyan"
	case ocispec.MediaTypeImageManifest, images.MediaTypeDockerSchema2Manifest:
		attrs[string(gographviz.FillColor)] = "yellow"
	default:
	}
	err := graph.AddNode(graph.Name, id, attrs)
	if err != nil {
		return "", err
	}
	return id, nil
}

func GenerateDescTree(ctx context.Context, provider content.Provider, desc ocispec.Descriptor) (*DescTree, error) {
	root := &DescNode{Descriptor: desc}
	t := &DescTree{
		descMap: make(map[digest.Digest]*DescNode),
	}
	if err := t.fill(ctx, provider, root); err != nil {
		return nil, err
	}
	t.root = t.descMap[desc.Digest]
	return t, nil
}

func (t *DescTree) fill(ctx context.Context, provider content.Provider, root *DescNode) error {

	children, err := images.Children(ctx, provider, root.Descriptor)
	if err != nil {
		return err
	}
	typ, err := getType(ctx, provider, root.Descriptor)
	if err != nil {
		return err
	}
	root.typ = typ
	for _, child := range children {
		chdNode := &DescNode{Descriptor: child}
		if err := t.fill(ctx, provider, chdNode); err != nil {
			return err
		}
		root.children = append(root.children, chdNode)
	}
	t.descMap[root.Digest] = root

	return nil
}

func (t *DescTree) Has(dig digest.Digest) bool {
	_, ok := t.descMap[dig]
	return ok
}

func (t *DescTree) Graphviz(path string, hideLeaf bool) error {
	graphAst, _ := gographviz.ParseString(`digraph G {}`)
	graph := gographviz.NewEscape()
	if err := gographviz.Analyse(graphAst, graph); err != nil {
		return err
	}
	queue := []*DescNode{t.root}
	for len(queue) > 0 {
		parent := queue[0]
		queue = queue[1:]
		pn, err := parent.addNode(graph)
		if err != nil {
			return err
		}
		for _, chd := range parent.children {
			if hideLeaf && !images.IsIndexType(chd.Descriptor.MediaType) && !images.IsManifestType(chd.Descriptor.MediaType) {
				continue
			}
			cn, err := chd.addNode(graph)
			if err != nil {
				return err
			}
			if err := graph.AddEdge(pn, cn, true, nil); err != nil {
				return err
			}
			queue = append(queue, chd)
		}
	}
	os.MkdirAll(filepath.Dir(path), 0755)
	return ioutil.WriteFile(path, []byte(graph.String()), 0755)

}

func getType(ctx context.Context, provider content.Provider, descriptor ocispec.Descriptor) (string, error) {
	if images.IsIndexType(descriptor.MediaType) {
		p, err := content.ReadBlob(ctx, provider, descriptor)
		if err != nil {
			return "", err
		}
		var index ocispec.Index
		if err := json.Unmarshal(p, &index); err != nil {
			return "", err
		}
		if v, ok := index.Annotations[AnnotationOCIBundleType]; ok {
			return v, nil
		}
		return "image-index", nil
	}
	if images.IsManifestType(descriptor.MediaType) {
		p, err := content.ReadBlob(ctx, provider, descriptor)
		if err != nil {
			return "", err
		}
		var manifest ocispec.Manifest
		if err := json.Unmarshal(p, &manifest); err != nil {
			return "", err
		}
		if v, ok := manifest.Annotations[AnnotationOCIBundleType]; ok {
			return v, nil
		}
		return "image", nil
	}
	if images.IsConfigType(descriptor.MediaType) {
		return "config", nil
	}
	return "layer", nil
}
