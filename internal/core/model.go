package core

import (
	"database/sql"
	"errors"
	"myt/internal/config"
	"myt/internal/constant"
	"myt/internal/database"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"myt/internal/render"
)

type Model struct {
	State            constant.AppState
	Configs          []config.Config
	ConfigCursor     int
	ErrorMsg         string
	Tee              string
	ConnectionSelect bool

	PasswordInput textinput.Model
	PromptTarget  string

	DB               *sql.DB
	Conn             *sql.Conn
	Databases        []string
	Tables           []string
	Columns          []string
	TableColumns     map[string][]string
	AutocompleteDict []string
	TxPending        bool

	FocusPanel   constant.Focus
	OutputFormat render.OutputFormat
	ShowHelp     bool

	DBCursor     int
	TableCursor  int
	ColumnCursor int

	SqlInput textarea.Model

	DBNet string

	TabMatches  []string
	TabMatchIdx int

	PromptYesAction func() tea.Cmd
	PromptNoAction  func() tea.Cmd
	PromptMsg       string
	PromptTitle     string
	PromptYesMsg    string
	PromptNoMsg     string
}

func NewModel(configs []config.Config, conSelect bool) *Model {
	ti := textarea.New()
	ti.Placeholder = "Write SQL query here..."
	ti.ShowLineNumbers = false
	ti.Prompt = ""
	ti.SetWidth(78)
	ti.SetHeight(6)
	ti.Blur()

	ti.FocusedStyle.Base = ti.FocusedStyle.Base.UnsetBackground()
	ti.FocusedStyle.Text = ti.FocusedStyle.Text.UnsetBackground()
	ti.FocusedStyle.Placeholder = ti.FocusedStyle.Placeholder.UnsetBackground()
	ti.FocusedStyle.CursorLine = ti.FocusedStyle.CursorLine.UnsetBackground()
	ti.BlurredStyle.Base = ti.BlurredStyle.Base.UnsetBackground()
	ti.BlurredStyle.Text = ti.BlurredStyle.Text.UnsetBackground()
	ti.BlurredStyle.Placeholder = ti.BlurredStyle.Placeholder.UnsetBackground()

	m := &Model{
		State:            constant.AppStateConfig,
		Configs:          configs,
		ConfigCursor:     0,
		ConnectionSelect: conSelect,
		SqlInput:         ti,
		OutputFormat:     render.Grid,
		ShowHelp:         false,
		FocusPanel:       constant.FocusTable,
	}

	if !conSelect {
		// ssh password
		if m.Configs[m.ConfigCursor].SSHUser != "" && m.Configs[m.ConfigCursor].SSHKey == "" && m.Configs[m.ConfigCursor].SSHPass == "" {
			m.SetPasswordSubmit("SSH")
			return m
		}

		// password
		if m.Configs[m.ConfigCursor].Pass == "" {
			m.SetPasswordSubmit("DB")
			return m
		}

		netType, err := m.GetNetType()
		if err != nil {
			m.ErrorMsg = err.Error()
		}
		err = m.InitConnection(m.Configs[m.ConfigCursor], netType)
		if err == nil {
			m.State = constant.AppStateDBSelect
		} else {
			m.ErrorMsg = err.Error()
		}
	}

	return m
}

func (m *Model) SetPasswordSubmit(target string) {
	m.State = constant.AppStatePassword
	m.PromptTarget = target
	m.PasswordInput = textinput.New()
	if target == "SSH" {
		m.PasswordInput.Placeholder = "Enter SSH Password"
	} else {
		m.PasswordInput.Placeholder = "Enter MySQL Password"
	}
	m.PasswordInput.EchoMode = textinput.EchoPassword
	m.PasswordInput.Focus()
}

func (m *Model) Init() tea.Cmd {
	if m.State == constant.AppStatePassword {
		return textinput.Blink
	}
	return nil
}

func (m *Model) GetNetType() (string, error) {
	netType := "tcp"
	if m.Configs[m.ConfigCursor].SSHHost != "" {
		// Generate a unique identifier for each SSH connection
		netType = "mysql+tcp+" + m.Configs[m.ConfigCursor].Name

		err := database.SetupSSH(m.Configs[m.ConfigCursor].SSHHost, m.Configs[m.ConfigCursor].SSHPort, m.Configs[m.ConfigCursor].SSHUser, m.Configs[m.ConfigCursor].SSHPass, m.Configs[m.ConfigCursor].SSHKey, netType)
		if err != nil {
			return "", err
		}
	}
	return netType, nil
}

func (m *Model) InitConnection(cfg config.Config, netType string) error {

	db, err := database.GetDatabase(cfg.Host, cfg.Port, cfg.User, cfg.Pass, netType, "", cfg.Charset)
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
		return errors.New("DB connection failed: " + err.Error())

	}

	databases, err := database.GetDatabases(db)
	if err != nil {
		return errors.New("Failed to list databases: " + err.Error())
	}

	m.DBNet = netType
	m.Databases = databases

	if len(databases) > 0 {
		defaultDB := databases[0]
		m.Close()
		newDB, _ := database.GetDatabase(cfg.Host, cfg.Port, cfg.User, cfg.Pass, netType, defaultDB, cfg.Charset)

		m.DB = newDB
		m.Conn, _ = database.GetConnection(newDB, m.Configs[m.ConfigCursor].ReadWrite)
		m.LoadMetadata(defaultDB)
	} else {
		m.DB = db
	}

	return nil
}

func (m *Model) Connect(cfg config.Config) {
	dbName := m.Databases[m.DBCursor]
	m.Close()

	db, _ := database.GetDatabase(
		m.Configs[m.ConfigCursor].Host, m.Configs[m.ConfigCursor].Port,
		m.Configs[m.ConfigCursor].User, m.Configs[m.ConfigCursor].Pass, m.DBNet, dbName, cfg.Charset)
	conn, _ := database.GetConnection(db, m.Configs[m.ConfigCursor].ReadWrite)
	m.DB = db
	m.Conn = conn
	m.LoadMetadata(dbName)
}

func (m *Model) Close() {
	if m.Conn != nil {
		m.Conn.Close()
	}
	if m.DB != nil {
		m.DB.Close()
	}
	m.TxPending = false
}

func (m *Model) LoadMetadata(dbName string) {
	m.TableColumns = make(map[string][]string)
	existT := make(map[string]bool)
	existC := make(map[string]bool)
	m.Tables = nil
	var columns []string

	query := "SELECT TABLE_NAME, COLUMN_NAME FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = ?"
	rows, err := m.DB.Query(query, dbName)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var tName, cName string
			if err := rows.Scan(&tName, &cName); err == nil {
				upperTName := strings.ToUpper(tName)
				m.TableColumns[upperTName] = append(m.TableColumns[upperTName], cName)

				if !existT[tName] {
					existT[tName] = true
					m.Tables = append(m.Tables, tName)
				}
				if !existC[cName] {
					existC[cName] = true
					columns = append(columns, cName)
				}
			}
		}
	}

	sort.Strings(m.Tables)
	sort.Strings(columns)

	var newDict []string
	existAll := make(map[string]bool)

	for _, t := range m.Tables {
		if !existAll[t] {
			existAll[t] = true
			newDict = append(newDict, t)
		}
	}

	for _, c := range columns {
		if !existAll[c] {
			existAll[c] = true
			newDict = append(newDict, c)
		}
	}

	for _, kw := range database.KEYWORDS {
		if !existAll[kw] {
			existAll[kw] = true
			newDict = append(newDict, kw)
		}
	}

	m.AutocompleteDict = newDict
	m.TableCursor = 0
	m.UpdateColumns()
}

func (m *Model) UpdateColumns() {
	m.Columns = nil
	m.ColumnCursor = 0
	if m.TableCursor < len(m.Tables) {
		tName := m.Tables[m.TableCursor]
		if cols, ok := m.TableColumns[strings.ToUpper(tName)]; ok {
			m.Columns = cols
		}
	}
}

func (m *Model) InsertStringToSQL(s string) {
	m.SqlInput.InsertString(s)
}
