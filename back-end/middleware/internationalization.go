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

type I18n struct {
	bundle       *i18n.Bundle
	localizers   map[string]*i18n.Localizer
	defaultLang  string
	supportedLangs []string
	mutex        sync.RWMutex
}

func NewI18n(defaultLang string, localesDir string) (*I18n, error) {
	langTag, err := language.Parse(defaultLang)
	if err != nil {
		return nil, fmt.Errorf("invalid default language %q: %w", defaultLang, err)
	}

	bundle := i18n.NewBundle(langTag)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	i18n := &I18n{
		bundle:      bundle,
		localizers:  make(map[string]*i18n.Localizer),
		defaultLang: defaultLang,
		supportedLangs: []string{defaultLang},
	}

	if err := i18n.loadLanguageFiles(localesDir); err != nil {
		return nil, err
	}

	return i18n, nil
}

func (i *I18n) loadLanguageFiles(localesDir string) error {
	files, err := ioutil.ReadDir(localesDir)
	if err != nil {
		return fmt.Errorf("failed to read locales directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		langCode := strings.TrimSuffix(file.Name(), ".json")

		filePath := filepath.Join(localesDir, file.Name())
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read language file %s: %w", filePath, err)
		}

		tmpFile, err := os.CreateTemp("", "i18n-*.json")
		if err != nil {
			return fmt.Errorf("failed to create temporary file: %w", err)
		}
		defer os.Remove(tmpFile.Name())

		if _, err := tmpFile.Write(data); err != nil {
			return fmt.Errorf("failed to write to temporary file: %w", err)
		}
		if err := tmpFile.Close(); err != nil {
			return fmt.Errorf("failed to close temporary file: %w", err)
		}
		if _, err := i.bundle.LoadMessageFile(tmpFile.Name()); err != nil {
			return fmt.Errorf("failed to load message file %s: %w", filePath, err)
		}
		_, err = language.Parse(langCode)
		if err != nil {
			return fmt.Errorf("invalid language code %q: %w", langCode, err)
		}

		i.mutex.Lock()
		i.localizers[langCode] = i18n.NewLocalizer(i.bundle, langCode)

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

func (i *I18n) Translate(messageID string, langCode string, templateData map[string]interface{}) string {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	localizer, ok := i.localizers[langCode]
	if !ok {
		localizer = i.localizers[i.defaultLang]
	}

	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: templateData,
	})

	if err != nil {
		return messageID
	}

	return msg
}

func (i *I18n) GetSupportedLanguages() []string {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	langs := make([]string, len(i.supportedLangs))
	copy(langs, i.supportedLangs)

	return langs
}

func (i *I18n) SetDefaultLocale(langCode string) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	i.defaultLang = langCode
}

func (i *I18n) GetDefaultLocale() string {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	return i.defaultLang
}

func I18nMiddleware(i18n *I18n) fiber.Handler {
	return func(c *fiber.Ctx) error {
		acceptLanguage := c.Get("Accept-Language")

		lang := parseAcceptLanguage(acceptLanguage, i18n.GetSupportedLanguages(), i18n.defaultLang)

		c.Locals("lang", lang)

		c.Locals("translate", func(messageID string, templateData map[string]interface{}) string {
			return i18n.Translate(messageID, lang, templateData)
		})

		return c.Next()
	}
}

func parseAcceptLanguage(acceptLanguage string, supportedLanguages []string, defaultLanguage string) string {
	if acceptLanguage == "" {
		return defaultLanguage
	}

	parts := strings.Split(acceptLanguage, ",")
	for _, part := range parts {
		langTag := strings.TrimSpace(strings.Split(part, ";")[0])

		for _, supported := range supportedLanguages {
			if strings.HasPrefix(langTag, supported) || strings.HasPrefix(supported, langTag) {
				return supported
			}
		}
	}

	return defaultLanguage
}

func TranslateErrorMessage(c *fiber.Ctx, messageID string, templateData map[string]interface{}) string {
	_, ok := c.Locals("lang").(string)
	if !ok {
	}

	translateFunc, ok := c.Locals("translate").(func(string, map[string]interface{}) string)
	if !ok {
		return messageID
	}

	return translateFunc(messageID, templateData)
}
