package builddefinitions

import (
	"fmt"

	"github.com/tektoncd/chains/pkg/chains/objects"
)

type TektonPipelineBuildType struct {
	BuildType string
	Pro       *objects.PipelineRunObject
}

type SLSAPipelineBuildType struct {
	BuildType string
	Pro       *objects.PipelineRunObject
}

// internalParameters adds the tekton feature flags that were enabled
// for the pipelinerun.
func (s SLSAPipelineBuildType) InternalParameters() map[string]any {
	internalParams := make(map[string]any)
	if s.Pro.Status.Provenance != nil && s.Pro.Status.Provenance.FeatureFlags != nil {
		internalParams["tekton-pipelines-feature-flags"] = *s.Pro.Status.Provenance.FeatureFlags
	}
	return internalParams
}

// externalParameters adds the pipeline run spec
func (s SLSAPipelineBuildType) ExternalParameters() map[string]any {
	return defaultExternalParameters(s.Pro)
}

func defaultExternalParameters(pro *objects.PipelineRunObject) map[string]any {
	externalParams := make(map[string]any)

	// add the origin of top level pipeline config
	// isRemotePipeline checks if the pipeline was fetched using a remote resolver
	isRemotePipeline := false
	if pro.Spec.PipelineRef != nil {
		if pro.Spec.PipelineRef.Resolver != "" && pro.Spec.PipelineRef.Resolver != "Cluster" {
			isRemotePipeline = true
		}
	}

	if p := pro.Status.Provenance; p != nil && p.RefSource != nil && isRemotePipeline {
		ref := ""
		for alg, hex := range p.RefSource.Digest {
			ref = fmt.Sprintf("%s:%s", alg, hex)
			break
		}
		buildConfigSource := map[string]string{
			"ref":        ref,
			"repository": p.RefSource.URI,
			"path":       p.RefSource.EntryPoint,
		}
		externalParams["buildConfigSource"] = buildConfigSource
	}
	externalParams["runSpec"] = pro.Spec
	return externalParams
}

func (t TektonPipelineBuildType) InternalParameters() map[string]any {
	internalParams := make(map[string]any)
	if t.Pro.Status.Provenance != nil && t.Pro.Status.Provenance.FeatureFlags != nil {
		internalParams["tekton-pipelines-feature-flags"] = *t.Pro.Status.Provenance.FeatureFlags
	}
	internalParams["labels"] = t.Pro.ObjectMeta.Labels
	internalParams["annotations"] = t.Pro.ObjectMeta.Annotations
	return internalParams
}

func (t TektonPipelineBuildType) ExternalParameters() map[string]any {
	return defaultExternalParameters(t.Pro)
}
