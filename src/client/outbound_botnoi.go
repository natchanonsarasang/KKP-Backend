package client

import (
	"errors"
	"go-fiber-template/domain/entities"
	"log"
	"os"

	"github.com/go-resty/resty/v2"
)

type OutboundBotnoiClient struct {
	client      *resty.Client
	baseURL     string
	accessToken string
}

type IOutboundBotnoiClient interface {
	MakeCall(payload entities.OutboundBotnoiDataModel) error
}

func NewOutboundBotnoiClient(token string, host string, port string) IOutboundBotnoiClient {
	return &OutboundBotnoiClient{
		client:      resty.New(),
		baseURL:     os.Getenv("OUTBOUND_URL"),
		accessToken: os.Getenv("OUTBOUND_ACCESS_TOKEN"),
	}
}

func (c *OutboundBotnoiClient) MakeCall(payload entities.OutboundBotnoiDataModel) error {
	// Implement the logic to make an outbound call using the Botnoi API
	// You can use an HTTP client to send a request to the Botnoi API endpoint with the payload
	log.Printf("Making call with payload: %+v\n", payload)

	resp, err := c.client.R().
		SetHeader("Authorization", "Bearer "+c.accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		Post(c.baseURL)

	if err != nil {
		log.Printf("Error making call: %v\n", err)
		return err
	}

	if resp.IsError() {
		log.Printf("Error response from Botnoi API: %s\n", resp.String())
		return errors.New("error response from Botnoi API: " + resp.String())
	}

	return nil
}
