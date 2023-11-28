package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	template "github.com/xichuan/yamlv1/template"
	"github.com/xichuan/yamlv1/yamlprocessor"
)

func main() {
	ctx := context.Background()
	templateUrl := "./template/template.yaml"
	targetNamespace := "cluster-manager"
	envUrl := "./template/env.conf"
	template, err := GetFromURL(ctx, templateUrl, targetNamespace, envUrl)
	if err != nil {
		errors.Wrapf(err, "fail to genergate template")
	}
	fmt.Println(template)
}

func GetFromURL(ctx context.Context, templateUrl string, targetNamespace string, envUrl string) (template.Template, error) {
	if templateUrl == "" {
		return nil, errors.New("invalid GetFromURL operation: missing templateURL value")
	}
	templateContent, err := getURLContent(ctx, templateUrl)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid GetFromURL operation for templateUrl")
	}
	if envUrl == "" {
		return nil, errors.Errorf("invalid getEnv operation: missing envURL value")
	}
	envContent, err := readConfFile(envUrl)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read file %q", envUrl)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "invalid GetFromURL operation for envUrl")
	}

	return template.NewTemplate(
		template.TemplateInput{
			RawArtifact:     templateContent,
			Processor:       yamlprocessor.NewSimpleProcessor(),
			TargetNamespace: "cluster-manager",
			Env:             envContent,
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

// read env.conf from filePath
func readConfFile(filePath string) (map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	confMap := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			confMap[key] = value
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return confMap, nil
}
