package handlers

import (
	"fmt"
	"html"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/nexora/backend/internal/database"
	"github.com/nexora/backend/internal/models"
	"github.com/nexora/backend/views/layout"
	"github.com/nexora/backend/views/pages"
)

// ===================== Page Handlers =====================

// RenderAuth handles GET / (login/register page)
func RenderAuth(c *fiber.Ctx) error {
	h := adaptor.HTTPHandler(templ.Handler(pages.AuthPage()))
	return h(c)
}

// RenderDashboard handles GET /dashboard
func RenderDashboard(c *fiber.Ctx) error {
	data := buildDashboardData()
	content := pages.DashboardPage(data)
	return renderPage(c, "Dashboard", "/dashboard", content)
}

func buildDashboardData() pages.DashboardData {
	if database.DB == nil {
		return pages.DashboardData{
			Stats: []pages.StatCard{
				{Label: "Ingresos totales", Value: "$0.00", Icon: "blue", Trend: "0%", Up: true},
				{Label: "Pedidos este mes", Value: "0", Icon: "green", Trend: "0%", Up: true},
				{Label: "Clientes activos", Value: "0", Icon: "yellow", Trend: "0%", Up: true},
			},
			Orders:   []pages.OrderRow{},
			Activity: []pages.ActivityItem{},
		}
	}

	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	startOfLastMonth := startOfMonth.AddDate(0, -1, 0)

	// 1. Ingresos totales (excluyendo cancelados)
	var totalRevenue float64
	database.DB.Model(&models.Order{}).
		Where("estado != ?", "cancelado").
		Select("COALESCE(SUM(total), 0)").
		Scan(&totalRevenue)

	var revenueThisMonth float64
	database.DB.Model(&models.Order{}).
		Where("created_at >= ? AND estado != ?", startOfMonth, "cancelado").
		Select("COALESCE(SUM(total), 0)").
		Scan(&revenueThisMonth)

	var revenueLastMonth float64
	database.DB.Model(&models.Order{}).
		Where("created_at >= ? AND created_at < ? AND estado != ?", startOfLastMonth, startOfMonth, "cancelado").
		Select("COALESCE(SUM(total), 0)").
		Scan(&revenueLastMonth)

	revTrend := "0%"
	revUp := true
	if revenueLastMonth > 0 {
		diff := (revenueThisMonth - revenueLastMonth) / revenueLastMonth * 100
		if diff >= 0 {
			revTrend = fmt.Sprintf("↑ %.1f%%", diff)
			revUp = true
		} else {
			revTrend = fmt.Sprintf("↓ %.1f%%", math.Abs(diff))
			revUp = false
		}
	} else if revenueThisMonth > 0 {
		revTrend = "↑ 100%"
		revUp = true
	}

	// 2. Pedidos este mes
	var ordersThisMonth int64
	database.DB.Model(&models.Order{}).
		Where("created_at >= ? AND estado != ?", startOfMonth, "cancelado").
		Count(&ordersThisMonth)

	var ordersLastMonth int64
	database.DB.Model(&models.Order{}).
		Where("created_at >= ? AND created_at < ? AND estado != ?", startOfLastMonth, startOfMonth, "cancelado").
		Count(&ordersLastMonth)

	ordersTrend := "0%"
	ordersUp := true
	if ordersLastMonth > 0 {
		diff := float64(ordersThisMonth-ordersLastMonth) / float64(ordersLastMonth) * 100
		if diff >= 0 {
			ordersTrend = fmt.Sprintf("↑ %.1f%%", diff)
			ordersUp = true
		} else {
			ordersTrend = fmt.Sprintf("↓ %.1f%%", math.Abs(diff))
			ordersUp = false
		}
	} else if ordersThisMonth > 0 {
		ordersTrend = fmt.Sprintf("+%d este mes", ordersThisMonth)
		ordersUp = true
	}

	// 3. Clientes registrados/activos
	var customerCount int64
	database.DB.Model(&models.Customer{}).Count(&customerCount)

	var customersThisMonth int64
	database.DB.Model(&models.Customer{}).
		Where("created_at >= ?", startOfMonth).
		Count(&customersThisMonth)

	custTrend := "Total"
	custUp := true
	if customersThisMonth > 0 {
		custTrend = fmt.Sprintf("+%d este mes", customersThisMonth)
		custUp = true
	}

	stats := []pages.StatCard{
		{
			Label: "Ingresos totales",
			Value: formatDashboardCurrency(totalRevenue),
			Icon:  "blue",
			Trend: revTrend,
			Up:    revUp,
		},
		{
			Label: "Pedidos este mes",
			Value: formatDashboardNumber(ordersThisMonth),
			Icon:  "green",
			Trend: ordersTrend,
			Up:    ordersUp,
		},
		{
			Label: "Clientes activos",
			Value: formatDashboardNumber(customerCount),
			Icon:  "yellow",
			Trend: custTrend,
			Up:    custUp,
		},
	}

	// 4. Pedidos recientes
	var recentOrders []models.Order
	database.DB.Model(&models.Order{}).
		Preload("Cliente").
		Order("created_at DESC").
		Limit(5).
		Find(&recentOrders)

	ordersRows := make([]pages.OrderRow, 0, len(recentOrders))
	for _, o := range recentOrders {
		orderID := o.NumeroOrden
		if orderID == "" {
			orderID = fmt.Sprintf("NX-%04d", o.ID)
		}
		if !strings.HasPrefix(orderID, "#") {
			orderID = "#" + orderID
		}

		clientName := strings.TrimSpace(o.ClienteNombre)
		if clientName == "" && o.Cliente != nil && o.Cliente.Nombre != "" {
			clientName = o.Cliente.Nombre
		}
		if clientName == "" {
			clientName = "Cliente Mostrador"
		}

		statusText, badge := getOrderStatusDisplay(o.Estado)

		ordersRows = append(ordersRows, pages.OrderRow{
			ID:     orderID,
			Client: clientName,
			Amount: formatDashboardCurrency(o.Total),
			Status: statusText,
			Date:   formatDashboardDate(o.CreatedAt),
			Badge:  badge,
		})
	}

	// 5. Actividad reciente (interleaving pedidos, alertas de stock y clientes)
	type rawActivityItem struct {
		timestamp time.Time
		item      pages.ActivityItem
	}
	var rawActivities []rawActivityItem

	// Actividades de pedidos (hasta 4)
	for i, o := range recentOrders {
		if i >= 4 {
			break
		}
		orderNum := o.NumeroOrden
		if orderNum == "" {
			orderNum = fmt.Sprintf("NX-%04d", o.ID)
		}
		if !strings.HasPrefix(orderNum, "#") {
			orderNum = "#" + orderNum
		}

		statusText, badge := getOrderStatusDisplay(o.Estado)
		color := "var(--accent)"
		if badge == "green" {
			color = "var(--success)"
		} else if badge == "red" {
			color = "var(--danger)"
		} else if badge == "yellow" {
			color = "var(--warning)"
		}

		rawActivities = append(rawActivities, rawActivityItem{
			timestamp: o.CreatedAt,
			item: pages.ActivityItem{
				Color: color,
				Text:  fmt.Sprintf("<strong>%s</strong> %s · %s", html.EscapeString(orderNum), html.EscapeString(strings.ToLower(statusText)), formatDashboardCurrency(o.Total)),
				Time:  formatDashboardRelativeTime(o.CreatedAt),
			},
		})
	}

	// Alertas de stock bajo (hasta 3)
	var lowStockProducts []models.Product
	database.DB.Model(&models.Product{}).
		Where("controla_inventario = ? AND stock <= stock_minimo AND estado = ?", true, "activo").
		Order("stock ASC").
		Limit(3).
		Find(&lowStockProducts)

	for _, p := range lowStockProducts {
		t := p.UpdatedAt
		if t.IsZero() {
			t = now
		}
		rawActivities = append(rawActivities, rawActivityItem{
			timestamp: t,
			item: pages.ActivityItem{
				Color: "var(--warning)",
				Text:  fmt.Sprintf("<strong>%s</strong> stock bajo (%d restantes)", html.EscapeString(p.Nombre), p.Stock),
				Time:  formatDashboardRelativeTime(t),
			},
		})
	}

	// Clientes registrados recientemente (hasta 2)
	var recentCustomers []models.Customer
	database.DB.Model(&models.Customer{}).
		Order("created_at DESC").
		Limit(2).
		Find(&recentCustomers)

	for _, cust := range recentCustomers {
		rawActivities = append(rawActivities, rawActivityItem{
			timestamp: cust.CreatedAt,
			item: pages.ActivityItem{
				Color: "var(--accent)",
				Text:  fmt.Sprintf("Nuevo cliente <strong>%s</strong>", html.EscapeString(cust.Nombre)),
				Time:  formatDashboardRelativeTime(cust.CreatedAt),
			},
		})
	}

	// Ordenar actividades por fecha más reciente
	sort.Slice(rawActivities, func(i, j int) bool {
		return rawActivities[i].timestamp.After(rawActivities[j].timestamp)
	})

	finalActivity := make([]pages.ActivityItem, 0, 6)
	for i, ra := range rawActivities {
		if i >= 6 {
			break
		}
		finalActivity = append(finalActivity, ra.item)
	}

	return pages.DashboardData{
		Stats:    stats,
		Orders:   ordersRows,
		Activity: finalActivity,
	}
}

func getOrderStatusDisplay(estado string) (text string, badge string) {
	switch strings.ToLower(strings.TrimSpace(estado)) {
	case "entregado", "completado":
		return "Completado", "green"
	case "enviado", "en_transito":
		return "En tránsito", "blue"
	case "confirmado", "procesamiento", "en_proceso":
		return "Confirmado", "blue"
	case "cancelado", "fallido":
		return "Cancelado", "red"
	case "reembolso", "reembolsado":
		return "Reembolsado", "red"
	case "pendiente":
		return "Pendiente", "yellow"
	default:
		if estado == "" {
			return "Pendiente", "yellow"
		}
		return strings.Title(estado), "yellow"
	}
}

func formatDashboardDate(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	orderDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	if orderDay.Equal(today) {
		return "Hoy " + t.Format("15:04")
	}
	if orderDay.Equal(today.AddDate(0, 0, -1)) {
		return "Ayer"
	}
	months := []string{
		"", "Ene", "Feb", "Mar", "Abr", "May", "Jun", "Jul", "Ago", "Sep", "Oct", "Nov", "Dic",
	}
	m := int(t.Month())
	if m >= 1 && m <= 12 {
		return fmt.Sprintf("%02d %s", t.Day(), months[m])
	}
	return t.Format("02/01/2006")
}

func formatDashboardRelativeTime(t time.Time) string {
	if t.IsZero() {
		return "reciente"
	}
	diff := time.Since(t)
	if diff < 0 {
		return "ahora"
	}
	if diff < time.Minute {
		return "hace un momento"
	}
	if diff < time.Hour {
		mins := int(diff.Minutes())
		if mins <= 1 {
			return "hace 1 min"
		}
		return fmt.Sprintf("hace %d min", mins)
	}
	if diff < 24*time.Hour {
		hrs := int(diff.Hours())
		if hrs <= 1 {
			return "hace 1 h"
		}
		return fmt.Sprintf("hace %d h", hrs)
	}
	if diff < 48*time.Hour {
		return "ayer"
	}
	days := int(diff.Hours() / 24)
	if days < 30 {
		return fmt.Sprintf("hace %d d", days)
	}
	return t.Format("02/01/2006")
}

func formatDashboardNumber(n int64) string {
	if n == 0 {
		return "0"
	}
	isNeg := n < 0
	if isNeg {
		n = -n
	}
	s := fmt.Sprintf("%d", n)
	var result []byte
	l := len(s)
	for i, b := range []byte(s) {
		if i > 0 && (l-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, b)
	}
	if isNeg {
		return "-" + string(result)
	}
	return string(result)
}

func formatDashboardCurrency(amount float64) string {
	isNeg := amount < 0
	if isNeg {
		amount = -amount
	}
	intPart := int64(amount)
	decPart := int64(math.Round((amount - float64(intPart)) * 100))
	if decPart >= 100 {
		intPart++
		decPart = 0
	}

	intStr := formatDashboardNumber(intPart)
	if isNeg {
		return fmt.Sprintf("-$%s.%02d", intStr, decPart)
	}
	return fmt.Sprintf("$%s.%02d", intStr, decPart)
}

// RenderProducts handles GET /products
func RenderProducts(c *fiber.Ctx) error {
	return renderPage(c, "Productos", "/products", pages.ProductsPage())
}

// RenderCategories handles GET /categories
func RenderCategories(c *fiber.Ctx) error {
	return renderPage(c, "Categorías", "/categories", pages.CategoriesPage())
}

// RenderCustomers handles GET /customers
func RenderCustomers(c *fiber.Ctx) error {
	content := pages.CustomersPage(pages.CustomersData{InitialPage: 1, InitialLimit: 20})
	return renderPage(c, "Clientes", "/customers", content)
}

// RenderLayaways handles GET /layaways
func RenderLayaways(c *fiber.Ctx) error {
	return renderPage(c, "Separados", "/layaways", pages.LayawaysPage())
}

// RenderUsers handles GET /users
func RenderUsers(c *fiber.Ctx) error {
	content := pages.UsersPage(pages.UsersData{InitialPage: 1})
	return renderPage(c, "Usuarios", "/users", content)
}

// RenderPOS handles GET /pos
func RenderPOS(c *fiber.Ctx) error {
	return renderPage(c, "Punto de Venta", "/pos", pages.POSPage())
}

// RenderProfile handles GET /profile
func RenderProfile(c *fiber.Ctx) error {
	return renderPage(c, "Perfil", "/profile", pages.ProfilePage())
}

// RenderSettings handles GET /settings
func RenderSettings(c *fiber.Ctx) error {
	return renderPage(c, "Configuración", "/settings", pages.SettingsPage())
}

// RenderRolesPermissions handles GET /roles-permissions
func RenderRolesPermissions(c *fiber.Ctx) error {
	return renderPage(c, "Roles y Permisos", "/roles-permissions", pages.RolesPermissionsPage())
}

// renderPage renders a complete page with base layout
func renderPage(c *fiber.Ctx, title, activeRoute string, content templ.Component) error {
	c.Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
	c.Set("Pragma", "no-cache")
	c.Set("Expires", "0")
	page := layout.Base(title, activeRoute, content)
	h := adaptor.HTTPHandler(templ.Handler(page))
	return h(c)
}

