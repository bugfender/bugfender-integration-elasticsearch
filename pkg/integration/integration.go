package integration

import (
	"context"
	"io/ioutil"
	"log"
	"time"

	"bugfender-integration-elasticsearch/pkg/backoff"
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

// Sync loops synchronizing forever, until cancelled
// Exits after retrying retries times with errors
func (i *Integration) Sync(ctx context.Context, retries uint) error {
	if i.verbose {
		log.Println("Sync started, press Ctrl-C to stop")
	}
	defer func() {
		i.saveState()
	}()
	nextStateSave := time.Now()
	for ctx.Err() == nil {
		boff := backoff.NewExponential(5*time.Second, 300*time.Second)
		var nErrors uint = 0
		for ctx.Err() == nil { // retry on error
			// save the state every 5 minutes
			if time.Now().After(nextStateSave) {
				i.saveState()
				nextStateSave = time.Now().Add(5 * time.Second)
			}
			err := i.syncOnePage(ctx)
			if err == nil {
				break
			}
			// wait and retry
			nErrors++
			log.Println("Trial", nErrors, "error:", err)
			if nErrors == retries {
				return err
			}
			boff.Wait(ctx)
		}
	}
	return ctx.Err()
}

// syncOnePage synchronizes one page of logs, returns error if something failed
func (i *Integration) syncOnePage(ctx context.Context) error {
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
