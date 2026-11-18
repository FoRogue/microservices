package internal

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestCatalogPriceIsExact(t *testing.T) {
	for _, test := range []struct {
		price string
		cents Money
	}{
		{"0.00", 0},
		{"49.90", 4990},
		{"99999999.99", maxCents},
	} {
		cents, err := priceInCents(test.price)
		if err != nil || cents != test.cents {
			t.Fatalf("price %s: got %d cents, error %v", test.price, cents, err)
		}
		output, err := json.Marshal(cents)
		if err != nil || string(output) != test.price {
			t.Fatalf("price %s: got JSON %s, error %v", test.price, output, err)
		}
	}
	for _, invalid := range []string{"", "-0.01", "1.001", "100000000.00"} {
		if _, err := priceInCents(invalid); !errors.Is(err, errDependency) {
			t.Fatalf("accepted invalid price %q, error %v", invalid, err)
		}
	}
}

func TestCreateRejectsInvalidItemsBeforeCallingCatalog(t *testing.T) {
	service := NewService(nil, nil)
	for _, items := range [][]ItemRequest{
		nil,
		{{ProductID: 1, Quantity: 0}},
		{{ProductID: 1, Quantity: 1}, {ProductID: 1, Quantity: 2}},
	} {
		if _, err := service.Create(context.Background(), 1, CreateOrderRequest{Items: items}); !errors.Is(err, errInvalidItems) {
			t.Fatalf("items %v: got %v, want errInvalidItems", items, err)
		}
	}
}
