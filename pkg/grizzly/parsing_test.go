package grizzly_test

import (
	"testing"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/stretchr/testify/require"
)

func TestValidateEnvelope(t *testing.T) {
	t.Run("Validate/Prepare", func(t *testing.T) {
		kind := "some-kind"
		metadata := map[string]any{
			"name": "a-name",
		}
		spec := map[string]any{
			"uid": "a-uid",
		}

		tests := []struct {
			Name          string
			Resource      grizzly.Resource
			ExpectedError string
		}{
			{
				Name: "missing kind",
				Resource: map[string]any{
					"metadata": metadata,
					"spec":     spec,
				},
				ExpectedError: "kind missing",
			},
			{
				Name: "missing metadata",
				Resource: map[string]any{
					"kind": kind,
					"spec": spec,
				},
				ExpectedError: "metadata missing",
			},
			{
				Name: "missing name",
				Resource: map[string]any{
					"kind":     kind,
					"metadata": map[string]any{},
					"spec":     spec,
				},
				ExpectedError: "metadata/name missing",
			},
			{
				Name: "missing spec",
				Resource: map[string]any{
					"kind":     kind,
					"metadata": metadata,
				},
				ExpectedError: "spec missing",
			},
			{
				Name: "empty spec",
				Resource: map[string]any{
					"kind":     kind,
					"metadata": metadata,
					"spec":     map[string]any{},
				},
				ExpectedError: "spec should not be empty",
			},
			{
				Name: "invalid spec",
				Resource: map[string]any{
					"kind":     kind,
					"metadata": metadata,
					"spec":     "a string spec",
				},
				ExpectedError: "spec is not a map",
			},
			{
				Name:          "empty resource",
				Resource:      map[string]any{},
				ExpectedError: "kind missing, metadata missing, spec missing",
			},
			{
				Name: "everything correct",
				Resource: map[string]any{
					"kind":     kind,
					"metadata": metadata,
					"spec":     spec,
				},
			},
		}
		for _, test := range tests {
			t.Run(test.Name, func(t *testing.T) {
				err := grizzly.ValidateEnvelope(test.Resource)
				if test.ExpectedError != "" {
					require.Error(t, err)
					require.Equal(t, err.Error(), "errors parsing resource: "+test.ExpectedError)
					return
				}
				require.NoError(t, err)
			})
		}
	})
}
