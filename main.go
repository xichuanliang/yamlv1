package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
	template "github.com/xichuan/yamlv1/template"
	"github.com/xichuan/yamlv1/yamlprocessor"
)

func main() {
	ctx := context.Background()
	templateUrl := "./template/template.yaml"
	targetNamespace := "cluster-manager"
	template, err := GetFromURL(ctx, templateUrl, targetNamespace)
	if err != nil {
		errors.Wrapf(err, "fail to genergate template")
	}
	fmt.Println(template)
}

func GetFromURL(ctx context.Context, templateUrl string, targetNamespace string) (template.Template, error) {
	if templateUrl == "" {
		return nil, errors.New("invalid GetFromURL operation: missing templateURL value")
	}
	content, err := getURLContent(ctx, templateUrl)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid GetFromURL operation")
	}
	return template.NewTemplate(
		template.TemplateInput{
			RawArtifact:     content,
			Processor:       yamlprocessor.NewSimpleProcessor(),
			TargetNamespace: "cluster-manager",
		},
	)
}

func getURLContent(ctx context.Context, templateURL string) ([]byte, error) {
	if templateURL == "" {
		return nil, errors.New("invalid GetFromURL operation: missing templateURL value")
	}
	content, err := os.ReadFile(templateURL)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read file %q", templateURL)
	}
	return content, nil
}
