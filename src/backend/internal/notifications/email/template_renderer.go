package email

import (
	"context"
	"fmt"
	htmltemplate "html/template"
	"log/slog"
	"strings"
	"sync"
	texttemplate "text/template"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
)

// EmailTemplateRenderer interface for rendering email templates
type EmailTemplateRenderer interface {
	RenderTemplate(ctx context.Context, templateID string, data *EmailTemplateData) (htmlContent, textContent string, err error)
	LoadTemplate(ctx context.Context, templateID string) (*EmailTemplate, error)
	ValidateTemplate(ctx context.Context, template *EmailTemplate) error
	ClearCache() error
}

// DefaultEmailTemplateRenderer implements EmailTemplateRenderer
type DefaultEmailTemplateRenderer struct {
	templates     map[string]*EmailTemplate
	htmlTemplates map[string]*htmltemplate.Template
	textTemplates map[string]*texttemplate.Template
	cache         sync.RWMutex
	logger        *slog.Logger
	config        *TemplateRendererConfig
}

// TemplateRendererConfig contains configuration for template rendering
type TemplateRendererConfig struct {
	CacheEnabled    bool          `json:"cache_enabled"`
	CacheTimeout    time.Duration `json:"cache_timeout"`
	MaxCacheSize    int           `json:"max_cache_size"`
	DefaultLanguage string        `json:"default_language"`
	BaseURL         string        `json:"base_url"`
}

// NewDefaultEmailTemplateRenderer creates a new template renderer
func NewDefaultEmailTemplateRenderer(logger *slog.Logger, config *TemplateRendererConfig) *DefaultEmailTemplateRenderer {
	renderer := &DefaultEmailTemplateRenderer{
		templates:     make(map[string]*EmailTemplate),
		htmlTemplates: make(map[string]*htmltemplate.Template),
		textTemplates: make(map[string]*texttemplate.Template),
		logger:        logger,
		config:        config,
	}

	// Load default templates
	renderer.loadDefaultTemplates()

	return renderer
}

// RenderTemplate renders an email template with the provided data
func (r *DefaultEmailTemplateRenderer) RenderTemplate(ctx context.Context, templateID string, data *EmailTemplateData) (htmlContent, textContent string, err error) {
	logger := r.logger.With("template_id", templateID, "correlation_id", data.CorrelationID)

	logger.Debug("Rendering email template")

	// Get template
	emailTemplate, err := r.LoadTemplate(ctx, templateID)
	if err != nil {
		return "", "", fmt.Errorf("failed to load template: %w", err)
	}

	// Enhance data with common variables
	enhancedData := r.enhanceTemplateData(data)

	// Render HTML content
	htmlContent, err = r.renderHTMLTemplate(emailTemplate.TemplateID, emailTemplate.HtmlTemplate, enhancedData)
	if err != nil {
		logger.Error("Failed to render HTML template", "error", err)
		return "", "", fmt.Errorf("failed to render HTML template: %w", err)
	}

	// Render text content
	textContent, err = r.renderTextTemplate(emailTemplate.TemplateID, emailTemplate.TextTemplate, enhancedData)
	if err != nil {
		logger.Error("Failed to render text template", "error", err)
		return "", "", fmt.Errorf("failed to render text template: %w", err)
	}

	logger.Info("Template rendered successfully",
		"html_length", len(htmlContent),
		"text_length", len(textContent))

	return htmlContent, textContent, nil
}

// LoadTemplate loads a template by ID
func (r *DefaultEmailTemplateRenderer) LoadTemplate(ctx context.Context, templateID string) (*EmailTemplate, error) {
	// Check cache first
	if r.config.CacheEnabled {
		r.cache.RLock()
		if template, exists := r.templates[templateID]; exists {
			r.cache.RUnlock()
			return template, nil
		}
		r.cache.RUnlock()
	}

	// Load from predefined templates
	template := r.getDefaultTemplate(templateID)
	if template == nil {
		return nil, domain.NewNotFoundError(fmt.Sprintf("template not found: %s", templateID), "")
	}

	// Cache the template
	if r.config.CacheEnabled {
		r.cache.Lock()
		r.templates[templateID] = template
		r.cache.Unlock()
	}

	return template, nil
}

// ValidateTemplate validates a template
func (r *DefaultEmailTemplateRenderer) ValidateTemplate(ctx context.Context, template *EmailTemplate) error {
	if template == nil {
		return domain.NewValidationError("template cannot be nil")
	}

	if template.TemplateID == "" {
		return domain.NewValidationError("template ID is required")
	}

	if template.EventType == "" {
		return domain.NewValidationError("event type is required")
	}

	if template.Subject == "" {
		return domain.NewValidationError("subject is required")
	}

	if template.HtmlTemplate == "" && template.TextTemplate == "" {
		return domain.NewValidationError("either HTML or text template is required")
	}

	// Validate HTML template syntax
	if template.HtmlTemplate != "" {
		_, err := htmltemplate.New("validation").Parse(template.HtmlTemplate)
		if err != nil {
			return domain.NewValidationError(fmt.Sprintf("invalid HTML template syntax: %v", err))
		}
	}

	// Validate text template syntax
	if template.TextTemplate != "" {
		_, err := texttemplate.New("validation").Parse(template.TextTemplate)
		if err != nil {
			return domain.NewValidationError(fmt.Sprintf("invalid text template syntax: %v", err))
		}
	}

	return nil
}

// ClearCache clears the template cache
func (r *DefaultEmailTemplateRenderer) ClearCache() error {
	r.cache.Lock()
	defer r.cache.Unlock()

	r.templates = make(map[string]*EmailTemplate)
	r.htmlTemplates = make(map[string]*htmltemplate.Template)
	r.textTemplates = make(map[string]*texttemplate.Template)

	r.logger.Info("Template cache cleared")
	return nil
}

// Private helper methods

// enhanceTemplateData adds common variables to template data
func (r *DefaultEmailTemplateRenderer) enhanceTemplateData(data *EmailTemplateData) map[string]interface{} {
	enhanced := make(map[string]interface{})

	// Copy original data
	if data.EventData != nil {
		for k, v := range data.EventData {
			enhanced[k] = v
		}
	}

	// Add enhanced fields
	enhanced["subscriber_name"] = data.SubscriberName
	enhanced["event_type"] = data.EventType
	enhanced["priority"] = data.Priority
	enhanced["event_description"] = data.EventDescription
	enhanced["entity_id"] = data.EntityID
	enhanced["user_id"] = data.UserID
	enhanced["timestamp"] = data.Timestamp
	enhanced["correlation_id"] = data.CorrelationID
	enhanced["action_url"] = data.ActionURL
	enhanced["unsubscribe_url"] = data.UnsubscribeURL

	// Add common variables
	enhanced["company_name"] = "International Center"
	enhanced["support_email"] = "support@international-center.app"
	enhanced["base_url"] = r.config.BaseURL
	enhanced["current_year"] = time.Now().Year()
	enhanced["current_date"] = time.Now().Format("January 2, 2006")
	
	// Add priority-based styling
	enhanced["priority_color"] = r.getPriorityColor(data.Priority)
	enhanced["priority_icon"] = r.getPriorityIcon(data.Priority)

	return enhanced
}

// renderHTMLTemplate renders HTML template content
func (r *DefaultEmailTemplateRenderer) renderHTMLTemplate(templateID, templateContent string, data map[string]interface{}) (string, error) {
	// Check cache
	if r.config.CacheEnabled {
		r.cache.RLock()
		if tmpl, exists := r.htmlTemplates[templateID]; exists {
			r.cache.RUnlock()
			return r.executeHTMLTemplate(tmpl, data)
		}
		r.cache.RUnlock()
	}

	// Parse template
	tmpl, err := htmltemplate.New(templateID).Funcs(r.getHTMLTemplateFuncs()).Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML template: %w", err)
	}

	// Cache template
	if r.config.CacheEnabled {
		r.cache.Lock()
		r.htmlTemplates[templateID] = tmpl
		r.cache.Unlock()
	}

	return r.executeHTMLTemplate(tmpl, data)
}

// renderTextTemplate renders text template content
func (r *DefaultEmailTemplateRenderer) renderTextTemplate(templateID, templateContent string, data map[string]interface{}) (string, error) {
	// Check cache
	if r.config.CacheEnabled {
		r.cache.RLock()
		if tmpl, exists := r.textTemplates[templateID]; exists {
			r.cache.RUnlock()
			return r.executeTextTemplate(tmpl, data)
		}
		r.cache.RUnlock()
	}

	// Parse template
	tmpl, err := texttemplate.New(templateID).Funcs(r.getTextTemplateFuncs()).Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse text template: %w", err)
	}

	// Cache template
	if r.config.CacheEnabled {
		r.cache.Lock()
		r.textTemplates[templateID] = tmpl
		r.cache.Unlock()
	}

	return r.executeTextTemplate(tmpl, data)
}

// executeHTMLTemplate executes HTML template with data
func (r *DefaultEmailTemplateRenderer) executeHTMLTemplate(tmpl *htmltemplate.Template, data map[string]interface{}) (string, error) {
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute HTML template: %w", err)
	}
	return buf.String(), nil
}

// executeTextTemplate executes text template with data
func (r *DefaultEmailTemplateRenderer) executeTextTemplate(tmpl *texttemplate.Template, data map[string]interface{}) (string, error) {
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute text template: %w", err)
	}
	return buf.String(), nil
}

// getHTMLTemplateFuncs returns HTML template functions
func (r *DefaultEmailTemplateRenderer) getHTMLTemplateFuncs() htmltemplate.FuncMap {
	return htmltemplate.FuncMap{
		"upper":      strings.ToUpper,
		"lower":      strings.ToLower,
		"title":      strings.Title,
		"formatDate": r.formatDate,
		"truncate":   r.truncateString,
		"default":    r.defaultValue,
	}
}

// getTextTemplateFuncs returns text template functions
func (r *DefaultEmailTemplateRenderer) getTextTemplateFuncs() texttemplate.FuncMap {
	return texttemplate.FuncMap{
		"upper":      strings.ToUpper,
		"lower":      strings.ToLower,
		"title":      strings.Title,
		"formatDate": r.formatDate,
		"truncate":   r.truncateString,
		"default":    r.defaultValue,
	}
}

// Template helper functions
func (r *DefaultEmailTemplateRenderer) formatDate(date string) string {
	parsedTime, err := time.Parse(time.RFC3339, date)
	if err != nil {
		return date
	}
	return parsedTime.Format("January 2, 2006 at 3:04 PM")
}

func (r *DefaultEmailTemplateRenderer) truncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}

func (r *DefaultEmailTemplateRenderer) defaultValue(value, defaultVal string) string {
	if value == "" {
		return defaultVal
	}
	return value
}

func (r *DefaultEmailTemplateRenderer) getPriorityColor(priority string) string {
	switch strings.ToLower(priority) {
	case "urgent":
		return "#dc3545" // Red
	case "high":
		return "#fd7e14" // Orange
	case "medium":
		return "#ffc107" // Yellow
	case "low":
		return "#28a745" // Green
	default:
		return "#6c757d" // Gray
	}
}

func (r *DefaultEmailTemplateRenderer) getPriorityIcon(priority string) string {
	switch strings.ToLower(priority) {
	case "urgent":
		return "🚨"
	case "high":
		return "⚠️"
	case "medium":
		return "📋"
	case "low":
		return "📝"
	default:
		return "ℹ️"
	}
}

// loadDefaultTemplates loads default email templates
func (r *DefaultEmailTemplateRenderer) loadDefaultTemplates() {
	// This would load templates from files or database in a real implementation
	// For now, we'll define them inline
	defaultTemplates := []*EmailTemplate{
		r.createBusinessInquiryTemplate(),
		r.createMediaInquiryTemplate(),
		r.createDonationInquiryTemplate(),
		r.createVolunteerInquiryTemplate(),
		r.createContentPublicationTemplate(),
		r.createSystemAlertTemplate(),
		r.createCapacityWarningTemplate(),
		r.createAdminActionTemplate(),
		r.createComplianceAlertTemplate(),
		r.createDefaultTemplate(),
	}

	for _, template := range defaultTemplates {
		r.templates[template.TemplateID] = template
	}

	r.logger.Info("Default email templates loaded", "count", len(defaultTemplates))
}

// getDefaultTemplate returns a default template by ID
func (r *DefaultEmailTemplateRenderer) getDefaultTemplate(templateID string) *EmailTemplate {
	if template, exists := r.templates[templateID]; exists {
		return template
	}
	
	// Return default template if specific template not found
	if defaultTemplate, exists := r.templates["default-notification-template"]; exists {
		return defaultTemplate
	}
	
	return nil
}

// Template creation methods
func (r *DefaultEmailTemplateRenderer) createBusinessInquiryTemplate() *EmailTemplate {
	return &EmailTemplate{
		TemplateID: "business-inquiry-template",
		EventType:  "inquiry-business",
		Subject:    "New Business Inquiry Received",
		HtmlTemplate: `
<html>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
	<div style="max-width: 600px; margin: 0 auto; padding: 20px;">
		<h2 style="color: {{.priority_color}};">{{.priority_icon}} New Business Inquiry</h2>
		<p>A new business inquiry has been received and requires your attention.</p>
		{{if .entity_id}}
		<p><strong>Inquiry ID:</strong> {{.entity_id}}</p>
		{{end}}
		<p><strong>Event Type:</strong> {{.event_type}}</p>
		<p><strong>Priority:</strong> {{upper .priority}}</p>
		<p><strong>Received:</strong> {{formatDate .timestamp}}</p>
		
		<div style="margin: 30px 0;">
			<a href="{{.action_url}}" style="background-color: #007bff; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px;">Review Inquiry</a>
		</div>
		
		<hr style="margin: 40px 0; border: 1px solid #eee;">
		<p style="font-size: 14px; color: #666;">
			This is an automated notification from {{.company_name}}.<br>
			If you no longer wish to receive these notifications, <a href="{{.unsubscribe_url}}">unsubscribe here</a>.
		</p>
	</div>
</body>
</html>`,
		TextTemplate: `
New Business Inquiry {{.priority_icon}}

A new business inquiry has been received and requires your attention.

{{if .entity_id}}Inquiry ID: {{.entity_id}}{{end}}
Event Type: {{.event_type}}
Priority: {{upper .priority}}
Received: {{formatDate .timestamp}}

Review the inquiry at: {{.action_url}}

---
This is an automated notification from {{.company_name}}.
To unsubscribe: {{.unsubscribe_url}}`,
		Variables: []string{"entity_id", "event_type", "priority", "timestamp", "action_url", "unsubscribe_url"},
	}
}

func (r *DefaultEmailTemplateRenderer) createDefaultTemplate() *EmailTemplate {
	return &EmailTemplate{
		TemplateID: "default-notification-template",
		EventType:  "default",
		Subject:    "Notification Alert",
		HtmlTemplate: `
<html>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
	<div style="max-width: 600px; margin: 0 auto; padding: 20px;">
		<h2 style="color: {{.priority_color}};">{{.priority_icon}} Notification Alert</h2>
		<p>You have a new notification that requires your attention.</p>
		
		<p><strong>Event Type:</strong> {{.event_type}}</p>
		<p><strong>Priority:</strong> {{upper .priority}}</p>
		<p><strong>Time:</strong> {{formatDate .timestamp}}</p>
		
		<div style="margin: 30px 0;">
			<a href="{{.action_url}}" style="background-color: #007bff; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px;">View Details</a>
		</div>
		
		<hr style="margin: 40px 0; border: 1px solid #eee;">
		<p style="font-size: 14px; color: #666;">
			This is an automated notification from {{.company_name}}.<br>
			<a href="{{.unsubscribe_url}}">Unsubscribe</a> from these notifications.
		</p>
	</div>
</body>
</html>`,
		TextTemplate: `
Notification Alert {{.priority_icon}}

You have a new notification that requires your attention.

Event Type: {{.event_type}}
Priority: {{upper .priority}}
Time: {{formatDate .timestamp}}

View details at: {{.action_url}}

---
This is an automated notification from {{.company_name}}.
To unsubscribe: {{.unsubscribe_url}}`,
		Variables: []string{"event_type", "priority", "timestamp", "action_url", "unsubscribe_url"},
	}
}

// Simplified template creation for other event types
func (r *DefaultEmailTemplateRenderer) createMediaInquiryTemplate() *EmailTemplate {
	template := r.createBusinessInquiryTemplate()
	template.TemplateID = "media-inquiry-template"
	template.EventType = "inquiry-media"
	template.Subject = "New Media Inquiry Received"
	template.HtmlTemplate = strings.ReplaceAll(template.HtmlTemplate, "Business Inquiry", "Media Inquiry")
	template.TextTemplate = strings.ReplaceAll(template.TextTemplate, "Business Inquiry", "Media Inquiry")
	return template
}

func (r *DefaultEmailTemplateRenderer) createDonationInquiryTemplate() *EmailTemplate {
	template := r.createBusinessInquiryTemplate()
	template.TemplateID = "donation-inquiry-template"
	template.EventType = "inquiry-donations"
	template.Subject = "New Donation Inquiry Received"
	template.HtmlTemplate = strings.ReplaceAll(template.HtmlTemplate, "Business Inquiry", "Donation Inquiry")
	template.TextTemplate = strings.ReplaceAll(template.TextTemplate, "Business Inquiry", "Donation Inquiry")
	return template
}

func (r *DefaultEmailTemplateRenderer) createVolunteerInquiryTemplate() *EmailTemplate {
	template := r.createBusinessInquiryTemplate()
	template.TemplateID = "volunteer-inquiry-template"
	template.EventType = "inquiry-volunteers"
	template.Subject = "New Volunteer Application Received"
	template.HtmlTemplate = strings.ReplaceAll(template.HtmlTemplate, "Business Inquiry", "Volunteer Application")
	template.TextTemplate = strings.ReplaceAll(template.TextTemplate, "Business Inquiry", "Volunteer Application")
	return template
}

func (r *DefaultEmailTemplateRenderer) createContentPublicationTemplate() *EmailTemplate {
	template := r.createBusinessInquiryTemplate()
	template.TemplateID = "content-publication-template"
	template.EventType = "event-registration"
	template.Subject = "New Content Published"
	template.HtmlTemplate = strings.ReplaceAll(template.HtmlTemplate, "Business Inquiry", "Content Publication")
	template.TextTemplate = strings.ReplaceAll(template.TextTemplate, "Business Inquiry", "Content Publication")
	return template
}

func (r *DefaultEmailTemplateRenderer) createSystemAlertTemplate() *EmailTemplate {
	template := r.createBusinessInquiryTemplate()
	template.TemplateID = "system-alert-template"
	template.EventType = "system-error"
	template.Subject = "System Alert - Immediate Action Required"
	template.HtmlTemplate = strings.ReplaceAll(template.HtmlTemplate, "Business Inquiry", "System Alert")
	template.TextTemplate = strings.ReplaceAll(template.TextTemplate, "Business Inquiry", "System Alert")
	return template
}

func (r *DefaultEmailTemplateRenderer) createCapacityWarningTemplate() *EmailTemplate {
	template := r.createBusinessInquiryTemplate()
	template.TemplateID = "capacity-warning-template"
	template.EventType = "capacity-alert"
	template.Subject = "Capacity Warning Alert"
	template.HtmlTemplate = strings.ReplaceAll(template.HtmlTemplate, "Business Inquiry", "Capacity Warning")
	template.TextTemplate = strings.ReplaceAll(template.TextTemplate, "Business Inquiry", "Capacity Warning")
	return template
}

func (r *DefaultEmailTemplateRenderer) createAdminActionTemplate() *EmailTemplate {
	template := r.createBusinessInquiryTemplate()
	template.TemplateID = "admin-action-template"
	template.EventType = "admin-action-required"
	template.Subject = "Admin Action Required"
	template.HtmlTemplate = strings.ReplaceAll(template.HtmlTemplate, "Business Inquiry", "Admin Action Required")
	template.TextTemplate = strings.ReplaceAll(template.TextTemplate, "Business Inquiry", "Admin Action Required")
	return template
}

func (r *DefaultEmailTemplateRenderer) createComplianceAlertTemplate() *EmailTemplate {
	template := r.createBusinessInquiryTemplate()
	template.TemplateID = "compliance-alert-template"
	template.EventType = "compliance-alert"
	template.Subject = "Compliance Alert - Review Required"
	template.HtmlTemplate = strings.ReplaceAll(template.HtmlTemplate, "Business Inquiry", "Compliance Alert")
	template.TextTemplate = strings.ReplaceAll(template.TextTemplate, "Business Inquiry", "Compliance Alert")
	return template
}