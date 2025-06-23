package amocrm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/2010kira2010/amocrm"
	"github.com/jasonlvhit/gocron"
	"go.uber.org/zap"

	"crm-dialer-integration/internal/models"
	"crm-dialer-integration/pkg/config"
)

type Service struct {
	client     amocrm.Client
	logger     *zap.Logger
	config     *config.Config
	tokenMutex sync.RWMutex
	tokenPath  string
}

// TokenStorage реализация хранилища токенов
type TokenStorage struct {
	service *Service
}

// NewTokenStorage создает новое хранилище токенов
func NewTokenStorage(service *Service) amocrm.TokenStorage {
	return &TokenStorage{service: service}
}

// GetToken получает токен из хранилища
func (ts *TokenStorage) GetToken() (amocrm.Token, error) {
	ts.service.tokenMutex.RLock()
	defer ts.service.tokenMutex.RUnlock()

	data, err := os.ReadFile(ts.service.tokenPath)
	if err != nil {
		return nil, err
	}

	var stored struct {
		AccessToken  string    `json:"access_token"`
		RefreshToken string    `json:"refresh_token"`
		TokenType    string    `json:"token_type"`
		ExpiresAt    time.Time `json:"expires_at"`
	}

	if err := json.Unmarshal(data, &stored); err != nil {
		return nil, err
	}

	return amocrm.NewToken(
		stored.AccessToken,
		stored.RefreshToken,
		stored.TokenType,
		stored.ExpiresAt,
	), nil
}

// SetToken сохраняет токен в хранилище
func (ts *TokenStorage) SetToken(token amocrm.Token) error {
	ts.service.tokenMutex.Lock()
	defer ts.service.tokenMutex.Unlock()

	stored := struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		ExpiresAt    int64  `json:"expires_at"`
	}{
		AccessToken:  token.AccessToken(),
		RefreshToken: token.RefreshToken(),
		TokenType:    token.TokenType(),
		ExpiresAt:    token.ExpiresAt().Unix(),
	}

	data, err := json.MarshalIndent(stored, "", "  ")
	if err != nil {
		return err
	}

	// Создаем директорию если не существует
	dir := filepath.Dir(ts.service.tokenPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(ts.service.tokenPath, data, 0600)
}

// NewService создает новый сервис AmoCRM
func NewService(cfg *config.Config, logger *zap.Logger) (*Service, error) {
	service := &Service{
		logger:    logger,
		config:    cfg,
		tokenPath: filepath.Join("/tmp", "amocrm_token.json"),
	}

	// Создаем хранилище токенов
	tokenStorage := NewTokenStorage(service)

	// Создаем клиент AmoCRM
	client := amocrm.NewWithStorage(
		tokenStorage,
		cfg.AmoCRMClientID,
		cfg.AmoCRMClientSecret,
		cfg.AmoCRMRedirectURI,
	)

	// Устанавливаем домен
	if err := client.SetDomain(cfg.AmoCRMDomain); err != nil {
		return nil, fmt.Errorf("failed to set domain: %w", err)
	}

	service.client = client

	// Пытаемся загрузить существующий токен
	if token, err := tokenStorage.GetToken(); err == nil {
		if err := client.SetToken(token); err != nil {
			logger.Warn("Failed to set loaded token", zap.Error(err))
		} else {
			logger.Info("Token loaded successfully",
				zap.Time("expires_at", token.ExpiresAt()))
		}
	}

	// Запускаем задачу проверки токена
	service.startTokenRefreshTask()

	return service, nil
}

// startTokenRefreshTask запускает задачу обновления токена
func (s *Service) startTokenRefreshTask() {
	gocron.Every(1).Hours().Do(func() {
		if err := s.CheckAndRefreshToken(); err != nil {
			s.logger.Error("Failed to refresh token", zap.Error(err))
		}
	})

	// Запускаем планировщик в отдельной горутине
	go func() {
		<-gocron.Start()
	}()
}

// CheckAndRefreshToken проверяет и обновляет токен при необходимости
func (s *Service) CheckAndRefreshToken() error {
	// Проверяем токен
	if err := s.client.CheckToken(); err != nil {
		s.logger.Info("Token needs refresh", zap.Error(err))

		// Токен автоматически обновится внутри клиента
		// благодаря кастомному TokenStorage
		return nil
	}

	return nil
}

// GetAuthURL возвращает URL для авторизации в AmoCRM
func (s *Service) GetAuthURL(state string) (*url.URL, error) {
	return s.client.AuthorizeURL(state, amocrm.PostMessageMode)
}

// ExchangeCode обменивает код авторизации на токены
func (s *Service) ExchangeCode(ctx context.Context, code string) error {
	token, err := s.client.TokenByCode(code)
	if err != nil {
		return fmt.Errorf("failed to exchange code: %w", err)
	}

	// Токен автоматически сохранится через TokenStorage
	s.logger.Info("Token received and saved",
		zap.Time("expires_at", token.ExpiresAt()))

	return nil
}

// GetLeads получает сделки из AmoCRM
func (s *Service) GetLeads(ctx context.Context, params map[string]string) ([]*amocrm.Lead, error) {
	options := &amocrm.GetLeadsOptions{}

	if limit, ok := params["limit"]; ok {
		fmt.Sscanf(limit, "%d", &options.Limit)
	}

	if page, ok := params["page"]; ok {
		fmt.Sscanf(page, "%d", &options.Page)
	}

	if query, ok := params["query"]; ok {
		options.Query = query
	}

	leads, err := s.client.GetLeads(options)
	if err != nil {
		return nil, fmt.Errorf("failed to get leads: %w", err)
	}

	return leads, nil
}

// GetLeadByID получает сделку по ID
func (s *Service) GetLeadByID(ctx context.Context, leadID int) (*amocrm.Lead, error) {
	leads, err := s.client.GetLeadsByID(leadID)
	if err != nil {
		return nil, fmt.Errorf("failed to get lead: %w", err)
	}

	if len(leads) == 0 {
		return nil, fmt.Errorf("lead not found")
	}

	return leads[0], nil
}

// UpdateLeads обновляет сделки
func (s *Service) UpdateLeads(ctx context.Context, leads []*amocrm.Lead) error {
	// Батчинг до 200 сущностей как указано в требованиях
	const batchSize = 200

	for i := 0; i < len(leads); i += batchSize {
		end := i + batchSize
		if end > len(leads) {
			end = len(leads)
		}

		batch := leads[i:end]
		updatedLeads, err := s.client.UpdateLeads(batch)
		if err != nil {
			return fmt.Errorf("failed to update leads batch %d-%d: %w", i, end, err)
		}

		s.logger.Info("Updated leads batch",
			zap.Int("from", i),
			zap.Int("to", end),
			zap.Int("updated", len(updatedLeads)))
	}

	return nil
}

// GetContacts получает контакты
func (s *Service) GetContacts(ctx context.Context, params map[string]string) ([]*amocrm.Contact, error) {
	options := &amocrm.GetContactsOptions{}

	if query, ok := params["query"]; ok {
		options.Query = query
	}

	if limit, ok := params["limit"]; ok {
		fmt.Sscanf(limit, "%d", &options.Limit)
	}

	if page, ok := params["page"]; ok {
		fmt.Sscanf(page, "%d", &options.Page)
	}

	contacts, err := s.client.GetContacts(options)
	if err != nil {
		return nil, fmt.Errorf("failed to get contacts: %w", err)
	}

	return contacts, nil
}

// GetContactByID получает контакт по ID
func (s *Service) GetContactByID(ctx context.Context, contactID int) (*amocrm.Contact, error) {
	contacts, err := s.client.GetContactsByID(contactID)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	if len(contacts) == 0 {
		return nil, fmt.Errorf("contact not found")
	}

	return contacts[0], nil
}

// GetCustomFields получает кастомные поля
func (s *Service) GetCustomFields(ctx context.Context, entityType string) ([]*models.AmoCRMField, error) {
	account, err := s.client.GetAccount()
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	var fields []*models.AmoCRMField

	switch entityType {
	case "leads":
		if account.CustomFields != nil && account.CustomFields.Leads != nil {
			for id, field := range account.CustomFields.Leads {
				fields = append(fields, &models.AmoCRMField{
					ID:         int64(id),
					Name:       field.Name,
					Type:       field.Type,
					EntityType: "leads",
				})
			}
		}

	case "contacts":
		if account.CustomFields != nil && account.CustomFields.Contacts != nil {
			for id, field := range account.CustomFields.Contacts {
				fields = append(fields, &models.AmoCRMField{
					ID:         int64(id),
					Name:       field.Name,
					Type:       field.Type,
					EntityType: "contacts",
				})
			}
		}

	default:
		return nil, fmt.Errorf("unsupported entity type: %s", entityType)
	}

	return fields, nil
}

// AddNote добавляет примечание к сущности
func (s *Service) AddNote(ctx context.Context, entityType string, entityID int, text string) error {
	note := &amocrm.Note{
		EntityID: entityID,
		Text:     text,
		NoteType: 4, // Обычное примечание
	}

	switch entityType {
	case "leads":
		note.ElementType = 2 // Тип для сделок
	case "contacts":
		note.ElementType = 1 // Тип для контактов
	default:
		return fmt.Errorf("unsupported entity type: %s", entityType)
	}

	notes, err := s.client.CreateNotes([]*amocrm.Note{note})
	if err != nil {
		return fmt.Errorf("failed to add note: %w", err)
	}

	if len(notes) > 0 {
		s.logger.Info("Note added successfully",
			zap.Int("note_id", notes[0].ID),
			zap.Int("entity_id", entityID))
	}

	return nil
}
