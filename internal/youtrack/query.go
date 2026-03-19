package youtrack

import "strings"

// BuildQuery constructs a YouTrack search query from individual filters.
// Multi-word values (e.g. state "In Progress") are wrapped in braces.
func BuildQuery(project, state, assignee, query string) string {
	var parts []string
	if project != "" {
		parts = append(parts, "project: "+project)
	}
	if state != "" {
		parts = append(parts, "State: {"+state+"}")
	}
	if assignee != "" {
		parts = append(parts, "Assignee: "+assignee)
	}
	if query != "" {
		parts = append(parts, query)
	}
	return strings.Join(parts, " ")
}
