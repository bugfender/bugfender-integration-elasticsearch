package dummy

import (
	"context"
	"log"

	"github.com/bugfender/bugfender-integration-elasticsearch/pkg/integration"
)

// ConsoleDestination prints logs to console
type ConsoleDestination struct {
}

var _ integration.LogWriter = ConsoleDestination{}

func NewConsoleDestination() ConsoleDestination {
	return ConsoleDestination{}
}
func (d ConsoleDestination) WriteLogs(_ context.Context, logs []integration.Log) error {
	for _, l := range logs {
		log.Println(l)
	}
	return nil
}
