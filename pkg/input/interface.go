package input

// Interface interface for command line input
type Interface interface {

	// PickValue pick a value
	PickValue(message string, defaultValue string, required bool, help string) (string, error)

	// PickPassword pick a password
	PickPassword(message string, help string) (string, error)

	// PickNameWithDefault pick a name with a default value
	PickNameWithDefault(names []string, message string, defaultValue string, help string) (string, error)

	// SelectNamesWithFilter selects zero or more names with a filter string
	SelectNamesWithFilter(names []string, message string, selectAll bool, filter string, help string) ([]string, error)

	// SelectNames selects zero or more names from the list
	SelectNames(names []string, message string, selectAll bool, help string) ([]string, error)

	// Confirm confirms an action
	Confirm(message string, defaultValue bool, help string) (bool, error)
}
