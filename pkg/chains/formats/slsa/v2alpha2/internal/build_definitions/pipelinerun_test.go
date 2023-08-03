package builddefinitions

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tektoncd/chains/pkg/chains/objects"
	"github.com/tektoncd/pipeline/pkg/apis/config"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

const pipelineFile = "pipeline.yaml"

func TestInternalParameters(t *testing.T) {
	pr := &v1beta1.PipelineRun{
		Status: v1beta1.PipelineRunStatus{
			PipelineRunStatusFields: v1beta1.PipelineRunStatusFields{
				Provenance: &v1beta1.Provenance{
					FeatureFlags: &config.FeatureFlags{
						RunningInEnvWithInjectedSidecars: true,
						EnableAPIFields:                  "stable",
						AwaitSidecarReadiness:            true,
						VerificationNoMatchPolicy:        "skip",
						EnableProvenanceInStatus:         true,
						ResultExtractionMethod:           "termination-message",
						MaxResultSize:                    4096,
					},
				},
			},
		},
	}

	slsaBuild := SLSAPipelineBuildType{
		Pro: objects.NewPipelineRunObject(pr),
	}

	want := map[string]any{
		"tekton-pipelines-feature-flags": config.FeatureFlags{
			RunningInEnvWithInjectedSidecars: true,
			EnableAPIFields:                  "stable",
			AwaitSidecarReadiness:            true,
			VerificationNoMatchPolicy:        "skip",
			EnableProvenanceInStatus:         true,
			ResultExtractionMethod:           "termination-message",
			MaxResultSize:                    4096,
		},
	}
	got := slsaBuild.InternalParameters()
	if d := cmp.Diff(want, got); d != "" {
		t.Fatalf("internalParameters (-want, +got):\n%s", d)
	}
}

func TestExternalParameters(t *testing.T) {
	pr := &v1beta1.PipelineRun{
		Spec: v1beta1.PipelineRunSpec{
			Params: v1beta1.Params{
				{
					Name:  "my-param",
					Value: v1beta1.ResultValue{Type: "string", StringVal: "string-param"},
				},
				{
					Name:  "my-array-param",
					Value: v1beta1.ResultValue{Type: "array", ArrayVal: []string{"my", "array"}},
				},
				{
					Name:  "my-empty-string-param",
					Value: v1beta1.ResultValue{Type: "string"},
				},
				{
					Name:  "my-empty-array-param",
					Value: v1beta1.ResultValue{Type: "array", ArrayVal: []string{}},
				},
			},
			PipelineRef: &v1beta1.PipelineRef{
				ResolverRef: v1beta1.ResolverRef{
					Resolver: "git",
				},
			},
		},
		Status: v1beta1.PipelineRunStatus{
			PipelineRunStatusFields: v1beta1.PipelineRunStatusFields{
				Provenance: &v1beta1.Provenance{
					RefSource: &v1beta1.RefSource{
						URI: "hello",
						Digest: map[string]string{
							"sha1": "abc123",
						},
						EntryPoint: pipelineFile,
					},
				},
			},
		},
	}

	slsaBuild := SLSAPipelineBuildType{
		Pro: objects.NewPipelineRunObject(pr),
	}

	want := map[string]any{
		"buildConfigSource": map[string]string{
			"path":       pipelineFile,
			"ref":        "sha1:abc123",
			"repository": "hello",
		},
		"runSpec": pr.Spec,
	}
	got := slsaBuild.ExternalParameters()
	if d := cmp.Diff(want, got); d != "" {
		t.Fatalf("externalParameters (-want, +got):\n%s", d)
	}
}
