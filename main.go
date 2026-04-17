package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/nexora/backend/internal/config"
	"github.com/nexora/backend/internal/database"
	"github.com/nexora/backend/internal/handlers"
	"github.com/nexora/backend/internal/routes"
)

func main() {
	// Cargar configuración
	cfg := config.Load()

	// Conectar a la base de datos
	if err := database.Connect(cfg); err != nil {
		log.Fatalf("Error conectando a la base de datos: %v", err)
	}

	// Ejecutar migraciones
	if err := database.Migrate(); err != nil {
		log.Fatalf("Error en migraciones: %v", err)
	}

	// Cargar datos iniciales
	if err := database.SeedData(); err != nil {
		log.Fatalf("Error cargando datos iniciales: %v", err)
	}

	// Crear app Fiber
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middlewares globales
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: false,
	}))

	// Servir archivos estáticos (CSS, JS, imágenes)
	app.Static("/static", "./static")
	app.Static("/public", "./public")

	// Servir la página de Login/Registro en la raíz
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendFile("./public/index.html")
	})

	// Rutas que renderizan páginas con Templ + HTMX
	app.Get("/dashboard", handlers.RenderDashboard)
	app.Get("/products", handlers.RenderProducts)
	app.Get("/categories", handlers.RenderCategories)
	app.Get("/customers", handlers.RenderCustomers)
	app.Get("/layaways", handlers.RenderLayaways)
	app.Get("/users", handlers.RenderUsers)
	app.Get("/pos", handlers.RenderPOS)
	app.Get("/profile", handlers.RenderProfile)
	app.Get("/settings", handlers.RenderSettings)
	app.Get("/roles-permissions", handlers.RenderRolesPermissions)

	// Configurar rutas API
	routes.Setup(app)

	// Iniciar servidor
	port := cfg.ServerPort
	if port == "" {
		port = "3000"
	}

	log.Printf("🚀 Servidor iniciado en http://localhost:%s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Error iniciando servidor: %v", err)
	}
}
