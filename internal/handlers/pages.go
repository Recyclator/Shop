package handlers

import (
	"github.com/a-h/templ"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
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
	println("DEBUG: RenderDashboard called")
	data := pages.DashboardData{
		Stats: []pages.StatCard{
			{Label: "Ingresos totales", Value: "$84,291", Icon: "blue", Trend: "↑ 18.4%", Up: true},
			{Label: "Pedidos este mes", Value: "1,847", Icon: "green", Trend: "↑ 7.2%", Up: true},
			{Label: "Clientes activos", Value: "24,309", Icon: "yellow", Trend: "↑ 12.1%", Up: true},
		},
		Orders: []pages.OrderRow{
			{ID: "#NX-4821", Client: "Carlos Rivera", Amount: "$249.00", Status: "Completado", Date: "27 Mar", Badge: "green"},
			{ID: "#NX-4820", Client: "Ana Gómez", Amount: "$89.99", Status: "En tránsito", Date: "26 Mar", Badge: "blue"},
			{ID: "#NX-4819", Client: "Luis Martínez", Amount: "$1,240.00", Status: "Pendiente", Date: "26 Mar", Badge: "yellow"},
		},
		Activity: []pages.ActivityItem{
			{Color: "var(--success)", Text: `<strong>#NX-4821</strong> completado · $249.00`, Time: "hace 3 min"},
			{Color: "var(--accent)", Text: `Nuevo cliente <strong>Laura M.</strong>`, Time: "hace 12 min"},
			{Color: "var(--warning)", Text: `<strong>Auriculares Pro</strong> stock bajo`, Time: "hace 28 min"},
		},
	}
	content := pages.DashboardPage(data)
	return renderPage(c, "Dashboard", "/dashboard", content)
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
	page := layout.Base(title, activeRoute, content)
	h := adaptor.HTTPHandler(templ.Handler(page))
	return h(c)
}
