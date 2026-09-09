package handlers

import (
	"strings"
	"testing"
	"time"

	"github.com/nexora/backend/internal/database"
)

func TestGetOrderStatusDisplay(t *testing.T) {
	tests := []struct {
		input       string
		expectedTxt string
		expectedBdg string
	}{
		{"entregado", "Completado", "green"},
		{"COMPLETADO", "Completado", "green"},
		{"enviado", "En tránsito", "blue"},
		{"en_transito", "En tránsito", "blue"},
		{"confirmado", "Confirmado", "blue"},
		{"en_proceso", "Confirmado", "blue"},
		{"cancelado", "Cancelado", "red"},
		{"fallido", "Cancelado", "red"},
		{"reembolso", "Reembolsado", "red"},
		{"pendiente", "Pendiente", "yellow"},
		{"", "Pendiente", "yellow"},
	}

	for _, tt := range tests {
		txt, bdg := getOrderStatusDisplay(tt.input)
		if txt != tt.expectedTxt || bdg != tt.expectedBdg {
			t.Errorf("getOrderStatusDisplay(%q) = (%q, %q); expected (%q, %q)",
				tt.input, txt, bdg, tt.expectedTxt, tt.expectedBdg)
		}
	}
}

func TestFormatDashboardNumber(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0"},
		{5, "5"},
		{999, "999"},
		{1000, "1,000"},
		{1847, "1,847"},
		{24309, "24,309"},
		{1000000, "1,000,000"},
		{-1500, "-1,500"},
	}

	for _, tt := range tests {
		actual := formatDashboardNumber(tt.input)
		if actual != tt.expected {
			t.Errorf("formatDashboardNumber(%d) = %q; expected %q", tt.input, actual, tt.expected)
		}
	}
}

func TestFormatDashboardCurrency(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0.0, "$0.00"},
		{249.00, "$249.00"},
		{89.99, "$89.99"},
		{1240.50, "$1,240.50"},
		{84291.00, "$84,291.00"},
		{-50.25, "-$50.25"},
	}

	for _, tt := range tests {
		actual := formatDashboardCurrency(tt.input)
		if actual != tt.expected {
			t.Errorf("formatDashboardCurrency(%f) = %q; expected %q", tt.input, actual, tt.expected)
		}
	}
}

func TestFormatDashboardDate(t *testing.T) {
	zero := time.Time{}
	if got := formatDashboardDate(zero); got != "-" {
		t.Errorf("formatDashboardDate(zero) = %q; want '-'", got)
	}

	now := time.Now()
	todayStr := formatDashboardDate(now)
	if !strings.HasPrefix(todayStr, "Hoy ") {
		t.Errorf("formatDashboardDate(now) = %q; want prefix 'Hoy '", todayStr)
	}

	yesterday := now.AddDate(0, 0, -1)
	yesterdayStr := formatDashboardDate(yesterday)
	if yesterdayStr != "Ayer" {
		t.Errorf("formatDashboardDate(yesterday) = %q; want 'Ayer'", yesterdayStr)
	}

	fixed := time.Date(2025, time.March, 15, 10, 0, 0, 0, time.UTC)
	fixedStr := formatDashboardDate(fixed)
	if fixedStr != "15 Mar" {
		t.Errorf("formatDashboardDate(fixed) = %q; want '15 Mar'", fixedStr)
	}
}

func TestFormatDashboardRelativeTime(t *testing.T) {
	zero := time.Time{}
	if got := formatDashboardRelativeTime(zero); got != "reciente" {
		t.Errorf("formatDashboardRelativeTime(zero) = %q; want 'reciente'", got)
	}

	now := time.Now()
	if got := formatDashboardRelativeTime(now.Add(-10 * time.Second)); got != "hace un momento" {
		t.Errorf("formatDashboardRelativeTime(-10s) = %q; want 'hace un momento'", got)
	}

	if got := formatDashboardRelativeTime(now.Add(-5 * time.Minute)); got != "hace 5 min" {
		t.Errorf("formatDashboardRelativeTime(-5m) = %q; want 'hace 5 min'", got)
	}

	if got := formatDashboardRelativeTime(now.Add(-3 * time.Hour)); got != "hace 3 h" {
		t.Errorf("formatDashboardRelativeTime(-3h) = %q; want 'hace 3 h'", got)
	}
}

func TestBuildDashboardDataNilDB(t *testing.T) {
	// Ensure safe fallback when DB is nil
	origDB := database.DB
	database.DB = nil
	defer func() {
		database.DB = origDB
	}()

	data := buildDashboardData()
	if len(data.Stats) != 3 {
		t.Fatalf("expected 3 stat cards, got %d", len(data.Stats))
	}
	if data.Stats[0].Value != "$0.00" {
		t.Errorf("expected $0.00 for revenue stat, got %s", data.Stats[0].Value)
	}
	if len(data.Orders) != 0 {
		t.Errorf("expected empty orders slice, got %d", len(data.Orders))
	}
	if len(data.Activity) != 0 {
		t.Errorf("expected empty activity slice, got %d", len(data.Activity))
	}
}
