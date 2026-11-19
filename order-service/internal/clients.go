package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Clients struct {
	http    *http.Client
	auth    string
	catalog string
}

func NewClients(cfg Config) *Clients {
	return &Clients{
		http: &http.Client{
			Timeout: 3 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		auth:    strings.TrimRight(cfg.AuthURL, "/"),
		catalog: strings.TrimRight(cfg.CatalogURL, "/"),
	}
}

func (c *Clients) UserID(ctx context.Context, authorization string) (int64, error) {
	var profile struct {
		ID int64 `json:"id"`
	}
	code, err := c.request(ctx, http.MethodGet, c.auth+"/api/me",
		http.Header{"Authorization": {authorization}}, &profile)
	if err == nil && code == http.StatusUnauthorized {
		return 0, errUnauthorized
	}
	if err != nil || code != http.StatusOK || profile.ID <= 0 {
		return 0, errDependency
	}
	return profile.ID, nil
}

func (c *Clients) Product(ctx context.Context, id int64) (OrderItem, error) {
	var product struct {
		ID    int64       `json:"id"`
		Name  string      `json:"name"`
		Price json.Number `json:"price"`
	}
	url := fmt.Sprintf("%s/api/products/%d", c.catalog, id)
	code, err := c.request(ctx, http.MethodGet, url, nil, &product)
	if err == nil && code == http.StatusNotFound {
		return OrderItem{}, fmt.Errorf("%w: %d", errUnknownProduct, id)
	}
	if err != nil || code != http.StatusOK || product.ID != id {
		return OrderItem{}, errDependency
	}
	price, err := priceInCents(product.Price.String())
	return OrderItem{ProductID: product.ID, ProductName: product.Name, UnitPrice: price}, err
}

func (c *Clients) request(ctx context.Context, method, url string, headers http.Header, response any) (int, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return 0, err
	}
	for key, values := range headers {
		req.Header[key] = values
	}
	res, err := c.http.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		err = json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(response)
	}
	return res.StatusCode, err
}
