// internationalization.go
package middleware

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// Supported languages
const (
	LangEN = "en" // English (default)
	LangVI = "vi" // Vietnamese
	LangZH = "zh" // Chinese
	LangJA = "ja" // Japanese
	LangKO = "ko" // Korean
	LangFR = "fr" // French
	LangDE = "de" // German
	LangES = "es" // Spanish
	LangIT = "it" // Italian
	LangRU = "ru" // Russian
)

// I18n holds internationalization resources
type I18n struct {
	bundle       *i18n.Bundle
	localizers   map[string]*i18n.Localizer
	defaultLang  string
	supportedLangs []string
	mutex        sync.RWMutex
}

// NewI18n creates a new internationalization instance
func NewI18n(defaultLang string, localesDir string) (*I18n, error) {
	// Create a new bundle with the default language
	langTag, err := language.Parse(defaultLang)
	if err != nil {
		return nil, fmt.Errorf("invalid default language %q: %w", defaultLang, err)
	}

	// Initialize bundle
	bundle := i18n.NewBundle(langTag)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// Create I18n instance
	i18n := &I18n{
		bundle:      bundle,
		localizers:  make(map[string]*i18n.Localizer),
		defaultLang: defaultLang,
		supportedLangs: []string{defaultLang},
	}

	// Load language files
	if err := i18n.loadLanguageFiles(localesDir); err != nil {
		return nil, err
	}

	return i18n, nil
}

// loadLanguageFiles loads all language files from the given directory
func (i *I18n) loadLanguageFiles(localesDir string) error {
	// Read language files from the directory
	files, err := ioutil.ReadDir(localesDir)
	if err != nil {
		return fmt.Errorf("failed to read locales directory: %w", err)
	}

	// Load each language file
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Check if file is a JSON file
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		// Extract language code from filename
		langCode := strings.TrimSuffix(file.Name(), ".json")

		// Load language file
		filePath := filepath.Join(localesDir, file.Name())
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read language file %s: %w", filePath, err)
		}

		// Create a temporary file for the bundle to read
		tmpFile, err := os.CreateTemp("", "i18n-*.json")
		if err != nil {
			return fmt.Errorf("failed to create temporary file: %w", err)
		}
		defer os.Remove(tmpFile.Name())

		// Write data to the temporary file
		if _, err := tmpFile.Write(data); err != nil {
			return fmt.Errorf("failed to write to temporary file: %w", err)
		}
		if err := tmpFile.Close(); err != nil {
			return fmt.Errorf("failed to close temporary file: %w", err)
		}		// Load the message file
		if _, err := i.bundle.LoadMessageFile(tmpFile.Name()); err != nil {
			return fmt.Errorf("failed to load message file %s: %w", filePath, err)
		}	// Validate language code
		_, err = language.Parse(langCode)
		if err != nil {
			return fmt.Errorf("invalid language code %q: %w", langCode, err)
		}

		i.mutex.Lock()
		i.localizers[langCode] = i18n.NewLocalizer(i.bundle, langCode)
		
		// Add to supported languages if not already included
		langExists := false
		for _, lang := range i.supportedLangs {
			if lang == langCode {
				langExists = true
				break
			}
		}
		
		if !langExists {
			i.supportedLangs = append(i.supportedLangs, langCode)
		}
		
		i.mutex.Unlock()
	}

	return nil
}

// Translate translates a message ID to the given language
func (i *I18n) Translate(messageID string, langCode string, templateData map[string]interface{}) string {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	// Get localizer for the requested language
	localizer, ok := i.localizers[langCode]
	if !ok {
		// Fall back to default language if requested language is not supported
		localizer = i.localizers[i.defaultLang]
	}

	// Translate message
	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: templateData,
	})

	if err != nil {
		// If translation fails, return the message ID as fallback
		return messageID
	}

	return msg
}

// GetSupportedLanguages returns the list of supported languages
func (i *I18n) GetSupportedLanguages() []string {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	
	// Return a copy to prevent modification of the original slice
	langs := make([]string, len(i.supportedLangs))
	copy(langs, i.supportedLangs)
	
	return langs
}

// SetDefaultLocale sets the default language for localization
func (i *I18n) SetDefaultLocale(langCode string) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	i.defaultLang = langCode
}

// GetDefaultLocale returns the default language for localization
func (i *I18n) GetDefaultLocale() string {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	return i.defaultLang
}

// I18nMiddleware creates a middleware for internationalization
func I18nMiddleware(i18n *I18n) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get language from Accept-Language header
		acceptLanguage := c.Get("Accept-Language")
		
		// Parse accept language
		lang := parseAcceptLanguage(acceptLanguage, i18n.GetSupportedLanguages(), i18n.defaultLang)
		
		// Store language in locals for use in handlers
		c.Locals("lang", lang)
		
		// Add translate function to locals
		c.Locals("translate", func(messageID string, templateData map[string]interface{}) string {
			return i18n.Translate(messageID, lang, templateData)
		})
		
		return c.Next()
	}
}

// parseAcceptLanguage parses the Accept-Language header and returns the best matching language
func parseAcceptLanguage(acceptLanguage string, supportedLanguages []string, defaultLanguage string) string {
	if acceptLanguage == "" {
		return defaultLanguage
	}
	
	// Split Accept-Language into parts
	parts := strings.Split(acceptLanguage, ",")
		// Parse each part
	for _, part := range parts {
		// Extract language tag
		langTag := strings.TrimSpace(strings.Split(part, ";")[0])
		
		// Try to match with supported languages
		for _, supported := range supportedLanguages {
			if strings.HasPrefix(langTag, supported) || strings.HasPrefix(supported, langTag) {
				return supported
			}
		}
	}
	
	return defaultLanguage
}

// TranslateErrorMessage translates an error message based on language context
func TranslateErrorMessage(c *fiber.Ctx, messageID string, templateData map[string]interface{}) string {
	// Get language from context - this is used by the translate function retrieved from context
	_, ok := c.Locals("lang").(string)
	if !ok {
		// Default to English if language not set - this is handled by the middleware
	}
	
	// Get translate function from context
	translateFunc, ok := c.Locals("translate").(func(string, map[string]interface{}) string)
	if !ok {
		// If translate function not available, return message ID as fallback
		return messageID
	}
	
	return translateFunc(messageID, templateData)
}
