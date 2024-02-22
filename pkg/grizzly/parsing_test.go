package grizzly_test

import (
	"testing"

	"github.com/grafana/grizzly/pkg/grafana"
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

func TestParseKindDetection(t *testing.T) {
	t.Run("Parse kind detection", func(t *testing.T) {

		registry := grizzly.NewRegistry(
			[]grizzly.Provider{
				grafana.NewProvider(),
			},
		)
		opts := grizzly.Opts{
			FolderUID: "general",
		}

		tests := []struct {
			Name          string
			InputFile     string
			ExpectedKind  string
			ExpectedError string
		}{
			{
				Name:         "json dashboard input, with envelope",
				InputFile:    "testdata/parsing/dashboard-with-envelope.json",
				ExpectedKind: "Dashboard",
			},
			{
				Name:         "json dashboard input, without envelope",
				InputFile:    "testdata/parsing/dashboard-without-envelope.json",
				ExpectedKind: "Dashboard",
			},
			{
				Name:         "yaml dashboard input, with envelope",
				InputFile:    "testdata/parsing/dashboard-with-envelope.yaml",
				ExpectedKind: "Dashboard",
			},
			{
				Name:         "yamljsonnet dashboard, without envelope",
				InputFile:    "testdata/parsing/dashboard-without-envelope.yaml",
				ExpectedKind: "Dashboard",
			},
			{
				Name:         "jsonnet dashboard, with envelope",
				InputFile:    "testdata/parsing/dashboard-with-envelope.jsonnet",
				ExpectedKind: "Dashboard",
			},
			{
				Name:         "jsonnet dashboard, without envelope",
				InputFile:    "testdata/parsing/dashboard-without-envelope.jsonnet",
				ExpectedKind: "Dashboard",
			},
			{
				Name:         "json datasource input, with envelope",
				InputFile:    "testdata/parsing/datasource-with-envelope.json",
				ExpectedKind: "Datasource",
			},
			{
				// This test assumes that Grizzly is not configured to detect the kind
				// of a datasource, thus resulting in an error.
				Name:          "json datasource input, without envelope",
				InputFile:     "testdata/parsing/datasource-without-envelope.json",
				ExpectedError: "",
			},
		}
		for _, test := range tests {
			t.Run(test.Name, func(t *testing.T) {
				resources, err := grizzly.Parse(registry, test.InputFile, &opts)
				if test.ExpectedError != "" {
					require.Error(t, err)
					require.Equal(t, err.Error(), test.ExpectedError)
					return
				}
				require.NoError(t, err)
				require.Equal(t, 1, len(resources), "Expected one resource from parsing")
			})
		}
	})
}
