package ontology

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"cobraknowledge.local/cobra-knowledge/internal/model"
)

var (
	ErrOntologyNotFound = errors.New("ontology not found")
	ErrVersionNotFound  = errors.New("ontology version not found")
	ErrBindingNotFound  = errors.New("ontology binding not found")
)

// Registry is the storage-neutral ontology lifecycle contract. The current FSRegistry
// is suitable for a single CobraKnowledge deployment; a PostgreSQL implementation can
// replace it later without changing Graph API, Planner, MCP, or WeKnora integration.
type Registry interface {
	RegisterVersion(ctx context.Context, o model.Ontology, actor, notes string) (model.OntologyVersionMeta, error)
	Publish(ctx context.Context, ontologyID, version, actor string) (model.OntologyVersionMeta, error)
	Activate(ctx context.Context, ontologyID, version, actor string) error
	Get(ctx context.Context, ontologyID, version string) (model.Ontology, model.OntologyVersionMeta, error)
	GetManifest(ctx context.Context, ontologyID string) (model.OntologyManifest, error)
	ListManifests(ctx context.Context) ([]model.OntologyManifest, error)
	BindKnowledgeBase(ctx context.Context, binding model.OntologyBinding) error
	GetBinding(ctx context.Context, knowledgeBaseID string) (model.OntologyBinding, error)
	ResolveForKnowledgeBase(ctx context.Context, knowledgeBaseID string) (model.OntologyResolution, error)
	ListAudit(ctx context.Context, limit int) ([]model.RegistryAuditEvent, error)
}

// FSRegistry stores immutable ontology payloads plus mutable manifests/bindings.
// No file path is exposed to consumers: callers address logical ontology IDs/versions.
type FSRegistry struct {
	Root string
	mu   sync.Mutex
}

func NewFSRegistry(root string) *FSRegistry {
	return &FSRegistry{Root: strings.TrimSpace(root)}
}

func (r *FSRegistry) RegisterVersion(_ context.Context, o model.Ontology, actor, notes string) (model.OntologyVersionMeta, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if strings.TrimSpace(r.Root) == "" {
		return model.OntologyVersionMeta{}, fmt.Errorf("registry root is required")
	}
	if strings.TrimSpace(o.ID) == "" || strings.TrimSpace(o.Domain) == "" || strings.TrimSpace(o.Version) == "" {
		return model.OntologyVersionMeta{}, fmt.Errorf("ontology id, domain and version are required")
	}
	if violations := ValidateOntology(o); hasValidationErrors(violations) {
		return model.OntologyVersionMeta{}, fmt.Errorf("ontology validation failed: %s", firstValidationError(violations))
	}
	path := r.versionPath(o.ID, o.Version)
	if _, err := os.Stat(path); err == nil {
		return model.OntologyVersionMeta{}, fmt.Errorf("immutable ontology version already exists: %s@%s", o.ID, o.Version)
	} else if !os.IsNotExist(err) {
		return model.OntologyVersionMeta{}, err
	}

	payload, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return model.OntologyVersionMeta{}, err
	}
	payload = append(payload, '\n')
	if err := writeAtomic(path, payload, 0o644); err != nil {
		return model.OntologyVersionMeta{}, err
	}
	sum := sha256.Sum256(payload)
	meta := model.OntologyVersionMeta{
		OntologyID:    o.ID,
		Domain:        o.Domain,
		Version:       o.Version,
		State:         model.RegistryCandidate,
		ContentSHA256: hex.EncodeToString(sum[:]),
		CreatedAt:     time.Now().UTC(),
		CreatedBy:     strings.TrimSpace(actor),
		Notes:         strings.TrimSpace(notes),
	}
	if parent := strings.TrimSpace(o.Metadata["parent_version"]); parent != "" {
		meta.ParentVersion = parent
	}
	manifest, err := r.loadManifestUnlocked(o.ID)
	if err != nil && !errors.Is(err, ErrOntologyNotFound) {
		return model.OntologyVersionMeta{}, err
	}
	if manifest.ID == "" {
		manifest = model.OntologyManifest{ID: o.ID, Domain: o.Domain}
	}
	if manifest.Domain != o.Domain {
		return model.OntologyVersionMeta{}, fmt.Errorf("ontology domain mismatch: registry=%s payload=%s", manifest.Domain, o.Domain)
	}
	manifest.Versions = append(manifest.Versions, meta)
	manifest.UpdatedAt = time.Now().UTC()
	sort.SliceStable(manifest.Versions, func(i, j int) bool { return manifest.Versions[i].CreatedAt.Before(manifest.Versions[j].CreatedAt) })
	if err := r.writeManifestUnlocked(manifest); err != nil {
		_ = os.Remove(path)
		return model.OntologyVersionMeta{}, err
	}
	_ = r.appendAuditUnlocked(model.RegistryAuditEvent{Actor: actor, Action: "register_version", OntologyID: o.ID, Version: o.Version, Details: map[string]interface{}{"sha256": meta.ContentSHA256}})
	return meta, nil
}

func (r *FSRegistry) Publish(ctx context.Context, ontologyID, version, actor string) (model.OntologyVersionMeta, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	o, meta, err := r.getUnlocked(ontologyID, version)
	if err != nil {
		return model.OntologyVersionMeta{}, err
	}
	if err := ensurePublishable(o); err != nil {
		return model.OntologyVersionMeta{}, err
	}
	manifest, err := r.loadManifestUnlocked(ontologyID)
	if err != nil {
		return model.OntologyVersionMeta{}, err
	}
	now := time.Now().UTC()
	found := false
	for i := range manifest.Versions {
		if manifest.Versions[i].Version != version {
			continue
		}
		found = true
		manifest.Versions[i].State = model.RegistryPublished
		manifest.Versions[i].PublishedAt = &now
		manifest.Versions[i].PublishedBy = strings.TrimSpace(actor)
		meta = manifest.Versions[i]
		break
	}
	if !found {
		return model.OntologyVersionMeta{}, ErrVersionNotFound
	}
	manifest.ActiveVersion = version
	manifest.UpdatedAt = now
	if err := r.writeManifestUnlocked(manifest); err != nil {
		return model.OntologyVersionMeta{}, err
	}
	_ = r.appendAuditUnlocked(model.RegistryAuditEvent{Actor: actor, Action: "publish", OntologyID: ontologyID, Version: version, Details: map[string]interface{}{"active_version": version}})
	_ = ctx // reserved for non-filesystem registry implementations
	return meta, nil
}

func (r *FSRegistry) Activate(_ context.Context, ontologyID, version, actor string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	manifest, err := r.loadManifestUnlocked(ontologyID)
	if err != nil {
		return err
	}
	meta, ok := findVersion(manifest, version)
	if !ok {
		return ErrVersionNotFound
	}
	if meta.State != model.RegistryPublished {
		return fmt.Errorf("only published versions can be activated: %s@%s", ontologyID, version)
	}
	previous := manifest.ActiveVersion
	manifest.ActiveVersion = version
	manifest.UpdatedAt = time.Now().UTC()
	if err := r.writeManifestUnlocked(manifest); err != nil {
		return err
	}
	return r.appendAuditUnlocked(model.RegistryAuditEvent{Actor: actor, Action: "activate", OntologyID: ontologyID, Version: version, Details: map[string]interface{}{"previous_active_version": previous}})
}

func (r *FSRegistry) Get(_ context.Context, ontologyID, version string) (model.Ontology, model.OntologyVersionMeta, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.getUnlocked(ontologyID, version)
}

func (r *FSRegistry) GetManifest(_ context.Context, ontologyID string) (model.OntologyManifest, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.loadManifestUnlocked(ontologyID)
}

func (r *FSRegistry) ListManifests(_ context.Context) ([]model.OntologyManifest, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	base := filepath.Join(r.Root, "ontologies")
	entries, err := os.ReadDir(base)
	if os.IsNotExist(err) {
		return []model.OntologyManifest{}, nil
	}
	if err != nil {
		return nil, err
	}
	out := make([]model.OntologyManifest, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		manifest, err := r.loadManifestUnlocked(entry.Name())
		if err == nil {
			out = append(out, manifest)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (r *FSRegistry) BindKnowledgeBase(_ context.Context, binding model.OntologyBinding) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	binding.KnowledgeBaseID = strings.TrimSpace(binding.KnowledgeBaseID)
	binding.OntologyID = strings.TrimSpace(binding.OntologyID)
	binding.Mode = strings.ToLower(strings.TrimSpace(binding.Mode))
	binding.Version = strings.TrimSpace(binding.Version)
	if binding.KnowledgeBaseID == "" || binding.OntologyID == "" {
		return fmt.Errorf("knowledge base id and ontology id are required")
	}
	manifest, err := r.loadManifestUnlocked(binding.OntologyID)
	if err != nil {
		return err
	}
	switch binding.Mode {
	case "active":
		if manifest.ActiveVersion == "" {
			return fmt.Errorf("ontology %s has no active published version", binding.OntologyID)
		}
		binding.Version = ""
	case "pinned":
		meta, ok := findVersion(manifest, binding.Version)
		if !ok {
			return ErrVersionNotFound
		}
		if meta.State != model.RegistryPublished {
			return fmt.Errorf("pinned binding requires a published version")
		}
	default:
		return fmt.Errorf("binding mode must be active or pinned")
	}
	binding.UpdatedAt = time.Now().UTC()
	path := r.bindingPath(binding.KnowledgeBaseID)
	if err := writeAtomicJSON(path, binding); err != nil {
		return err
	}
	return r.appendAuditUnlocked(model.RegistryAuditEvent{Actor: binding.UpdatedBy, Action: "bind_knowledge_base", OntologyID: binding.OntologyID, Version: binding.Version, KnowledgeBaseID: binding.KnowledgeBaseID, Details: map[string]interface{}{"mode": binding.Mode}})
}

func (r *FSRegistry) GetBinding(_ context.Context, knowledgeBaseID string) (model.OntologyBinding, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.loadBindingUnlocked(knowledgeBaseID)
}

func (r *FSRegistry) ResolveForKnowledgeBase(ctx context.Context, knowledgeBaseID string) (model.OntologyResolution, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	binding, err := r.loadBindingUnlocked(knowledgeBaseID)
	if err != nil {
		return model.OntologyResolution{}, err
	}
	manifest, err := r.loadManifestUnlocked(binding.OntologyID)
	if err != nil {
		return model.OntologyResolution{}, err
	}
	version := binding.Version
	if binding.Mode == "active" {
		version = manifest.ActiveVersion
	}
	if version == "" {
		return model.OntologyResolution{}, fmt.Errorf("no active ontology version for knowledge base %s", knowledgeBaseID)
	}
	o, meta, err := r.getUnlocked(binding.OntologyID, version)
	if err != nil {
		return model.OntologyResolution{}, err
	}
	if meta.State != model.RegistryPublished {
		return model.OntologyResolution{}, fmt.Errorf("resolved ontology version is not published: %s@%s", binding.OntologyID, version)
	}
	_ = ctx
	return model.OntologyResolution{Ontology: o, Binding: binding, Version: meta}, nil
}

func (r *FSRegistry) ListAudit(_ context.Context, limit int) ([]model.RegistryAuditEvent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	path := filepath.Join(r.Root, "audit", "events.jsonl")
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return []model.RegistryAuditEvent{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var events []model.RegistryAuditEvent
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var event model.RegistryAuditEvent
		if json.Unmarshal(scanner.Bytes(), &event) == nil {
			events = append(events, event)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 100
	}
	if len(events) > limit {
		events = events[len(events)-limit:]
	}
	// newest first for operator use
	for i, j := 0, len(events)-1; i < j; i, j = i+1, j-1 {
		events[i], events[j] = events[j], events[i]
	}
	return events, nil
}

func (r *FSRegistry) getUnlocked(ontologyID, version string) (model.Ontology, model.OntologyVersionMeta, error) {
	manifest, err := r.loadManifestUnlocked(ontologyID)
	if err != nil {
		return model.Ontology{}, model.OntologyVersionMeta{}, err
	}
	meta, ok := findVersion(manifest, version)
	if !ok {
		return model.Ontology{}, model.OntologyVersionMeta{}, ErrVersionNotFound
	}
	data, err := os.ReadFile(r.versionPath(ontologyID, version))
	if err != nil {
		return model.Ontology{}, model.OntologyVersionMeta{}, err
	}
	sum := sha256.Sum256(data)
	if got := hex.EncodeToString(sum[:]); meta.ContentSHA256 != "" && got != meta.ContentSHA256 {
		return model.Ontology{}, model.OntologyVersionMeta{}, fmt.Errorf("ontology payload checksum mismatch: %s@%s", ontologyID, version)
	}
	var o model.Ontology
	if err := json.Unmarshal(data, &o); err != nil {
		return model.Ontology{}, model.OntologyVersionMeta{}, err
	}
	return o, meta, nil
}

func (r *FSRegistry) loadManifestUnlocked(ontologyID string) (model.OntologyManifest, error) {
	path := r.manifestPath(ontologyID)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return model.OntologyManifest{}, ErrOntologyNotFound
	}
	if err != nil {
		return model.OntologyManifest{}, err
	}
	var manifest model.OntologyManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return model.OntologyManifest{}, err
	}
	return manifest, nil
}

func (r *FSRegistry) writeManifestUnlocked(manifest model.OntologyManifest) error {
	return writeAtomicJSON(r.manifestPath(manifest.ID), manifest)
}

func (r *FSRegistry) loadBindingUnlocked(kbID string) (model.OntologyBinding, error) {
	data, err := os.ReadFile(r.bindingPath(kbID))
	if os.IsNotExist(err) {
		return model.OntologyBinding{}, ErrBindingNotFound
	}
	if err != nil {
		return model.OntologyBinding{}, err
	}
	var binding model.OntologyBinding
	if err := json.Unmarshal(data, &binding); err != nil {
		return model.OntologyBinding{}, err
	}
	return binding, nil
}

func (r *FSRegistry) appendAuditUnlocked(event model.RegistryAuditEvent) error {
	event.At = time.Now().UTC()
	event.ID = model.StableID("audit", event.At.Format(time.RFC3339Nano), event.Action+"\x00"+event.OntologyID+"\x00"+event.Version+"\x00"+event.KnowledgeBaseID)
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	path := filepath.Join(r.Root, "audit", "events.jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(append(data, '\n'))
	return err
}

func (r *FSRegistry) manifestPath(id string) string {
	return filepath.Join(r.Root, "ontologies", sanitize(id), "manifest.json")
}
func (r *FSRegistry) versionPath(id, version string) string {
	return filepath.Join(r.Root, "ontologies", sanitize(id), "versions", sanitize(version)+".json")
}
func (r *FSRegistry) bindingPath(kbID string) string {
	return filepath.Join(r.Root, "bindings", sanitize(kbID)+".json")
}

func findVersion(manifest model.OntologyManifest, version string) (model.OntologyVersionMeta, bool) {
	for _, meta := range manifest.Versions {
		if meta.Version == version {
			return meta, true
		}
	}
	return model.OntologyVersionMeta{}, false
}

func ensurePublishable(o model.Ontology) error {
	if o.Status != model.StatusApproved {
		return fmt.Errorf("only approved ontology snapshots can be published")
	}
	if violations := ValidateOntology(o); hasValidationErrors(violations) {
		return fmt.Errorf("ontology validation failed: %s", firstValidationError(violations))
	}
	for _, item := range o.ReviewQueue {
		if item.Risk == "high" && item.Status != "approved" {
			return fmt.Errorf("high-risk review item %s is not approved", item.ID)
		}
	}
	return nil
}

func hasValidationErrors(v []Violation) bool {
	for _, issue := range v {
		if issue.Severity == "error" {
			return true
		}
	}
	return false
}
func firstValidationError(v []Violation) string {
	for _, issue := range v {
		if issue.Severity == "error" {
			return issue.Message
		}
	}
	return "unknown validation error"
}

func writeAtomicJSON(path string, value interface{}) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return writeAtomic(path, append(data, '\n'), 0o644)
}
func writeAtomic(path string, data []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(mode); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

func sanitize(s string) string {
	s = strings.TrimSpace(s)
	replacer := strings.NewReplacer("/", "_", "\\", "_", " ", "_")
	if s == "" {
		return "default"
	}
	return replacer.Replace(s)
}
