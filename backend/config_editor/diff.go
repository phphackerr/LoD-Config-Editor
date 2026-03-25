package config_editor

import (
	"fmt"
	"sort"

	"gopkg.in/ini.v1"
)

type DiffStatus string

const (
	DiffStatusAdded    DiffStatus = "added"
	DiffStatusModified DiffStatus = "modified"
	DiffStatusDeleted  DiffStatus = "deleted"
)

type KeyDiff struct {
	Key    string     `json:"key"`
	Old    string     `json:"old"`
	New    string     `json:"new"`
	Status DiffStatus `json:"status"`
}

type SectionDiff struct {
	Section string     `json:"section"`
	Status  DiffStatus `json:"status"`
	Keys    []KeyDiff  `json:"keys"`
}

// CheckConfigDiff compares the current in-memory config with the on-disk file
// and returns only externally changed sections/keys.
func (e *ConfigEditor) CheckConfigDiff() ([]SectionDiff, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.config == nil || e.config.Path() == "" {
		return nil, ErrConfigNotLoaded
	}

	diskCfg := &GameConfig{}
	if err := diskCfg.Load(e.config.Path()); err != nil {
		return nil, fmt.Errorf("failed to load config from disk: %w", err)
	}

	bySection := make(map[string]*SectionDiff)
	ensureSection := func(name string) *SectionDiff {
		if existing, ok := bySection[name]; ok {
			return existing
		}
		d := &SectionDiff{
			Section: name,
			Status:  DiffStatusModified,
			Keys:    make([]KeyDiff, 0, 4),
		}
		bySection[name] = d
		return d
	}

	// Added/modified sections and keys (disk compared to in-memory).
	for _, diskSection := range diskCfg.file.Sections() {
		secName := diskSection.Name()
		memSection, memSectionExists := getSectionIfExists(e.config.file, secName)

		if !memSectionExists {
			sectionDiff := ensureSection(secName)
			sectionDiff.Status = DiffStatusAdded
			for _, key := range diskSection.Keys() {
				sectionDiff.Keys = append(sectionDiff.Keys, KeyDiff{
					Key:    key.Name(),
					Old:    "",
					New:    key.Value(),
					Status: DiffStatusAdded,
				})
			}
			continue
		}

		for _, key := range diskSection.Keys() {
			diskVal := key.Value()
			if memSection.HasKey(key.Name()) {
				memVal := memSection.Key(key.Name()).String()
				if diskVal != memVal {
					sectionDiff := ensureSection(secName)
					sectionDiff.Keys = append(sectionDiff.Keys, KeyDiff{
						Key:    key.Name(),
						Old:    memVal,
						New:    diskVal,
						Status: DiffStatusModified,
					})
				}
				continue
			}

			sectionDiff := ensureSection(secName)
			sectionDiff.Keys = append(sectionDiff.Keys, KeyDiff{
				Key:    key.Name(),
				Old:    "",
				New:    diskVal,
				Status: DiffStatusAdded,
			})
		}
	}

	// Deleted sections and keys.
	for _, memSection := range e.config.file.Sections() {
		secName := memSection.Name()
		diskSection, diskSectionExists := getSectionIfExists(diskCfg.file, secName)
		if !diskSectionExists {
			sectionDiff := ensureSection(secName)
			sectionDiff.Status = DiffStatusDeleted
			sectionDiff.Keys = sectionDiff.Keys[:0]
			continue
		}

		for _, key := range memSection.Keys() {
			if diskSection.HasKey(key.Name()) {
				continue
			}

			sectionDiff := ensureSection(secName)
			sectionDiff.Keys = append(sectionDiff.Keys, KeyDiff{
				Key:    key.Name(),
				Old:    key.Value(),
				New:    "",
				Status: DiffStatusDeleted,
			})
		}
	}

	diffs := make([]SectionDiff, 0, len(bySection))
	for _, sectionDiff := range bySection {
		sort.Slice(sectionDiff.Keys, func(i, j int) bool {
			return sectionDiff.Keys[i].Key < sectionDiff.Keys[j].Key
		})
		diffs = append(diffs, *sectionDiff)
	}

	sort.Slice(diffs, func(i, j int) bool {
		return diffs[i].Section < diffs[j].Section
	})

	return diffs, nil
}

// ApplyChangesToConfig applies external changes into in-memory config and saves it.
func (e *ConfigEditor) ApplyChangesToConfig(diff []SectionDiff) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.config == nil || e.config.Path() == "" {
		return ErrConfigNotLoaded
	}

	for _, sectionDiff := range diff {
		if sectionDiff.Section == "" {
			return fmt.Errorf("section name is empty")
		}

		if sectionDiff.Status == DiffStatusDeleted {
			e.config.file.DeleteSection(sectionDiff.Section)
			continue
		}

		section, exists := getSectionIfExists(e.config.file, sectionDiff.Section)
		if !exists {
			created, err := e.config.file.NewSection(sectionDiff.Section)
			if err != nil {
				return fmt.Errorf("failed to create section %s: %w", sectionDiff.Section, err)
			}
			section = created
		}

		for _, keyDiff := range sectionDiff.Keys {
			if keyDiff.Key == "" {
				return fmt.Errorf("key name is empty in section %s", sectionDiff.Section)
			}

			switch keyDiff.Status {
			case DiffStatusAdded, DiffStatusModified:
				section.Key(keyDiff.Key).SetValue(keyDiff.New)
			case DiffStatusDeleted:
				section.DeleteKey(keyDiff.Key)
			default:
				return fmt.Errorf("unknown diff status for %s.%s: %s", sectionDiff.Section, keyDiff.Key, keyDiff.Status)
			}
		}
	}

	e.config.rebuildKeyMap()
	return e.config.Save()
}

func getSectionIfExists(file *ini.File, name string) (*ini.Section, bool) {
	if !file.HasSection(name) {
		return nil, false
	}

	section, err := file.GetSection(name)
	if err != nil {
		return nil, false
	}

	return section, true
}

// DiscardExternalChanges overwrites the on-disk config with current in-memory state.
func (e *ConfigEditor) DiscardExternalChanges() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.config == nil || e.config.Path() == "" {
		return ErrConfigNotLoaded
	}

	err := e.config.Save()
	if err != nil {
		return fmt.Errorf("не удалось перезаписать конфиг на диске: %w", err)
	}
	return nil
}
