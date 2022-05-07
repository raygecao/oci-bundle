package action

import (
	"context"
	"io/ioutil"

	"gopkg.in/yaml.v3"

	"ocibundle/bundle"
)

func Bundle(ctx context.Context, bundlePath string) error {
	ct, err := ioutil.ReadFile(bundlePath)
	if err != nil {
		return err
	}
	bd := new(bundle.Bundle)
	if err := yaml.Unmarshal(ct, &bd); err != nil {
		return err
	}
	return bd.Upload(ctx)
}
