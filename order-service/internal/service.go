package internal

import (
	"context"
	"errors"
	"sort"
	"time"
)

var (
	errUnauthorized   = errors.New("valid bearer token is required")
	errDependency     = errors.New("upstream service unavailable")
	errInvalidItems   = errors.New("provide 1 to 20 distinct products with positive IDs and quantities from 1 to 1000")
	errUnknownProduct = errors.New("product not found")
	errInvalidTotal   = errors.New("order total is outside the payment limit")
)

type Service struct {
	repo    *Repository
	clients *Clients
}

func NewService(repo *Repository, clients *Clients) *Service {
	return &Service{repo: repo, clients: clients}
}

func (s *Service) Create(ctx context.Context, userID int64, req CreateOrderRequest) (Order, error) {
	if len(req.Items) == 0 || len(req.Items) > 20 {
		return Order{}, errInvalidItems
	}
	seen := make(map[int64]bool, len(req.Items))
	for _, requested := range req.Items {
		if requested.ProductID <= 0 || requested.Quantity < 1 || requested.Quantity > 1000 || seen[requested.ProductID] {
			return Order{}, errInvalidItems
		}
		seen[requested.ProductID] = true
	}
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	order := Order{UserID: userID, Status: pending, Items: make([]OrderItem, 0, len(req.Items))}
	for _, requested := range req.Items {
		item, err := s.clients.Product(ctx, requested.ProductID)
		if err != nil {
			return Order{}, err
		}
		item.Quantity = requested.Quantity
		order.Total += item.UnitPrice * Money(item.Quantity)
		if order.Total > maxCents {
			return Order{}, errInvalidTotal
		}
		order.Items = append(order.Items, item)
	}
	sort.Slice(order.Items, func(i, j int) bool {
		return order.Items[i].ProductID < order.Items[j].ProductID
	})
	return s.repo.Create(ctx, order)
}

func (s *Service) Get(ctx context.Context, userID int64, id string) (Order, error) {
	return s.repo.Get(ctx, userID, id)
}

func (s *Service) Pay(ctx context.Context, userID int64, id string) (Order, error) {
	order, err := s.repo.Get(ctx, userID, id)
	if err != nil || order.Status != pending {
		return order, err
	}
	if err := s.repo.RequestPayment(ctx, order); err != nil {
		return Order{}, err
	}
	return order, nil
}
