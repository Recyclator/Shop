package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/nexora/backend/internal/handlers"
	"github.com/nexora/backend/internal/middleware"
)

func Setup(app *fiber.App) {
	// Grupo API
	api := app.Group("/api")

	// ========== AUTH ==========
	auth := api.Group("/auth")
	auth.Post("/register", handlers.Register)
	auth.Post("/login", handlers.Login)
	auth.Post("/refresh", handlers.RefreshToken)
	auth.Post("/logout", handlers.Logout)
	auth.Post("/change-password", middleware.AuthMiddleware(), handlers.ChangePassword)
	auth.Get("/me", middleware.AuthMiddleware(), handlers.GetCurrentUser)
	auth.Get("/profile", middleware.AuthMiddleware(), handlers.GetFullProfile)
	auth.Put("/profile", middleware.AuthMiddleware(), handlers.UpdateProfile)
	auth.Put("/theme", middleware.AuthMiddleware(), handlers.UpdateTheme)

	// ========== USERS ==========
	users := api.Group("/users", middleware.AuthMiddleware())
	users.Get("/", handlers.GetUsers)
	users.Get("/search", handlers.SearchUsers)
	users.Post("/", handlers.CreateUser)
	users.Get("/:id", handlers.GetUserByID)
	users.Put("/:id", handlers.UpdateUser)
	users.Delete("/:id", handlers.DeleteUser)
	users.Post("/:id/roles", handlers.AssignRoles)

	// ========== PRODUCTS ==========
	products := api.Group("/products", middleware.AuthMiddleware())
	products.Get("/", handlers.GetProducts)
	products.Get("/featured", handlers.GetFeaturedProducts)
	products.Get("/search", handlers.SearchProducts)
	products.Get("/category/:id", handlers.GetProductsByCategory)
	products.Get("/sku/:sku", handlers.GetProductBySKU)
	products.Get("/barcode/:code", middleware.AuthMiddleware(), handlers.GetProductByBarcode)
	products.Post("/", handlers.CreateProduct)
	products.Get("/:id", handlers.GetProductByID)
	products.Put("/:id", handlers.UpdateProduct)
	products.Delete("/:id", handlers.DeleteProduct)
	products.Post("/:id/qr", handlers.GenerateProductQR)
	products.Post("/:id/stock/adjust", handlers.AdjustProductStock)

	// ========== PRODUCT VARIANTS ==========
	products.Get("/:id/variants", handlers.GetVariantsByProduct)
	products.Post("/:id/variants", handlers.CreateVariant)
	products.Put("/:id/variants/:var_id", handlers.UpdateVariant)
	products.Delete("/:id/variants/:var_id", handlers.DeleteVariant)

	// ========== CATEGORIES ==========
	categories := api.Group("/categories")
	categories.Get("/", handlers.GetCategories)
	categories.Get("/:id", handlers.GetCategoryByID)
	categories.Post("/", middleware.AuthMiddleware(), handlers.CreateCategory)
	categories.Put("/:id", middleware.AuthMiddleware(), handlers.UpdateCategory)
	categories.Delete("/:id", middleware.AuthMiddleware(), handlers.DeleteCategory)

	// ========== ATTRIBUTES ==========
	attributes := api.Group("/attributes")
	attributes.Get("/global", handlers.GetGlobalAttributes)
	attributes.Get("/category/:id", handlers.GetAttributesByCategory)
	attributes.Post("/category/:id", middleware.AuthMiddleware(), handlers.CreateAttribute)
	attributes.Put("/:id", middleware.AuthMiddleware(), handlers.UpdateAttribute)
	attributes.Delete("/:id", middleware.AuthMiddleware(), handlers.DeleteAttribute)

	// ========== ORDERS ==========
	orders := api.Group("/orders", middleware.AuthMiddleware())
	orders.Get("/", handlers.GetOrders)
	orders.Post("/", handlers.CreateOrder)
	orders.Get("/:id", handlers.GetOrderByID)
	orders.Put("/:id", handlers.UpdateOrder)
	orders.Post("/:id/cancel", handlers.CancelOrder)
	orders.Post("/pos", handlers.QuickPOSSale)

	// ========== POS ==========
	pos := api.Group("/pos", middleware.AuthMiddleware())
	pos.Get("/search", handlers.POSSearch)
	pos.Get("/products", handlers.POSSearchSimple)

	// ========== CUSTOMERS ==========
	customers := api.Group("/customers", middleware.AuthMiddleware())
	customers.Get("/", handlers.GetCustomers)
	customers.Get("/cedula/:cedula", handlers.GetCustomerByCedula)
	customers.Post("/", handlers.CreateCustomer)
	customers.Get("/:id", handlers.GetCustomerByID)
	customers.Put("/:id", handlers.UpdateCustomer)
	customers.Delete("/:id", handlers.DeleteCustomer)

	// ========== LAYAWAYS (SEPARADOS) ==========
	layaways := api.Group("/layaways", middleware.AuthMiddleware())
	layaways.Get("/", handlers.GetLayaways)
	layaways.Post("/", handlers.CreateLayaway)
	layaways.Get("/:id", handlers.GetLayawayByID)
	layaways.Put("/:id", handlers.UpdateLayawayStatus)
	layaways.Post("/:id/cancel", handlers.CancelLayaway)
	layaways.Post("/:id/payments", handlers.AddPayment)
	layaways.Get("/:id/payments", handlers.GetLayawayPayments)
	layaways.Post("/check-expired", handlers.ExpiredLayaways)
	layaways.Delete("/:id", handlers.DeleteLayaway)

	// ========== SYSTEM ==========
	system := api.Group("/system", middleware.AuthMiddleware())
	system.Post("/refresh-permissions", handlers.RefreshPermissions)

	// ========== ROLES ==========
	roles := api.Group("/roles", middleware.AuthMiddleware())
	roles.Get("/", handlers.GetRoles)
	roles.Post("/", handlers.CreateRole)
	roles.Get("/:id", handlers.GetRoleByID)
	roles.Put("/:id", handlers.UpdateRole)
	roles.Delete("/:id", handlers.DeleteRole)
	roles.Post("/:id/permissions", handlers.AssignPermissionsToRole)

	// ========== PERMISSIONS ==========
	permissions := api.Group("/permissions", middleware.AuthMiddleware())
	permissions.Get("/", handlers.GetPermissions)
	permissions.Get("/search", handlers.SearchPermissions)
	permissions.Get("/:id", handlers.GetPermissionByID)

	// User permissions
	api.Get("/auth/permissions", middleware.AuthMiddleware(), handlers.GetMyPermissions)

	// ========== INVOICES ==========
	invoices := api.Group("/invoices", middleware.AuthMiddleware())
	invoices.Get("/", handlers.GetInvoices)
	invoices.Get("/:id", handlers.GetInvoiceByID)
	invoices.Post("/", handlers.CreateInvoiceFromOrder)
	invoices.Post("/:id/send", handlers.SendInvoiceToDIAN)
	invoices.Get("/:id/xml", handlers.GetInvoiceXML)

	// ========== DASHBOARD ==========
	api.Get("/dashboard/stock-alerts", middleware.AuthMiddleware(), handlers.GetStockAlerts)
}
