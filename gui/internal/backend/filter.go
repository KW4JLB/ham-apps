package backend

import (
	"strings"

	"github.com/kw4jlb/ham-apps/gui/internal/port"
)

// FilterApps returns the subset of apps that match the given category and
// search string. Both filters are applied simultaneously.
//
// category: exact match against AppInfo.Category; empty string or "All" means
// no category filter (show all categories).
//
// search: case-insensitive substring match across Name, Category, and
// Description fields; empty string means no search filter.
func FilterApps(apps []port.AppInfo, category, search string) []port.AppInfo {
	searchLower := strings.ToLower(search)
	filterCat := category != "" && category != "All"

	var result []port.AppInfo
	for _, app := range apps {
		// Category filter.
		if filterCat && app.Category != category {
			continue
		}
		// Search filter: case-insensitive across Name, Category, Description, and Slug.
		if searchLower != "" {
			slugLower := strings.ToLower(app.Slug)
			nameLower := strings.ToLower(app.Name)
			catLower := strings.ToLower(app.Category)
			descLower := strings.ToLower(app.Description)
			if !strings.Contains(slugLower, searchLower) &&
				!strings.Contains(nameLower, searchLower) &&
				!strings.Contains(catLower, searchLower) &&
				!strings.Contains(descLower, searchLower) {
				continue
			}
		}
		result = append(result, app)
	}
	return result
}
