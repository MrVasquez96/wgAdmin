package lang

import "fmt"

// Global translations map - populated by translation files
var languages = map[string]Translations{
	"no-nb": englishTranslations,
	"en":    norwegianTranslations,
}

// AvailableLanguages maps language codes to display names for UI
var AvailableLanguages = map[string]string{
	"en":    "English",
	"no-nb": "Norsk Bokmål ",
}

// DefaultLanguage is the fallback language
const DefaultLanguage = "en"

// Translations is a map of translation keys to their text values
type Translations map[string]string

// Language holds translations and provides lookup methods
type Language struct {
	code         string
	translations Translations
}

// New creates a new Language instance for the specified language code
// If the language code is not found, it falls back to DefaultLanguage
func New(langCode string) *Language {
	trans, ok := languages[langCode]
	if !ok {
		trans = languages[DefaultLanguage]
		langCode = DefaultLanguage
	}

	return &Language{
		code:         langCode,
		translations: trans,
	}
}

// T (Translate) looks up a translation key and returns the translated string
// If the key is not found, it falls back to English, then returns [key] format
func (l *Language) T(key string) string {
	// Try current language
	if val, ok := l.translations[key]; ok {
		return val
	}

	// Fallback to English if key missing in current language
	if l.code != DefaultLanguage {
		if val, ok := languages[DefaultLanguage][key]; ok {
			return val
		}
	}

	// Return key itself in brackets if completely missing (helps identify missing translations)
	return "[" + key + "]"
}

// Tf (Translate with format) looks up and formats a translation with fmt.Sprintf
func (l *Language) Tf(key string, args ...interface{}) string {
	return fmt.Sprintf(l.T(key), args...)
}

// Code returns the current language code
func (l *Language) Code() string {
	return l.code
}
