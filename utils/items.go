package utils

type Item struct {
	title       string
	description string
	runnable    bool
}

// Title returns the primary command text so list rendering and selection can
// surface what will be executed.
func (i Item) Title() string { return i.title }

// Description returns secondary metadata so the list can show usage context.
func (i Item) Description() string { return i.description }

// FilterValue returns the text used for matching so filtering behaves like
// command search.
func (i Item) FilterValue() string { return i.title }

// Runnable indicates whether selecting this item should trigger command
// execution, allowing safe informational rows.
func (i Item) Runnable() bool { return i.runnable }

// NewItem creates a runnable history item because real commands should be
// executable from the chooser.
func NewItem(title, description string) Item {
	return Item{
		title:       title,
		description: description,
		runnable:    true,
	}
}

// NewNoticeItem creates a non-runnable informational row so fallback messages
// cannot be executed by accident.
func NewNoticeItem(title, description string) Item {
	return Item{
		title:       title,
		description: description,
		runnable:    false,
	}
}
