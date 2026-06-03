package constant

type AppState int

const (
	AppStateConfig AppState = iota
	AppStatePassword
	AppStateMain
)

type Focus int

const (
	FocusDB Focus = iota
	FocusTable
	FocusColumn
	FocusEditor
)
