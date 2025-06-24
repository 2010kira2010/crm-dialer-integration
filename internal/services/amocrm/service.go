package amocrm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/2010kira2010/amocrm"
	"github.com/jasonlvhit/gocron"
	"go.uber.org/zap"

	"crm-dialer-integration/internal/models"
	"crm-dialer-integration/pkg/config"
)

type Service struct {
	client       amocrm.Client
	logger       *zap.Logger
	config       *config.Config
	tokenManager *TokenManager
}

// TokenStored структура для хранения токенов
type TokenStored struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// NewService создает новый сервис AmoCRM
func NewService(cfg *config.Config, logger *zap.Logger) (*Service, error) {
	servicePath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	tokenPath := filepath.Join(filepath.Dir(servicePath), "amocrm_token.json")
	service := &Service{
		logger:       logger,
		config:       cfg,
		tokenManager: NewTokenManager(tokenPath, logger),
	}

	// Создаем клиент AmoCRM
	client := amocrm.New(
		cfg.AmoCRMClientID,
		cfg.AmoCRMClientSecret,
		cfg.AmoCRMRedirectURI,
	)

	// Устанавливаем домен
	if err := client.SetDomain(cfg.AmoCRMDomain); err != nil {
		return nil, fmt.Errorf("failed to set domain: %w", err)
	}

	service.client = client

	// Загружаем или получаем токен
	if err := service.initializeToken(); err != nil {
		logger.Error("Failed to initialize token", zap.Error(err))
		// Не возвращаем ошибку, так как токен может быть получен позже через OAuth
	}

	// Запускаем задачу проверки токена
	service.startTokenRefreshTask()

	return service, nil
}

// initializeToken загружает существующий токен или получает новый через код авторизации
func (s *Service) initializeToken() error {
	// Пытаемся загрузить существующий токен
	if token, err := s.tokenManager.LoadToken(); err == nil && token != nil {
		if err := s.client.SetToken(token); err != nil {
			return fmt.Errorf("failed to set loaded token: %w", err)
		}
		s.logger.Info("Token loaded successfully",
			zap.Time("expires_at", token.ExpiresAt()))
		return nil
	}

	// Если токена нет и есть код авторизации, получаем токен
	if s.config.AmoCRMAuthCode != "" {
		token, err := s.client.TokenByCode(s.config.AmoCRMAuthCode)
		if err != nil {
			return fmt.Errorf("failed to get token by code: %w", err)
		}

		// Сохраняем токен
		if err := s.tokenManager.SaveToken(token); err != nil {
			s.logger.Error("Failed to save token", zap.Error(err))
		}

		s.logger.Info("Token obtained and saved",
			zap.Time("expires_at", token.ExpiresAt()))
		return nil
	}

	return fmt.Errorf("no token available and no auth code provided")
}

// loadToken загружает токен из файла
func (s *Service) loadToken() (amocrm.Token, error) {
	return s.tokenManager.LoadToken()
}

// saveToken сохраняет токен в файл
func (s *Service) saveToken(token amocrm.Token) error {
	return s.tokenManager.SaveToken(token)
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

		// После вызова CheckToken токен должен обновиться внутри клиента
		// Загружаем обновленный токен из файла (если библиотека его сохранила)
		// или пытаемся получить его другим способом

		// Поскольку библиотека автоматически обновляет токен,
		// мы можем попробовать сделать тестовый запрос
		values := url.Values{}
		values.Add("limit", "1")
		_, _, statusCode := s.client.Leads().GetLeads(values)

		if statusCode == 200 || statusCode == 204 {
			// Токен успешно обновлен, сохраняем его
			// К сожалению, без метода GetToken мы не можем получить обновленный токен
			// Это ограничение библиотеки
			s.logger.Info("Token refreshed successfully")
		} else {
			return fmt.Errorf("failed to refresh token, status code: %d", statusCode)
		}
	}

	return nil
}

// GetAuthURL возвращает URL для авторизации в AmoCRM
func (s *Service) GetAuthURL(state string) string {
	authURL, _ := s.client.AuthorizeURL(state, amocrm.PostMessageMode)
	return authURL.String()
}

// ExchangeCode обменивает код авторизации на токены
func (s *Service) ExchangeCode(ctx context.Context, code string) error {
	token, err := s.client.TokenByCode(code)
	if err != nil {
		return fmt.Errorf("failed to exchange code: %w", err)
	}

	// Сохраняем токен
	if err := s.saveToken(token); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	s.logger.Info("Token received and saved",
		zap.Time("expires_at", token.ExpiresAt()))

	return nil
}

// GetToken возвращает текущий токен (для внутреннего использования)
func (s *Service) GetToken() amocrm.Token {
	// Поскольку в библиотеке нет метода GetToken, возвращаем загруженный токен
	token, err := s.loadToken()
	if err != nil {
		s.logger.Error("Failed to load token", zap.Error(err))
		return nil
	}
	return token
}

// SetTokens устанавливает токены (для загрузки сохраненных токенов)
func (s *Service) SetTokens(tokens *TokenStored) error {
	token := amocrm.NewToken(
		tokens.AccessToken,
		tokens.RefreshToken,
		tokens.TokenType,
		tokens.ExpiresAt,
	)

	if err := s.client.SetToken(token); err != nil {
		return err
	}

	// Сохраняем в файл
	return s.saveToken(token)
}

// LoadTokens загружает токены из файла (статический метод для внешнего использования)
func LoadTokens(ctx context.Context) (*TokenStored, error) {
	servicePath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	tokenPath := filepath.Join(filepath.Dir(servicePath), "amocrm_token.json")

	data, err := os.ReadFile(tokenPath)
	if err != nil {
		return nil, err
	}

	var stored TokenStored
	if err := json.Unmarshal(data, &stored); err != nil {
		return nil, err
	}

	return &stored, nil
}

// GetLeads получает сделки из AmoCRM
func (s *Service) GetLeads(ctx context.Context, params map[string]string) ([]*amocrm.Lead, error) {
	values := url.Values{}

	// Добавляем связанные сущности
	values.Add("with", "contacts")

	// Добавляем параметры из запроса
	if limit, ok := params["limit"]; ok {
		values.Add("limit", limit)
	} else {
		values.Add("limit", "50")
	}

	if page, ok := params["page"]; ok {
		values.Add("page", page)
	} else {
		values.Add("page", "1")
	}

	if query, ok := params["query"]; ok {
		values.Add("query", query)
	}

	// Фильтры
	if statusID, ok := params["filter[status_id]"]; ok {
		values.Add("filter[statuses][0][status_id]", statusID)
	}

	if pipelineID, ok := params["filter[pipeline_id]"]; ok {
		values.Add("filter[statuses][0][pipeline_id]", pipelineID)
	}

	if responsibleUserID, ok := params["filter[responsible_user_id]"]; ok {
		values.Add("filter[responsible_user_id]", responsibleUserID)
	}

	// Сортировка
	values.Add("order[id]", "desc")

	// Вызываем API
	resLeads, err, statusCode := s.client.Leads().GetLeads(values)

	if err != nil && statusCode != 204 {
		s.logger.Error("Failed to get leads",
			zap.Error(err),
			zap.Int("status_code", statusCode))
		return nil, fmt.Errorf("failed to get leads: %w", err)
	}

	if statusCode == 204 || resLeads == nil {
		// Нет контента
		return []*amocrm.Lead{}, nil
	}

	if err == io.EOF {
		return []*amocrm.Lead{}, nil
	}

	if resLeads != nil && resLeads.Embedded.Leads != nil {
		return resLeads.Embedded.Leads, nil
	}

	return []*amocrm.Lead{}, nil
}

// GetLeadByID получает сделку по ID
func (s *Service) GetLeadByID(ctx context.Context, leadID int) (*amocrm.Lead, error) {
	lead, err, statusCode := s.client.Leads().GetLead(strconv.Itoa(leadID))

	if err != nil {
		s.logger.Error("Failed to get lead",
			zap.Error(err),
			zap.Int("lead_id", leadID),
			zap.Int("status_code", statusCode))
		return nil, fmt.Errorf("failed to get lead: %w", err)
	}

	if statusCode == 404 {
		return nil, fmt.Errorf("lead not found")
	}

	return lead, nil
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
		updatedLeads, err, statusCode := s.client.Leads().Update(batch)

		if err != nil {
			s.logger.Error("Failed to update leads batch",
				zap.Error(err),
				zap.Int("from", i),
				zap.Int("to", end),
				zap.Int("status_code", statusCode))
			return fmt.Errorf("failed to update leads batch %d-%d: %w", i, end, err)
		}

		s.logger.Info("Updated leads batch",
			zap.Int("from", i),
			zap.Int("to", end),
			zap.Int("updated", len(updatedLeads)),
			zap.Int("status_code", statusCode))
	}

	return nil
}

// GetContacts получает контакты
func (s *Service) GetContacts(ctx context.Context, params map[string]string) ([]*amocrm.Contact, error) {
	values := url.Values{}

	// Добавляем параметры
	if query, ok := params["query"]; ok {
		values.Add("query", query)
	}

	if limit, ok := params["limit"]; ok {
		values.Add("limit", limit)
	} else {
		values.Add("limit", "50")
	}

	if page, ok := params["page"]; ok {
		values.Add("page", page)
	} else {
		values.Add("page", "1")
	}

	// Вызываем API
	resContacts, err, statusCode := s.client.Contacts().GetContacts(values)

	if err != nil && statusCode != 204 {
		s.logger.Error("Failed to get contacts",
			zap.Error(err),
			zap.Int("status_code", statusCode))
		return nil, fmt.Errorf("failed to get contacts: %w", err)
	}

	if statusCode == 204 || resContacts == nil {
		// Нет контента
		return []*amocrm.Contact{}, nil
	}

	if err == io.EOF {
		return []*amocrm.Contact{}, nil
	}

	if resContacts != nil && resContacts.Embedded.Contacts != nil {
		return resContacts.Embedded.Contacts, nil
	}

	return []*amocrm.Contact{}, nil
}

// GetContactByID получает контакт по ID
func (s *Service) GetContactByID(ctx context.Context, contactID int) (*amocrm.Contact, error) {
	contact, err, statusCode := s.client.Contacts().GetContact(strconv.Itoa(contactID))

	if err != nil {
		s.logger.Error("Failed to get contact",
			zap.Error(err),
			zap.Int("contact_id", contactID),
			zap.Int("status_code", statusCode))
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	if statusCode == 404 {
		return nil, fmt.Errorf("contact not found")
	}

	return contact, nil
}

// GetCustomFields получает кастомные поля
func (s *Service) GetCustomFields(ctx context.Context, entityType string) ([]*models.AmoCRMField, error) {
	// В библиотеке нет прямого метода для получения полей
	// Обычно они приходят вместе с аккаунтом или сущностями
	// Для демонстрации возвращаем пустой массив
	// В реальном приложении нужно будет получать поля из сущностей или через отдельный endpoint

	s.logger.Warn("GetCustomFields not fully implemented",
		zap.String("entity_type", entityType))

	return []*models.AmoCRMField{}, nil
}

// AddNote добавляет примечание к сущности
func (s *Service) AddNote(ctx context.Context, entityType string, entityID int, text string) error {
	note := &amocrm.Notes{
		EntityID: entityID,
		NoteType: "common", // Обычное примечание
		Params: &amocrm.NotesParams{
			Text: text,
		},
		CreatedAt: int(time.Now().Unix()),
		UpdatedAt: int(time.Now().Unix()),
	}

	switch entityType {
	case "leads":
		// Для сделок
		notes, err, statusCode := s.client.Leads().AddNotes([]*amocrm.Notes{note})
		if err != nil {
			s.logger.Error("Failed to add note to lead",
				zap.Error(err),
				zap.Int("entity_id", entityID),
				zap.Int("status_code", statusCode))
			return fmt.Errorf("failed to add note: %w", err)
		}

		if len(notes) > 0 {
			s.logger.Info("Note added successfully to lead",
				zap.Int("note_id", notes[0].ID),
				zap.Int("entity_id", entityID))
		}

	case "contacts":
		// Для контактов
		notes, err, statusCode := s.client.Contacts().AddNotes([]*amocrm.Notes{note})
		if err != nil {
			s.logger.Error("Failed to add note to contact",
				zap.Error(err),
				zap.Int("entity_id", entityID),
				zap.Int("status_code", statusCode))
			return fmt.Errorf("failed to add note: %w", err)
		}

		if len(notes) > 0 {
			s.logger.Info("Note added successfully to contact",
				zap.Int("note_id", notes[0].ID),
				zap.Int("entity_id", entityID))
		}

	default:
		return fmt.Errorf("unsupported entity type: %s", entityType)
	}

	return nil
}
