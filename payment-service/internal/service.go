package internal

import "context"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreatePayment(ctx context.Context, req CreatePaymentRequest) (Payment, error) {
	return s.repo.CreatePayment(ctx, req, paymentStatus(req.Amount))
}

func paymentStatus(amount Money) string {
	if amount.sign() <= 0 {
		return StatusFailed
	}
	return StatusSucceeded
}

func (s *Service) GetPayment(ctx context.Context, id int64) (Payment, error) {
	return s.repo.GetPayment(ctx, id)
}

func (s *Service) ListPayments(ctx context.Context, orderID string) ([]Payment, error) {
	return s.repo.ListPayments(ctx, orderID)
}
