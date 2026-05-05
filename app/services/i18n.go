package services

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type I18nService struct {
	translations map[string]map[string]string
	mu           sync.RWMutex
}

var (
	i18nInstance *I18nService
	i18nOnce     sync.Once
)

func GetI18n() *I18nService {
	i18nOnce.Do(func() {
		i18nInstance = &I18nService{
			translations: make(map[string]map[string]string),
		}
		i18nInstance.loadTranslations()
	})
	return i18nInstance
}

func (s *I18nService) loadTranslations() {
	s.mu.Lock()
	defer s.mu.Unlock()

	files, _ := filepath.Glob("app/i18n/*.json")
	for _, file := range files {
		lang := filepath.Base(file)[:2] // assumes en.json, fr.json
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		var trans map[string]string
		if err := json.Unmarshal(data, &trans); err == nil {
			s.translations[lang] = trans
		}
	}
}

func (s *I18nService) T(ctx context.Context, key string) string {
	lang, ok := ctx.Value("lang").(string)
	if !ok || lang == "" {
		lang = "en"
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if trans, ok := s.translations[lang]; ok {
		if val, ok := trans[key]; ok {
			return val
		}
	}

	// Fallback to English
	if trans, ok := s.translations["en"]; ok {
		if val, ok := trans[key]; ok {
			return val
		}
	}

	return key
}
