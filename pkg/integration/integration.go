package integration

import (
	"context"
	"io/ioutil"
	"log"
	"time"
)

type Integration struct {
	bugfenderClient *Client
	destination     LogWriter
	verbose         bool
	stateFile       string
}

type LogWriter interface {
	WriteLogs(context.Context, []Log) error
}

// New creates a new integration from the bugfenderClient to the destination
func New(bugfenderClient *Client, destination LogWriter, verbose bool, stateFile string) (*Integration, error) {
	return &Integration{
		bugfenderClient: bugfenderClient,
		destination:     destination,
		verbose:         verbose,
		stateFile:       stateFile,
	}, nil
}

func (i *Integration) Sync(ctx context.Context) error {
	if i.verbose {
		log.Println("Sync started, press Ctrl-C to stop")
	}
	defer func() {
		i.saveState()
	}()
	nextStateSave := time.Now()
	for ctx.Err() == nil {
		// save the state every 5 minutes
		if time.Now().After(nextStateSave) {
			i.saveState()
			nextStateSave = time.Now().Add(5 * time.Second)
		}
		// get a page from Bugfender
		logs, err := i.bugfenderClient.GetNextPage(ctx)
		if err != nil {
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		// put it in the destination
		err = i.destination.WriteLogs(ctx, logs)
		if err != nil {
			return err
		}
		if i.verbose {
			log.Printf("Wrote %d logs", len(logs))
		}
	}
	return ctx.Err()
}

func (i *Integration) saveState() {
	if i.verbose {
		log.Println("Saving state")
	}
	err := ioutil.WriteFile(i.stateFile, i.bugfenderClient.GetState(), 0600)
	if err != nil {
		log.Fatalln("error saving state file:", err)
	}
}
