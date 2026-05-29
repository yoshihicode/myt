package config

// Config YAML
type Config struct {
	Name      string `yaml:"name"`
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	User      string `yaml:"user"`
	Pass      string `yaml:"password"`
	SSHHost   string `yaml:"ssh_host"`
	SSHPort   int    `yaml:"ssh_port"`
	SSHUser   string `yaml:"ssh_user"`
	SSHPass   string `yaml:"ssh_password"`
	SSHKey    string `yaml:"ssh_key"`
	Charset   string `yaml:"charset"`
	Tee       string `yaml:"tee"`
	ReadWrite bool   `yaml:"read_write"`
}
