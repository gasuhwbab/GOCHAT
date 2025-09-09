package proto

import (
	"fmt"
	"time"
)

func stamp() string {
	return time.Now().Format("15:04")
}

func FormatUserMessage(nick, text string) string {
	return fmt.Sprintf("[%s] %s: %s", stamp(), nick, text)
}

func FormatSystemMessage(text string) string {
	return fmt.Sprintf("[%s] %s", stamp(), text)
}

func FormatPrivateMessage(from, to, text string) string {
	return fmt.Sprintf("[%s] (private) %s -> %s: %s", stamp(), from, to, text)
}
