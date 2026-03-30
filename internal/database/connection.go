package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/nexora/backend/internal/config"
	"github.com/nexora/backend/internal/models"
)

var DB *gorm.DB

func Connect(cfg *config.Config) error {
	var err error

	// Usar SQLite para desarrollo local
	dbPath := "./nexora.db"

	// Eliminar db existente si hay problemas
	if _, err := os.Stat(dbPath); err == nil {
		log.Println("ℹ️  Usando base de datos existente")
	}

	dbConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC().Add(-5 * time.Hour)
		},
	}

	DB, err = gorm.Open(sqlite.Open(dbPath), dbConfig)
	if err != nil {
		return fmt.Errorf("error al conectar a SQLite: %w", err)
	}

	log.Println("✅ Conexión a SQLite establecida")
	return nil
}

func Migrate() error {
	log.Println("🔄 Ejecutando migraciones...")

	err := DB.AutoMigrate(
		&models.Permission{},
		&models.Role{},
		&models.RolePermission{},
		&models.User{},
		&models.Category{},
		&models.Product{},
		&models.ProductImage{},
		&models.ProductVariant{},
		&models.Tag{},
		&models.Order{},
		&models.OrderItem{},
		&models.Invoice{},
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
