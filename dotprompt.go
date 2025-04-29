package dotprompt

import (
	"bytes"
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

var (
	validDataTypes      = []string{"string", "number", "bool", "datetime", "object"}
	invalidCharsRegex   = regexp.MustCompile(`([^A-Za-z0-9 \-\r\n]*)`)
	multipleSpacesRegex = regexp.MustCompile(`[\s\r\n]+`)
)

// PromptError represents an error related to prompt processing.
type PromptError struct {
	Message string
}

// Error returns the error message associated with the PromptError.
func (e PromptError) Error() string {
	return e.Message
}

// OutputFormat represents the format of output, such as text or JSON.
type OutputFormat int

// UnmarshalYAML unmarshals a YAML node into an OutputFormat value, supporting "text" and "json".
// Returns an error if format is invalid.
func (of *OutputFormat) UnmarshalYAML(value *yaml.Node) error {
	switch strings.ToLower(value.Value) {
	case "text":
		*of = Text
	case "json":
		*of = Json
	default:
		return fmt.Errorf("invalid output format: %s", value.Value)
	}
	return nil
}

// MarshalYAML marshals the OutputFormat into a YAML-compatible representation.
// Returns a string representation of the format ("text" or "json") or an error if the format is invalid.
func (of OutputFormat) MarshalYAML() (interface{}, error) {
	switch of {
	case Text:
		return "text", nil
	case Json:
		return "json", nil
	default:
		return nil, fmt.Errorf("invalid output format: %v", of)
	}
}

const (

	// Text represents the plain text output format.
	Text OutputFormat = iota

	// Json represents the JSON output format.
	Json
)

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

// PromptFile represents the structure of a file containing a prompt configuration and multiple associated prompts.
type PromptFile struct {
	Name     string              `yaml:"name,omitempty"`
	Model    string              `yaml:"model,omitempty"`
	Config   PromptConfig        `yaml:"config"`
	Prompts  Prompts             `yaml:"prompts"`
	FewShots []FewShotPromptPair `yaml:"fewShots,omitempty"`
}

// PromptConfig represents the configuration options for a prompt, including temperature, max tokens, output
// format, and input schema.
type PromptConfig struct {
	Temperature  *float32      `yaml:"temperature,omitempty"`
	MaxTokens    *int          `yaml:"maxTokens,omitempty"`
	OutputFormat OutputFormat  `yaml:"outputFormat"`
	Input        InputSchema   `yaml:"input"`
	Output       *OutputSchema `yaml:"output,omitempty"`
}

// InputSchema represents the schema for input parameters and their default values.
type InputSchema struct {
	Parameters map[string]string      `yaml:"parameters"`
	Default    map[string]interface{} `yaml:"default,omitempty"`
}

type OutputSchema struct {
	Format OutputFormat `yaml:"format"`
}

// Prompts represents a set of system and user prompts.
type Prompts struct {
	System string `yaml:"system,omitempty"`
	User   string `yaml:"user"`
}

// FewShotPromptPair represents a pair of user prompt and the corresponding response.
type FewShotPromptPair struct {
	User     string `yaml:"user"`
	Response string `yaml:"response"`
}

// NewPromptFileFromFile reads a file from the specified path, processes its content, and returns a PromptFile
// structure or an error.
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

// NewPromptFile creates a new PromptFile from the provided name and prompt data.
// It validates the input, configures the prompt file, and returns an error if any issues are encountered.
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

	// Ensure that the output configuration is set, if not, then set it to the config format to ensure backward
	// compatibility.
	if promptFile.Config.Output == nil {
		promptFile.Config.Output = &OutputSchema{
			Format: promptFile.Config.OutputFormat,
		}
	}

	// Check that the output format is the same between the two locations it can be defined
	promptFile.Config.OutputFormat = promptFile.Config.Output.Format

	return promptFile, nil
}

// GetSystemPrompt generates the system prompt string using the provided template values, appending
// JSON format instructions if required.
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

// GetUserPrompt generates a user prompt string based on a provided template and a set of values.
// It utilizes the 'Prompts.User' template within the PromptFile and replaces template placeholders with
// corresponding values from the input map.
func (pf *PromptFile) GetUserPrompt(values map[string]interface{}) (string, error) {
	return pf.generatePrompt(pf.Prompts.User, values)
}

// generatePrompt generates a prompt by rendering a given template with provided values, utilizing the liquid
// templating engine. Returns the rendered prompt string or an error in case of failure.
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

// parseAndValidateParameters parses input parameters against the configuration, providing default values and validation.
func (pf *PromptFile) parseAndValidateParameters(values map[string]interface{}) (map[string]interface{}, error) {
	bindings := make(map[string]interface{})

	if values == nil {
		values = make(map[string]interface{})
	}

	// Iterate over the prompt file parameters and extract the values from the user provided collection
	for key := range pf.Config.Input.Parameters {
		// Get the current key without the `optional` suffix, then get the parameters type
		keyWithoutOptionalSuffix := strings.TrimSuffix(key, "?")
		parameterType := strings.ToLower(pf.Config.Input.Parameters[key])
		if value, ok := values[keyWithoutOptionalSuffix]; ok {
			// If the parameter value is an object which implements the fmt.Stringer interface then use this
			// to convert the object to its string representation. Otherwise, generate a string version of the
			// object with keys.
			//
			// If the value is not an object, then set the binding as the value directly
			if stringerValue, ok := value.(fmt.Stringer); ok && parameterType == "object" {
				bindings[keyWithoutOptionalSuffix] = fmt.Sprintf("%s", stringerValue.String())
			} else if parameterType == "object" {
				bindings[keyWithoutOptionalSuffix] = fmt.Sprintf("%+v", value)
			} else {
				bindings[keyWithoutOptionalSuffix] = value
			}
		} else if defaultValue, ok := pf.Config.Input.Default[keyWithoutOptionalSuffix]; ok {
			// If no value was provided by the user, but a default exists, then use the default
			bindings[keyWithoutOptionalSuffix] = defaultValue
		} else if !strings.HasSuffix(key, "?") {
			// User has not provided a value for a required parameter
			return nil, &PromptError{
				Message: fmt.Sprintf("no value provided for parameter %s", key),
			}
		}
	}

	// Iterate over all the set values and make sure that their types conform to the prompt file defined
	// types
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
			if !isNumeric(value) {
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
		}
	}

	return bindings, nil
}

// ToFile serializes the PromptFile and writes it to a specified file.
// Returns an error if the serialization or file write operation fails.
func (pf *PromptFile) ToFile(name string) error {
	content, err := pf.Serialize()
	if err != nil {
		return err
	}

	err = os.WriteFile(name, content, 0600)
	if err != nil {
		return &PromptError{
			Message: fmt.Sprintf("failed to write prompt file: %v", err),
		}
	}

	return nil
}

// Serialize serializes the PromptFile into a byte slice in YAML format and returns it, or an error if serialization
// fails.
func (pf *PromptFile) Serialize() ([]byte, error) {
	var b bytes.Buffer

	encoder := yaml.NewEncoder(&b)
	encoder.SetIndent(2)

	err := encoder.Encode(&pf)
	if err != nil {
		return nil, &PromptError{
			Message: fmt.Sprintf("failed to marshal prompt file: %v", err),
		}
	}

	return b.Bytes(), nil
}

// cleanName sanitizes the provided name string by removing invalid characters, replacing multiple spaces with a hyphen,
// trimming leading and trailing hyphens, and converting the result to lowercase.
func cleanName(name string) string {
	strippedName := invalidCharsRegex.ReplaceAllString(name, "")
	trimmedName := strings.Trim(multipleSpacesRegex.ReplaceAllString(strippedName, "-"), "-")

	return strings.ToLower(trimmedName)
}

// isNumeric checks if a value is of numeric type and returns true if it is.
func isNumeric(value interface{}) bool {
	switch value.(type) {
	case int, int8, int16, int32, int64:
		return true
	//case uint, uint8, uint16, uint32, uint64:
	//	return true
	case float32, float64:
		return true
	}
	return false
}
