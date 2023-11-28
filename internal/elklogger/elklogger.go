package elklogger

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"go.elastic.co/ecszap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config holds configuration for the ELK logger
type Config struct {
	ElasticsearchURL string
	Context          context.Context
}

var (
	logQueue  = make(chan *zapcore.Entry, 1000)
	elkConfig Config
	wg        sync.WaitGroup
)

// Init initializes the ELK logger with the provided configuration
func Init(cfg Config) {
	elkConfig = cfg

	encoderConfig := ecszap.NewDefaultEncoderConfig()
	core := ecszap.NewCore(encoderConfig, zapcore.AddSync(&logBuffer{}), zap.DebugLevel)
	logger := zap.New(core, zap.AddCaller())

	wg.Add(1)
	go sendLogsToELK(elkConfig.Context)

	zap.ReplaceGlobals(logger)
}

// logBuffer is an implementation of zapcore.WriteSyncer
type logBuffer struct{}

func (lb *logBuffer) Write(p []byte) (n int, err error) {
	var entry zapcore.Entry
	if err := json.Unmarshal(p, &entry); err != nil {
		return 0, err
	}

	logQueue <- &entry
	return len(p), nil
}

func (lb *logBuffer) Sync() error {
	return nil
}

// sendLogsToELK sends logs to Elasticsearch in a separate goroutine
func sendLogsToELK(ctx context.Context) {
	defer wg.Done()
	for {
		select {
		case entry := <-logQueue:
			jsonValue, _ := json.Marshal(entry)
			_, err := http.Post(elkConfig.ElasticsearchURL+"/your-index/_doc", "application/json", bytes.NewBuffer(jsonValue))
			if err != nil {
				// handle error
			}
		case <-ctx.Done():
			return
		}
	}
}

// Shutdown waits for all logs to be sent before shutting down
func Shutdown() {
	close(logQueue)
	wg.Wait()
}
