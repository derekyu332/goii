package ether

import (
	"context"
	"encoding/json"
	"github.com/derekyu332/goii/helper/logger"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"time"
)

const (
	ETHER_REDIAL_DELAY = 5 * time.Second
)

type EtherHandler func(topic string, body []byte) error

type EtherSubscriber struct {
	client  *ethclient.Client
	dialUrl string
	done    chan bool
	handler EtherHandler
}

func NewSubscriber(dialUrl string, handler EtherHandler) *EtherSubscriber {
	newSubscriber := &EtherSubscriber{
		dialUrl: dialUrl,
		done:    make(chan bool),
		handler: handler,
	}
	err := newSubscriber.initSubscriber()

	if err != nil {
		panic(err)
	}

	return newSubscriber
}

func (this *EtherSubscriber) initSubscriber() error {
	var err error
	this.client, err = ethclient.Dial(this.dialUrl)

	if err != nil {
		return err
	}

	go this.subscribeNewHead()

	return nil
}

func (this *EtherSubscriber) subscribeNewHead() error {
	headers := make(chan *types.Header)
	sub, err := this.client.SubscribeNewHead(context.Background(), headers)

	if err != nil {
		logger.Warning("SubscribeNewHead failed %v", err.Error())
		return err
	}

	isLost := false

	for {
		select {
		case err := <-sub.Err():
			logger.Warning("SubscribeNewHead failed %v", err.Error())
			isLost = true
			break
		case header := <-headers:
			logger.Info("Got New Block %v, Number %v", header.Hash().Hex(), header.Number.String())
			body, _ := json.Marshal(header)
			go this.handler("new-heads", body)
		}

		if isLost {
			logger.Warning("WebSocket Lost")
			break
		}
	}

	for {
		select {
		case <-this.done:
			logger.Warning("WebSocket Closed.")
			return nil
		case <-time.After(ETHER_REDIAL_DELAY):
			logger.Warning("WebSocket Retrying Init...")

			if err = this.initSubscriber(); err == nil {
				return nil
			}
		}
	}

	return nil
}
