package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/nexora/backend/internal/config"
	"github.com/nexora/backend/internal/models"
	"github.com/nexora/backend/internal/utils"
)

var DB *gorm.DB

func Connect(cfg *config.Config) error {
	var err error

	dsn := cfg.GetDSN()

	dbConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC().Add(-5 * time.Hour)
		},
	}

	DB, err = gorm.Open(postgres.Open(dsn), dbConfig)
	if err != nil {
		return fmt.Errorf("error al conectar a PostgreSQL: %w", err)
	}

	log.Println("✅ Conexión a PostgreSQL establecida")
	return nil
}

func Migrate() error {
	log.Println("🔄 Ejecutando migraciones...")

	err := DB.AutoMigrate(
		&models.Permission{},
		&models.Role{},
		&models.RolePermission{},
		&models.User{},
		&models.Customer{},
		&models.Category{},
		&models.Product{},
		&models.ProductImage{},
		&models.ProductVariant{},
		&models.VariantAttributeValue{},
		&models.ProductAttribute{},
		&models.StockMovement{},
		&models.Tag{},
		&models.Order{},
		&models.OrderItem{},
		&models.Invoice{},
		&models.Layaway{},
		&models.Payment{},
	)

	if err != nil {
		return fmt.Errorf("error en migraciones: %w", err)
	}

	log.Println("✅ Migraciones completadas")
	return nil
}

func SeedData() error {
	log.Println("🌱 Verificando datos iniciales...")

	var count int64
	DB.Model(&models.Role{}).Count(&count)
	if count > 0 {
		log.Println("ℹ️  Los datos ya existen, omitiendo seed")
		return nil
	}

	permissions := models.GetDefaultPermissions()
	if err := DB.Create(&permissions).Error; err != nil {
		return fmt.Errorf("error al crear permisos: %w", err)
	}

	roles := models.GetDefaultRoles()
	if err := DB.Create(&roles).Error; err != nil {
		return fmt.Errorf("error al crear roles: %w", err)
	}

	if err := models.AssignDefaultPermissions(DB); err != nil {
		return fmt.Errorf("error al asignar permisos: %w", err)
	}

	// Crear usuario superadmin de prueba
	var userCount int64
	DB.Model(&models.User{}).Count(&userCount)
	if userCount == 0 {
		hashedPassword, err := utils.HashPassword("superadmin123")
		if err != nil {
			return fmt.Errorf("error al hash contraseña: %w", err)
		}

		var superadminRole models.Role
		if err := DB.Where("nombre = ?", "superadmin").First(&superadminRole).Error; err != nil {
			return fmt.Errorf("error al buscar rol superadmin: %w", err)
		}

		superadmin := models.User{
			Email:           "superadmin@nexora.com",
			Password:        hashedPassword,
			Nombre:          "Super",
			Apellido:        "Admin",
			Telefono:        "3001234567",
			TipoDocumento:   "CC",
			NumeroDocumento: "1234567890",
			Activo:          true,
			Roles:           []models.Role{superadminRole},
		}

		if err := DB.Create(&superadmin).Error; err != nil {
			return fmt.Errorf("error al crear usuario superadmin: %w", err)
		}

		log.Println("✅ Usuario superadmin creado: superadmin@nexora.com / superadmin123")
	}

	log.Println("✅ Datos iniciales creados")
	return nil
}

func Close() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
