package internal

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMoneyPreservesDecimalJSON(t *testing.T) {
	const decimal = "99999999.990000000001"
	var price Money
	if err := json.Unmarshal([]byte(decimal), &price); err != nil {
		t.Fatal(err)
	}
	output, err := json.Marshal(price)
	if err != nil {
		t.Fatal(err)
	}
	if string(output) != decimal || price.databaseValue() != decimal {
		t.Fatalf("decimal changed: JSON %s, database %s", output, price.databaseValue())
	}
}

func TestCreateProductRejectsInvalidPrices(t *testing.T) {
	handler := NewHandler(nil).Routes()
	for _, price := range []string{`"1.25"`, "-0.001", "-1e-20"} {
		t.Run(price, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/products",
				strings.NewReader(`{"name":"test","price":`+price+`}`))
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("got status %d, want %d", response.Code, http.StatusBadRequest)
			}
		})
	}
}
