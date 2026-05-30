package core

import (
	"database/sql"
	"fmt"
	"myt/internal/config"
	"myt/internal/database"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"

	"myt/internal/render"
)

type AppState int

const (
	SelectConfig AppState = iota
	Main
)

type Model struct {
	State            AppState
	Configs          []config.Config
	ConfigCursor     int
	ErrorMsg         string
	Tee              string
	ConnectionSelect bool

	DB               *sql.DB
	Conn             *sql.Conn
	Databases        []string
	Tables           []string
	Columns          []string
	TableColumns     map[string][]string
	AutocompleteDict []string
	TxPending        bool

	FocusSQL     bool
	SchemaPane   int
	OutputFormat render.OutputFormat
	ShowHelp     bool

	DBCursor     int
	TableCursor  int
	ColumnCursor int

	SqlInput textarea.Model

	DBUser string
	DBPass string
	DBHost string
	DBPort int
	DBNet  string

	TabMatches  []string
	TabMatchIdx int
}

func NewModel(configs []config.Config, conSelect bool) *Model {
	ti := textarea.New()
	ti.Placeholder = "Write your SQL query here..."
	ti.SetHeight(5)
	ti.SetWidth(70)
	ti.ShowLineNumbers = false
	ti.Prompt = ""
	ti.Blur()

	m := &Model{
		State:            SelectConfig,
		Configs:          configs,
		ConfigCursor:     0,
		ConnectionSelect: conSelect,
		FocusSQL:         false,
		SchemaPane:       0,
		SqlInput:         ti,
		OutputFormat:     render.Grid,
		ShowHelp:         false,
	}

	if len(configs) == 1 {
		err := m.InitConnection(configs[0])
		if err == nil {
			m.State = Main
		} else {
			m.ErrorMsg = err.Error()
		}
	}

	return m
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) InitConnection(cfg config.Config) error {
	netType := "tcp"
	if cfg.SSHHost != "" {
		// Generate a unique identifier for each SSH connection
		netType = fmt.Sprintf("mysql+tcp+%s", cfg.Name)
		err := database.SetupSSH(cfg.SSHHost, cfg.SSHPort, cfg.SSHUser, cfg.SSHPass, cfg.SSHKey, netType)
		if err != nil {
			return err
		}
	}

	db, err := database.GetDatabase(cfg.Host, cfg.Port, cfg.User, cfg.Pass, netType, "", cfg.Charset)
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
		return fmt.Errorf("DB connection failed: %v", err)
	}

	databases, err := database.GetDatabases(db)
	if err != nil {
		return fmt.Errorf("Failed to list databases: %v", err)
	}

	m.DBUser = cfg.User
	m.DBPass = cfg.Pass
	m.DBHost = cfg.Host
	m.DBPort = cfg.Port
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

	db, _ := database.GetDatabase(m.DBHost, m.DBPort, m.DBUser, m.DBPass, m.DBNet, dbName, cfg.Charset)
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

	query := fmt.Sprintf("SELECT TABLE_NAME, COLUMN_NAME FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = '%s'", dbName)
	rows, err := m.DB.Query(query)
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

	keywords := make([]string, len(database.KEYWORDS))
	copy(keywords, database.KEYWORDS)
	sort.Strings(keywords)

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

	for _, kw := range keywords {
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
