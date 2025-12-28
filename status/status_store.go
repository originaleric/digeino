package status

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/originaleric/digeino/config"
	"github.com/originaleric/digeino/webhook"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// StatusStore 状态存储接口
type StatusStore interface {
	CreateExecution(executionID, appName, requestID string) *ExecutionRecord
	GetExecution(executionID string) (*ExecutionRecord, bool)
	AddStatus(executionID string, status webhook.ExecutionStatus) bool
	SetResult(executionID string, result webhook.Message) bool
	ListExecutions(page, pageSize int) ([]*ExecutionRecord, int)
}

// ExecutionRecord 执行记录
type ExecutionRecord struct {
	ExecutionID   string
	AppName       string
	RequestID     string
	StartTime     time.Time
	EndTime       *time.Time
	Status        string
	StatusHistory []webhook.ExecutionStatus
	Result        *webhook.Message
	Error         string
	mu            sync.RWMutex
}

// memoryStatusStore 内存实现
type memoryStatusStore struct {
	mu         sync.RWMutex
	executions map[string]*ExecutionRecord
	maxRecords int
}

func NewMemoryStatusStore(maxRecords int) StatusStore {
	return &memoryStatusStore{
		executions: make(map[string]*ExecutionRecord),
		maxRecords: maxRecords,
	}
}

func (s *memoryStatusStore) CreateExecution(executionID, appName, requestID string) *ExecutionRecord {
	s.mu.Lock()
	defer s.mu.Unlock()

	record := &ExecutionRecord{
		ExecutionID:   executionID,
		AppName:       appName,
		RequestID:     requestID,
		StartTime:     time.Now(),
		Status:        "running",
		StatusHistory: make([]webhook.ExecutionStatus, 0),
	}
	s.executions[executionID] = record
	return record
}

func (s *memoryStatusStore) GetExecution(executionID string) (*ExecutionRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, exists := s.executions[executionID]
	return record, exists
}

func (s *memoryStatusStore) AddStatus(executionID string, status webhook.ExecutionStatus) bool {
	s.mu.RLock()
	record, exists := s.executions[executionID]
	s.mu.RUnlock()

	if !exists {
		return false
	}

	record.mu.Lock()
	defer record.mu.Unlock()

	record.StatusHistory = append(record.StatusHistory, status)
	if status.Type == "complete" {
		if status.Status == "error" {
			record.Status = "failed"
			record.Error = status.Error
		} else {
			record.Status = "completed"
		}
		now := time.Now()
		record.EndTime = &now
	}
	return true
}

func (s *memoryStatusStore) SetResult(executionID string, result webhook.Message) bool {
	s.mu.RLock()
	record, exists := s.executions[executionID]
	s.mu.RUnlock()

	if !exists {
		return false
	}

	record.mu.Lock()
	defer record.mu.Unlock()
	record.Result = &result
	return true
}

func (s *memoryStatusStore) ListExecutions(page, pageSize int) ([]*ExecutionRecord, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	records := make([]*ExecutionRecord, 0, len(s.executions))
	for _, r := range s.executions {
		records = append(records, r)
	}
	return records, len(records)
}

// MySQL 模型
type MySQLExecutionModel struct {
	ID            uint32     `gorm:"primaryKey;column:id;autoIncrement"`
	ExecutionID   string     `gorm:"column:execution_id;size:64;not null;uniqueIndex"`
	AppName       string     `gorm:"column:app_name;size:100;not null;index"`
	RequestID     string     `gorm:"column:request_id;size:100;not null;index"`
	StartTime     time.Time  `gorm:"column:start_time;not null"`
	EndTime       *time.Time `gorm:"column:end_time"`
	Status        string     `gorm:"column:status;size:32;not null"`
	Error         string     `gorm:"column:error;type:text"`
	ResultRole    string     `gorm:"column:result_role;size:32"`
	ResultContent string     `gorm:"column:result_content;type:text"`
	CreatedAt     time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

type MySQLExecutionStatusModel struct {
	ID           uint32 `gorm:"primaryKey;column:id;autoIncrement"`
	ExecutionID  string `gorm:"column:execution_id;size:64;not null;index"`
	Type         string `gorm:"column:type;size:32;not null"`
	TimestampMs  int64  `gorm:"column:timestamp_ms;not null"`
	NodeKey      string `gorm:"column:node_key;size:100"`
	NodeType     string `gorm:"column:node_type;size:50"`
	Status       string `gorm:"column:status;size:32;not null"`
	Error        string `gorm:"column:error;type:text"`
	AppName      string `gorm:"column:app_name;size:100;not null;index"`
	RequestID    string `gorm:"column:request_id;size:100;not null;index"`
	DataFlowJSON string `gorm:"column:data_flow;type:text"`
	ControlJSON  string `gorm:"column:control_flow;type:text"`
	CreatedAt    time.Time
}

type mysqlStatusStore struct {
	db          *gorm.DB
	execTable   string
	statusTable string
}

func NewMySQLStatusStore(cfg config.StatusStoreConfig) (StatusStore, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.MySQL.User, cfg.MySQL.Password, cfg.MySQL.Host, cfg.MySQL.Port, cfg.MySQL.Database)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &mysqlStatusStore{
		db:          db,
		execTable:   cfg.MySQL.ExecTable,
		statusTable: cfg.MySQL.StatusTable,
	}, nil
}

func (s *mysqlStatusStore) CreateExecution(executionID, appName, requestID string) *ExecutionRecord {
	now := time.Now()
	model := &MySQLExecutionModel{
		ExecutionID: executionID,
		AppName:     appName,
		RequestID:   requestID,
		StartTime:   now,
		Status:      "running",
	}
	s.db.Table(s.execTable).Create(model)
	return &ExecutionRecord{
		ExecutionID: executionID,
		AppName:     appName,
		RequestID:   requestID,
		StartTime:   now,
		Status:      "running",
	}
}

func (s *mysqlStatusStore) GetExecution(executionID string) (*ExecutionRecord, bool) {
	var model MySQLExecutionModel
	if err := s.db.Table(s.execTable).Where("execution_id = ?", executionID).First(&model).Error; err != nil {
		return nil, false
	}
	return &ExecutionRecord{
		ExecutionID: model.ExecutionID,
		AppName:     model.AppName,
		RequestID:   model.RequestID,
		StartTime:   model.StartTime,
		EndTime:     model.EndTime,
		Status:      model.Status,
		Error:       model.Error,
	}, true
}

func (s *mysqlStatusStore) AddStatus(executionID string, status webhook.ExecutionStatus) bool {
	var dataJson, controlJson string
	if status.DataFlow != nil {
		b, _ := json.Marshal(status.DataFlow)
		dataJson = string(b)
	}
	if status.ControlFlow != nil {
		b, _ := json.Marshal(status.ControlFlow)
		controlJson = string(b)
	}

	statusModel := &MySQLExecutionStatusModel{
		ExecutionID:  executionID,
		Type:         status.Type,
		TimestampMs:  status.Timestamp,
		NodeKey:      status.NodeKey,
		NodeType:     status.NodeType,
		Status:       status.Status,
		Error:        status.Error,
		AppName:      status.AppName,
		RequestID:    status.RequestID,
		DataFlowJSON: dataJson,
		ControlJSON:  controlJson,
	}
	s.db.Table(s.statusTable).Create(statusModel)

	if status.Type == "complete" {
		update := map[string]interface{}{
			"status":   "completed",
			"end_time": time.Now(),
		}
		if status.Status == "error" {
			update["status"] = "failed"
			update["error"] = status.Error
		}
		s.db.Table(s.execTable).Where("execution_id = ?", executionID).Updates(update)
	}
	return true
}

func (s *mysqlStatusStore) SetResult(executionID string, result webhook.Message) bool {
	update := map[string]interface{}{
		"result_role":    result.Role,
		"result_content": result.Content,
	}
	s.db.Table(s.execTable).Where("execution_id = ?", executionID).Updates(update)
	return true
}

func (s *mysqlStatusStore) ListExecutions(page, pageSize int) ([]*ExecutionRecord, int) {
	var models []MySQLExecutionModel
	var total int64
	s.db.Table(s.execTable).Count(&total)
	s.db.Table(s.execTable).Offset((page - 1) * pageSize).Limit(pageSize).Find(&models)

	records := make([]*ExecutionRecord, 0, len(models))
	for _, m := range models {
		records = append(records, &ExecutionRecord{
			ExecutionID: m.ExecutionID,
			AppName:     m.AppName,
			RequestID:   m.RequestID,
			StartTime:   m.StartTime,
			EndTime:     m.EndTime,
			Status:      m.Status,
		})
	}
	return records, int(total)
}

func GetDefaultStore() StatusStore {
	cfg := config.Get().Status.Store
	if cfg.Type == "memory" {
		return NewMemoryStatusStore(1000)
	}
	store, _ := NewMySQLStatusStore(cfg)
	return store
}
