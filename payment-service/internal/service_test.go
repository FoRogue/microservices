package internal

import "testing"

func TestPaymentStatus(t *testing.T) {
	for _, test := range []struct {
		amount Money
		want   string
	}{
		{"0.01", StatusSucceeded},
		{"1e-20", StatusSucceeded},
		{"-0.01", StatusFailed},
		{"-0.000e+5", StatusFailed},
		{"0", StatusFailed},
		{"", StatusFailed},
	} {
		t.Run(string(test.amount), func(t *testing.T) {
			if got := paymentStatus(test.amount); got != test.want {
				t.Fatalf("got %q, want %q", got, test.want)
			}
		})
	}
}
