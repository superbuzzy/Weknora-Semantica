package promotion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

var (
	ErrCandidateNotFound = errors.New("promotion candidate not found")
	ErrVersionConflict   = errors.New("promotion candidate version conflict")
)

type Store interface {
	Create(ctx context.Context, candidate Candidate) error
	Get(ctx context.Context, id string) (Candidate, error)
	Update(ctx context.Context, candidate Candidate, expectedVersion int64) error
	List(ctx context.Context, workspaceID string, limit int) ([]Candidate, error)
}

type FSStore struct {
	Root string
	mu   sync.Mutex
}

func NewFSStore(root string) *FSStore { return &FSStore{Root: strings.TrimSpace(root)} }

func (s *FSStore) Create(_ context.Context, candidate Candidate) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.ensureRoot(); err != nil { return err }
	path, err := s.candidatePath(candidate.ID)
	if err != nil { return err }
	if _, err := os.Stat(path); err == nil { return fmt.Errorf("promotion candidate already exists: %s", candidate.ID) } else if !os.IsNotExist(err) { return err }
	return writeJSONAtomic(path, candidate)
}

func (s *FSStore) Get(_ context.Context, id string) (Candidate, error) {
	s.mu.Lock(); defer s.mu.Unlock(); return s.getUnlocked(id)
}

func (s *FSStore) Update(_ context.Context, candidate Candidate, expectedVersion int64) error {
	s.mu.Lock(); defer s.mu.Unlock()
	current, err := s.getUnlocked(candidate.ID); if err != nil { return err }
	if current.Version != expectedVersion { return ErrVersionConflict }
	if candidate.Version != expectedVersion+1 { return fmt.Errorf("candidate version must increment exactly once") }
	path, err := s.candidatePath(candidate.ID); if err != nil { return err }
	return writeJSONAtomic(path, candidate)
}

func (s *FSStore) List(_ context.Context, workspaceID string, limit int) ([]Candidate, error) {
	s.mu.Lock(); defer s.mu.Unlock()
	if err := s.ensureRoot(); err != nil { return nil, err }
	entries, err := os.ReadDir(filepath.Join(s.Root, "candidates")); if err != nil { return nil, err }
	items := make([]Candidate, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" { continue }
		var candidate Candidate
		raw, err := os.ReadFile(filepath.Join(s.Root, "candidates", entry.Name()))
		if err != nil || json.Unmarshal(raw, &candidate) != nil { continue }
		if strings.TrimSpace(workspaceID) != "" && candidate.WorkspaceID != strings.TrimSpace(workspaceID) { continue }
		items = append(items, candidate)
	}
	sort.Slice(items, func(i, j int) bool { if items[i].UpdatedAt.Equal(items[j].UpdatedAt) { return items[i].ID < items[j].ID }; return items[i].UpdatedAt.After(items[j].UpdatedAt) })
	if limit <= 0 { limit = 100 }; if limit > 500 { limit = 500 }; if len(items) > limit { items = items[:limit] }
	return items, nil
}

func (s *FSStore) getUnlocked(id string) (Candidate, error) {
	var candidate Candidate
	path, err := s.candidatePath(id); if err != nil { return candidate, err }
	raw, err := os.ReadFile(path); if os.IsNotExist(err) { return candidate, ErrCandidateNotFound }; if err != nil { return candidate, err }
	if err := json.Unmarshal(raw, &candidate); err != nil { return candidate, fmt.Errorf("decode promotion candidate: %w", err) }
	return candidate, nil
}

func (s *FSStore) ensureRoot() error {
	if strings.TrimSpace(s.Root) == "" { return fmt.Errorf("promotion store root is required") }
	return os.MkdirAll(filepath.Join(s.Root, "candidates"), 0o700)
}

func (s *FSStore) candidatePath(id string) (string, error) {
	id = strings.TrimSpace(id); if id == "" { return "", fmt.Errorf("candidate id is required") }
	for _, r := range id { if !(r >= 'a' && r <= 'z') && !(r >= 'A' && r <= 'Z') && !(r >= '0' && r <= '9') && r != '-' && r != '_' { return "", fmt.Errorf("invalid candidate id") } }
	return filepath.Join(s.Root, "candidates", id+".json"), nil
}

func writeJSONAtomic(path string, value interface{}) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil { return err }
	raw, err := json.MarshalIndent(value, "", "  "); if err != nil { return err }; raw = append(raw, '\n')
	tmp, err := os.CreateTemp(filepath.Dir(path), ".promotion-*.tmp"); if err != nil { return err }
	tmpName := tmp.Name(); defer os.Remove(tmpName)
	if err := tmp.Chmod(0o600); err != nil { _ = tmp.Close(); return err }
	if _, err := tmp.Write(raw); err != nil { _ = tmp.Close(); return err }
	if err := tmp.Sync(); err != nil { _ = tmp.Close(); return err }
	if err := tmp.Close(); err != nil { return err }
	return os.Rename(tmpName, path)
}
