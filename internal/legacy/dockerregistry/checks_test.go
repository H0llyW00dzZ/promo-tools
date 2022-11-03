/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package inventory_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	grafeaspb "google.golang.org/genproto/googleapis/grafeas/v1"

	reg "sigs.k8s.io/promo-tools/v3/internal/legacy/dockerregistry"
	"sigs.k8s.io/promo-tools/v3/internal/legacy/dockerregistry/registry"
	"sigs.k8s.io/promo-tools/v3/internal/legacy/dockerregistry/schema"
	"sigs.k8s.io/promo-tools/v3/types/image"
)

func TestImageRemovalCheck(t *testing.T) {
	srcRegName := image.Registry("gcr.io/foo")
	srcRegName2 := image.Registry("gcr.io/foo2")
	destRegName := image.Registry("gcr.io/bar")
	destRC := registry.Context{
		Name:           destRegName,
		ServiceAccount: "robot",
	}
	srcRC := registry.Context{
		Name:           srcRegName,
		ServiceAccount: "robot",
		Src:            true,
	}
	srcRC2 := registry.Context{
		Name:           srcRegName2,
		ServiceAccount: "robot",
		Src:            true,
	}
	registries := []registry.Context{destRC, srcRC}
	registries2 := []registry.Context{destRC, srcRC, srcRC2}

	imageA := registry.Image{
		Name: "a",
		Dmap: registry.DigestTags{
			"sha256:000": {"0.9"},
		},
	}

	imageA2 := registry.Image{
		Name: "a",
		Dmap: registry.DigestTags{
			"sha256:111": {"0.9"},
		},
	}

	imageB := registry.Image{
		Name: "b",
		Dmap: registry.DigestTags{
			"sha256:000": {"0.9"},
		},
	}

	tests := []struct {
		name            string
		check           reg.ImageRemovalCheck
		masterManifests []schema.Manifest
		pullManifests   []schema.Manifest
		expected        error
	}{
		{
			"Empty manifests",
			reg.ImageRemovalCheck{},
			[]schema.Manifest{},
			[]schema.Manifest{},
			nil,
		},
		{
			"Same manifests",
			reg.ImageRemovalCheck{},
			[]schema.Manifest{
				{
					Registries: registries,
					Images: []registry.Image{
						imageA,
					},
					SrcRegistry: &srcRC,
				},
			},
			[]schema.Manifest{
				{
					Registries: registries,
					Images: []registry.Image{
						imageA,
					},
					SrcRegistry: &srcRC,
				},
			},
			nil,
		},
		{
			"Different manifests",
			reg.ImageRemovalCheck{},
			[]schema.Manifest{
				{
					Registries: registries,
					Images: []registry.Image{
						imageA,
					},
					SrcRegistry: &srcRC,
				},
			},
			[]schema.Manifest{
				{
					Registries: registries,
					Images: []registry.Image{
						imageB,
					},
					SrcRegistry: &srcRC,
				},
			},
			fmt.Errorf(
				"the following images were removed in this pull request: a",
			),
		},
		{
			"Promoting same image from different registry",
			reg.ImageRemovalCheck{},
			[]schema.Manifest{
				{
					Registries: registries2,
					Images: []registry.Image{
						imageA,
					},
					SrcRegistry: &srcRC,
				},
			},
			[]schema.Manifest{
				{
					Registries: registries2,
					Images: []registry.Image{
						imageA,
					},
					SrcRegistry: &srcRC2,
				},
			},
			nil,
		},
		{
			"Promoting image with same name and different digest",
			reg.ImageRemovalCheck{},
			[]schema.Manifest{
				{
					Registries: registries,
					Images: []registry.Image{
						imageA,
					},
					SrcRegistry: &srcRC,
				},
			},
			[]schema.Manifest{
				{
					Registries: registries,
					Images: []registry.Image{
						imageA2,
					},
					SrcRegistry: &srcRC,
				},
			},
			fmt.Errorf(
				"the following images were removed in this pull request: a",
			),
		},
	}

	for _, test := range tests {
		// TODO: Why are we not checking errors here?
		//nolint:errcheck
		masterEdges, _ := reg.ToPromotionEdges(test.masterManifests)
		// TODO: Why are we not checking errors here?
		//nolint:errcheck
		pullEdges, _ := reg.ToPromotionEdges(test.pullManifests)
		got := test.check.Compare(masterEdges, pullEdges)
		require.Equal(t, test.expected, got)
	}
}

func TestImageSizeCheck(t *testing.T) {
	srcRegName := image.Registry("gcr.io/foo")
	destRegName := image.Registry("gcr.io/bar")

	destRC := registry.Context{
		Name:           destRegName,
		ServiceAccount: "robot",
	}

	srcRC := registry.Context{
		Name:           srcRegName,
		ServiceAccount: "robot",
		Src:            true,
	}

	registries := []registry.Context{destRC, srcRC}

	image1 := registry.Image{
		Name: "foo",
		Dmap: registry.DigestTags{
			"sha256:000": {"0.9"},
		},
	}
	image2 := registry.Image{
		Name: "bar",
		Dmap: registry.DigestTags{
			"sha256:111": {"0.9"},
		},
	}

	tests := []struct {
		name       string
		check      reg.ImageSizeCheck
		manifests  []schema.Manifest
		imageSizes map[image.Digest]int
		expected   error
	}{
		{
			"Image size under the max size",
			reg.ImageSizeCheck{
				MaxImageSize:    1,
				DigestImageSize: make(reg.DigestImageSize),
			},
			[]schema.Manifest{
				{
					Registries: registries,
					Images: []registry.Image{
						image1,
					},
					SrcRegistry: &srcRC,
				},
			},
			map[image.Digest]int{
				"sha256:000": reg.MBToBytes(1),
			},
			nil,
		},
		{
			"Image size over the max size",
			reg.ImageSizeCheck{
				MaxImageSize:    1,
				DigestImageSize: make(reg.DigestImageSize),
			},
			[]schema.Manifest{
				{
					Registries: registries,
					Images: []registry.Image{
						image1,
					},
					SrcRegistry: &srcRC,
				},
			},
			map[image.Digest]int{
				"sha256:000": reg.MBToBytes(5),
			},
			reg.ImageSizeError{
				1,
				map[string]int{
					"foo": reg.MBToBytes(5),
				},
				map[string]int{},
			},
		},
		{
			"Multiple images over the max size",
			reg.ImageSizeCheck{
				MaxImageSize:    1,
				DigestImageSize: make(reg.DigestImageSize),
			},
			[]schema.Manifest{
				{
					Registries: registries,
					Images: []registry.Image{
						image1,
						image2,
					},
					SrcRegistry: &srcRC,
				},
			},
			map[image.Digest]int{
				"sha256:000": reg.MBToBytes(5),
				"sha256:111": reg.MBToBytes(10),
			},
			reg.ImageSizeError{
				1,
				map[string]int{
					"foo": reg.MBToBytes(5),
					"bar": reg.MBToBytes(10),
				},
				map[string]int{},
			},
		},
		{
			"Image sizes are <= 0",
			reg.ImageSizeCheck{
				MaxImageSize:    1,
				DigestImageSize: make(reg.DigestImageSize),
			},
			[]schema.Manifest{
				{
					Registries: registries,
					Images: []registry.Image{
						image1,
						image2,
					},
					SrcRegistry: &srcRC,
				},
			},
			map[image.Digest]int{
				"sha256:000": 0,
				"sha256:111": reg.MBToBytes(-5),
			},
			reg.ImageSizeError{
				1,
				map[string]int{},
				map[string]int{
					"foo": 0,
					"bar": reg.MBToBytes(-5),
				},
			},
		},
	}

	for _, test := range tests {
		// TODO: Why are we not checking errors here?
		//nolint:errcheck
		test.check.PullEdges, _ = reg.ToPromotionEdges(test.manifests)
		require.Equal(t, len(test.imageSizes), len(test.check.PullEdges))

		for edge := range test.check.PullEdges {
			test.check.DigestImageSize[edge.Digest] = test.imageSizes[edge.Digest]
		}

		got := test.check.Run()
		require.Equal(t, test.expected, got)
	}
}

// TestImageVulnCheck uses a fake populateRequests function and a fake
// vulnerability producer. The fake vulnerability producer simply returns the
// vulnerability occurrences that have been mapped to a given PromotionEdge in
// order to simulate running the real check without having to make an api call
// to the Container Analysis Service.
func TestImageVulnCheck(t *testing.T) {
	edge1 := reg.PromotionEdge{
		SrcImageTag: reg.ImageTag{
			Name: "foo",
		},
		Digest: "sha256:000",
		DstImageTag: reg.ImageTag{
			Name: "foo",
		},
	}
	edge2 := reg.PromotionEdge{
		SrcImageTag: reg.ImageTag{
			Name: "bar",
		},
		Digest: "sha256:111",
		DstImageTag: reg.ImageTag{
			Name: "bar/1",
		},
	}
	edge3 := reg.PromotionEdge{
		SrcImageTag: reg.ImageTag{
			Name: "bar",
		},
		Digest: "sha256:111",
		DstImageTag: reg.ImageTag{
			Name: "bar/2",
		},
	}

	mkVulnProducerFake := func(
		edgeVulnOccurrences map[image.Digest][]*grafeaspb.Occurrence,
	) reg.ImageVulnProducer {
		return func(
			edge reg.PromotionEdge,
		) ([]*grafeaspb.Occurrence, error) {
			return edgeVulnOccurrences[edge.Digest], nil
		}
	}

	tests := []struct {
		name              string
		severityThreshold int
		edges             map[reg.PromotionEdge]interface{}
		vulnerabilities   map[image.Digest][]*grafeaspb.Occurrence
		expected          error
	}{
		{
			"Severity under threshold",
			int(grafeaspb.Severity_HIGH),
			map[reg.PromotionEdge]interface{}{
				edge1: nil,
				edge2: nil,
			},
			map[image.Digest][]*grafeaspb.Occurrence{
				"sha256:000": {
					{
						Details: &grafeaspb.Occurrence_Vulnerability{
							Vulnerability: &grafeaspb.VulnerabilityOccurrence{
								Severity:     grafeaspb.Severity_LOW,
								FixAvailable: true,
							},
						},
					},
				},
				"sha256:111": {
					{
						Details: &grafeaspb.Occurrence_Vulnerability{
							Vulnerability: &grafeaspb.VulnerabilityOccurrence{
								Severity:     grafeaspb.Severity_MEDIUM,
								FixAvailable: true,
							},
						},
					},
				},
			},
			nil,
		},
		{
			"Severity at threshold",
			int(grafeaspb.Severity_HIGH),
			map[reg.PromotionEdge]interface{}{
				edge1: nil,
				edge2: nil,
			},
			map[image.Digest][]*grafeaspb.Occurrence{
				"sha256:000": {
					{
						Details: &grafeaspb.Occurrence_Vulnerability{
							Vulnerability: &grafeaspb.VulnerabilityOccurrence{
								Severity:     grafeaspb.Severity_HIGH,
								FixAvailable: true,
							},
						},
					},
				},
				"sha256:111": {
					{
						Details: &grafeaspb.Occurrence_Vulnerability{
							Vulnerability: &grafeaspb.VulnerabilityOccurrence{
								Severity:     grafeaspb.Severity_HIGH,
								FixAvailable: true,
							},
						},
					},
				},
			},
			fmt.Errorf("VulnerabilityCheck: The following vulnerable images were found:\n" +
				"    bar@sha256:111 [1 fixable severe vulnerabilities, 1 total]\n" +
				"    foo@sha256:000 [1 fixable severe vulnerabilities, 1 total]"),
		},
		{
			"Severity above threshold",
			int(grafeaspb.Severity_MEDIUM),
			map[reg.PromotionEdge]interface{}{
				edge1: nil,
				edge2: nil,
			},
			map[image.Digest][]*grafeaspb.Occurrence{
				"sha256:000": {
					{
						Details: &grafeaspb.Occurrence_Vulnerability{
							Vulnerability: &grafeaspb.VulnerabilityOccurrence{
								Severity:     grafeaspb.Severity_HIGH,
								FixAvailable: true,
							},
						},
					},
				},
				"sha256:111": {
					{
						Details: &grafeaspb.Occurrence_Vulnerability{
							Vulnerability: &grafeaspb.VulnerabilityOccurrence{
								Severity:     grafeaspb.Severity_CRITICAL,
								FixAvailable: true,
							},
						},
					},
					{
						Details: &grafeaspb.Occurrence_Vulnerability{
							Vulnerability: &grafeaspb.VulnerabilityOccurrence{
								Severity:     grafeaspb.Severity_HIGH,
								FixAvailable: true,
							},
						},
					},
				},
			},
			fmt.Errorf("VulnerabilityCheck: The following vulnerable images were found:\n" +
				"    bar@sha256:111 [2 fixable severe vulnerabilities, 2 total]\n" +
				"    foo@sha256:000 [1 fixable severe vulnerabilities, 1 total]"),
		},
		{
			"Multiple edges with same source image",
			int(grafeaspb.Severity_MEDIUM),
			map[reg.PromotionEdge]interface{}{
				edge2: nil,
				edge3: nil,
			},
			map[image.Digest][]*grafeaspb.Occurrence{
				"sha256:111": {
					{
						Details: &grafeaspb.Occurrence_Vulnerability{
							Vulnerability: &grafeaspb.VulnerabilityOccurrence{
								Severity:     grafeaspb.Severity_HIGH,
								FixAvailable: true,
							},
						},
					},
				},
			},
			fmt.Errorf("VulnerabilityCheck: The following vulnerable images were found:\n" +
				"    bar@sha256:111 [1 fixable severe vulnerabilities, 1 total]"),
		},
		{
			"Multiple vulnerabilities with no fix",
			int(grafeaspb.Severity_MEDIUM),
			map[reg.PromotionEdge]interface{}{
				edge1: nil,
			},
			map[image.Digest][]*grafeaspb.Occurrence{
				"sha256:000": {
					{
						Details: &grafeaspb.Occurrence_Vulnerability{
							Vulnerability: &grafeaspb.VulnerabilityOccurrence{
								Severity:     grafeaspb.Severity_HIGH,
								FixAvailable: false,
							},
						},
					},
					{
						Details: &grafeaspb.Occurrence_Vulnerability{
							Vulnerability: &grafeaspb.VulnerabilityOccurrence{
								Severity:     grafeaspb.Severity_CRITICAL,
								FixAvailable: false,
							},
						},
					},
				},
			},
			nil,
		},
	}

	for _, test := range tests {
		sc := reg.SyncContext{}
		check := reg.MKImageVulnCheck(
			&sc,
			test.edges,
			test.severityThreshold,
			mkVulnProducerFake(test.vulnerabilities),
		)
		got := check.Run()
		require.Equal(t, test.expected, got)
	}
}
