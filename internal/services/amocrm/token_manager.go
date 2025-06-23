package amocrm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/2010kira2010/amocrm"
	"go.uber.org/zap"
)

// TokenManager управляет токенами AmoCRM с учетом ограничений библиотеки
type TokenManager struct {
	mu         sync.RWMutex
	tokenPath  string
	logger     *zap.Logger
	lastUpdate time.Time
}

// NewTokenManager создает новый менеджер токенов
func NewTokenManager(tokenPath string, logger *zap.Logger) *TokenManager {
	return &TokenManager{
		tokenPath: tokenPath,
		logger:    logger,
	}
}

// SaveToken сохраняет токен в файл
func (tm *TokenManager) SaveToken(token amocrm.Token) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	stored := TokenStored{
		AccessToken:  token.AccessToken(),
		RefreshToken: token.RefreshToken(),
		TokenType:    token.TokenType(),
		ExpiresAt:    token.ExpiresAt(),
	}

	data, err := json.MarshalIndent(stored, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	// Создаем директорию если не существует
	dir := filepath.Dir(tm.tokenPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(tm.tokenPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	tm.lastUpdate = time.Now()
	tm.logger.Info("Token saved successfully",
		zap.Time("expires_at", token.ExpiresAt()))

	return nil
}

// LoadToken загружает токен из файла
func (tm *TokenManager) LoadToken() (amocrm.Token, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	data, err := os.ReadFile(tm.tokenPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read token file: %w", err)
	}

	var stored TokenStored
	if err := json.Unmarshal(data, &stored); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %w", err)
	}

	return amocrm.NewToken(
		stored.AccessToken,
		stored.RefreshToken,
		stored.TokenType,
		stored.ExpiresAt,
	), nil
}

// IsTokenExpired проверяет, истек ли токен
func (tm *TokenManager) IsTokenExpired() bool {
	token, err := tm.LoadToken()
	if err != nil {
		return true
	}

	// Проверяем с запасом в 5 минут
	return time.Now().Add(5 * time.Minute).After(token.ExpiresAt())
}

// GetLastUpdateTime возвращает время последнего обновления токена
func (tm *TokenManager) GetLastUpdateTime() time.Time {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.lastUpdate
}

// DeleteToken удаляет файл с токеном
func (tm *TokenManager) DeleteToken() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if err := os.Remove(tm.tokenPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete token file: %w", err)
	}

	tm.logger.Info("Token file deleted")
	return nil
}

// BackupToken создает резервную копию токена
func (tm *TokenManager) BackupToken() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	data, err := os.ReadFile(tm.tokenPath)
	if err != nil {
		return fmt.Errorf("failed to read token file: %w", err)
	}

	backupPath := fmt.Sprintf("%s.backup.%d", tm.tokenPath, time.Now().Unix())
	if err := os.WriteFile(backupPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	tm.logger.Info("Token backed up", zap.String("backup_path", backupPath))
	return nil
}

// RestoreFromBackup восстанавливает токен из последней резервной копии
func (tm *TokenManager) RestoreFromBackup() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	dir := filepath.Dir(tm.tokenPath)
	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	var latestBackup string
	var latestTime int64

	// Находим самую свежую резервную копию
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".backup" {
			info, err := file.Info()
			if err != nil {
				continue
			}
			if info.ModTime().Unix() > latestTime {
				latestTime = info.ModTime().Unix()
				latestBackup = filepath.Join(dir, file.Name())
			}
		}
	}

	if latestBackup == "" {
		return fmt.Errorf("no backup files found")
	}

	data, err := os.ReadFile(latestBackup)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	if err := os.WriteFile(tm.tokenPath, data, 0600); err != nil {
		return fmt.Errorf("failed to restore token file: %w", err)
	}

	tm.logger.Info("Token restored from backup", zap.String("backup_path", latestBackup))
	return nil
}
