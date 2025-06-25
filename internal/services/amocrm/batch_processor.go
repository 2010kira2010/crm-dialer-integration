package amocrm

import (
	"context"
	"sync"
	"time"

	"github.com/2010kira2010/amocrm"
	"go.uber.org/zap"
)

// LeadUpdate представляет обновление лида
type LeadUpdate struct {
	LeadID     int
	Fields     map[int]interface{}
	StatusID   int
	PipelineID int
	ReceivedAt time.Time
}

// LeadBatchProcessor обрабатывает обновления лидов батчами
type LeadBatchProcessor struct {
	service       *Service
	logger        *zap.Logger
	batchSize     int
	batchInterval time.Duration

	mu     sync.Mutex
	leads  map[int]*LeadUpdate // key is LeadID
	ticker *time.Ticker
	stopCh chan struct{}
	wg     sync.WaitGroup
}

// Start запускает обработчик
func (p *LeadBatchProcessor) Start(ctx context.Context) {
	p.ticker = time.NewTicker(p.batchInterval)
	p.stopCh = make(chan struct{})
	p.wg.Add(1)

	go func() {
		defer p.wg.Done()
		for {
			select {
			case <-p.ticker.C:
				p.processBatch(ctx)
			case <-p.stopCh:
				// Process remaining leads before stopping
				p.processBatch(ctx)
				return
			case <-ctx.Done():
				p.processBatch(ctx)
				return
			}
		}
	}()
}

// Stop останавливает обработчик
func (p *LeadBatchProcessor) Stop() {
	if p.ticker != nil {
		p.ticker.Stop()
	}
	close(p.stopCh)
	p.wg.Wait()
}

// AddLead добавляет лид в очередь на обработку
func (p *LeadBatchProcessor) AddLead(update *LeadUpdate) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Если лид уже есть в батче, обновляем его данные
	// (последнее обновление выигрывает)
	if existing, ok := p.leads[update.LeadID]; ok {
		// Merge fields
		if existing.Fields == nil {
			existing.Fields = make(map[int]interface{})
		}
		for k, v := range update.Fields {
			existing.Fields[k] = v
		}

		// Update status/pipeline if provided
		if update.StatusID > 0 {
			existing.StatusID = update.StatusID
		}
		if update.PipelineID > 0 {
			existing.PipelineID = update.PipelineID
		}

		existing.ReceivedAt = update.ReceivedAt
	} else {
		p.leads[update.LeadID] = update
	}

	// Если достигли размера батча, обрабатываем немедленно
	if len(p.leads) >= p.batchSize {
		go p.processBatch(context.Background())
	}
}

// processBatch обрабатывает накопленные обновления
func (p *LeadBatchProcessor) processBatch(ctx context.Context) {
	p.mu.Lock()
	if len(p.leads) == 0 {
		p.mu.Unlock()
		return
	}

	// Копируем и очищаем текущий батч
	batch := make(map[int]*LeadUpdate)
	for k, v := range p.leads {
		batch[k] = v
	}
	p.leads = make(map[int]*LeadUpdate)
	p.mu.Unlock()

	p.logger.Info("Processing lead batch",
		zap.Int("batch_size", len(batch)))

	// Группируем обновления по типу операции для оптимизации
	var leadsToUpdate []*amocrm.Lead

	// Сначала получаем все лиды одним запросом
	leadIDs := make([]int, 0, len(batch))
	for leadID := range batch {
		leadIDs = append(leadIDs, leadID)
	}

	// Для каждого лида в батче
	for leadID, update := range batch {
		// Получаем текущие данные лида
		lead, err := p.service.GetLeadByID(ctx, leadID)
		if err != nil {
			p.logger.Error("Failed to get lead",
				zap.Int("lead_id", leadID),
				zap.Error(err))
			continue
		}

		// Обновляем поля
		if update.Fields != nil && len(update.Fields) > 0 {
			if lead.CustomFieldsValues == nil {
				lead.CustomFieldsValues = make([]*amocrm.CustomsFields, 0)
			}

			// Обновляем существующие поля или добавляем новые
			for fieldID, value := range update.Fields {
				updated := false
				for _, field := range lead.CustomFieldsValues {
					if field.FieldID == fieldID {
						field.Values = []*amocrm.CustomsFieldsValues{
							{Value: value},
						}
						updated = true
						break
					}
				}

				// Если поле не найдено, добавляем новое
				if !updated {
					lead.CustomFieldsValues = append(lead.CustomFieldsValues, &amocrm.CustomsFields{
						FieldID: fieldID,
						Values: []*amocrm.CustomsFieldsValues{
							{Value: value},
						},
					})
				}
			}
		}

		// Обновляем статус
		if update.StatusID > 0 {
			lead.StatusID = update.StatusID
		}

		// Обновляем воронку
		if update.PipelineID > 0 {
			lead.PipelineID = update.PipelineID
		}

		// Обновляем время изменения
		lead.UpdatedAt = int(time.Now().Unix())

		leadsToUpdate = append(leadsToUpdate, lead)
	}

	// Отправляем батч обновлений в AmoCRM
	if len(leadsToUpdate) > 0 {
		if err := p.service.UpdateLeads(ctx, leadsToUpdate); err != nil {
			p.logger.Error("Failed to update leads batch",
				zap.Error(err),
				zap.Int("batch_size", len(leadsToUpdate)))

			// В случае ошибки можно реализовать повторную попытку
			// или сохранение в dead letter queue
		} else {
			p.logger.Info("Successfully updated leads batch",
				zap.Int("batch_size", len(leadsToUpdate)))
		}
	}
}
