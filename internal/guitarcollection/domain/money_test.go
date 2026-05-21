package domain

import "testing"

func TestNewMoney_AcceptsValidEUR(t *testing.T) {
	m, err := NewMoney(199900, EUR)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if m.Amount() != 199900 {
		t.Errorf("amount: want 199900, got %d", m.Amount())
	}
	if m.Currency() != EUR {
		t.Errorf("currency: want EUR, got %s", m.Currency())
	}
}

func TestNewMoney_AcceptsValidUSD(t *testing.T) {
	m, err := NewMoney(50000, USD)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if m.Currency() != USD {
		t.Errorf("currency: want USD, got %s", m.Currency())
	}
}

func TestNewMoney_NormalisesCurrencyCase(t *testing.T) {
	m, err := NewMoney(1000, Currency("eur"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if m.Currency() != EUR {
		t.Errorf("currency should be normalised to EUR, got %s", m.Currency())
	}
}

func TestNewMoney_RejectsNegativeAmount(t *testing.T) {
	_, err := NewMoney(-1, EUR)
	if err == nil {
		t.Fatalf("expected validation error for negative amount")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestNewMoney_RejectsUnsupportedCurrency(t *testing.T) {
	_, err := NewMoney(1000, Currency("GBP"))
	if err == nil {
		t.Fatalf("expected validation error for unsupported currency")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestMoney_Equals(t *testing.T) {
	a, _ := NewMoney(1000, EUR)
	b, _ := NewMoney(1000, EUR)
	c, _ := NewMoney(1000, USD)
	d, _ := NewMoney(2000, EUR)
	if !a.Equals(b) {
		t.Errorf("a should equal b")
	}
	if a.Equals(c) {
		t.Errorf("a should not equal c (different currency)")
	}
	if a.Equals(d) {
		t.Errorf("a should not equal d (different amount)")
	}
}
