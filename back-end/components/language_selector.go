package components

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/LTPPPP/TracePost-larvaeChain/middleware"
	"github.com/gofiber/fiber/v2"
)

// Language represents a supported language
type Language struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	NativeName  string `json:"nativeName"`
	FlagEmoji   string `json:"flagEmoji"`
	Enabled     bool   `json:"enabled"`
	Percentage  int    `json:"percentage"` // Translation coverage percentage
}

// LanguageSelectorConfig holds configuration for the language selector component
type LanguageSelectorConfig struct {
	DefaultLanguage string
	Persist         bool   // Whether to persist language selection in cookies
	CookieName      string // Name of the cookie to store language preference
	CookieMaxAge    int    // Max age of the cookie in seconds
}

// LanguageSelector provides language selection functionality
type LanguageSelector struct {
	Languages []Language
	Config    LanguageSelectorConfig
	i18n      *middleware.I18n
}

// NewLanguageSelector creates a new language selector
func NewLanguageSelector(i18n *middleware.I18n, config LanguageSelectorConfig) *LanguageSelector {
	if config.CookieName == "" {
		config.CookieName = "lang_preference"
	}
	if config.CookieMaxAge == 0 {
		config.CookieMaxAge = 30 * 24 * 60 * 60 // 30 days default
	}

	// Define supported languages
	languages := []Language{
		{
			Code:        "en",
			Name:        "English",
			NativeName:  "English",
			FlagEmoji:   "🇺🇸",
			Enabled:     true,
			Percentage:  100,
		},
		{
			Code:        "vi",
			Name:        "Vietnamese",
			NativeName:  "Tiếng Việt",
			FlagEmoji:   "🇻🇳",
			Enabled:     true,
			Percentage:  100,
		},
		{
			Code:        "zh",
			Name:        "Chinese",
			NativeName:  "中文",
			FlagEmoji:   "🇨🇳",
			Enabled:     false,
			Percentage:  0,
		},
		{
			Code:        "ja",
			Name:        "Japanese",
			NativeName:  "日本語",
			FlagEmoji:   "🇯🇵",
			Enabled:     false,
			Percentage:  0,
		},
		{
			Code:        "ko",
			Name:        "Korean",
			NativeName:  "한국어",
			FlagEmoji:   "🇰🇷",
			Enabled:     false,
			Percentage:  0,
		},
		{
			Code:        "fr",
			Name:        "French",
			NativeName:  "Français",
			FlagEmoji:   "🇫🇷",
			Enabled:     false,
			Percentage:  0,
		},
		{
			Code:        "es",
			Name:        "Spanish",
			NativeName:  "Español",
			FlagEmoji:   "🇪🇸",
			Enabled:     false,
			Percentage:  0,
		},
	}

	return &LanguageSelector{
		Languages: languages,
		Config:    config,
		i18n:      i18n,
	}
}

// GetAvailableLanguages returns all available languages
func (ls *LanguageSelector) GetAvailableLanguages() []Language {
	return ls.Languages
}

// GetEnabledLanguages returns only enabled languages
func (ls *LanguageSelector) GetEnabledLanguages() []Language {
	var enabled []Language
	for _, lang := range ls.Languages {
		if lang.Enabled {
			enabled = append(enabled, lang)
		}
	}
	return enabled
}

// SetLanguage sets the active language
func (ls *LanguageSelector) SetLanguage(c *fiber.Ctx, langCode string) error {
	// Validate language code
	valid := false
	for _, lang := range ls.Languages {
		if lang.Code == langCode && lang.Enabled {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("invalid or disabled language code: %s", langCode)
	}

	// Update i18n
	ls.i18n.SetDefaultLocale(langCode)

	// Set cookie if persistence is enabled
	if ls.Config.Persist {
		c.Cookie(&fiber.Cookie{
			Name:     ls.Config.CookieName,
			Value:    langCode,
			MaxAge:   ls.Config.CookieMaxAge,
			Path:     "/",
			HTTPOnly: true,
			SameSite: "Lax",
		})
	}

	return nil
}

// GetCurrentLanguage returns the current active language
func (ls *LanguageSelector) GetCurrentLanguage(c *fiber.Ctx) Language {
	// Get language from cookie if available
	langCode := c.Cookies(ls.Config.CookieName, ls.i18n.GetDefaultLocale())

	// Find language details
	for _, lang := range ls.Languages {
		if lang.Code == langCode {
			return lang
		}
	}

	// Return default language if not found
	for _, lang := range ls.Languages {
		if lang.Code == ls.Config.DefaultLanguage {
			return lang
		}
	}

	// Fallback to first enabled language
	for _, lang := range ls.Languages {
		if lang.Enabled {
			return lang
		}
	}

	// Last resort fallback to first language
	return ls.Languages[0]
}

// RegisterRoutes registers the language selector API routes
func (ls *LanguageSelector) RegisterRoutes(app *fiber.App) {
	app.Get("/api/languages", ls.HandleGetLanguages)
	app.Post("/api/languages/:code", ls.HandleSetLanguage)
	app.Get("/api/languages/current", ls.HandleGetCurrentLanguage)
}

// HandleGetLanguages handles the request to get all available languages
func (ls *LanguageSelector) HandleGetLanguages(c *fiber.Ctx) error {
	return c.JSON(ls.GetEnabledLanguages())
}

// HandleSetLanguage handles the request to set the active language
func (ls *LanguageSelector) HandleSetLanguage(c *fiber.Ctx) error {
	langCode := c.Params("code")
	if err := ls.SetLanguage(c, langCode); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Language updated successfully",
		"code":    langCode,
	})
}

// HandleGetCurrentLanguage handles the request to get the current language
func (ls *LanguageSelector) HandleGetCurrentLanguage(c *fiber.Ctx) error {
	return c.JSON(ls.GetCurrentLanguage(c))
}

// RenderLanguageSelector returns HTML for a language selector dropdown
func (ls *LanguageSelector) RenderLanguageSelector(c *fiber.Ctx) string {
	currentLang := ls.GetCurrentLanguage(c)
	enabledLangs := ls.GetEnabledLanguages()

	// Convert languages to JSON for the client-side script
	langsJSON, _ := json.Marshal(enabledLangs)

	// Build HTML for the language selector
	html := `
<div class="language-selector">
  <div class="current-language" id="current-language">
    <span class="flag">${currentLang.FlagEmoji}</span>
    <span class="label">${currentLang.NativeName}</span>
    <span class="arrow">▼</span>
  </div>
  <div class="language-dropdown" id="language-dropdown">
    ${languageOptions}
  </div>
</div>

<style>
  .language-selector {
    position: relative;
    display: inline-block;
    font-family: Arial, sans-serif;
  }
  .current-language {
    display: flex;
    align-items: center;
    cursor: pointer;
    padding: 8px 12px;
    border: 1px solid #ddd;
    border-radius: 4px;
    background-color: #fff;
  }
  .flag {
    margin-right: 8px;
    font-size: 1.2em;
  }
  .arrow {
    margin-left: 8px;
    font-size: 0.8em;
  }
  .language-dropdown {
    display: none;
    position: absolute;
    top: 100%;
    left: 0;
    right: 0;
    background-color: #fff;
    border: 1px solid #ddd;
    border-top: none;
    border-radius: 0 0 4px 4px;
    z-index: 100;
    max-height: 200px;
    overflow-y: auto;
  }
  .language-option {
    display: flex;
    align-items: center;
    padding: 8px 12px;
    cursor: pointer;
  }
  .language-option:hover {
    background-color: #f5f5f5;
  }
  .language-option.active {
    background-color: #e9f5ff;
  }
  .lang-progress {
    margin-left: auto;
    font-size: 0.8em;
    color: #888;
  }
</style>

<script>
  document.addEventListener('DOMContentLoaded', function() {
    const currentLangEl = document.getElementById('current-language');
    const dropdownEl = document.getElementById('language-dropdown');
    const languages = ${langsJSON};
    const currentLangCode = '${currentLang.Code}';
    
    // Toggle dropdown
    currentLangEl.addEventListener('click', function() {
      dropdownEl.style.display = dropdownEl.style.display === 'block' ? 'none' : 'block';
    });
    
    // Close dropdown when clicking outside
    document.addEventListener('click', function(e) {
      if (!e.target.closest('.language-selector')) {
        dropdownEl.style.display = 'none';
      }
    });
    
    // Handle language selection
    document.querySelectorAll('.language-option').forEach(option => {
      option.addEventListener('click', function() {
        const langCode = this.getAttribute('data-lang');
        
        // Send request to change language
        fetch('/api/languages/' + langCode, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          }
        })
        .then(response => response.json())
        .then(data => {
          if (data.success) {
            // Reload page to apply new language
            window.location.reload();
          }
        })
        .catch(error => {
          console.error('Error changing language:', error);
        });
      });
    });
  });
</script>
`

	// Build language options HTML
	var languageOptions string
	for _, lang := range enabledLangs {
		activeClass := ""
		if lang.Code == currentLang.Code {
			activeClass = " active"
		}

		option := fmt.Sprintf(`
<div class="language-option%s" data-lang="%s">
  <span class="flag">%s</span>
  <span class="label">%s</span>
  <span class="lang-progress">%d%%</span>
</div>
`, activeClass, lang.Code, lang.FlagEmoji, lang.NativeName, lang.Percentage)

		languageOptions += option
	}

	// Replace placeholder with actual options
	html = strings.Replace(html, "${languageOptions}", languageOptions, 1)
	html = strings.Replace(html, "${currentLang.FlagEmoji}", currentLang.FlagEmoji, 1)
	html = strings.Replace(html, "${currentLang.NativeName}", currentLang.NativeName, 1)
	html = strings.Replace(html, "${langsJSON}", string(langsJSON), 1)
	html = strings.Replace(html, "${currentLang.Code}", currentLang.Code, 1)

	return html
}
