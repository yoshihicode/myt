package main

import (
	"flag"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/yaml.v3"

	"myt/internal/core"
	"myt/internal/types"
)

func main() {
	confPath := flag.String("conf", "", "Path to YAML config file (e.g., myt.yaml)")

	host := flag.String("host", "127.0.0.1", "MySQL host address")
	port := flag.Int("port", 3306, "MySQL port")
	user := flag.String("user", "", "MySQL username")
	pass := flag.String("pass", "", "MySQL password")

	rw := flag.Bool("rw", false, "Enable read-write mode (Caution: Allows INSERT/UPDATE/DELETE with autocommit=0)")
	charset := flag.String("charset", "utf8mb4", "Character set for the connection")

	sshHost := flag.String("ssh-host", "", "SSH bastion host address (e.g., 192.168.1.10)")
	sshPort := flag.Int("ssh-port", 22, "SSH port")
	sshUser := flag.String("ssh-user", "", "SSH username")
	sshPass := flag.String("ssh-pass", "", "SSH password")
	sshKey := flag.String("ssh-key", "", "Path to SSH private key (e.g., ~/.ssh/id_rsa)")
	flag.Parse()

	var configs []types.Config

	if *confPath != "" {
		data, err := os.ReadFile(*confPath)
		if err != nil {
			log.Fatalf("Failed to load the configuration file: %v", err)
		}
		if err := yaml.Unmarshal(data, &configs); err != nil {
			log.Fatalf("Failed to parse YAML: %v", err)
		}

		for i := range configs {
			if configs[i].Port == 0 {
				configs[i].Port = 3306
			}
			if configs[i].SSHHost != "" && configs[i].SSHPort == 0 {
				configs[i].SSHPort = 22
			}
		}
	} else {

		configs = append(configs, types.Config{
			Name:    "CLI Connection",
			Host:    *host,
			Port:    *port,
			User:    *user,
			Pass:    *pass,
			Charset: *charset,
			SSHHost: *sshHost,
			SSHPort: *sshPort,
			SSHUser: *sshUser,
			SSHPass: *sshPass,
			SSHKey:  *sshKey,
		})
	}

	m := core.NewModel(configs, *rw)
	defer func() {
		m.Close()
	}()

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
