package yamlprocessor

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/drone/envsubst/v2"
	"github.com/drone/envsubst/v2/parse"
)

type Processor interface {
	GetVariables([]byte) ([]string, error)
	GetVariableMap([]byte) (map[string]*string, error)
	Process([]byte, func(string) (string, error)) ([]byte, error)
}

type SimpleProcessor struct {
}

var _ Processor = &SimpleProcessor{}

func NewSimpleProcessor() *SimpleProcessor {
	return &SimpleProcessor{}
}

func (tp *SimpleProcessor) GetVariables(rawArtifact []byte) ([]string, error) {
	variables, err := tp.GetVariableMap(rawArtifact)
	if err != nil {
		return nil, err
	}
	varNames := make([]string, 0, len(variables))
	for k := range variables {
		varNames = append(varNames, k)
	}
	sort.Strings(varNames)
	return varNames, nil
}

func (tp *SimpleProcessor) GetVariableMap(rawArtifact []byte) (map[string]*string, error) {
	strArtifact := convertLegacyVars(string(rawArtifact))
	variables, err := inspectVariables(strArtifact)
	if err != nil {
		return nil, err
	}
	varMap := make(map[string]*string, len(variables))
	for k, v := range variables {
		if v == "" {
			varMap[k] = nil
		} else {
			v := v
			varMap[k] = &v
		}
	}
	return varMap, nil
}

func (tp *SimpleProcessor) Process(rawArtifact []byte, variablesClient func(string) (string, error)) ([]byte, error) {
	tmp := convertLegacyVars(string(rawArtifact))
	variables, err := inspectVariables(tmp)
	if err != nil {
		return rawArtifact, err
	}
	var missingVariables []string
	for name, defaultVale := range variables {
		_, err := variablesClient(name)
		if err != nil && defaultVale == "" {
			missingVariables = append(missingVariables, name)
			continue
		}
	}
	if len(missingVariables) > 0 {
		return rawArtifact, &errMissingariables{missingVariables}
	}

	tmp, err = envsubst.Eval(tmp, func(s string) string {
		v, _ := variablesClient(s)
		return v
	})
	if err != nil {
		return rawArtifact, err
	}
	return []byte(tmp), err
}

type errMissingariables struct {
	Missing []string
}

func (e *errMissingariables) Error() string {
	sort.Strings(e.Missing)
	return fmt.Sprintf(
		"value for variables [%s] is not set. Please set the value using os environment variables or the clusterctl config file",
		strings.Join(e.Missing, ", "),
	)
}

var legacyVariableRegEx = regexp.MustCompile(`(\${(\s+([A-Za-z0-9_$]+)\s+)})|(\${(\s+([A-Za-z0-9_$]+))})|(\${(([A-Za-z0-9_$]+)\s+)})`)
var whitespaceRegEx = regexp.MustCompile(`\s`)

func convertLegacyVars(data string) string {
	return legacyVariableRegEx.ReplaceAllStringFunc(data, func(s string) string {
		return whitespaceRegEx.ReplaceAllString(s, "")
	})
}

func inspectVariables(data string) (map[string]string, error) {
	variables := make(map[string]string)
	t, err := parse.Parse(data)
	if err != nil {
		return nil, err
	}
	traverse(t.Root, variables)
	return variables, nil
}

func traverse(root parse.Node, variables map[string]string) {
	switch v := root.(type) {
	case *parse.ListNode:
		for _, ln := range v.Nodes {
			traverse(ln, variables)
		}
	case *parse.FuncNode:
		if _, ok := variables[v.Param]; !ok {
			b := strings.Builder{}
			for _, a := range v.Args {
				switch w := a.(type) {
				case *parse.FuncNode:
					b.WriteString(fmt.Sprintf("${%s}", w.Param))
				case *parse.TextNode:
					b.WriteString(w.Value)
				}
			}
			variables[v.Param] = b.String()
		}
	}
}
