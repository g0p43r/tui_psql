package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/g0p43r/tui_psql/internal/domain"
)

const (
	appDir       = "tui_psql"
	profilesFile = "profiles.json"
)

func LoadProfiles() ([]domain.ConnectionProfile, error) {
	path, err := profilesPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read profiles: %w", err)
	}

	var profiles []domain.ConnectionProfile
	if err := json.Unmarshal(data, &profiles); err != nil {
		return nil, fmt.Errorf("decode profiles: %w", err)
	}

	for i := range profiles {
		profiles[i] = NormalizeProfile(profiles[i])
	}

	return profiles, nil
}

func SaveProfile(profile domain.ConnectionProfile) ([]domain.ConnectionProfile, error) {
	profile = NormalizeProfile(profile)
	profile.Password = ""

	profiles, err := LoadProfiles()
	if err != nil {
		return nil, err
	}

	index := slices.IndexFunc(profiles, func(candidate domain.ConnectionProfile) bool {
		return candidate.Name == profile.Name
	})

	if index >= 0 {
		profiles[index] = profile
	} else {
		profiles = append(profiles, profile)
	}

	path, err := profilesPath()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create config dir: %w", err)
	}

	data, err := json.MarshalIndent(profiles, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode profiles: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return nil, fmt.Errorf("write profiles: %w", err)
	}

	return profiles, nil
}

func DeleteProfile(name string) ([]domain.ConnectionProfile, error) {
	profiles, err := LoadProfiles()
	if err != nil {
		return nil, err
	}

	index := slices.IndexFunc(profiles, func(candidate domain.ConnectionProfile) bool {
		return candidate.Name == name
	})
	if index < 0 {
		return profiles, nil
	}

	profiles = append(profiles[:index], profiles[index+1:]...)

	path, err := profilesPath()
	if err != nil {
		return nil, err
	}

	if len(profiles) == 0 {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("remove profiles file: %w", err)
		}
		return profiles, nil
	}

	data, err := json.MarshalIndent(profiles, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode profiles: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return nil, fmt.Errorf("write profiles: %w", err)
	}

	return profiles, nil
}

func NormalizeProfile(profile domain.ConnectionProfile) domain.ConnectionProfile {
	profile.Password = ""
	if profile.SSLMode == "" {
		profile.SSLMode = "disable"
	}
	if profile.Name == "" {
		profile.Name = fmt.Sprintf("%s@%s:%s/%s", profile.User, profile.Host, profile.Port, profile.Database)
	}
	return profile
}

func profilesPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}

	return filepath.Join(configDir, appDir, profilesFile), nil
}
