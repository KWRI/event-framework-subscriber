package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	ps "cloud.google.com/go/pubsub"
)

var (
	proj      string
	mu        sync.Mutex
	connected int
	port      string
)

type contextKey string

func getClient(ctx context.Context) *ps.Client {

	if proj == "" {
		panic("Can't find Google cloud project used for pubsub.")
	}

	c, err := ps.NewClient(ctx, proj)

	if err != nil {
		panic(err)
	}
	return c
}

func subscribe(ctx context.Context, subscription string) {

	client := getClient(ctx)
	sub := client.Subscription(subscription)

	err := sub.Receive(ctx, func(ctx context.Context, msg *ps.Message) {
		val := ctx.Value(contextKey("conn"))
		if val == nil {
			return
		}
		conn, ok := val.(net.Conn)

		if !ok {
			return
		}
		data := msg.Data
		ins := []byte(fmt.Sprintf(`, "published_at":"%s", "subscription_name":"%s"}`, msg.PublishTime.Format(time.RFC3339), subscription))
		closingBraceIdx := bytes.LastIndexByte(data, '}')
		data = append(data[:closingBraceIdx], ins...)
		conn.Write(data)
		msg.Ack()
	})

	if err != nil {
		panic(err)
	}

}

func initFlag() int {

	var (
		secret string
	)

	flags := flag.NewFlagSet("event-framework-subscriber", flag.ExitOnError)
	flags.Usage = usage
	flags.StringVar(&proj, "project", "", "Google Cloud Pub Sub Project Name")
	flags.StringVar(&secret, "credentials-file", "", "Google Secret json file location")
	flags.StringVar(&port, "port", "9001", "TCP Port number default 9001")
	if err := flags.Parse(os.Args[1:]); err != nil {
		flags.Usage()
		return 1
	}

	if proj == "" {
		flags.Usage()
		return 1
	}

	if secret != "" && os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", secret)
	}
	return -1
}

func execute() int {
	if retVal := initFlag(); retVal != -1 {
		return retVal
	}

	ctx, cancel := context.WithCancel(context.Background())

	fmt.Println("Launching server...")
	ln, err := net.Listen("tcp", ":9001")
	if err != nil {
		cancel()
	}

	for {
		select {
		case <-(ctx).Done():
			return 0
		default:
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println("Error accepting: ", err.Error())
			}
			go readConnectionMessage(conn)
		}
		fmt.Println("loop called")
	}
}

func readConnectionMessage(conn net.Conn) {
	mu.Lock()
	connected++
	mu.Unlock()
	fmt.Println("Listening message.", "Current connection number #", connected)
	var buff bytes.Buffer
	test := make([]byte, 1)
	delim := []byte(";")

	ctx := context.WithValue(context.Background(), contextKey("conn"), conn)
	ctx, cancel := context.WithCancel(ctx)
	for {
		conn.Read(test)
		buff.Write(test)
		if test[0] == delim[0] {
			message := buff.String()
			buff.Reset()
			if message == "quit!;" || message == ";" {
				cancel()
				conn.Close()
				mu.Lock()
				connected--
				fmt.Println("Connection closed number of current connection is #", connected)
				mu.Unlock()
				return
			}
			if message != "" {
				go processMessage(ctx, message)
			}
		}
	}
}

func processMessage(ctx context.Context, message string) {
	if strings.HasPrefix(message, "subscribe:") {
		sub := strings.TrimPrefix(message, "subscribe:")
		sub = strings.TrimSuffix(sub, ";")
		subscribe(ctx, sub)
	}
}

func main() {
	os.Exit(execute())
}

func usage() {
	fmt.Fprintf(os.Stderr, helpText)
}

const helpText = `
  KW Event Framework Subscriber, subscribe google pubsubs events
  Usage: event-framework-subscriber -project=value -credentials-file=value -port=value

    Options:

	-project=""            	Google Cloud Pub Sub Project Name
	-credentials-file="" 	Google Secret Credentials json file location
	-port="9001"            TCP Port number default 9001
`
