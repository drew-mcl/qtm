package session

import (
	"context"
	"encoding/json"
	"fmt"
	"qtm/pkg/suite"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdSessionManager struct {
	SessionID  string
	etcdClient *clientv3.Client
	prefix     string
	timeout    time.Duration
	username   string
}

func NewEtcdSessionManager(endpoints []string, prefix, username string) (*EtcdSessionManager, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: endpoints,
	})
	if err != nil {
		return nil, err
	}

	return &EtcdSessionManager{
		etcdClient: cli,
		SessionID:  "ERR:PLACEHOLDER",
		prefix:     prefix,
		username:   username,
		timeout:    5 * time.Second,
	}, nil
}

// GetSessions returns a list of session IDs. The core use is create a list for a user to select from.
func (e *EtcdSessionManager) GetSessions() ([]string, error) {
	sessionListKey := fmt.Sprintf("%s/sessionList", e.prefix)
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	resp, err := e.etcdClient.Get(ctx, sessionListKey)
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return []string{}, nil
	}

	var sessionList map[string]struct{}
	if err := json.Unmarshal(resp.Kvs[0].Value, &sessionList); err != nil {
		return nil, err
	}

	var sessions []string
	for sessionID := range sessionList {
		sessions = append(sessions, sessionID)
	}

	return sessions, nil
}

// CreateSession creates a new session with the given session ID and registers it in etcd.
func (e *EtcdSessionManager) RegisterNewSession(sessionID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	e.SetSessionID(sessionID)

	sessionData := SessionData{
		SessionID:     e.SessionID,
		Username:      e.username,
		Apps:          make(map[string]AppData),
		Endpoints:     make(map[string]string),
		ConfigChanges: make([]ConfigChange, 0),
	}

	jsonData, err := json.Marshal(sessionData)
	if err != nil {
		return err
	}

	_, err = e.etcdClient.Put(ctx, fmt.Sprintf("%s/sessions/%s", e.prefix, e.SessionID), string(jsonData))
	if err != nil {
		return err
	}

	return e.updateSessionList(sessionID, true)
}

// RemoveSession removes the session and all its data from etcd.
func (e *EtcdSessionManager) RemoveSession() error {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	_, err := e.etcdClient.Delete(ctx, fmt.Sprintf("%s/sessions/%s", e.prefix, e.SessionID), clientv3.WithPrefix())
	if err != nil {
		return err
	}

	return e.updateSessionList(e.SessionID, false)
}

// SetSessionID sets the session ID.
func (e *EtcdSessionManager) SetSessionID(sessionID string) {
	e.SessionID = sessionID
}

// ValidateSession checks if the session exists in etcd.
func (e *EtcdSessionManager) ValidateSession() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	resp, err := e.etcdClient.Get(ctx, fmt.Sprintf("%s/sessions/%s", e.prefix, e.SessionID))
	if err != nil {
		return false, err
	}

	return len(resp.Kvs) > 0, nil
}

// AddApp adds an app to the session.
func (e *EtcdSessionManager) AddApp(app suite.SuiteItem, version string) error {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	appData := AppData{
		version: version,
		obj:     app,
	}

	jsonData, err := json.Marshal(appData)
	if err != nil {
		return err
	}

	_, err = e.etcdClient.Put(ctx, fmt.Sprintf("%s/sessions/%s/apps/%s", e.prefix, e.SessionID, app.Name), string(jsonData))
	if err != nil {
		return err
	}

	return nil
}

// RemoveApp removes an app from the session.
func (e *EtcdSessionManager) RemoveApp(appName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	_, err := e.etcdClient.Delete(ctx, fmt.Sprintf("%s/sessions/%s/apps/%s", e.prefix, e.SessionID, appName))
	if err != nil {
		return err
	}

	return nil
}

// AddEndpoint adds an endpoint to the session.
func (e *EtcdSessionManager) AddEndpoint(endpointName, address string) error {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	_, err := e.etcdClient.Put(ctx, fmt.Sprintf("%s/sessions/%s/endpoints/%s", e.prefix, e.SessionID, endpointName), address)
	if err != nil {
		return err
	}

	return nil
}

// AddConfigAdjustment adds a config adjustment to the session.
func (e *EtcdSessionManager) AddConfigAdjustment(app, filename, data string) error {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	configChange := ConfigChange{
		App:       app,
		Filename:  filename,
		Data:      data,
		Timestamp: time.Now().String(),
	}

	jsonData, err := json.Marshal(configChange)
	if err != nil {
		return err
	}

	_, err = e.etcdClient.Put(ctx, fmt.Sprintf("%s/sessions/%s/config_changes/%s", e.prefix, e.SessionID, configChange.Timestamp), string(jsonData))
	if err != nil {
		return err
	}

	return nil
}

// IsEmpty checks if the session is empty.
func (e *EtcdSessionManager) IsEmpty() bool {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	resp, err := e.etcdClient.Get(ctx, fmt.Sprintf("%s/sessions/%s/apps", e.prefix, e.SessionID))
	if err != nil {
		return true
	}

	return len(resp.Kvs) == 0
}

// GetAppVersion returns the version of the app.
func (e *EtcdSessionManager) GetAppVersion(appName string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	resp, err := e.etcdClient.Get(ctx, fmt.Sprintf("%s/sessions/%s/apps/%s/version", e.prefix, e.SessionID, appName))
	if err != nil {
		return "", err
	}

	if len(resp.Kvs) == 0 {
		return "", fmt.Errorf("no app found for %s", appName)
	}

	return string(resp.Kvs[0].Value), nil
}

func (e *EtcdSessionManager) CreateSessionID() (string, error) {
	return "temp", nil
}

func (e *EtcdSessionManager) updateSessionList(sessionID string, add bool) error {
	sessionListKey := fmt.Sprintf("%s/sessionList", e.prefix)
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	// Retrieve the current session list
	resp, err := e.etcdClient.Get(ctx, sessionListKey)
	if err != nil {
		return err
	}

	sessionList := make(map[string]struct{})
	if len(resp.Kvs) > 0 {
		if err := json.Unmarshal(resp.Kvs[0].Value, &sessionList); err != nil {
			return err
		}
	}

	// Update the session list
	if add {
		sessionList[sessionID] = struct{}{}
	} else {
		delete(sessionList, sessionID)
	}

	// Save the updated list back to etcd
	updatedList, err := json.Marshal(sessionList)
	if err != nil {
		return err
	}

	_, err = e.etcdClient.Put(ctx, sessionListKey, string(updatedList))
	return err
}
