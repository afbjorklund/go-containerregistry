// Copyright 2018 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package crane

import (
	"fmt"
	"os"

	legacy "github.com/google/go-containerregistry/pkg/legacy/tarball"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

// Tag applied to images that were pulled by digest. This denotes that the
// image was (probably) never tagged with this, but lets us avoid applying the
// ":latest" tag which might be misleading.
const iWasADigestTag = "i-was-a-digest"

// Pull returns a v1.Image of the remote image src.
func Pull(src string, opt ...Option) (v1.Image, error) {
	o := makeOptions(opt...)
	ref, err := name.ParseReference(src, o.name...)
	if err != nil {
		return nil, fmt.Errorf("parsing tag %q: %v", src, err)
	}

	return remote.Image(ref, o.remote...)
}

// Save writes the v1.Image img as a tarball at path with tag src.
func Save(img v1.Image, src, path string) error {
	ref, err := name.ParseReference(src)
	if err != nil {
		return fmt.Errorf("parsing ref %q: %v", src, err)
	}

	// WriteToFile wants a tag to write to the tarball, but we might have
	// been given a digest.
	// If the original ref was a tag, use that. Otherwise, if it was a
	// digest, tag the image with :i-was-a-digest instead.
	tag, ok := ref.(name.Tag)
	if !ok {
		d, ok := ref.(name.Digest)
		if !ok {
			return fmt.Errorf("ref wasn't a tag or digest")
		}
		if d.TagStr() != "" {
			tag = d.Tag()
		} else {
			tag = d.Repository.Tag(iWasADigestTag)
		}
	}

	return tarball.WriteToFile(path, tag, img)
}

// PullLayer returns the given layer from a registry.
func PullLayer(ref string, opt ...Option) (v1.Layer, error) {
	o := makeOptions(opt...)
	digest, err := name.NewDigest(ref, o.name...)
	if err != nil {
		return nil, err
	}

	return remote.Layer(digest, o.remote...)
}

// SaveLegacy writes the v1.Image img as a legacy tarball at path with tag src.
func SaveLegacy(img v1.Image, src, path string) error {
	ref, err := name.ParseReference(src)
	if err != nil {
		return fmt.Errorf("parsing ref %q: %v", src, err)
	}

	w, err := os.Create(path)
	if err != nil {
		return err
	}
	defer w.Close()

	return legacy.Write(ref, img, w)
}

// SaveOCI writes the v1.Image img as an OCI Image Layout at path. If a layout
// already exists at that path, it will add the image to the index.
func SaveOCI(img v1.Image, path string) error {
	p, err := layout.FromPath(path)
	if err != nil {
		p, err = layout.Write(path, empty.Index)
		if err != nil {
			return err
		}
	}
	return p.AppendImage(img)
}
