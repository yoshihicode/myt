# MySQL/MariaDB Tiny Tui Client (`myt`)
![myt](screenshot.gif)  
`myt` is terminal based MySQL/MariaDB client.

## ⚙️ Features
* **🛡️ Safety First:** Default Read-Only mode and enforced `autocommit=0` in Read-Write mode to protect your data. 
* **🔒 Built-in SSH Tunneling:** Easily connect to remote databases through secure bastion hosts using passwords or SSH private keys.
* **🤖 Auto-Completion:** Case-insensitive autocomplete for standard SQL keywords, table names, and columns.
* **📂 Multi-Format Rendering:** View or seamlessly log results to a file in **GRID, MARKDOWN, CSV, or JSON** formats.
* **🧩 No Dependencies:** Just a single binary with zero external runtime requirements.
* **🖥️ Cross-Platform:** Available for Windows, macOS, and Linux.

## 💾 Download
Download prebuilt binaries from the latest release:  
👉 [Get the latest binaries](https://github.com/yoshihicode/myt/releases/latest)

## 📦 Installation
### 🐧 Linux
```bash
wget https://github.com/yoshihicode/myt/releases/latest/download/myt_linux_amd64.tar.gz
tar -xzvf myt_linux_amd64.tar.gz
sudo mv myt /usr/local/bin/

# Run
myt -host=127.0.0.1 -port=3306 -user=root -pass=your_password
```

### 🍎🍺 macOS / Homebrew
```bash
brew tap yoshihicode/tap
brew install myt

# Run
myt -host=127.0.0.1=-port=3306 -user=root -pass=your_password
```

### 🪟 Windows
```powershell
Invoke-WebRequest -OutFile myt_windows_amd64.tar.gz https://github.com/yoshihicode/myt/releases/latest/download/myt_windows_amd64.tar.gz
tar -xzvf myt_windows_amd64.tar.gz

# Run
.\myt.exe -host=127.0.0.1 -port=3306 -user=root -pass=your_password
```

## 🚀 Quick start
### Direct Connection
```bash
myt -host=127.0.0.1 -port=3306 -user=root -pass=your_password
```
### SSH tunnel (Password Authentication)
```bash
myt -host=10.0.0.5 -user=db_user -pass=db_pass -ssh-host=192.168.1.10 -ssh-port=22 -ssh-user=bastion_user -ssh-pass=bastion_pass
```
### SSH tunnel (SSH Key Authentication)
```bash
myt -host=10.0.0.5 -user=db_user -pass=db_pass -ssh-host=192.168.1.10 -ssh-user=bastion_user -ssh-key=$HOME/.ssh/id_rsa
```
### Configuration File

Manage multiple database environments easily by utilizing a configuration file. You can define various environments (e.g., local, staging, production) and launch `myt` with the `-conf` flag.

```bash
myt -conf myt.yaml
```
#### Fields
|Field|Type|Description|Default|Note|
| --- | --- | --- | --- | --- |
| name | string | Required. A unique name for this connection configuration. || Used as the dynamic network name. |
| host | string | MySQL server host address. |||
| port | int | MySQL server port number.| 3306 ||
| user | string | MySQL username. |||
| pass | string | MySQL password. || Leave empty if no password. |
| charset| string | Character set for the MySQL connection. | utf8mb4 ||
| tee | string | Output file path for query results (appended), mimicking the MySQL `tee` command. |||
| ssh_host| string | SSH bastion host address for tunneling. |||
| ssh_port| int | SSH bastion server port number. | 22 ||
| ssh_user| string | SSH username for the bastion server. |||
| ssh_pass| string | SSH password for the bastion server. || Used if ssh_key is empty. |
| ssh_key| string | Path to your SSH private key. || e.g. /home/user/.ssh/id_rsa |
| read_write| bool | Enable read-write mode  | false | Allows DML with autocommit=0 |

#### Yaml
```yml
- name: production-db
  host: 10.0.0.5
  port: 3306
  user: root
  pass: prod_secret_pass
  charset: utf8mb4
  ssh_host: 192.168.1.10
  ssh_port: 22
  ssh_user: ubuntu
  ssh_key: /home/user/.ssh/id_rsa

- name: staging-db
  host: 127.0.0.1
  port: 3306
  user: root
  pass: local_pass
  tee:  sql.logs
  readw_rite: true
```

## 📘 Usage
```
Usage of myt:
  -charset string
        Character set for the connection (default "utf8mb4")
  -conf string
        Path to YAML config file (e.g., myt.yaml)
  -host string
        MySQL host address (default "127.0.0.1")
  -pass string
        MySQL password
  -port int
        MySQL port (default 3306)
  -rw
        Enable read-write mode (Caution: Allows DML with autocommit=0)
  -ssh-host string
        SSH bastion host address (e.g., 192.168.1.10)
  -ssh-key string
        Path to SSH private key (e.g., ~/.ssh/id_rsa)
  -ssh-pass string
        SSH password
  -ssh-port int
        SSH port (default 22)
  -ssh-user string
        SSH username
  -tee string
        Output file path for query results (Appends results)
  -user string
        MySQL username
```

## ⌨ Key Bindings
### Global Shortcuts
| Key | Action |
| --- | --- |
| Tab | Switch panel focus between Schema Panel and SQL Input Panel |
| Ctrl+L | Clear terminal result screen |
| Ctrl+R | Reload metadata/schema from the current database |
| Ctrl+C | Exit Application |
| Ctrl+H | Toggle Help window |
| Esc | Disconnect and return to the connection selection page (only when started with a configuration file) |

### Schema Panel
| Key | Action |
| --- | --- |
| ← / → | Switch schema focus (Database ➔ Table ➔ Column) |
|↑ / ↓ | Move the panel cursor up and down |
|Enter | Select item / Insert the highlighted name directly into the SQL editor |

### SQL Panel
| Key | Action |
| --- | --- |
| Ctrl+N or [Ctrl+Space] | Trigger autocomplete / cycle through candidate matches |
| Ctrl+F | Switch output formats cyclically (GRID ➔ MARKDOWN ➔ CSV ➔ JSON) |
| Ctrl+E | Execute the written SQL query (supports multi-queries separated by `;`) |
| Ctrl+U | Clear all text inside the SQL editor |

