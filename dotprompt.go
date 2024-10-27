package dotprompt

import (
	"fmt"
	"gopkg.in/osteele/liquid.v1"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
)

var validDataTypes = []string{"string", "number", "bool", "datetime", "object"}

type PromptError struct {
	Message string
}

func (e *PromptError) Error() string {
	return e.Message
}

type OutputFormat int

func (of *OutputFormat) UnmarshalYAML(value *yaml.Node) error {
	switch strings.ToLower(value.Value) {
	case "text":
		*of = Text
	case "json":
		*of = Json
	default:
		return value.Decode(&of)
	}
	return nil
}

func (of *OutputFormat) String() string {
	switch *of {
	case Text:
		return "text"
	case Json:
		return "json"
	default:
		return "unknown"
	}
}

const (
	Text OutputFormat = iota
	Json
)

type PromptFile struct {
	Name     string              `yaml:"name"`
	Config   PromptConfig        `yaml:"config"`
	Prompts  Prompts             `yaml:"prompts"`
	FewShots []FewShotPromptPair `yaml:"fewShots"`
}

type PromptConfig struct {
	Temperature  *float32     `yaml:"temperature"`
	MaxTokens    *int         `yaml:"maxTokens"`
	OutputFormat OutputFormat `yaml:"outputFormat"`
	Input        InputSchema  `yaml:"input"`
}

type InputSchema struct {
	Parameters map[string]string      `yaml:"parameters"`
	Default    map[string]interface{} `yaml:"default"`
}

type Prompts struct {
	System string `yaml:"system"`
	User   string `yaml:"user"`
}

type FewShotPromptPair struct {
	User     string `yaml:"user"`
	Response string `yaml:"response"`
}

func NewPromptFileFromFile(path string) (*PromptFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	fileName := strings.ToLower(filepath.Base(path))
	extension := filepath.Ext(fileName)
	promptFileName := strings.TrimSuffix(fileName, extension)

	return NewPromptFile(promptFileName, data)
}

func NewPromptFile(name string, data []byte) (*PromptFile, error) {
	promptFile := &PromptFile{}
	err := yaml.Unmarshal(data, promptFile)
	if err != nil {
		return nil, &PromptError{
			Message: fmt.Sprintf("failed to parse prompt file: %v", err),
		}
	}

	if len(promptFile.Prompts.User) == 0 {
		return nil, &PromptError{
			Message: "no user prompt template was provided in the prompt file",
		}
	}

	if len(promptFile.Name) == 0 {
		promptFile.Name = name
	}

	promptFile.Name = cleanName(promptFile.Name)

	if len(promptFile.Name) == 0 {
		return nil, &PromptError{
			Message: "the prompt file name, once cleaned, is empty",
		}
	}

	for key, paramType := range promptFile.Config.Input.Parameters {
		if !slices.Contains(validDataTypes, paramType) {
			return nil, &PromptError{
				Message: fmt.Sprintf("invalid data type for parameter %s: %s", key, paramType),
			}
		}
	}

	return promptFile, nil
}

func (pf *PromptFile) GetSystemPrompt(values map[string]interface{}) (string, error) {
	systemPrompt := pf.Prompts.System
	if pf.Config.OutputFormat == Json &&
		!strings.Contains(strings.ToLower(systemPrompt), "json") &&
		!strings.Contains(strings.ToLower(pf.Prompts.User), "json") {

		promptSuffix := "Please provide the response in JSON"

		if len(systemPrompt) == 0 {
			systemPrompt = promptSuffix
		} else {
			systemPrompt += " " + promptSuffix
		}
	}

	return pf.generatePrompt(systemPrompt, values)
}

func (pf *PromptFile) GetUserPrompt(values map[string]interface{}) (string, error) {
	return pf.generatePrompt(pf.Prompts.User, values)
}

func (pf *PromptFile) generatePrompt(template string, values map[string]interface{}) (string, error) {
	engine := liquid.NewEngine()
	bindings, err := pf.parseAndValidateParameters(values)
	if err != nil {
		return "", err
	}

	prompt, err := engine.ParseAndRenderString(template, bindings)
	if err != nil {
		return "", &PromptError{
			Message: fmt.Sprintf("failed to render prompt: %v", err),
		}
	}

	return prompt, nil
}

func (pf *PromptFile) parseAndValidateParameters(values map[string]interface{}) (map[string]interface{}, error) {
	bindings := make(map[string]interface{})

	for key := range pf.Config.Input.Parameters {
		if value, ok := values[key]; ok {
			bindings[key] = value
		} else if defaultValue, ok := pf.Config.Input.Default[key]; ok {
			bindings[key] = defaultValue
		} else if !strings.HasSuffix(key, "?") {
			return nil, &PromptError{
				Message: fmt.Sprintf("no value provided for parameter %s", key),
			}
		}
	}

	for key, value := range bindings {
		expectedType := pf.Config.Input.Parameters[key]
		switch expectedType {
		case "string":
			if _, ok := value.(string); !ok {
				return nil, &PromptError{
					Message: fmt.Sprintf("parameter %s is not a string", key),
				}
			}
			break
		case "number":
			if _, ok := value.(float64); !ok {
				return nil, &PromptError{
					Message: fmt.Sprintf("parameter %s is not a number", key),
				}
			}
			break
		case "bool":
			if _, ok := value.(bool); !ok {
				return nil, &PromptError{
					Message: fmt.Sprintf("parameter %s is not a bool", key),
				}
			}
			break
		case "datetime":
			if _, ok := value.(time.Time); !ok {
				return nil, &PromptError{
					Message: fmt.Sprintf("parameter %s is not a datetime", key),
				}
			}
			break
		case "object":
			if _, ok := value.(fmt.Stringer); !ok {
				return nil, &PromptError{
					Message: fmt.Sprintf("parameter %s does not implement the fmt.Stringer interface", key),
				}
			}
		}
	}

	return bindings, nil
}

func cleanName(name string) string {
	invalidCharsRegex := regexp.MustCompile(`([^A-Za-z0-9 \-\r\n]*)`)
	multipleSpacesRegex := regexp.MustCompile(`[\s\r\n]+`)

	strippedName := invalidCharsRegex.ReplaceAllString(name, "")
	trimmedName := strings.Trim(multipleSpacesRegex.ReplaceAllString(strippedName, "-"), "-")

	return strings.ToLower(trimmedName)
}
