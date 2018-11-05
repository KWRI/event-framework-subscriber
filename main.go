package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	ps "cloud.google.com/go/pubsub"
)

var (
	proj string
)

func GetClient(ctx context.Context) *ps.Client {

	if proj == "" {
		panic("Can't find Google cloud project used for pubsub.")
	}
	c, err := ps.NewClient(ctx, proj)

	if err != nil {
		panic(err)
	}
	return c
}

func subscribe(ctx context.Context, subscription string, ch chan *ps.Message) {

	client := GetClient(ctx)
	sub := client.Subscription(subscription)

	err := sub.Receive(ctx, func(ctx context.Context, msg *ps.Message) {
		ch <- msg
		msg.Ack()
	})

	if err != nil {
		panic(err)
	}

}

type subscriptions []string

func (s *subscriptions) String() string {
	return strings.Join(*s, " ")
}

func (s *subscriptions) Set(value string) error {
	for _, v := range strings.Split(value, " ") {
		if v != "" {
			s.appendIfMissing(strings.ToLower(v))
		}
	}
	return nil
}

func (s *subscriptions) appendIfMissing(value string) {
	for _, existing := range *s {
		if existing == value {
			return
		}
	}

	*s = append(*s, value)
}

func execute() int {
	var (
		subs   subscriptions
		secret string
	)

	flags := flag.NewFlagSet("event-framework-subscriber", flag.ExitOnError)
	flags.Usage = usage
	flags.StringVar(&proj, "project", "", "Google Cloud Pub Sub Project Name")
	flags.Var(&subs, "subscriptions", "Subscription Names separated by space")
	flags.StringVar(&secret, "credentials-file", "", "Google Secret json file location")

	if err := flags.Parse(os.Args[1:]); err != nil {
		flags.Usage()
		return 1
	}

	if proj == "" && len(subs) == 0 {
		flags.Usage()
		return 1
	}

	if secret != "" && os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", secret)
	}

	ch := make(chan *ps.Message)
	ctx := context.Background()
	for _, subscription := range subs {
		go subscribe(ctx, subscription, ch)
	}

	for {
		select {
		case a := <-ch:
			data := a.Data
			ins := []byte(fmt.Sprintf(`, "published_at":"%s"}`, a.PublishTime.Format(time.RFC3339)))
			closingBraceIdx := bytes.LastIndexByte(data, '}')
			data = append(data[:closingBraceIdx], ins...)
			fmt.Println(string(data))
		}
	}
	return 0
}

func main() {
	os.Exit(execute())
}

func usage() {
	fmt.Fprintf(os.Stderr, helpText)
}

const helpText = `
KW Event Framework Subscriber, subscribe google pubsubs events

Usage: event-framework-subscriber -project=value -subscriptions=values -credentials-file=value

Options:

	-project=""            	Google Cloud Pub Sub Project Name
	-subscriptions=""		Subscription Names separated by space
	-credentials-file="" 	Google Secret Credentials json file location
`
