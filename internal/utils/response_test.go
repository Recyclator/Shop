package utils

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestPaginatedResponse(t *testing.T) {
	app := fiber.New()
	app.Get("/test-page", func(c *fiber.Ctx) error {
		items := []string{"item1", "item2"}
		return Paginated(c, fiber.StatusOK, items, 2, 10, 45)
	})

	req := httptest.NewRequest("GET", "/test-page", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("error realizando request: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("esperaba status 200, obtuvo %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var paginated PaginatedResponse
	if err := json.Unmarshal(body, &paginated); err != nil {
		t.Fatalf("error deserializando JSON: %v", err)
	}

	if !paginated.Success {
		t.Fatal("esperaba success: true")
	}

	if paginated.Meta.TotalPages != 5 { // 45 items con limit 10 = 5 páginas
		t.Fatalf("esperaba total_pages 5, obtuvo %d", paginated.Meta.TotalPages)
	}

	if paginated.Meta.Page != 2 {
		t.Fatalf("esperaba page 2, obtuvo %d", paginated.Meta.Page)
	}
}

func TestPaginatedZeroDivisionSafety(t *testing.T) {
	app := fiber.New()
	app.Get("/test-zero", func(c *fiber.Ctx) error {
		// Probar edge case que anteriormente provocaba división por cero (limit=0, page=-1)
		return Paginated(c, fiber.StatusOK, []string{}, -1, 0, 0)
	})

	req := httptest.NewRequest("GET", "/test-zero", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("error realizando request: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("esperaba status 200, obtuvo %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var paginated PaginatedResponse
	if err := json.Unmarshal(body, &paginated); err != nil {
		t.Fatalf("error deserializando JSON: %v", err)
	}

	if paginated.Meta.Limit != 20 { // Debe autoprotegerse usando default 20
		t.Fatalf("esperaba default limit 20, obtuvo %d", paginated.Meta.Limit)
	}

	if paginated.Meta.Page != 1 { // Debe autoprotegerse usando default 1
		t.Fatalf("esperaba default page 1, obtuvo %d", paginated.Meta.Page)
	}

	if paginated.Meta.TotalPages != 0 {
		t.Fatalf("esperaba total_pages 0 para total 0, obtuvo %d", paginated.Meta.TotalPages)
	}
}
