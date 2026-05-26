package core

import (
	"database/sql"
	"fmt"
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

// Config YAML
type Config struct {
	Name    string `yaml:"name"`
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	User    string `yaml:"user"`
	Pass    string `yaml:"pass"`
	SSHHost string `yaml:"ssh_host"`
	SSHPort int    `yaml:"ssh_port"`
	SSHUser string `yaml:"ssh_user"`
	SSHPass string `yaml:"ssh_pass"`
	SSHKey  string `yaml:"ssh_key"`
}

type Model struct {
	State        AppState
	Configs      []Config
	ConfigCursor int
	ErrorMsg     string // 接続失敗時のメッセージ

	DB               *sql.DB
	Conn             *sql.Conn
	Databases        []string
	Tables           []string
	Columns          []string
	TableColumns     map[string][]string
	AutocompleteDict []string

	FocusSQL     bool
	SchemaPane   int
	OutputFormat render.OutputFormat
	ShowHelp     bool
	ReadWrite    bool

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

func NewModel(configs []Config, rw bool) *Model {
	ti := textarea.New()
	ti.Placeholder = "Write your SQL query here..."
	ti.SetHeight(5)
	ti.SetWidth(70)
	ti.ShowLineNumbers = false
	ti.Prompt = ""
	ti.Blur()

	m := &Model{
		State:        SelectConfig,
		Configs:      configs,
		ConfigCursor: 0,
		FocusSQL:     false,
		SchemaPane:   0,
		SqlInput:     ti,
		OutputFormat: render.FormatGrid,
		ShowHelp:     false,
		ReadWrite:    rw,
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

func (m *Model) InitConnection(cfg Config) error {
	netType := "tcp"
	if cfg.SSHHost != "" {
		// Generate a unique identifier for each SSH connection
		netType = fmt.Sprintf("mysql+tcp+%s", cfg.Name)
		err := database.SetupSSHWithNetType(cfg.SSHHost, cfg.SSHPort, cfg.SSHUser, cfg.SSHPass, cfg.SSHKey, netType)
		if err != nil {
			return err
		}
	}

	db, err := database.GetDatabase(cfg.Host, cfg.Port, cfg.User, cfg.Pass, netType, "")
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
		newDB, _ := database.GetDatabase(cfg.Host, cfg.Port, cfg.User, cfg.Pass, netType, defaultDB)

		m.DB = newDB
		m.Conn, _ = database.GetConnection(newDB, m.ReadWrite)
		m.LoadMetadata(defaultDB)
	} else {
		m.DB = db
	}

	return nil
}

func (m *Model) Connect() {
	dbName := m.Databases[m.DBCursor]
	m.Close()

	db, _ := database.GetDatabase(m.DBHost, m.DBPort, m.DBUser, m.DBPass, m.DBNet, dbName)
	conn, _ := database.GetConnection(db, m.ReadWrite)
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
}

func (m *Model) LoadMetadata(dbName string) {
	dictMap := make(map[string]bool)
	for _, kw := range database.KEYWORDS {
		dictMap[kw] = true
	}
	m.TableColumns = make(map[string][]string)
	seenTables := make(map[string]bool)
	m.Tables = nil
	query := fmt.Sprintf("SELECT TABLE_NAME, COLUMN_NAME FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = '%s'", dbName)
	rows, err := m.DB.Query(query)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var tName, cName string
			if err := rows.Scan(&tName, &cName); err == nil {
				dictMap[tName] = true
				dictMap[cName] = true
				upperTName := strings.ToUpper(tName)
				m.TableColumns[upperTName] = append(m.TableColumns[upperTName], cName)
				if !seenTables[tName] {
					seenTables[tName] = true
					m.Tables = append(m.Tables, tName)
				}
			}
		}
	}
	sort.Strings(m.Tables)
	var newDict []string
	for k := range dictMap {
		newDict = append(newDict, k)
	}
	sort.Strings(newDict)
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
