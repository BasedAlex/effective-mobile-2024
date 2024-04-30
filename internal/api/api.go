package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type Car struct {
	RegNum string `json:"regNum"`
	Mark   string `json:"mark"`
	Model  string `json:"model"`
	Year   int    `json:"year"`
	Owner  People `json:"owner"`
}

type People struct {
	Name       string `json:"name"`
	Surname    string `json:"surname"`
	Patronymic string `json:"patronymic"`
}

type Client struct {
	host string
	port string
}

func New(host, port string) Client {
	return Client{
		host: host,
		port: port,
	}
}

func (c *Client) GetInfo(ctx context.Context, regNum string) (Car, error) {
	addr := net.JoinHostPort(c.host, c.port)

	url := fmt.Sprintf("http://%s/info?regNum=%s", addr, regNum)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Car{},fmt.Errorf("create request: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return Car{},fmt.Errorf("get response: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return Car{}, fmt.Errorf("incorrect status code: %d", res.StatusCode)
	}

	var car Car

	err = json.NewDecoder(res.Body).Decode(&car)
	if err != nil {
		return Car{},fmt.Errorf("json decode: %w", err)
	}
	// log which car we got from API
	log.Info(car)

	return car, nil
}