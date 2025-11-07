package store

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
	"sync"
)

// Plugin represents a stored plugin record
// Fields align with OpenAPI's Plugin schema
// status: installed | running | stopped | disabled
// checksum: optional sha256 digest

type Plugin struct {
	ID       string
	Name     string
	Version  string
	Status   string
	Checksum string
}

type Store struct {
	mu      sync.RWMutex
	plugins map[string]Plugin
}

func New() *Store {
	return &Store{plugins: make(map[string]Plugin)}
}

func sha(id, version, source string) string {
	h := sha256.Sum256([]byte(id + "@" + version + "#" + source))
	return "sha256:" + hex.EncodeToString(h[:])
}

// Install creates or overwrites a plugin record with installed status
func (s *Store) Install(id, version, sourceURL string) Plugin {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := Plugin{
		ID:       id,
		Name:     id, // default to id as name in absence of manifest
		Version:  version,
		Status:   "installed",
		Checksum: sha(id, version, sourceURL),
	}
	s.plugins[id] = p
	return p
}

// SetStatus updates the status of an existing plugin
func (s *Store) SetStatus(id, status string) (Plugin, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.plugins[id]
	if !ok {
		return Plugin{}, false
	}
	p.Status = status
	s.plugins[id] = p
	return p, true
}

// Upgrade updates the version (and checksum) of an existing plugin
func (s *Store) Upgrade(id, version string) (Plugin, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.plugins[id]
	if !ok {
		return Plugin{}, false
	}
	p.Version = version
	p.Status = "installed"
	p.Checksum = sha(id, version, "")
	s.plugins[id] = p
	return p, true
}

// Uninstall removes a plugin record
func (s *Store) Uninstall(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.plugins[id]; !ok {
		return false
	}
	delete(s.plugins, id)
	return true
}

// List returns plugins filtered by status/name and paginated
func (s *Store) List(status, name string, page, pageSize int) (items []Plugin, total int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	buf := make([]Plugin, 0, len(s.plugins))
	for _, p := range s.plugins {
		if status != "" && p.Status != status {
			continue
		}
		if name != "" && !strings.Contains(strings.ToLower(p.Name), strings.ToLower(name)) && !strings.Contains(strings.ToLower(p.ID), strings.ToLower(name)) {
			continue
		}
		buf = append(buf, p)
	}
	// stable sort by id
	sort.Slice(buf, func(i, j int) bool { return buf[i].ID < buf[j].ID })
	total = len(buf)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	start := (page - 1) * pageSize
	if start >= total {
		return []Plugin{}, total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return buf[start:end], total
}