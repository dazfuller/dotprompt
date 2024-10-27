package dotprompt

import (
	"strings"
	"testing"
)

func TestReadingValidPromptFile(t *testing.T) {
	promptFile, err := NewPromptFileFromFile("prompts/example.prompt")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if promptFile.Name != "example" {
		t.Errorf("Expected name to be 'example', got '%s'", promptFile.Name)
	}

	if promptFile.Config.Temperature == nil {
		t.Error("Expected temperature to be set")
	} else if *promptFile.Config.Temperature != 0.7 {
		t.Errorf("Expected temperature to be 0.7, got '%f'", *promptFile.Config.Temperature)
	}

	if promptFile.Config.OutputFormat != Text {
		t.Errorf("Expected output format to be text, got '%s'", promptFile.Config.OutputFormat.String())
	}

	if *promptFile.Config.MaxTokens != 500 {
		t.Errorf("Expected max tokens to be 500, got '%d'", *promptFile.Config.MaxTokens)
	}

	if len(promptFile.Config.Input.Parameters) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(promptFile.Config.Input.Parameters))
	}

	if topic, ok := promptFile.Config.Input.Parameters["topic"]; !ok {
		t.Error("Expected parameter 'topic' to be set")
	} else if topic != "string" {
		t.Errorf("Expected parameter 'topic' to be 'string', got '%s'", topic)
	}

	if style, ok := promptFile.Config.Input.Parameters["style?"]; !ok {
		t.Error("Expected parameter 'style?' to be set")
	} else if style != "string" {
		t.Errorf("Expected parameter 'style?' to be 'string', got '%s'", style)
	}

	if len(promptFile.Config.Input.Default) != 1 {
		t.Errorf("Expected 1 default, got %d", len(promptFile.Config.Input.Default))
	}

	if topicDefault, ok := promptFile.Config.Input.Default["topic"]; !ok {
		t.Error("Expected default 'topic' to be set")
	} else if topicDefault != "social media" {
		t.Errorf("Expected default 'topic' to be 'social media', got '%s'", topicDefault)
	}

	if promptFile.Prompts.System != "You are a helpful research assistant who will provide descriptive responses for a given topic and how it impacts society" {
		t.Errorf("Expected system prompt to be 'You are a helpful research assistant who will provide descriptive responses for a given topic and how it impacts society', got '%s'", promptFile.Prompts.System)
	}

	if !strings.HasPrefix(promptFile.Prompts.User, "Explain the impact of {{ topic }}") {
		t.Errorf("Expected user prompt to start with 'Explain the impact of {{ topic }}', got '%s'", promptFile.Prompts.User)
	}

	if len(promptFile.FewShots) != 3 {
		t.Errorf("Expected 3 few shots, got %d", len(promptFile.FewShots))
	}

	if promptFile.FewShots[0].User != "What is Bluetooth" {
		t.Errorf("Expected first few shot user message to be 'What is Bluetooth', got '%s'", promptFile.FewShots[0].User)
	}

	if promptFile.FewShots[2].Response != "AI is used in virtual assistants like Siri and Alexa, which understand and respond to voice commands." {
		t.Errorf("Expected third few shot response to be 'AI is used in virtual assistants like Siri and Alexa, which understand and respond to voice commands.', got '%s'", promptFile.FewShots[2].Response)
	}
}

func TestGenerateUserPrompt(t *testing.T) {
	promptFile, err := NewPromptFileFromFile("prompts/example.prompt")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	userPrompt, err := promptFile.GetUserPrompt(map[string]interface{}{
		"topic": "bluetooth",
	})
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if userPrompt != "Explain the impact of bluetooth on how we engage with technology as a society\n" {
		t.Errorf("Expected user prompt to be 'Explain the impact of bluetooth on how we engage with technology as a society', got '%s'", userPrompt)
	}
}

func TestGenerateSystemPrompt(t *testing.T) {
	promptFile, err := NewPromptFileFromFile("prompts/example.prompt")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	systemPrompt, err := promptFile.GetSystemPrompt(map[string]interface{}{})
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if systemPrompt != "You are a helpful research assistant who will provide descriptive responses for a given topic and how it impacts society" {
		t.Errorf("Expected system prompt to be 'You are a helpful research assistant who will provide descriptive responses for a given topic and how it impacts society', got '%s'", systemPrompt)
	}
}
