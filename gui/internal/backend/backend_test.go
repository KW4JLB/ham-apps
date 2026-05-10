// Package backend_test contains unit tests for the backend package.
// Tests are written TDD-first (red phase) before the implementation exists.
package backend_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kw4jlb/ham-apps/gui/internal/backend"
	"github.com/kw4jlb/ham-apps/gui/internal/port"
	// Compile-time interface check for runner (UT-30) lives in runner_test.go.
	// Compile-time interface check for repository (UT-29) is below.
)

// Compile-time check: FilesystemRepository must implement port.AppRepository (UT-29).
var _ port.AppRepository = (*backend.FilesystemRepository)(nil)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// makeMetadataFile creates apps/<slug>/metadata inside dir with the given fields.
func makeMetadataFile(t *testing.T, dir, slug string, fields map[string]string) {
	t.Helper()
	appDir := filepath.Join(dir, "apps", slug)
	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatalf("makeMetadataFile MkdirAll: %v", err)
	}
	var sb strings.Builder
	for k, v := range fields {
		sb.WriteString(k + "=" + v + "\n")
	}
	if err := os.WriteFile(filepath.Join(appDir, "metadata"), []byte(sb.String()), 0644); err != nil {
		t.Fatalf("makeMetadataFile WriteFile: %v", err)
	}
}

// makeDescriptionFile creates apps/<slug>/description with the given text.
func makeDescriptionFile(t *testing.T, dir, slug, text string) {
	t.Helper()
	appDir := filepath.Join(dir, "apps", slug)
	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatalf("makeDescriptionFile MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(appDir, "description"), []byte(text), 0644); err != nil {
		t.Fatalf("makeDescriptionFile WriteFile: %v", err)
	}
}

// makeValidHamappsDir creates a minimal valid hamappsDir structure in tmpDir.
func makeValidHamappsDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, sub := range []string{"apps", "scripts"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0755); err != nil {
			t.Fatalf("makeValidHamappsDir MkdirAll %s: %v", sub, err)
		}
	}
	if err := os.MkdirAll(filepath.Join(dir, "data"), 0755); err != nil {
		t.Fatalf("makeValidHamappsDir MkdirAll data: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "data", "categories"), []byte("digital-modes|Digital Modes|desc\n"), 0644); err != nil {
		t.Fatalf("makeValidHamappsDir WriteFile categories: %v", err)
	}
	return dir
}

// newRepo is a shorthand for creating a FilesystemRepository for testing.
func newRepo(hamappsDir, statusDir string) *backend.FilesystemRepository {
	return &backend.FilesystemRepository{
		HamappsDir: hamappsDir,
		StatusDir:  statusDir,
	}
}

// ---------------------------------------------------------------------------
// UT-01: metadata parser — all fields populated
// ---------------------------------------------------------------------------

func TestParseMetadata_AllFields(t *testing.T) {
	dir := t.TempDir()
	makeMetadataFile(t, dir, "wsjtx", map[string]string{
		"name":     "WSJT-X",
		"category": "digital-modes",
		"website":  "https://wsjt.sourceforge.io",
		"tags":     "ft8,ft4,wspr",
		"min-os":   "debian:11",
	})

	info, err := backend.ParseMetadata(filepath.Join(dir, "apps", "wsjtx"))
	if err != nil {
		t.Fatalf("ParseMetadata: unexpected error: %v", err)
	}
	if info.Name != "WSJT-X" {
		t.Errorf("Name: got %q, want %q", info.Name, "WSJT-X")
	}
	if info.Category != "digital-modes" {
		t.Errorf("Category: got %q, want %q", info.Category, "digital-modes")
	}
	if info.Website != "https://wsjt.sourceforge.io" {
		t.Errorf("Website: got %q, want %q", info.Website, "https://wsjt.sourceforge.io")
	}
	if len(info.Tags) != 3 || info.Tags[0] != "ft8" {
		t.Errorf("Tags: got %v, want [ft8 ft4 wspr]", info.Tags)
	}
	if info.MinOS != "debian:11" {
		t.Errorf("MinOS: got %q, want %q", info.MinOS, "debian:11")
	}
}

// ---------------------------------------------------------------------------
// UT-02: metadata parser — missing optional fields
// ---------------------------------------------------------------------------

func TestParseMetadata_MissingOptionalFields(t *testing.T) {
	dir := t.TempDir()
	makeMetadataFile(t, dir, "minapp", map[string]string{
		"name":     "MinApp",
		"category": "tools",
	})

	info, err := backend.ParseMetadata(filepath.Join(dir, "apps", "minapp"))
	if err != nil {
		t.Fatalf("ParseMetadata: unexpected error: %v", err)
	}
	if info.Website != "" {
		t.Errorf("Website should be empty, got %q", info.Website)
	}
	if len(info.Tags) != 0 {
		t.Errorf("Tags should be empty, got %v", info.Tags)
	}
	if info.MinOS != "" {
		t.Errorf("MinOS should be empty, got %q", info.MinOS)
	}
}

// ---------------------------------------------------------------------------
// UT-03: metadata parser — malformed line ignored
// ---------------------------------------------------------------------------

func TestParseMetadata_MalformedLineIgnored(t *testing.T) {
	dir := t.TempDir()
	appDir := filepath.Join(dir, "apps", "broken")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	content := "name=GoodApp\nNO_EQUALS_SIGN_HERE\ncategory=tools\n"
	if err := os.WriteFile(filepath.Join(appDir, "metadata"), []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	info, err := backend.ParseMetadata(appDir)
	if err != nil {
		t.Fatalf("ParseMetadata: unexpected error: %v", err)
	}
	if info.Name != "GoodApp" {
		t.Errorf("Name: got %q, want %q", info.Name, "GoodApp")
	}
	if info.Category != "tools" {
		t.Errorf("Category: got %q, want %q", info.Category, "tools")
	}
}

// ---------------------------------------------------------------------------
// UT-04: category parser — standard file (uses actual data/categories)
// ---------------------------------------------------------------------------

func TestLoadCategories_StandardFile(t *testing.T) {
	// Locate the repo root relative to this test file's package.
	// The test binary runs in the package directory: gui/internal/backend/
	// We go up 3 levels: backend -> internal -> gui -> ham-apps (repo root).
	repoRoot := filepath.Join("..", "..", "..")
	hamappsDir, err := filepath.Abs(repoRoot)
	if err != nil {
		t.Fatalf("Abs: %v", err)
	}

	repo := newRepo(hamappsDir, t.TempDir())
	cats, err := repo.LoadCategories()
	if err != nil {
		t.Fatalf("LoadCategories: %v", err)
	}
	if len(cats) != 9 {
		t.Errorf("category count: got %d, want 9", len(cats))
	}
	if cats[0].ID != "digital-modes" {
		t.Errorf("first category ID: got %q, want %q", cats[0].ID, "digital-modes")
	}
}

// ---------------------------------------------------------------------------
// UT-05: category parser — comment lines skipped
// ---------------------------------------------------------------------------

func TestLoadCategories_CommentLinesSkipped(t *testing.T) {
	dir := t.TempDir()
	content := "# this is a comment\ndigital-modes|Digital Modes|desc\n"
	if err := os.MkdirAll(filepath.Join(dir, "data"), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "data", "categories"), []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	repo := newRepo(dir, t.TempDir())
	cats, err := repo.LoadCategories()
	if err != nil {
		t.Fatalf("LoadCategories: %v", err)
	}
	if len(cats) != 1 {
		t.Errorf("category count: got %d, want 1 (comment should be skipped)", len(cats))
	}
}

// ---------------------------------------------------------------------------
// UT-06: IsInstalled — file exists
// ---------------------------------------------------------------------------

func TestIsInstalled_FileExists(t *testing.T) {
	statusDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(statusDir, "wsjtx"), []byte{}, 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	repo := newRepo(t.TempDir(), statusDir)
	if !repo.IsInstalled("wsjtx") {
		t.Error("IsInstalled: expected true, got false")
	}
}

// ---------------------------------------------------------------------------
// UT-07: IsInstalled — file absent
// ---------------------------------------------------------------------------

func TestIsInstalled_FileAbsent(t *testing.T) {
	statusDir := t.TempDir()
	repo := newRepo(t.TempDir(), statusDir)
	if repo.IsInstalled("wsjtx") {
		t.Error("IsInstalled: expected false, got true")
	}
}

// ---------------------------------------------------------------------------
// UT-08: LoadApps — discovers all apps
// ---------------------------------------------------------------------------

func TestLoadApps_DiscoversAllApps(t *testing.T) {
	dir := makeValidHamappsDir(t)
	slugs := []string{"wsjtx", "fldigi", "direwolf"}
	for _, slug := range slugs {
		makeMetadataFile(t, dir, slug, map[string]string{"name": slug, "category": "digital-modes"})
		makeDescriptionFile(t, dir, slug, "Short description\nFull text")
	}

	repo := newRepo(dir, t.TempDir())
	apps, errs := repo.LoadApps()
	if len(errs) != 0 {
		t.Errorf("LoadApps errs: got %d errors: %v", len(errs), errs)
	}
	if len(apps) != 3 {
		t.Errorf("LoadApps: got %d apps, want 3", len(apps))
	}
	found := make(map[string]bool)
	for _, a := range apps {
		found[a.Slug] = true
	}
	for _, slug := range slugs {
		if !found[slug] {
			t.Errorf("LoadApps: slug %q not found in results", slug)
		}
	}
}

// ---------------------------------------------------------------------------
// UT-09: LoadApps — missing description file
// ---------------------------------------------------------------------------

func TestLoadApps_MissingDescriptionFile(t *testing.T) {
	dir := makeValidHamappsDir(t)
	makeMetadataFile(t, dir, "nodesc", map[string]string{"name": "NoDesc", "category": "tools"})
	// No description file created.

	repo := newRepo(dir, t.TempDir())
	apps, _ := repo.LoadApps()
	if len(apps) != 1 {
		t.Fatalf("LoadApps: got %d apps, want 1", len(apps))
	}
	if apps[0].Description != "" {
		t.Errorf("Description: got %q, want empty string", apps[0].Description)
	}
}

// ---------------------------------------------------------------------------
// UT-10: LoadApps — missing metadata file (partial-failure semantics)
// ---------------------------------------------------------------------------

func TestLoadApps_MissingMetadataFile(t *testing.T) {
	dir := makeValidHamappsDir(t)

	// Valid app.
	makeMetadataFile(t, dir, "valid", map[string]string{"name": "Valid", "category": "tools"})
	makeDescriptionFile(t, dir, "valid", "Valid app")

	// Invalid app: directory exists but no metadata file.
	invalidDir := filepath.Join(dir, "apps", "invalid")
	if err := os.MkdirAll(invalidDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	repo := newRepo(dir, t.TempDir())
	apps, errs := repo.LoadApps()
	if len(apps) != 1 {
		t.Errorf("LoadApps apps: got %d, want 1 (valid app only)", len(apps))
	}
	if len(errs) == 0 {
		t.Error("LoadApps errs: expected at least one error for missing metadata, got none")
	}
}

// ---------------------------------------------------------------------------
// UT-11: app list filtering — by category
// ---------------------------------------------------------------------------

func TestFilterApps_ByCategory(t *testing.T) {
	apps := []port.AppInfo{
		{Slug: "wsjtx", Name: "WSJT-X", Category: "digital-modes"},
		{Slug: "fldigi", Name: "Fldigi", Category: "digital-modes"},
		{Slug: "direwolf", Name: "Direwolf", Category: "packet-aprs"},
	}
	result := backend.FilterApps(apps, "digital-modes", "")
	if len(result) != 2 {
		t.Errorf("FilterApps by category: got %d, want 2", len(result))
	}
}

// ---------------------------------------------------------------------------
// UT-12: app list filtering — by search text (name match)
// ---------------------------------------------------------------------------

func TestFilterApps_ByNameCaseInsensitive(t *testing.T) {
	apps := []port.AppInfo{
		{Slug: "wsjtx", Name: "WSJT-X", Category: "digital-modes"},
		{Slug: "fldigi", Name: "Fldigi", Category: "digital-modes"},
		{Slug: "direwolf", Name: "Direwolf", Category: "packet-aprs"},
	}
	result := backend.FilterApps(apps, "", "wsjtx")
	if len(result) != 1 {
		t.Errorf("FilterApps by name: got %d, want 1", len(result))
	}
	if result[0].Slug != "wsjtx" {
		t.Errorf("FilterApps by name: got slug %q, want %q", result[0].Slug, "wsjtx")
	}
}

// ---------------------------------------------------------------------------
// UT-13: app list filtering — by search text (description match)
// ---------------------------------------------------------------------------

func TestFilterApps_ByDescription(t *testing.T) {
	apps := []port.AppInfo{
		{Slug: "wsjtx", Name: "WSJT-X", Category: "digital-modes", Description: "Weak signal digital modes by K1JT"},
		{Slug: "fldigi", Name: "Fldigi", Category: "digital-modes", Description: "Multi-mode software modem"},
	}
	result := backend.FilterApps(apps, "", "weak signal")
	if len(result) != 1 {
		t.Errorf("FilterApps by description: got %d, want 1", len(result))
	}
	if result[0].Slug != "wsjtx" {
		t.Errorf("FilterApps by description: got slug %q, want wsjtx", result[0].Slug)
	}
}

// ---------------------------------------------------------------------------
// UT-14: app list filtering — combined category + search
// ---------------------------------------------------------------------------

func TestFilterApps_CombinedCategoryAndSearch(t *testing.T) {
	apps := []port.AppInfo{
		{Slug: "wsjtx", Name: "WSJT-X", Category: "digital-modes"},
		{Slug: "fldigi", Name: "Fldigi", Category: "digital-modes"},
		{Slug: "direwolf", Name: "Direwolf", Category: "packet-aprs"},
		{Slug: "xastir", Name: "Xastir", Category: "packet-aprs"},
	}
	result := backend.FilterApps(apps, "digital-modes", "fldigi")
	if len(result) != 1 {
		t.Errorf("FilterApps combined: got %d, want 1", len(result))
	}
	if result[0].Slug != "fldigi" {
		t.Errorf("FilterApps combined: got slug %q, want fldigi", result[0].Slug)
	}
}

// ---------------------------------------------------------------------------
// UT-15: version reading — file present
// ---------------------------------------------------------------------------

func TestReadVersion_FilePresent(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "version"), []byte("0.3.0\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	repo := newRepo(dir, t.TempDir())
	v := repo.ReadVersion()
	if v != "0.3.0" {
		t.Errorf("ReadVersion: got %q, want %q", v, "0.3.0")
	}
}

// ---------------------------------------------------------------------------
// UT-16: version reading — file absent
// ---------------------------------------------------------------------------

func TestReadVersion_FileAbsent(t *testing.T) {
	dir := t.TempDir()
	repo := newRepo(dir, t.TempDir())
	v := repo.ReadVersion()
	if v != "dev" {
		t.Errorf("ReadVersion: got %q, want %q", v, "dev")
	}
}

// ---------------------------------------------------------------------------
// UT-17: install state write — mark installed
// ---------------------------------------------------------------------------

func TestMarkInstalled(t *testing.T) {
	statusDir := t.TempDir()
	repo := newRepo(t.TempDir(), statusDir)
	if err := repo.MarkInstalled("wsjtx"); err != nil {
		t.Fatalf("MarkInstalled: %v", err)
	}
	if _, err := os.Stat(filepath.Join(statusDir, "wsjtx")); err != nil {
		t.Errorf("MarkInstalled: state file not created: %v", err)
	}
}

// ---------------------------------------------------------------------------
// UT-18: install state write — mark uninstalled
// ---------------------------------------------------------------------------

func TestMarkUninstalled(t *testing.T) {
	statusDir := t.TempDir()
	// Create the state file first.
	if err := os.WriteFile(filepath.Join(statusDir, "wsjtx"), []byte{}, 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	repo := newRepo(t.TempDir(), statusDir)
	if err := repo.MarkUninstalled("wsjtx"); err != nil {
		t.Fatalf("MarkUninstalled: %v", err)
	}
	if _, err := os.Stat(filepath.Join(statusDir, "wsjtx")); !os.IsNotExist(err) {
		t.Error("MarkUninstalled: state file still exists or unexpected error")
	}

	// Second call should not error (already absent).
	if err := repo.MarkUninstalled("wsjtx"); err != nil {
		t.Errorf("MarkUninstalled (already absent): got error %v, want nil", err)
	}
}

// ---------------------------------------------------------------------------
// UT-24: slug validation — valid slug passes
// ---------------------------------------------------------------------------

func TestValidateSlug_Valid(t *testing.T) {
	validSlugs := []string{"wsjtx", "wsjt-x", "wsjtx2", "my_app", "a", "wsjtx-2"}
	for _, slug := range validSlugs {
		if err := backend.ValidateSlug(slug); err != nil {
			t.Errorf("ValidateSlug(%q): unexpected error: %v", slug, err)
		}
	}
}

// ---------------------------------------------------------------------------
// UT-25: slug validation — invalid slug rejected
// ---------------------------------------------------------------------------

func TestValidateSlug_Invalid(t *testing.T) {
	invalidSlugs := []string{
		"../../../etc/passwd",
		"",
		"-starts-with-dash",
		"has space",
		"has/slash",
		"has\nnewline",
		"has$dollar",
	}
	for _, slug := range invalidSlugs {
		if err := backend.ValidateSlug(slug); err == nil {
			t.Errorf("ValidateSlug(%q): expected error, got nil", slug)
		}
	}
}

// ---------------------------------------------------------------------------
// UT-26: ValidateHamappsDir — valid directory passes
// ---------------------------------------------------------------------------

func TestValidateHamappsDir_Valid(t *testing.T) {
	dir := makeValidHamappsDir(t)
	if err := backend.ValidateHamappsDir(dir); err != nil {
		t.Errorf("ValidateHamappsDir: unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// UT-27: ValidateHamappsDir — relative path rejected
// ---------------------------------------------------------------------------

func TestValidateHamappsDir_RelativePath(t *testing.T) {
	err := backend.ValidateHamappsDir("./ham-apps")
	if err == nil {
		t.Fatal("ValidateHamappsDir: expected error for relative path, got nil")
	}
	if !strings.Contains(err.Error(), "absolute path") {
		t.Errorf("ValidateHamappsDir: error %q should mention 'absolute path'", err.Error())
	}
}

// ---------------------------------------------------------------------------
// UT-28: ValidateHamappsDir — missing required subdirectory rejected
// ---------------------------------------------------------------------------

func TestValidateHamappsDir_MissingSubdir(t *testing.T) {
	dir := t.TempDir()
	// Create apps/ and scripts/ but NOT data/categories.
	for _, sub := range []string{"apps", "scripts"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
	}
	err := backend.ValidateHamappsDir(dir)
	if err == nil {
		t.Fatal("ValidateHamappsDir: expected error for missing data/categories, got nil")
	}
	if !strings.Contains(err.Error(), "data/categories") && !strings.Contains(err.Error(), "categories") {
		t.Errorf("ValidateHamappsDir: error %q should mention missing path", err.Error())
	}
}

// ---------------------------------------------------------------------------
// LoadIcon tests
// ---------------------------------------------------------------------------

// TestLoadIcon_Absent verifies nil is returned (not error) when icon.png is missing.
func TestLoadIcon_Absent(t *testing.T) {
	dir := makeValidHamappsDir(t)
	makeMetadataFile(t, dir, "noicon", map[string]string{"name": "NoIcon", "category": "tools"})
	repo := newRepo(dir, t.TempDir())
	data := repo.LoadIcon("noicon")
	if data != nil {
		t.Errorf("LoadIcon absent: expected nil, got %d bytes", len(data))
	}
}

// TestLoadIcon_Present verifies that an existing icon.png is loaded.
func TestLoadIcon_Present(t *testing.T) {
	dir := makeValidHamappsDir(t)
	iconData := []byte{0x89, 0x50, 0x4E, 0x47} // PNG magic bytes
	iconPath := filepath.Join(dir, "apps", "iconapp", "icon.png")
	if err := os.MkdirAll(filepath.Dir(iconPath), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(iconPath, iconData, 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	makeMetadataFile(t, dir, "iconapp", map[string]string{"name": "IconApp", "category": "tools"})

	repo := newRepo(dir, t.TempDir())
	data := repo.LoadIcon("iconapp")
	if data == nil {
		t.Fatal("LoadIcon present: expected data, got nil")
	}
	if len(data) < 4 {
		t.Errorf("LoadIcon present: got %d bytes, want at least 4", len(data))
	}
}

// TestLoadIcon_PathTraversal verifies that a path-traversal slug is rejected.
func TestLoadIcon_PathTraversal(t *testing.T) {
	dir := makeValidHamappsDir(t)
	repo := newRepo(dir, t.TempDir())
	// This slug would pass ValidateSlug (it won't), so we test with a slug that
	// was somehow let through — the icon loader must also check the path.
	// We use a slug with a valid format that resolves to a traversal after Join+Clean.
	// Because ValidateSlug blocks "../", we test the confinement logic directly.
	data := repo.LoadIcon("../../../etc/passwd")
	if data != nil {
		t.Error("LoadIcon path traversal: expected nil, got data")
	}
}

// ---------------------------------------------------------------------------
// Summary description parsing
// ---------------------------------------------------------------------------

// TestLoadApps_SummaryAndDescription verifies that Summary is the first line of description.
func TestLoadApps_SummaryAndDescription(t *testing.T) {
	dir := makeValidHamappsDir(t)
	makeMetadataFile(t, dir, "summtest", map[string]string{"name": "SummTest", "category": "tools"})
	makeDescriptionFile(t, dir, "summtest", "First line is summary\nSecond line is extra\n")

	repo := newRepo(dir, t.TempDir())
	apps, _ := repo.LoadApps()
	if len(apps) != 1 {
		t.Fatalf("LoadApps: got %d apps, want 1", len(apps))
	}
	if apps[0].Summary != "First line is summary" {
		t.Errorf("Summary: got %q, want %q", apps[0].Summary, "First line is summary")
	}
	if !strings.Contains(apps[0].Description, "Second line") {
		t.Errorf("Description: expected to contain 'Second line', got %q", apps[0].Description)
	}
}

// ---------------------------------------------------------------------------
// FilterApps — "All" category shows everything
// ---------------------------------------------------------------------------

func TestFilterApps_AllCategory(t *testing.T) {
	apps := []port.AppInfo{
		{Slug: "wsjtx", Name: "WSJT-X", Category: "digital-modes"},
		{Slug: "direwolf", Name: "Direwolf", Category: "packet-aprs"},
	}
	result := backend.FilterApps(apps, "All", "")
	if len(result) != 2 {
		t.Errorf("FilterApps All: got %d, want 2", len(result))
	}
	result2 := backend.FilterApps(apps, "", "")
	if len(result2) != 2 {
		t.Errorf("FilterApps empty category: got %d, want 2", len(result2))
	}
}
