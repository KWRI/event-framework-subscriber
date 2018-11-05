package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	ps "cloud.google.com/go/pubsub"
)

var (
	proj string
)

func GetClient() (*ps.Client, context.Context) {

	if proj == "" {
		panic("Can't find Google cloud project used for pubsub.")
	}

	ctx := context.Background()
	c, err := ps.NewClient(ctx, proj)
	if err != nil {
		panic(err)
	}
	return c, ctx
}

func subscribe(subscription string, ch chan []byte) {

	client, ctx := GetClient()
	sub := client.Subscription(subscription)

	cctx, cancel := context.WithCancel(ctx)

	err := sub.Receive(cctx, func(ctx context.Context, msg *ps.Message) {
		ch <- msg.Data
		msg.Ack()
	})
	if err != nil {
		cancel()
		panic(err)
	}

}

func main() {
	p := flag.String("project", "", "Google Cloud Pub Sub Project")
	subscriptionString := flag.String("subscriptions", "", "Subscription Names")
	flag.Parse()
	proj = *p
	subscriptions := strings.Split(*subscriptionString, " ")

	ch := make(chan []byte)
	for _, subscription := range subscriptions {
		go subscribe(subscription, ch)
	}

	for {
		a := <-ch
		fmt.Println(string(a))
	}

}
