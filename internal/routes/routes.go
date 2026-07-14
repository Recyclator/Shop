package routes

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/nexora/backend/internal/handlers"
	"github.com/nexora/backend/internal/middleware"
)

func Setup(app *fiber.App) {
	// Grupo API
	api := app.Group("/api")

	// ========== AUTH ==========
	auth := api.Group("/auth")
	auth.Post("/register", middleware.RateLimiterMiddleware(5, time.Hour), handlers.Register)
	auth.Post("/login", middleware.RateLimiterMiddleware(10, time.Minute), handlers.Login)
	auth.Post("/refresh", middleware.RateLimiterMiddleware(20, time.Minute), handlers.RefreshToken)
	auth.Post("/logout", handlers.Logout)
	auth.Post("/change-password", middleware.AuthMiddleware(), handlers.ChangePassword)
	auth.Get("/me", middleware.AuthMiddleware(), handlers.GetCurrentUser)
	auth.Get("/profile", middleware.AuthMiddleware(), handlers.GetFullProfile)
	auth.Put("/profile", middleware.AuthMiddleware(), handlers.UpdateProfile)
	auth.Put("/theme", middleware.AuthMiddleware(), handlers.UpdateTheme)

	// ========== USERS ==========
	users := api.Group("/users", middleware.AuthMiddleware())
	users.Get("/", middleware.RequirePermissionAny("users.read", "users.list"), handlers.GetUsers)
	users.Get("/search", middleware.RequirePermissionAny("users.read", "users.search"), handlers.SearchUsers)
	users.Post("/", middleware.RequirePermissionAny("users.create"), handlers.CreateUser)
	users.Get("/:id", middleware.RequirePermissionAny("users.read"), handlers.GetUserByID)
	users.Put("/:id", middleware.RequirePermissionAny("users.update"), handlers.UpdateUser)
	users.Delete("/:id", middleware.RequirePermissionAny("users.delete"), handlers.DeleteUser)
	users.Post("/:id/roles", middleware.RequirePermissionAny("roles.assign"), handlers.AssignRoles)

	// ========== PRODUCTS ==========
	products := api.Group("/products", middleware.AuthMiddleware())
	products.Get("/", middleware.RequirePermissionAny("products.read"), handlers.GetProducts)
	products.Get("/featured", handlers.GetFeaturedProducts) // público
	products.Get("/search", handlers.SearchProducts) // público
	products.Get("/category/:id", handlers.GetProductsByCategory) // público
	products.Get("/sku/:sku", handlers.GetProductBySKU) // público
	products.Get("/barcode/:code", middleware.AuthMiddleware(), middleware.RequirePermissionAny("products.read"), handlers.GetProductByBarcode)
	products.Post("/", middleware.RequirePermissionAny("products.create"), handlers.CreateProduct)
	products.Get("/:id", handlers.GetProductByID) // público
	products.Put("/:id", middleware.RequirePermissionAny("products.update"), handlers.UpdateProduct)
	products.Delete("/:id", middleware.RequirePermissionAny("products.delete"), handlers.DeleteProduct)
	products.Post("/:id/qr", middleware.RequirePermissionAny("products.read"), handlers.GenerateProductQR)
	products.Post("/:id/stock/adjust", middleware.RequirePermissionAny("products.update"), handlers.AdjustProductStock)

	// ========== PRODUCT VARIANTS ==========
	products.Get("/:id/variants", middleware.RequirePermissionAny("products.read"), handlers.GetVariantsByProduct)
	products.Post("/:id/variants", middleware.RequirePermissionAny("products.create"), handlers.CreateVariant)
	products.Put("/:id/variants/:var_id", middleware.RequirePermissionAny("products.update"), handlers.UpdateVariant)
	products.Delete("/:id/variants/:var_id", middleware.RequirePermissionAny("products.delete"), handlers.DeleteVariant)
	products.Post("/:id/variants/generate", middleware.RequirePermissionAny("products.create"), handlers.GenerateVariants)
	products.Put("/:id/variants/bulk", middleware.RequirePermissionAny("products.update"), handlers.BulkUpdateVariants)
	products.Post("/:id/variants/:var_id/images", middleware.RequirePermissionAny("products.update"), handlers.UploadVariantImages)

	// ========== PRODUCT IMAGES ==========
	products.Get("/:id/images", middleware.RequirePermissionAny("products.read"), handlers.GetProductImages)
	products.Post("/:id/images", middleware.RequirePermissionAny("products.update"), handlers.UploadProductImages)
	products.Delete("/:id/images/:imgId", middleware.RequirePermissionAny("products.update"), handlers.DeleteProductImage)
	products.Put("/:id/images/:imgId/principal", middleware.RequirePermissionAny("products.update"), handlers.SetPrincipalImage)
	products.Put("/:id/images/reorder", middleware.RequirePermissionAny("products.update"), handlers.ReorderProductImages)

	// ========== CATEGORIES ==========
	categories := api.Group("/categories")
	categories.Get("/", handlers.GetCategories) // público
	categories.Get("/:id", handlers.GetCategoryByID) // público
	categories.Post("/", middleware.AuthMiddleware(), middleware.RequirePermissionAny("categories.create"), handlers.CreateCategory)
	categories.Put("/:id", middleware.AuthMiddleware(), middleware.RequirePermissionAny("categories.update"), handlers.UpdateCategory)
	categories.Delete("/:id", middleware.AuthMiddleware(), middleware.RequirePermissionAny("categories.delete"), handlers.DeleteCategory)

	// ========== ATTRIBUTES ==========
	attributes := api.Group("/attributes")
	attributes.Get("/global", handlers.GetGlobalAttributes) // público
	attributes.Get("/category/:id", handlers.GetAttributesByCategory) // público
	attributes.Post("/category/:id", middleware.AuthMiddleware(), middleware.RequirePermissionAny("attributes.create"), handlers.CreateAttribute)
	attributes.Put("/:id", middleware.AuthMiddleware(), middleware.RequirePermissionAny("attributes.update"), handlers.UpdateAttribute)
	attributes.Delete("/:id", middleware.AuthMiddleware(), middleware.RequirePermissionAny("attributes.delete"), handlers.DeleteAttribute)

	// ========== ORDERS ==========
	orders := api.Group("/orders", middleware.AuthMiddleware())
	orders.Get("/", middleware.RequirePermissionAny("orders.read"), handlers.GetOrders)
	orders.Post("/", middleware.RequirePermissionAny("orders.create"), handlers.CreateOrder)
	orders.Get("/:id", middleware.RequirePermissionAny("orders.read"), handlers.GetOrderByID)
	orders.Put("/:id", middleware.RequirePermissionAny("orders.update"), handlers.UpdateOrder)
	orders.Post("/:id/cancel", middleware.RequirePermissionAny("orders.cancel"), handlers.CancelOrder)
	orders.Post("/pos", middleware.RequirePermissionAny("orders.create"), handlers.QuickPOSSale)

	// ========== POS ==========
	pos := api.Group("/pos", middleware.AuthMiddleware(), middleware.RequirePermissionAny("pos.use"))
	pos.Get("/search", handlers.POSSearch)
	pos.Get("/products", handlers.POSSearchSimple)

	// ========== CUSTOMERS ==========
	customers := api.Group("/customers", middleware.AuthMiddleware())
	customers.Get("/", middleware.RequirePermissionAny("customers.read"), handlers.GetCustomers)
	customers.Get("/cedula/:cedula", middleware.RequirePermissionAny("customers.read"), handlers.GetCustomerByCedula)
	customers.Post("/", middleware.RequirePermissionAny("customers.create"), handlers.CreateCustomer)
	customers.Get("/:id", middleware.RequirePermissionAny("customers.read"), handlers.GetCustomerByID)
	customers.Put("/:id", middleware.RequirePermissionAny("customers.update"), handlers.UpdateCustomer)
	customers.Delete("/:id", middleware.RequirePermissionAny("customers.delete"), handlers.DeleteCustomer)

	// ========== LAYAWAYS (SEPARADOS) ==========
	layaways := api.Group("/layaways", middleware.AuthMiddleware())
	layaways.Get("/", middleware.RequirePermissionAny("layaways.read"), handlers.GetLayaways)
	layaways.Post("/", middleware.RequirePermissionAny("layaways.create"), handlers.CreateLayaway)
	layaways.Get("/:id", middleware.RequirePermissionAny("layaways.read"), handlers.GetLayawayByID)
	layaways.Put("/:id", middleware.RequirePermissionAny("layaways.update"), handlers.UpdateLayawayStatus)
	layaways.Post("/:id/cancel", middleware.RequirePermissionAny("layaways.cancel"), handlers.CancelLayaway)
	layaways.Post("/:id/payments", middleware.RequirePermissionAny("layaways.payment"), handlers.AddPayment)
	layaways.Get("/:id/payments", middleware.RequirePermissionAny("layaways.read"), handlers.GetLayawayPayments)
	layaways.Post("/check-expired", middleware.RequirePermissionAny("layaways.read"), handlers.ExpiredLayaways)
	layaways.Delete("/:id", middleware.RequirePermissionAny("layaways.delete"), handlers.DeleteLayaway)

	// ========== SYSTEM ==========
	system := api.Group("/system", middleware.AuthMiddleware())
	system.Post("/refresh-permissions", middleware.RequirePermissionAny("system.admin"), handlers.RefreshPermissions)

	// ========== ROLES ==========
	roles := api.Group("/roles", middleware.AuthMiddleware())
	roles.Get("/", middleware.RequirePermissionAny("roles.read"), handlers.GetRoles)
	roles.Post("/", middleware.RequirePermissionAny("roles.create"), handlers.CreateRole)
	roles.Get("/:id", middleware.RequirePermissionAny("roles.read"), handlers.GetRoleByID)
	roles.Put("/:id", middleware.RequirePermissionAny("roles.update"), handlers.UpdateRole)
	roles.Delete("/:id", middleware.RequirePermissionAny("roles.delete"), handlers.DeleteRole)
	roles.Post("/:id/permissions", middleware.RequirePermissionAny("roles.assign"), handlers.AssignPermissionsToRole)
	// ========== PERMISSIONS ==========
	permissions := api.Group("/permissions", middleware.AuthMiddleware())
	permissions.Get("/", middleware.RequirePermissionAny("permissions.read"), handlers.GetPermissions)
	permissions.Get("/search", middleware.RequirePermissionAny("permissions.read"), handlers.SearchPermissions)
	permissions.Get("/:id", middleware.RequirePermissionAny("permissions.read"), handlers.GetPermissionByID)

	// User permissions
	api.Get("/auth/permissions", middleware.AuthMiddleware(), handlers.GetMyPermissions)

	// ========== INVOICES ==========
	invoices := api.Group("/invoices", middleware.AuthMiddleware())
	invoices.Get("/", middleware.RequirePermissionAny("invoices.read"), handlers.GetInvoices)
	invoices.Get("/:id", middleware.RequirePermissionAny("invoices.read"), handlers.GetInvoiceByID)
	invoices.Post("/", middleware.RequirePermissionAny("invoices.create"), handlers.CreateInvoiceFromOrder)
	invoices.Post("/:id/send", middleware.RequirePermissionAny("invoices.send"), handlers.SendInvoiceToDIAN)
	invoices.Get("/:id/xml", middleware.RequirePermissionAny("invoices.read"), handlers.GetInvoiceXML)

	// ========== DASHBOARD ==========
	api.Get("/dashboard/stock-alerts", middleware.AuthMiddleware(), handlers.GetStockAlerts)
}
