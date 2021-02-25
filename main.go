package main

import (
	"context"
	"crypto/tls"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/namsral/flag"

	"bugfender-integration-elasticsearch/pkg/dummy"
	"bugfender-integration-elasticsearch/pkg/elasticsearch"
	"bugfender-integration-elasticsearch/pkg/integration"
)

func main() {
	var (
		clientID               string
		clientSecret           string
		apiURL                 string
		appID                  int64
		esIndex                string
		esNodes                string
		esUsername, esPassword string
		consoleOutput          bool
		stateFile              string
		insecureSkipTLSVerify  bool
		verbose                bool
	)
	flag.String(flag.DefaultConfigFlagname, "", "path to config file")
	// Bugfender parameters
	flag.StringVar(&clientID, "client-id", "", "OAuth client ID to connect to Bugfender (mandatory)")
	flag.StringVar(&clientSecret, "client-secret", "", "OAuth client secret to connect to Bugfender (mandatory)")
	flag.Int64Var(&appID, "app-id", 0, "Bugfender app ID (mandatory)")
	flag.StringVar(&apiURL, "api-url", "https://dashboard.bugfender.com", "Bugfender API URL (only necessary for on-premises)")
	// Elasticsearch parameters
	flag.StringVar(&esIndex, "es-index", "", "Elasticsearch index to write to (default: logs)")
	flag.StringVar(&esNodes, "es-nodes", "", "List of Elasticsearch nodes (multiple nodes can be specified, separated by spaces)")
	flag.StringVar(&esUsername, "es-username", "", "Username to connect to Elasticsearch")
	flag.StringVar(&esPassword, "es-password", "", "Password to connect to Elasticsearch")
	// Console output
	flag.BoolVar(&consoleOutput, "console-output", false, "Print logs to console instead of Elasticsearch (for debugging)")
	// other
	flag.StringVar(&stateFile, "state-file", "", "File to restore and save state, to resume sync (recommended)")
	flag.BoolVar(&insecureSkipTLSVerify, "insecure-skip-tls-verify", false, "Skip TLS certificate verification (insecure)")
	flag.BoolVar(&verbose, "verbose", false, "Verbose messages")
	flag.Parse()

	// parameter validation
	if clientID == "" || clientSecret == "" || appID == 0 {
		flag.Usage()
		os.Exit(1)
	}
	parsedURL, err := url.Parse(apiURL)
	if err != nil {
		log.Fatal("invalid apiurl:", err)
	}

	if insecureSkipTLSVerify {
		// #nosec G402 this is intended, user specified -insecure-skip-tls-verify flag
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	// connect to Bugfender
	state, err := ioutil.ReadFile(stateFile) // #nosec G304 user intends to load this file
	if perr, ok := err.(*os.PathError); ok && perr.Err.(syscall.Errno) == syscall.ENOENT {
		// missing file, ignore
	} else if err != nil {
		log.Fatal("can not open state file:", err)
	}
	bf, err := integration.NewBugfenderClient(&integration.Config{
		OAuthClientID:     clientID,
		OAuthClientSecret: clientSecret,
		ApiUrl:            parsedURL,
	}, appID, state)
	if err != nil {
		log.Fatal("error initializing Bugfender client", err)
	}
	var destination integration.LogWriter
	if consoleOutput {
		destination = dummy.NewConsoleDestination()
	}
	// connect to Elasticsearch
	if esIndex != "" && esNodes != "" {
		destination, err = elasticsearch.NewClient(esIndex, strings.Split(esNodes, " "), esUsername, esPassword)
		if err != nil {
			log.Fatal("error initializing Elasticsearch client:", err)
		}
	}
	if destination == nil {
		log.Fatal("No destination specified")
	}
	// run integration
	i, err := integration.New(bf, destination, verbose, stateFile)
	if err != nil {
		log.Fatal("error initializing integration:", err)
	}

	// trap SIGINT to trigger a shutdown.
	ctx, cancelFunc := context.WithCancel(context.Background())
	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-exitSignal
		log.Println("")
		if verbose {
			log.Println("- Ctrl+C pressed in Terminal, closing")
		}
		cancelFunc()
	}()

	err = i.Sync(ctx)
	if ctx.Err() != context.Canceled && err != nil {
		log.Fatal(err)
	}
}
