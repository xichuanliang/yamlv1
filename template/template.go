package template

import (
	"github.com/pkg/errors"
	"github.com/xichuan/yamlv1/yamlprocessor"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/cluster-api/util/yaml"
)

type Template interface {
	Variables() []string
	VariableMap() map[string]*string
	Objs() []unstructured.Unstructured
}

type template struct {
	variables       []string
	variableMap     map[string]*string
	targetNamespace string
	objs            []unstructured.Unstructured
}

var _ Template = &template{}

func (t *template) Variables() []string {
	return t.variables
}

func (t *template) VariableMap() map[string]*string {
	return t.variableMap
}

func (t *template) Objs() []unstructured.Unstructured {
	return t.objs
}

type TemplateInput struct {
	RawArtifact     []byte
	Env             map[string]string
	Processor       yamlprocessor.Processor
	TargetNamespace string
}

func NewTemplate(input TemplateInput) (Template, error) {
	variables, err := input.Processor.GetVariables(input.RawArtifact)
	if err != nil {
		return nil, err
	}
	variableMap, err := input.Processor.GetVariableMap(input.RawArtifact)
	if err != nil {
		return nil, err
	}
	for key, val := range input.Env {
		v := val
		variableMap[key] = &v
	}
	mapping := func(key string) (string, error) {
		if variableMap[key] == nil {
			return "", errors.Errorf("Failed to get value for variable %q", key)
		}
		return *variableMap[key], nil
	}

	processedYaml, err := input.Processor.Process(input.RawArtifact, mapping)
	if err != nil {
		return nil, err
	}
	objs, err := yaml.ToUnstructured(processedYaml)
	if err != nil {
		return nil, err
	}
	return &template{
		variables:       variables,
		variableMap:     variableMap,
		targetNamespace: input.TargetNamespace,
		objs:            objs,
	}, nil
}
