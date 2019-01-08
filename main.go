package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	ps "cloud.google.com/go/pubsub"
)

var (
	proj      string
	mu        sync.Mutex
	connected int
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
		conn := val.(net.Conn)
		data := msg.Data
		ins := []byte(fmt.Sprintf(`, "published_at":"%s", "subscription_name":"%s"}`, msg.PublishTime.Format(time.RFC3339), subscription))
		closingBraceIdx := bytes.LastIndexByte(data, '}')
		data = append(data[:closingBraceIdx], ins...)
		log.Println("Process message", string(data))
		_, err := conn.Write(data)
		if err == nil {
			msg.Ack()
		}
	})

	if err != nil {
		reportError(ctx, err)
	}
}

func reportError(ctx context.Context, err error) {
	val := ctx.Value(contextKey("conn"))
	if val == nil {
		return
	}
	conn := val.(net.Conn)
	conn.Write([]byte(fmt.Sprintf("error:%s", err.Error())))
}

func initFlag() int {

	var (
		secret string
	)

	flags := flag.NewFlagSet("event-framework-subscriber", flag.ExitOnError)
	flags.Usage = usage
	flags.StringVar(&proj, "project", "", "Google Cloud Pub Sub Project Name")
	flags.StringVar(&secret, "credentials-file", "", "Google Secret json file location")
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

func HandleSIGINTKILL() chan os.Signal {
	sig := make(chan os.Signal, 1)

	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	return sig
}

func execute() int {
	if retVal := initFlag(); retVal != -1 {
		return retVal
	}
	closed := false
	ctx, cancel := context.WithCancel(context.Background())

	log.Println("Launching server...")

	ln, err := net.Listen("unix", "/tmp/kw-eventsubscriber.sock")
	if err != nil {
		log.Println(err)
		cancel()
	}
	go func() {
		<-HandleSIGINTKILL()
		log.Println("Received termination signal")
		closed = true
		ln.Close()
		cancel()
		os.Exit(0)
	}()

	defer cancel()

	for {
		select {
		case <-(ctx).Done():
			fmt.Println("Stopped")
			closed = true
			ln.Close()
			return 0
		default:
			log.Println("Waiting for connection")
			conn, err := ln.Accept()
			if err != nil {
				if closed {
					return 0
				}
				log.Println("Error accepting: ", err.Error(), closed)
				continue
			}
			go readConnectionMessage(conn)
		}
	}

}

func readConnectionMessage(conn net.Conn) {
	mu.Lock()
	connected++
	mu.Unlock()
	log.Println("Listening message.", "Current connection number #", connected)

	var buff bytes.Buffer
	chunks := make([]byte, 1)
	delim := []byte(";")

	ctx := context.WithValue(context.Background(), contextKey("conn"), conn)
	ctx, cancel := context.WithCancel(ctx)
	for {
		conn.Read(chunks)
		buff.Write(chunks)

		if chunks[0] == delim[0] {
			message := buff.String()
			buff.Reset()
			if message == "quit!;" || message == ";" {
				cancel()
				conn.Close()
				mu.Lock()
				connected--
				log.Println("Connection closed number of current connection is #", connected)
				mu.Unlock()
				return
			}
			if message != "" {
				go processMessage(ctx, message)
			}
		} else {
			if string(chunks) == "" {
				cancel()
				conn.Close()
				mu.Lock()
				connected--
				log.Println("Connection closed number of current connection is #", connected)
				mu.Unlock()
				return
			}

		}
	}
}

func processMessage(ctx context.Context, message string) {
	fmt.Println(message)
	if strings.HasPrefix(message, "subscribe:") {
		sub := strings.TrimPrefix(message, "subscribe:")
		sub = strings.TrimSuffix(sub, ";")
		log.Println("Subscribing ", sub)
		go subscribe(ctx, sub)
	}
}

func main() {
	log.SetOutput(os.Stderr)
	os.Exit(execute())
}

func usage() {
	fmt.Fprintf(os.Stderr, helpText)
}

const helpText = `
  KW Event Framework Subscriber, subscribe google pubsubs events
  Usage: event-framework-subscriber -project=value -credentials-file=value

    Options:

	-project=""            	Google Cloud Pub Sub Project Name
	-credentials-file="" 	Google Secret Credentials json file location
`
