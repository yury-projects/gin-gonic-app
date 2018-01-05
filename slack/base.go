package slack

// CommandInterface - should be an "abstract" interface that is implemented by different commands
type CommandInterface interface {
	IsValidText(string) bool
	GetCommandResponse() interface{}
}
