package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/nexora/backend/internal/config"
	"github.com/nexora/backend/internal/database"
	"github.com/nexora/backend/internal/models"
)

type DIANService struct {
	cfg *config.Config
}

func NewDIANService() *DIANService {
	return &DIANService{
		cfg: config.Load(),
	}
}

func (s *DIANService) GenerateInvoiceXML(invoice *models.Invoice) (string, error) {
	var sb strings.Builder

	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	sb.WriteString("\n")
	sb.WriteString(`<Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2"`)
	sb.WriteString("\n")
	sb.WriteString(`  xmlns:cac="urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2"`)
	sb.WriteString("\n")
	sb.WriteString(`  xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2"`)
	sb.WriteString("\n")
	sb.WriteString(`  xmlns:ds="http://www.w3.org/2000/09/xmldsig#"`)
	sb.WriteString("\n")
	sb.WriteString(`  xmlns:ext="urn:oasis:names:specification:ubl:schema:xsd:CommonExtensionComponents-2"`)
	sb.WriteString("\n")
	sb.WriteString(`  xmlns:dian="dian:gov:co:facturaelectronica:Structures-2-1">`)
	sb.WriteString("\n")

	sb.WriteString("  <cbc:UBLVersionID>2.1</cbc:UBLVersionID>\n")
	sb.WriteString("  <cbc:CustomizationID>urn:cenit:specification:ubl:invoice:CII:2.1</cbc:CustomizationID>\n")
	sb.WriteString(fmt.Sprintf("  <cbc:ProfileID>DIAN 2.1</cbc:ProfileID>\n"))
	sb.WriteString(fmt.Sprintf("  <cbc:ProfileExecutionID>%s</cbc:ProfileExecutionID>\n", s.cfg.Env))
	sb.WriteString(fmt.Sprintf("  <cbc:ID>%s</cbc:ID>\n", invoice.NumeroCompleto))
	sb.WriteString(fmt.Sprintf("  <cbc:IssueDate>%s</cbc:IssueDate>\n", invoice.FechaEmision.Format("2006-01-02")))
	sb.WriteString(fmt.Sprintf("  <cbc:IssueTime>%s</cbc:IssueTime>\n", invoice.HoraEmision))
	sb.WriteString(fmt.Sprintf("  <cbc:InvoiceTypeCode>01</cbc:InvoiceTypeCode>\n"))
	sb.WriteString(fmt.Sprintf("  <cbc:DocumentCurrencyCode>COP</cbc:DocumentCurrencyCode>\n"))
	sb.WriteString(fmt.Sprintf("  <cbc:LineCountNumeric>%d</cbc:LineCountNumeric>\n", len(invoice.Orden.Items)))

	sb.WriteString("  <cac:AccountingSupplierParty>\n")
	sb.WriteString("    <cac:Party>\n")
	sb.WriteString(fmt.Sprintf("      <cac:PartyName><cbc:Name>%s</cbc:Name></cac:PartyName>\n", s.cfg.CompanyName))
	sb.WriteString("      <cac:PhysicalLocation>\n")
	sb.WriteString("        <cac:Address>\n")
	sb.WriteString(fmt.Sprintf("          <cbc:ID>%s</cbc:ID>\n", s.cfg.CompanyCityCode))
	sb.WriteString(fmt.Sprintf("          <cbc:CityName>%s</cbc:CityName>\n", s.cfg.CompanyCity))
	sb.WriteString(fmt.Sprintf("          <cbc:CountrySubentity>%s</cbc:CountrySubentity>\n", s.cfg.CompanyDepartment))
	sb.WriteString("          <cac:AddressLine><cbc:Line></cbc:Line></cac:AddressLine>\n")
	sb.WriteString("          <cac:Country>\n")
	sb.WriteString("            <cbc:IdentificationCode>CO</cbc:IdentificationCode>\n")
	sb.WriteString("            <cbc:Name>Colombia</cbc:Name>\n")
	sb.WriteString("          </cac:Country>\n")
	sb.WriteString("        </cac:Address>\n")
	sb.WriteString("      </cac:PhysicalLocation>\n")
	sb.WriteString("      <cac:PartyTaxScheme>\n")
	sb.WriteString(fmt.Sprintf("        <cbc:RegistrationName>%s</cbc:RegistrationName>\n", s.cfg.CompanyName))
	sb.WriteString(fmt.Sprintf("        <cbc:CompanyID>%s</cbc:CompanyID>\n", s.cfg.CompanyNIT))
	sb.WriteString("        <cbc:TaxLevelCode>04</cbc:TaxLevelCode>\n")
	sb.WriteString("        <cac:Country><cbc:IdentificationCode>CO</cbc:IdentificationCode></cac:Country>\n")
	sb.WriteString("      </cac:PartyTaxScheme>\n")
	sb.WriteString(fmt.Sprintf("      <cac:Contact><cbc:Telephone>%s</cbc:Telephone><cbc:ElectronicMail>%s</cbc:ElectronicMail></cac:Contact>\n", s.cfg.CompanyPhone, s.cfg.CompanyEmail))
	sb.WriteString("    </cac:Party>\n")
	sb.WriteString("  </cac:AccountingSupplierParty>\n")

	sb.WriteString("  <cac:AccountingCustomerParty>\n")
	sb.WriteString("    <cac:Party>\n")
	sb.WriteString(fmt.Sprintf("      <cac:PartyName><cbc:Name>%s</cbc:Name></cac:PartyName>\n", invoice.ClienteNombre))
	sb.WriteString("      <cac:PartyTaxScheme>\n")
	sb.WriteString(fmt.Sprintf("        <cbc:RegistrationName>%s</cbc:RegistrationName>\n", invoice.ClienteNombre))
	sb.WriteString(fmt.Sprintf("        <cbc:CompanyID>%s</cbc:CompanyID>\n", invoice.ClienteNIT))
	sb.WriteString("        <cbc:TaxLevelCode>ZZ</cbc:TaxLevelCode>\n")
	sb.WriteString("        <cac:Country><cbc:IdentificationCode>CO</cbc:IdentificationCode></cac:Country>\n")
	sb.WriteString("      </cac:PartyTaxScheme>\n")
	sb.WriteString("    </cac:Party>\n")
	sb.WriteString("  </cac:AccountingCustomerParty>\n")

	sb.WriteString("  <cac:TaxTotal>\n")
	sb.WriteString(fmt.Sprintf("    <cbc:TaxAmount currencyID=\"COP\">%.2f</cbc:TaxAmount>\n", invoice.ImpuestoMonto))
	sb.WriteString("    <cbc:TaxEvidenceIndicator>true</cbc:TaxEvidenceIndicator>\n")
	sb.WriteString("    <cac:TaxSubtotal>\n")
	sb.WriteString(fmt.Sprintf("      <cbc:TaxableAmount currencyID=\"COP\">%.2f</cbc:TaxableAmount>\n", invoice.BaseImponible))
	sb.WriteString(fmt.Sprintf("      <cbc:TaxAmount currencyID=\"COP\">%.2f</cbc:TaxAmount>\n", invoice.ImpuestoMonto))
	sb.WriteString(fmt.Sprintf("      <cbc:Percent>%.2f</cbc:Percent>\n", invoice.ImpuestoPorcentaje))
	sb.WriteString("      <cac:TaxCategory><cbc:ID>S</cbc:ID></cac:TaxCategory>\n")
	sb.WriteString("    </cac:TaxSubtotal>\n")
	sb.WriteString("  </cac:TaxTotal>\n")

	sb.WriteString("  <cac:LegalMonetaryTotal>\n")
	sb.WriteString(fmt.Sprintf("    <cbc:LineExtensionAmount currencyID=\"COP\">%.2f</cbc:LineExtensionAmount>\n", invoice.Subtotal))
	sb.WriteString(fmt.Sprintf("    <cbc:TaxExclusiveAmount currencyID=\"COP\">%.2f</cbc:TaxExclusiveAmount>\n", invoice.BaseImponible))
	sb.WriteString(fmt.Sprintf("    <cbc:TaxInclusiveAmount currencyID=\"COP\">%.2f</cbc:TaxInclusiveAmount>\n", invoice.Total))
	sb.WriteString(fmt.Sprintf("    <cbc:PayableAmount currencyID=\"COP\">%.2f</cbc:PayableAmount>\n", invoice.Total))
	sb.WriteString("  </cac:LegalMonetaryTotal>\n")

	for i, item := range invoice.Orden.Items {
		sb.WriteString(fmt.Sprintf("  <cac:InvoiceLine>\n"))
		sb.WriteString(fmt.Sprintf("    <cbc:ID>%d</cbc:ID>\n", i+1))
		sb.WriteString(fmt.Sprintf("    <cbc:InvoicedQuantity unitCode=\"EA\">%d</cbc:InvoicedQuantity>\n", item.Cantidad))
		sb.WriteString(fmt.Sprintf("    <cbc:LineExtensionAmount currencyID=\"COP\">%.2f</cbc:LineExtensionAmount>\n", item.Total))
		sb.WriteString("    <cac:Item>\n")
		sb.WriteString(fmt.Sprintf("      <cbc:Description>%s</cbc:Description>\n", item.NombreProducto))
		sb.WriteString("    </cac:Item>\n")
		sb.WriteString("    <cac:Price>\n")
		sb.WriteString(fmt.Sprintf("      <cbc:PriceAmount currencyID=\"COP\">%.2f</cbc:PriceAmount>\n", item.PrecioUnitario))
		sb.WriteString("      <cbc:BaseQuantity unitCode=\"EA\">1</cbc:BaseQuantity>\n")
		sb.WriteString("    </cac:Price>\n")
		sb.WriteString("  </cac:InvoiceLine>\n")
	}

	sb.WriteString("</Invoice>\n")

	return sb.String(), nil
}

func (s *DIANService) GenerateUUID() string {
	data := make([]byte, 16)
	rand.Read(data)
	hash := sha256.Sum256(data)
	return base64.StdEncoding.EncodeToString(hash[:])[:36]
}

func (s *DIANService) GenerateInvoiceNumber(prefix string, numero uint) string {
	return fmt.Sprintf("%s%010d", prefix, numero)
}

func (s *DIANService) CreateInvoiceFromOrder(order *models.Order) (*models.Invoice, error) {
	var count int64
	database.DB.Model(&models.Invoice{}).Count(&count)

	now := time.Now()
	invoice := &models.Invoice{
		OrdenID:            order.ID,
		Prefijo:            "FV",
		Numero:             uint(count) + 1,
		NumeroCompleto:     s.GenerateInvoiceNumber("FV", uint(count)+1),
		FechaEmision:       now,
		HoraEmision:        now.Format("15:04:05"),
		Estado:             "borrador",
		EstadoDIAN:         "no_enviada",
		UUID:               s.GenerateUUID(),
		Subtotal:           order.Subtotal,
		DescuentoMonto:     order.Descuento,
		BaseImponible:      order.Subtotal - order.Descuento,
		ImpuestoMonto:      order.Impuesto,
		ImpuestoPorcentaje: order.ImpuestoPorcentaje,
		CostoEnvio:         order.CostoEnvio,
		Total:              order.Total,
		ClienteNIT:         order.ClienteDocumento,
		ClienteNombre:      order.ClienteNombre,
		ClienteDireccion:   order.DireccionEnvio,
		ClienteCiudad:      order.CiudadEnvio,
		ClienteTelefono:    order.ClienteTelefono,
		ClienteEmail:       order.ClienteEmail,
		ClienteRegimen:     "comun",
		MetodoPago:         "contado",
		MedioPago:          order.MetodoPago,
	}

	if order.FechaEntregaEsperada != nil {
		invoice.FechaVencimiento = order.FechaEntregaEsperada
	} else {
		dueDate := now.Add(30 * 24 * time.Hour)
		invoice.FechaVencimiento = &dueDate
	}

	if err := database.DB.Create(invoice).Error; err != nil {
		return nil, fmt.Errorf("error creando factura: %v", err)
	}

	xmlContent, err := s.GenerateInvoiceXML(invoice)
	if err != nil {
		return nil, err
	}
	invoice.XMLContent = xmlContent
	database.DB.Save(invoice)

	return invoice, nil
}

func (s *DIANService) SendToDIAN(invoice *models.Invoice) error {
	invoice.EstadoDIAN = "enviada"
	database.DB.Save(invoice)

	invoice.EstadoDIAN = "aceptada"
	now := time.Now()
	invoice.FechaRespuestaDIAN = &now
	database.DB.Save(invoice)

	return nil
}
