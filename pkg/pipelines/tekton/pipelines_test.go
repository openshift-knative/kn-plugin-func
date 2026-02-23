package tekton

import (
	"context"
	"testing"

	"github.com/tektoncd/pipeline/pkg/apis/config"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/utils/set"
)

func TestNodeS2IPipeline(t *testing.T) {
	myScheme := runtime.NewScheme()
	if err := tektonv1.AddToScheme(myScheme); err != nil {
		t.Fatal(err)
	}
	codecs := serializer.NewCodecFactory(myScheme)
	decode := codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(GetNodeJSPipeline()), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	pipeline, ok := obj.(*tektonv1.Pipeline)
	if !ok {
		t.Fatalf("unexpected type: %T", obj)
	}
	t.Logf("successfully decoded pipeline: %s\n", pipeline.Name)

	// Run deeper validations on the pipeline
	flags, err := config.NewFeatureFlagsFromMap(map[string]string{
		"enable-api-fields": "alpha",
	})
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		FeatureFlags: flags,
	}
	ctx := config.ToContext(context.Background(), cfg)
	pipeline.SetDefaults(ctx)
	apiErr := pipeline.Validate(ctx)
	if apiErr != nil {
		t.Fatalf("%+v\n", apiErr)
	}

	var passedParams = set.New[string]()
	for _, param := range pipeline.PipelineSpec().Tasks[0].Params {
		passedParams.Insert(param.Name)
	}

	var expectedParams = set.New[string]()
	for _, param := range pipeline.PipelineSpec().Tasks[0].TaskSpec.Params {
		expectedParams.Insert(param.Name)
	}

	missingParams := expectedParams.Difference(passedParams)
	if missingParams.Len() > 0 {
		t.Error("missing params:", missingParams)
	}
	superfluousParams := passedParams.Difference(expectedParams)
	if superfluousParams.Len() > 0 {
		t.Error("superfluous params:", superfluousParams)
	}
}
