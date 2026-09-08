package handlers

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nexora/backend/internal/database"
	"github.com/nexora/backend/internal/middleware"
	"github.com/nexora/backend/internal/models"
	"github.com/nexora/backend/internal/utils"
)

const maxUploadFileSize = 10 * 1024 * 1024 // 10MB

// validateUploadedImage inspects file size, extension, and magic bytes
func validateUploadedImage(file *multipart.FileHeader) (string, error) {
	if file.Size > maxUploadFileSize {
		return "", fmt.Errorf("el archivo %s excede el tamaño máximo permitido de 10MB", file.Filename)
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".webp": "image/webp",
		".gif":  "image/gif",
	}

	expectedMime, ok := allowedExts[ext]
	if !ok {
		return "", fmt.Errorf("formato de archivo no permitido: %s", ext)
	}

	f, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("no se pudo abrir el archivo %s", file.Filename)
	}
	defer f.Close()

	// Sniff first 512 bytes
	buffer := make([]byte, 512)
	n, err := f.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("error al leer contenido del archivo %s", file.Filename)
	}

	detectedMime := strings.Split(http.DetectContentType(buffer[:n]), ";")[0]
	if detectedMime != expectedMime {
		return "", fmt.Errorf("el contenido del archivo no coincide con su extensión (%s != %s)", detectedMime, expectedMime)
	}

	return ext, nil
}

// UploadProductImages handles multiple image uploads for a product
// POST /api/products/:id/images
func UploadProductImages(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("products.update") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso")
	}

	productID, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de producto inválido")
	}

	var product models.Product
	if result := database.DB.First(&product, productID); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "producto no encontrado")
	}

	// Parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_FORM", "error al procesar formulario")
	}

	files := form.File["images"]
	if len(files) == 0 {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "NO_FILES", "no se enviaron imágenes")
	}

	// Ensure upload directory exists
	uploadDir := "./static/uploads/products"
	os.MkdirAll(uploadDir, os.ModePerm)

	// Check if product already has a principal image
	var existingPrincipal int64
	database.DB.Model(&models.ProductImage{}).Where("producto_id = ? AND es_principal = ?", productID, true).Count(&existingPrincipal)

	// Get current max order
	var maxOrden int
	database.DB.Model(&models.ProductImage{}).Where("producto_id = ?", productID).Select("COALESCE(MAX(orden), -1)").Scan(&maxOrden)

	var savedImages []models.ProductImage

	for _, file := range files {
		// Validate file type and magic bytes
		ext, err := validateUploadedImage(file)
		if err != nil {
			continue // Skip invalid/unsupported/suspicious files
		}

		// Generate cryptographically unique, unpredictable filename
		filename := fmt.Sprintf("%d_%s%s", productID, uuid.New().String(), ext)
		filePath := filepath.Join(uploadDir, filename)

		// Save file
		if err := c.SaveFile(file, filePath); err != nil {
			continue // Skip files that fail to save
		}

		// Determine if this should be the principal image
		isPrincipal := (existingPrincipal == 0 && len(savedImages) == 0)

		image := models.ProductImage{
			ProductoID:  uint(productID),
			URL:         "/static/uploads/products/" + filename,
			Alt:         product.Nombre,
			EsPrincipal: isPrincipal,
			Orden:       maxOrden + 1 + len(savedImages),
		}

		if result := database.DB.Create(&image); result.Error == nil {
			savedImages = append(savedImages, image)
		}

		// Update product's main image if this is the principal
		if isPrincipal {
			database.DB.Model(&product).Update("imagen_principal", image.URL)
		}
	}

	if len(savedImages) == 0 {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "NO_SAVED", "no se pudo guardar ninguna imagen válida")
	}

	database.InvalidateProductCache(uint(productID))

	return utils.Success(c, fiber.StatusCreated, fmt.Sprintf("%d imágenes subidas", len(savedImages)), savedImages)
}

// UploadVariantImages handles multiple image uploads for a product variant
// POST /api/products/:id/variants/:var_id/images
func UploadVariantImages(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("products.update") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso")
	}

	productID, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de producto inválido")
	}

	varID, err := c.ParamsInt("var_id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de variante inválido")
	}

	var variant models.ProductVariant
	if result := database.DB.Where("id = ? AND producto_id = ?", varID, productID).First(&variant); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "variante no encontrada")
	}

	form, err := c.MultipartForm()
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_FORM", "error al procesar formulario")
	}

	files := form.File["images"]
	if len(files) == 0 {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "NO_FILES", "no se enviaron imágenes")
	}

	uploadDir := "./static/uploads/variants"
	os.MkdirAll(uploadDir, os.ModePerm)

	var savedURLs []string
	for _, file := range files {
		ext, err := validateUploadedImage(file)
		if err != nil {
			continue
		}

		filename := fmt.Sprintf("%d_%s%s", varID, uuid.New().String(), ext)
		filePath := filepath.Join(uploadDir, filename)

		if err := c.SaveFile(file, filePath); err != nil {
			continue
		}

		savedURLs = append(savedURLs, "/static/uploads/variants/"+filename)
	}

	if len(savedURLs) == 0 {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "NO_SAVED", "no se pudo guardar ninguna imagen válida")
	}

	existing := variant.Imagen
	var finalURLs []string
	if existing != "" {
		finalURLs = strings.Split(existing, ";")
	}
	finalURLs = append(finalURLs, savedURLs...)
	variant.Imagen = strings.Join(finalURLs, ";")

	database.DB.Model(&variant).Update("imagen", variant.Imagen)

	database.InvalidateProductCache(uint(productID))

	return utils.Success(c, fiber.StatusOK, "imágenes de variante subidas", finalURLs)
}

// GetProductImages returns all images for a product
// GET /api/products/:id/images
func GetProductImages(c *fiber.Ctx) error {
	productID, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de producto inválido")
	}

	var images []models.ProductImage
	if err := database.DB.Where("producto_id = ?", productID).Order("orden ASC").Find(&images).Error; err != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "DB_ERROR", "error al obtener imágenes")
	}

	return utils.SuccessData(c, fiber.StatusOK, images)
}

// DeleteProductImage deletes a single product image
// DELETE /api/products/:id/images/:imgId
func DeleteProductImage(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("products.update") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso")
	}

	productID, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de producto inválido")
	}

	imgID, err := c.ParamsInt("imgId")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de imagen inválido")
	}

	var image models.ProductImage
	if result := database.DB.Where("id = ? AND producto_id = ?", imgID, productID).First(&image); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "imagen no encontrada")
	}

	wasPrincipal := image.EsPrincipal

	// Safely delete the file from disk (prevent path traversal)
	if strings.HasPrefix(image.URL, "/static/uploads/") {
		cleanRel := filepath.Clean("." + image.URL)
		if strings.HasPrefix(cleanRel, "static/uploads/") {
			_ = os.Remove(cleanRel)
		}
	}

	// Delete from DB
	if result := database.DB.Unscoped().Delete(&image); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusInternalServerError, "DELETE_ERROR", "error al eliminar imagen")
	}

	// If it was the principal, promote the next image
	if wasPrincipal {
		var nextImage models.ProductImage
		if result := database.DB.Where("producto_id = ?", productID).Order("orden ASC").First(&nextImage); result.Error == nil {
			database.DB.Model(&nextImage).Update("es_principal", true)
			database.DB.Model(&models.Product{}).Where("id = ?", productID).Update("imagen_principal", nextImage.URL)
		} else {
			// No more images
			database.DB.Model(&models.Product{}).Where("id = ?", productID).Update("imagen_principal", "")
		}
	}

	database.InvalidateProductCache(uint(productID))

	return utils.SuccessMessage(c, fiber.StatusOK, "imagen eliminada")
}

// SetPrincipalImage sets a specific image as the product's principal image
// PUT /api/products/:id/images/:imgId/principal
func SetPrincipalImage(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("products.update") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso")
	}

	productID, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de producto inválido")
	}

	imgID, err := c.ParamsInt("imgId")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de imagen inválido")
	}

	var image models.ProductImage
	if result := database.DB.Where("id = ? AND producto_id = ?", imgID, productID).First(&image); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "imagen no encontrada")
	}

	// Remove current principal
	database.DB.Model(&models.ProductImage{}).Where("producto_id = ?", productID).Update("es_principal", false)

	// Set new principal
	database.DB.Model(&image).Update("es_principal", true)
	database.DB.Model(&models.Product{}).Where("id = ?", productID).Update("imagen_principal", image.URL)

	database.InvalidateProductCache(uint(productID))

	return utils.SuccessMessage(c, fiber.StatusOK, "imagen principal actualizada")
}

// ReorderProductImages reorders images for a product
// PUT /api/products/:id/images/reorder
func ReorderProductImages(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	productID, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de producto inválido")
	}

	var input struct {
		ImageIDs []uint `json:"image_ids"`
	}
	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_BODY", "formato inválido")
	}

	for i, imgID := range input.ImageIDs {
		database.DB.Model(&models.ProductImage{}).Where("id = ? AND producto_id = ?", imgID, productID).Update("orden", i)
	}

	database.InvalidateProductCache(uint(productID))

	return utils.SuccessMessage(c, fiber.StatusOK, "orden actualizado")
}

// GenerateVariants auto-generates variants from attribute combinations
// POST /api/products/:id/variants/generate
func GenerateVariants(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("products.create") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso")
	}

	productID, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de producto inválido")
	}

	var product models.Product
	if result := database.DB.First(&product, productID); result.Error != nil {
		return utils.ErrorWithCode(c, fiber.StatusNotFound, "NOT_FOUND", "producto no encontrado")
	}

	type GenerateInput struct {
		Attributes []struct {
			AttributeID uint     `json:"attribute_id"`
			Values      []string `json:"values"`
		} `json:"attributes"`
	}

	var input GenerateInput
	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_BODY", "formato inválido")
	}

	if len(input.Attributes) == 0 {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "NO_ATTRIBUTES", "debe especificar al menos un atributo")
	}

	// Load attribute names
	attrMap := make(map[uint]string)
	var attrIDs []uint
	for _, a := range input.Attributes {
		attrIDs = append(attrIDs, a.AttributeID)
	}
	var attrs []models.ProductAttribute
	database.DB.Where("id IN ?", attrIDs).Find(&attrs)
	for _, a := range attrs {
		attrMap[a.ID] = a.Nombre
	}

	// Generate cartesian product of all attribute values
	type AttrVal struct {
		AttributeID uint
		Value       string
	}

	var combinations [][]AttrVal
	combinations = append(combinations, []AttrVal{})

	for _, attr := range input.Attributes {
		var newCombinations [][]AttrVal
		for _, combo := range combinations {
			for _, val := range attr.Values {
				newCombo := make([]AttrVal, len(combo))
				copy(newCombo, combo)
				newCombo = append(newCombo, AttrVal{AttributeID: attr.AttributeID, Value: val})
				newCombinations = append(newCombinations, newCombo)
			}
		}
		combinations = newCombinations
	}

	var createdVariants []models.VariantResponse
	variantCount := 0

	for _, combo := range combinations {
		// Build SKU and name
		skuParts := []string{product.SKU}
		nameParts := []string{}
		for _, av := range combo {
			skuParts = append(skuParts, av.Value)
			attrName := attrMap[av.AttributeID]
			nameParts = append(nameParts, fmt.Sprintf("%s: %s", attrName, av.Value))
		}

		// Build SKU - replace spaces with dashes and uppercase
		sku := ""
		for i, part := range skuParts {
			clean := ""
			for _, r := range part {
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
					clean += string(r)
				}
			}
			if i > 0 {
				sku += "-"
			}
			sku += clean
		}

		// Check if variant with same SKU already exists
		var existing models.ProductVariant
		if result := database.DB.Where("sku = ?", sku).First(&existing); result.RowsAffected > 0 {
			continue // Skip duplicates
		}

		variant := models.ProductVariant{
			ProductoID:  product.ID,
			SKU:         sku,
			Nombre:      product.Nombre + " (" + joinStrings(nameParts, ", ") + ")",
			Stock:       0,
			StockMinimo: product.StockMinimo,
			Activa:      true,
			Orden:       variantCount,
		}

		if result := database.DB.Create(&variant); result.Error != nil {
			continue
		}

		// Create attribute values for this variant
		for _, av := range combo {
			attrValue := models.VariantAttributeValue{
				VariantID:   variant.ID,
				AttributeID: av.AttributeID,
				Value:       av.Value,
			}
			database.DB.Create(&attrValue)
		}

		database.DB.Preload("Atributos").First(&variant, variant.ID)
		createdVariants = append(createdVariants, variant.ToResponse(product.Precio))
		variantCount++
	}

	// Update product to indicate it has variants
	if variantCount > 0 {
		database.DB.Model(&product).Update("tiene_variantes", true)
	}

	return utils.Success(c, fiber.StatusCreated, fmt.Sprintf("%d variantes generadas", variantCount), createdVariants)
}

// BulkUpdateVariants updates multiple variants at once
// PUT /api/products/:id/variants/bulk
func BulkUpdateVariants(c *fiber.Ctx) error {
	currentUser := middleware.GetUser(c)
	if currentUser == nil {
		return utils.ErrorWithCode(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "no autorizado")
	}

	if !currentUser.HasPermission("products.update") {
		return utils.ErrorWithCode(c, fiber.StatusForbidden, "FORBIDDEN", "no tienes permiso")
	}

	productID, err := c.ParamsInt("id")
	if err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_ID", "ID de producto inválido")
	}

	type BulkVariantUpdate struct {
		ID             uint     `json:"id"`
		SKU            string   `json:"sku"`
		Barcode        string   `json:"barcode"`
		PrecioOverride *float64 `json:"precio_override"`
		Stock          *int     `json:"stock"`
		Imagen         *string  `json:"imagen"`
		Activa         *bool    `json:"activa"`
	}

	var input struct {
		Variants []BulkVariantUpdate `json:"variants"`
	}
	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorWithCode(c, fiber.StatusBadRequest, "INVALID_BODY", "formato inválido")
	}

	updated := 0
	for _, v := range input.Variants {
		updates := make(map[string]interface{})
		if v.SKU != "" {
			updates["sku"] = v.SKU
		}
		if v.Barcode != "" {
			updates["barcode"] = v.Barcode
		}
		if v.PrecioOverride != nil {
			updates["precio_override"] = *v.PrecioOverride
		}
		if v.Stock != nil {
			updates["stock"] = *v.Stock
		}
		if v.Imagen != nil {
			updates["imagen"] = *v.Imagen
		}
		if v.Activa != nil {
			updates["activa"] = *v.Activa
		}
		if len(updates) > 0 {
			result := database.DB.Model(&models.ProductVariant{}).Where("id = ? AND producto_id = ?", v.ID, productID).Updates(updates)
			if result.RowsAffected > 0 {
				updated++
			}
		}
	}

	return utils.Success(c, fiber.StatusOK, fmt.Sprintf("%d variantes actualizadas", updated), nil)
}

func joinStrings(parts []string, sep string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += sep
		}
		result += p
	}
	return result
}
