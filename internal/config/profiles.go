package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/g0p43r/tui_psql/internal/domain"
	"github.com/g0p43r/tui_psql/internal/errs"
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
		return nil, errs.E(errs.CodeConfig, "config.LoadProfiles.ReadFile", "Failed to read saved profiles.", err)
	}

	var profiles []domain.ConnectionProfile
	if err := json.Unmarshal(data, &profiles); err != nil {
		return nil, errs.E(errs.CodeConfig, "config.LoadProfiles.Unmarshal", "Saved profiles file is invalid.", err)
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
		return nil, errs.E(errs.CodeConfig, "config.SaveProfile.MkdirAll", "Failed to create config directory.", err)
	}

	data, err := json.MarshalIndent(profiles, "", "  ")
	if err != nil {
		return nil, errs.E(errs.CodeConfig, "config.SaveProfile.Marshal", "Failed to encode profiles.", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return nil, errs.E(errs.CodeConfig, "config.SaveProfile.WriteFile", "Failed to write profiles.", err)
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
			return nil, errs.E(errs.CodeConfig, "config.DeleteProfile.Remove", "Failed to remove profiles file.", err)
		}
		return profiles, nil
	}

	data, err := json.MarshalIndent(profiles, "", "  ")
	if err != nil {
		return nil, errs.E(errs.CodeConfig, "config.DeleteProfile.Marshal", "Failed to encode profiles.", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return nil, errs.E(errs.CodeConfig, "config.DeleteProfile.WriteFile", "Failed to write profiles.", err)
	}

	return profiles, nil
}

func NormalizeProfile(profile domain.ConnectionProfile) domain.ConnectionProfile {
	profile.Password = ""
	profile.SSLMode = strings.ToLower(strings.TrimSpace(profile.SSLMode))
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
		return "", errs.E(errs.CodeConfig, "config.profilesPath.UserConfigDir", "Failed to resolve user config directory.", err)
	}

	return filepath.Join(configDir, appDir, profilesFile), nil
}
