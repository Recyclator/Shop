package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/nexora/backend/internal/config"
	"github.com/nexora/backend/internal/database"
	"github.com/nexora/backend/internal/handlers"
	customMiddleware "github.com/nexora/backend/internal/middleware"
	"github.com/nexora/backend/internal/routes"
)

func main() {
	// Cargar configuración
	cfg := config.Load()

	// Inicializar Casbin
	if err := customMiddleware.InitCasbin(cfg.CasbinModelPath, cfg.CasbinPolicyPath); err != nil {
		log.Fatalf("Error inicializando Casbin: %v", err)
	}

	// Conectar a la base de datos
	if err := database.Connect(cfg); err != nil {
		log.Fatalf("Error conectando a la base de datos: %v", err)
	}

	// Conectar a Redis (opcional en desarrollo, log de advertencia si falla)
	if err := database.ConnectRedis(cfg); err != nil {
		log.Printf("⚠️  Advertencia: No se pudo conectar a Redis: %v. Caching y rate limiting distribuido desactivados.", err)
	} else {
		defer database.CloseRedis()
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
			// Log interno del error completo para debugging
			log.Printf("[ERROR] %d: %v", code, err)
			// Respuesta genérica al cliente, nunca exponer detalles internos
			msg := "Error interno del servidor"
			switch code {
			case fiber.StatusBadRequest:
				msg = "Solicitud inv\u00e1lida"
			case fiber.StatusUnauthorized:
				msg = "No autorizado"
			case fiber.StatusForbidden:
				msg = "Acceso prohibido"
			case fiber.StatusNotFound:
				msg = "Recurso no encontrado"
			case fiber.StatusConflict:
				msg = "Conflicto de datos"
			case fiber.StatusTooManyRequests:
				msg = "Demasiadas solicitudes"
			}
			return c.Status(code).JSON(fiber.Map{
				"error": msg,
			})
		},
	})

	// Middlewares globales
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(customMiddleware.SecurityHeadersMiddleware())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.FrontendURL + "," + cfg.AppURL,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
		MaxAge:           86400,
	}))

	// Servir archivos estáticos (CSS, JS, imágenes)
	app.Static("/static", "./static", fiber.Static{
		CacheDuration: 0,
		MaxAge:        0,
	})

	// Servir la página de Login/Registro en la raíz
	app.Get("/", handlers.RenderAuth)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "ok",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"version":   "1.0.0",
		})
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
