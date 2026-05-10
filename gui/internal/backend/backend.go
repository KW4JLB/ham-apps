// Package backend provides the filesystem-backed implementation of port.AppRepository.
// It has no dependency on Fyne — it is pure Go and can be tested without a display.
package backend

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kw4jlb/ham-apps/gui/internal/port"
)

// FilesystemRepository implements port.AppRepository by reading the ham-apps
// directory layout from the local filesystem.
//
// HamappsDir is the root of the ham-apps installation (same as $HAMAPPS_DIR).
// StatusDir is the directory where install-state files are stored
// (normally ~/.local/share/ham-apps/installed/).
type FilesystemRepository struct {
	HamappsDir string
	StatusDir  string
}

// ---------------------------------------------------------------------------
// port.AppRepository implementation
// ---------------------------------------------------------------------------

// LoadApps scans the apps/ subdirectory of HamappsDir and returns all valid
// apps. Per-app errors (invalid slug, missing metadata) are collected into the
// second return value; the first return value contains all apps that could be
// parsed successfully.
func (r *FilesystemRepository) LoadApps() ([]port.AppInfo, []error) {
	appsDir := filepath.Join(r.HamappsDir, "apps")
	entries, err := os.ReadDir(appsDir)
	if err != nil {
		return nil, []error{fmt.Errorf("LoadApps: cannot read apps dir %q: %w", appsDir, err)}
	}

	var apps []port.AppInfo
	var errs []error

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		slug := entry.Name()
		if err := ValidateSlug(slug); err != nil {
			errs = append(errs, fmt.Errorf("LoadApps: skipping invalid slug %q: %w", slug, err))
			continue
		}

		appDir := filepath.Join(appsDir, slug)
		info, err := ParseMetadata(appDir)
		if err != nil {
			errs = append(errs, fmt.Errorf("LoadApps: slug %q: %w", slug, err))
			continue
		}
		info.Slug = slug

		// Load description (optional — no error if absent).
		descPath := filepath.Join(appDir, "description")
		if descBytes, err := os.ReadFile(descPath); err == nil {
			desc := string(descBytes)
			info.Description = desc
			if idx := strings.IndexByte(desc, '\n'); idx >= 0 {
				info.Summary = strings.TrimSpace(desc[:idx])
			} else {
				info.Summary = strings.TrimSpace(desc)
			}
		}

		// Check install state.
		info.Installed = r.IsInstalled(slug)

		// Record icon path if it exists.
		iconPath := filepath.Join(appDir, "icon.png")
		if _, err := os.Stat(iconPath); err == nil {
			info.IconPath = iconPath
		}

		apps = append(apps, info)
	}

	return apps, errs
}

// LoadCategories parses data/categories (pipe-delimited; # comments skipped).
func (r *FilesystemRepository) LoadCategories() ([]port.Category, error) {
	path := filepath.Join(r.HamappsDir, "data", "categories")
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("LoadCategories: %w", err)
	}
	defer f.Close()

	var cats []port.Category
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "|", 3)
		if len(parts) < 2 {
			continue
		}
		cat := port.Category{
			ID:          strings.TrimSpace(parts[0]),
			DisplayName: strings.TrimSpace(parts[1]),
		}
		if len(parts) == 3 {
			cat.Description = strings.TrimSpace(parts[2])
		}
		cats = append(cats, cat)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("LoadCategories: scanner: %w", err)
	}
	return cats, nil
}

// IsInstalled returns true if the state file for the given slug exists.
func (r *FilesystemRepository) IsInstalled(slug string) bool {
	_, err := os.Stat(filepath.Join(r.StatusDir, slug))
	return err == nil
}

// MarkInstalled creates the install-state file for slug.
func (r *FilesystemRepository) MarkInstalled(slug string) error {
	if err := os.MkdirAll(r.StatusDir, 0755); err != nil {
		return fmt.Errorf("MarkInstalled: mkdir: %w", err)
	}
	path := filepath.Join(r.StatusDir, slug)
	if err := os.WriteFile(path, []byte{}, 0644); err != nil {
		return fmt.Errorf("MarkInstalled: write: %w", err)
	}
	return nil
}

// MarkUninstalled removes the install-state file for slug.
// Returns nil if the file is already absent.
func (r *FilesystemRepository) MarkUninstalled(slug string) error {
	err := os.Remove(filepath.Join(r.StatusDir, slug))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("MarkUninstalled: %w", err)
	}
	return nil
}

// ReadVersion reads the version file and returns the trimmed version string.
// Returns "dev" if the file is absent or unreadable.
func (r *FilesystemRepository) ReadVersion() string {
	data, err := os.ReadFile(filepath.Join(r.HamappsDir, "version"))
	if err != nil {
		return "dev"
	}
	return strings.TrimSpace(string(data))
}

// ScriptsDir returns the absolute path to the scripts/ directory.
func (r *FilesystemRepository) ScriptsDir() string {
	return filepath.Join(r.HamappsDir, "scripts")
}

// LoadIcon returns the raw bytes of apps/<slug>/icon.png, or nil if the file
// is absent or the path escapes the apps/ directory.
func (r *FilesystemRepository) LoadIcon(slug string) []byte {
	appsBase := filepath.Clean(filepath.Join(r.HamappsDir, "apps"))
	iconPath := filepath.Clean(filepath.Join(appsBase, slug, "icon.png"))

	// SC-04: confine path to apps/ subdirectory.
	if !strings.HasPrefix(iconPath, appsBase+string(filepath.Separator)) {
		return nil
	}

	data, err := os.ReadFile(iconPath)
	if err != nil {
		return nil
	}
	return data
}

// ---------------------------------------------------------------------------
// Exported helpers (used by tests and other packages)
// ---------------------------------------------------------------------------

var slugPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

// ValidateSlug returns an error if slug does not match the allowed pattern.
// Valid slugs: start with alphanumeric, then any mix of alphanum, dash, underscore.
func ValidateSlug(slug string) error {
	if !slugPattern.MatchString(slug) {
		return fmt.Errorf("invalid app slug: %q (must match ^[a-zA-Z0-9][a-zA-Z0-9_-]*$)", slug)
	}
	return nil
}

// ValidateHamappsDir ensures that dir is an absolute path containing the
// expected ham-apps directory structure.
func ValidateHamappsDir(dir string) error {
	if !filepath.IsAbs(dir) {
		return fmt.Errorf("HAMAPPS_DIR must be an absolute path, got: %q", dir)
	}
	required := []string{"apps", "scripts", filepath.Join("data", "categories")}
	for _, rel := range required {
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			return fmt.Errorf("HAMAPPS_DIR %q missing required path %q", dir, rel)
		}
	}
	return nil
}

// ParseMetadata reads apps/<slug>/metadata from appDir and returns an AppInfo
// with all populated fields. The Slug field is NOT set here (caller sets it).
func ParseMetadata(appDir string) (port.AppInfo, error) {
	path := filepath.Join(appDir, "metadata")
	f, err := os.Open(path)
	if err != nil {
		return port.AppInfo{}, fmt.Errorf("ParseMetadata: open %q: %w", path, err)
	}
	defer f.Close()

	var info port.AppInfo
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			// Malformed line: skip silently (UT-03).
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		switch key {
		case "name":
			info.Name = val
		case "category":
			info.Category = val
		case "website":
			info.Website = val
		case "tags":
			if val != "" {
				rawTags := strings.Split(val, ",")
				for _, tag := range rawTags {
					trimmed := strings.TrimSpace(tag)
					if trimmed != "" {
						info.Tags = append(info.Tags, trimmed)
					}
				}
			}
		case "min-os":
			info.MinOS = val
		}
	}
	if err := scanner.Err(); err != nil {
		return port.AppInfo{}, fmt.Errorf("ParseMetadata: scan %q: %w", path, err)
	}
	return info, nil
}
