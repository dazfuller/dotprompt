package dotprompt

type Loader interface {
	Load() ([]PromptFile, error)
}
