package ether

import (
	"context"
	"encoding/json"
	"github.com/derekyu332/goii/helper/logger"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"time"
)

const (
	ETHER_REDIAL_DELAY = 5 * time.Second
)

type EtherHandler func(topic string, body []byte) error

type EtherSubscriber struct {
	dialUrl         string
	filterAddresses []string
	done            chan bool
	handler         EtherHandler
}

func NewSubscriber(dialUrl string, filterAddresses []string, handler EtherHandler) *EtherSubscriber {
	newSubscriber := &EtherSubscriber{
		dialUrl:         dialUrl,
		filterAddresses: filterAddresses,
		done:            make(chan bool),
		handler:         handler,
	}

	err := newSubscriber.initHeadSubscriber()

	if err != nil {
		panic(err)
	}

	if len(filterAddresses) > 0 {
		err = newSubscriber.initLogsSubscriber()

		if err != nil {
			panic(err)
		}
	}

	return newSubscriber
}

func (this *EtherSubscriber) initHeadSubscriber() error {
	client, err := ethclient.Dial(this.dialUrl)

	if err != nil {
		return err
	}

	go this.subscribeHead(client)

	return nil
}

func (this *EtherSubscriber) subscribeHead(client *ethclient.Client) error {
	headers := make(chan *types.Header)
	sub, err := client.SubscribeNewHead(context.Background(), headers)

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

			if err = this.initHeadSubscriber(); err == nil {
				return nil
			}
		}
	}

	return nil
}

func (this *EtherSubscriber) initLogsSubscriber() error {
	client, err := ethclient.Dial(this.dialUrl)

	if err != nil {
		return err
	}

	go this.subscribeLogs(client)

	return nil
}

func (this *EtherSubscriber) subscribeLogs(client *ethclient.Client) error {
	filterlogs := make(chan types.Log)
	var addresses []common.Address

	for _, address := range this.filterAddresses {
		addresses = append(addresses, common.HexToAddress(address))
	}

	sub, err := client.SubscribeFilterLogs(context.Background(), ethereum.FilterQuery{
		Addresses: addresses,
	}, filterlogs)

	if err != nil {
		logger.Warning("SubscribeFilterLogs failed %v", err.Error())
		return err
	}

	isLost := false

	for {
		select {
		case err := <-sub.Err():
			logger.Warning("SubscribeFilterLogs failed %v", err.Error())
			isLost = true
			break
		case log := <-filterlogs:
			logger.Info("Got New Transaction %v, BlockNumber %v", log.TxHash, log.BlockNumber)
			body, _ := json.Marshal(log)
			go this.handler("filter-logs", body)
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

			if err = this.initLogsSubscriber(); err == nil {
				return nil
			}
		}
	}

	return nil
}
