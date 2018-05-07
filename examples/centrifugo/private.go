// Private channel subscription example.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/centrifugal/centrifuge-go"
)

// In production you need to receive credentials from application backend.
func credentials() *centrifuge.Credentials {
	// Never show secret to client of your application. Keep it on your application backend only.
	secret := "secret"
	// Application user ID.
	user := "42"
	// Exp as string.
	exp := centrifuge.Exp(1000000000)
	// Empty info.
	info := ""
	// Generate sign so Centrifugo server can trust connection parameters received from client.
	sign := centrifuge.GenerateClientSign(secret, user, exp, info)

	return &centrifuge.Credentials{
		User: user,
		Exp:  exp,
		Info: info,
		Sign: sign,
	}
}

type eventHandler struct{}

func (h *eventHandler) OnPrivateSub(c *centrifuge.Client, e centrifuge.PrivateSubEvent) (centrifuge.PrivateSign, error) {
	// Here we allow everyone to subscribe on private channel.
	// To reject subscription we could return any error from this func.
	// In most real application secret key must not be kept on client side
	// and here must be request to your backend to get channel sign.
	info := ""
	sign := centrifuge.GenerateChannelSign("secret", e.ClientID, e.Channel, info)
	privateSign := centrifuge.PrivateSign{Sign: sign, Info: info}
	return privateSign, nil
}

type subEventHandler struct{}

func (h *subEventHandler) OnSubscribeSuccess(sub *centrifuge.Subscription, e centrifuge.SubscribeSuccessEvent) {
	log.Println(fmt.Sprintf("Successfully subscribed on private channel %s", sub.Channel()))
	os.Exit(0)
}

func (h *subEventHandler) OnSubscribeError(sub *centrifuge.Subscription, e centrifuge.SubscribeErrorEvent) {
	log.Println(fmt.Sprintf("Error subscribing to private channel %s: %v", sub.Channel(), e.Error))
	os.Exit(1)
}

func newConnection() *centrifuge.Client {
	creds := credentials()
	wsURL := "ws://localhost:8000/connection/websocket"

	handler := &eventHandler{}
	events := centrifuge.NewEventHub()
	events.OnPrivateSub(handler)

	c := centrifuge.New(wsURL, events, centrifuge.DefaultConfig())
	c.SetCredentials(creds)

	err := c.Connect()
	if err != nil {
		log.Fatalln(err)
	}
	return c
}

func main() {
	log.Println("Start program")
	c := newConnection()
	defer c.Close()

	subEvents := centrifuge.NewSubscriptionEventHub()
	subEventHandler := &subEventHandler{}
	subEvents.OnSubscribeSuccess(subEventHandler)
	subEvents.OnSubscribeError(subEventHandler)

	// Subscribe on private channel.
	c.Subscribe("$public:chat", subEvents)

	select {}
}
