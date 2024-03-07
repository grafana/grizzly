package grizzly_test

import (
	"fmt"
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
				m := map[string]any(test.Resource)
				err := grizzly.ValidateEnvelope(m)
				if test.ExpectedError != "" {
					require.Error(t, err)
					require.Equal(t, "errors parsing resource: "+test.ExpectedError, err.Error())
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
				&grafana.Provider{},
			},
		)
		tests := []struct {
			Name              string
			InputFile         string
			ExpectedKind      string
			ExpectedError     string
			ExpectedResources int
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
				Name:              "json dashboards input, with envelope",
				InputFile:         "testdata/parsing/dashboards-with-envelope.json",
				ExpectedKind:      "Dashboard",
				ExpectedResources: 2,
			},
			{
				Name:         "yaml dashboard input, with envelope",
				InputFile:    "testdata/parsing/dashboard-with-envelope.yaml",
				ExpectedKind: "Dashboard",
			},
			{
				Name:         "yaml dashboard input, without envelope",
				InputFile:    "testdata/parsing/dashboard-without-envelope.yaml",
				ExpectedKind: "Dashboard",
			},
			{
				Name:              "yaml dashboards input, with envelope",
				InputFile:         "testdata/parsing/dashboards-with-envelope.yaml",
				ExpectedKind:      "Dashboard",
				ExpectedResources: 2,
			},
			{
				Name:         "jsonnet dashboard input, with envelope",
				InputFile:    "testdata/parsing/dashboard-with-envelope.jsonnet",
				ExpectedKind: "Dashboard",
			},
			{
				Name:         "jsonnet dashboard input, without envelope",
				InputFile:    "testdata/parsing/dashboard-without-envelope.jsonnet",
				ExpectedKind: "Dashboard",
			},
			{
				Name:              "jsonnet dashboards input, with envelope",
				InputFile:         "testdata/parsing/dashboards-with-envelope.jsonnet",
				ExpectedKind:      "Dashboard",
				ExpectedResources: 2,
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
				ExpectedError: "parse error in 'testdata/parsing/datasource-without-envelope.json': found invalid object (at .): errors parsing resource: kind missing, metadata missing, spec missing\n\naccess: proxy\nisDefault: true\njsonData:\n    httpMethod: GET\ntype: prometheus\nurl: http://localhost/prometheus/\n",
			},
		}

		parser := grizzly.DefaultParser(registry, nil, nil)
		parseOpts := grizzly.ParserOptions{
			DefaultResourceKind: "",
			DefaultFolderUID:    "General",
		}

		for _, test := range tests {
			t.Run(test.Name, func(t *testing.T) {
				resources, err := parser.Parse(test.InputFile, parseOpts)
				if test.ExpectedError != "" {
					require.Error(t, err)
					require.Equal(t, test.ExpectedError, err.Error())
					return
				}
				require.NoError(t, err)
				if test.ExpectedResources == 0 { // i.e. the default, which actually means 1
					require.Len(t, resources, 1, "Expected one resource from parsing")
				} else {
					require.Equal(t, test.ExpectedResources, len(resources), fmt.Sprintf("Expected %d resources from parsing", test.ExpectedResources))
				}
			})
		}
	})
}
