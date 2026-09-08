package services

import (
	"fmt"
	"log"
	"time"

	"github.com/nexora/backend/internal/database"
	"github.com/nexora/backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ProcessExpiredLayaways marks expired layaways as "vencido" and restores their reserved stock
func ProcessExpiredLayaways() (int64, error) {
	now := time.Now()

	tx := database.DB.Begin()
	var expiredLayaways []models.Layaway
	if err := tx.Where("estado = ? AND fecha_vencimiento < ?", "activo", now).Find(&expiredLayaways).Error; err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("error al buscar separados vencidos: %w", err)
	}

	for _, lay := range expiredLayaways {
		var product models.Product
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&product, lay.ProductID).Error; err == nil && product.ControlaInventario {
			_ = tx.Model(&product).Update("stock", gorm.Expr("stock + ?", lay.Cantidad)).Error
		}
	}

	result := tx.Model(&models.Layaway{}).
		Where("estado = ? AND fecha_vencimiento < ?", "activo", now).
		Update("estado", "vencido")

	if result.Error != nil {
		tx.Rollback()
		return 0, fmt.Errorf("error al actualizar estado de separados vencidos: %w", result.Error)
	}

	if err := tx.Commit().Error; err != nil {
		return 0, fmt.Errorf("error al confirmar transacción de separados vencidos: %w", err)
	}

	return result.RowsAffected, nil
}

// StartLayawayExpirationWorker starts a periodic background worker to process expired layaways
func StartLayawayExpirationWorker(interval time.Duration) {
	go func() {
		// Run once on startup after 5 seconds to let the application finish booting
		time.Sleep(5 * time.Second)
		if count, err := ProcessExpiredLayaways(); err != nil {
			log.Printf("⚠️ [WORKER] Error inicial en verificación de separados vencidos: %v", err)
		} else if count > 0 {
			log.Printf("ℹ️ [WORKER] %d separados vencidos procesados y stock restaurado en arranque", count)
		}

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			if count, err := ProcessExpiredLayaways(); err != nil {
				log.Printf("⚠️ [WORKER] Error en verificación periódica de separados vencidos: %v", err)
			} else if count > 0 {
				log.Printf("ℹ️ [WORKER] %d separados vencidos procesados y stock restaurado", count)
			}
		}
	}()
}
