package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type QueryResult struct {
	Columns []string
	Rows    []map[string]interface{}
	Message string
}

var (
	sshClientsMu sync.Mutex
	sshClients   = map[string]*ssh.Client{}
)

// replaceSSHClient registers client as the active tunnel for netType,
// closing whatever client was previously registered for it.
func replaceSSHClient(netType string, client *ssh.Client) {
	sshClientsMu.Lock()
	old := sshClients[netType]
	sshClients[netType] = client
	sshClientsMu.Unlock()
	if old != nil {
		old.Close()
	}
}

func SetupSSH(sshHost string, sshPort int, sshUser, sshPass, sshKey, netType string) error {
	var authMethods []ssh.AuthMethod
	if sshKey != "" {
		keyData, err := os.ReadFile(expandHome(sshKey))
		if err != nil {
			return errors.New("Failed to load the SSH private key: " + err.Error())
		}
		signer, err := ssh.ParsePrivateKey(keyData)
		if err != nil {
			return errors.New("Failed to parse the SSH private key: " + err.Error())
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}
	if sshPass != "" {
		authMethods = append(authMethods, ssh.Password(sshPass))
		authMethods = append(authMethods, ssh.KeyboardInteractive(
			func(sshUser, instruction string, questions []string, echos []bool) ([]string, error) {
				answers := make([]string, len(questions))
				for i := range questions {
					answers[i] = sshPass
				}
				return answers, nil
			},
		))

	}
	if len(authMethods) == 0 {
		return errors.New("SSH connection requires either a password or a private key")
	}

	hostKeyCallback, err := knownHostsCallback()
	if err != nil {
		return err
	}

	sshConfig := &ssh.ClientConfig{
		User:            sshUser,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
	}

	sshClient, err := ssh.Dial("tcp", sshHost+":"+strconv.Itoa(sshPort), sshConfig)
	if err != nil {
		return errors.New("Failed to establish SSH connection to the bastion server: " + err.Error())
	}

	replaceSSHClient(netType, sshClient)

	mysql.RegisterDialContext(netType, func(ctx context.Context, addr string) (net.Conn, error) {
		return sshClient.Dial("tcp", addr)
	})

	return nil
}

func GetDatabases(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SHOW DATABASES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err == nil {
			databases = append(databases, dbName)
		}
	}
	return databases, nil
}

func GetDatabase(host string, port int, user, pass, netType string, dbName string, charset string) (*sql.DB, error) {
	dns := user + ":" + pass + "@" + netType + "(" + host + ":" + strconv.Itoa(port) + ")/" + dbName + "?charset=" + charset

	db, err := sql.Open("mysql", dns)
	if err != nil {
		if db != nil {
			db.Close()
		}
		return nil, err
	}
	return db, nil
}

func GetConnection(db *sql.DB, readWrite bool) (*sql.Conn, error) {

	conn, err := db.Conn(context.Background())
	if err != nil {
		if conn != nil {
			conn.Close()
		}
		if db != nil {
			db.Close()
		}
		return nil, err
	}
	if readWrite {
		_, err := conn.ExecContext(context.Background(), "SET autocommit=0")
		if err != nil {
			if conn != nil {
				conn.Close()
			}
			if db != nil {
				db.Close()
			}
			return nil, err
		}
	} else {
		_, err := conn.ExecContext(context.Background(), "SET TRANSACTION READ ONLY")
		if err != nil {
			errStr := err.Error()
			// `SET TRANSACTION READ ONLY` is not supported under 5.6.5, so we ignore this error if it occurs.
			if !strings.Contains(errStr, "syntax") && !strings.Contains(errStr, "read only") {
				if conn != nil {
					conn.Close()
				}
				if db != nil {
					db.Close()
				}
				return nil, err
			}
		}
	}
	return conn, nil
}

func ExecuteQuery(ctx context.Context, conn *sql.Conn, query string) (*QueryResult, error) {
	rows, err := conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	if len(cols) == 0 {
		return &QueryResult{Message: "Query OK. (No result set returned)"}, nil
	}

	var results []map[string]interface{}
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		valPtrs := make([]interface{}, len(cols))
		for i := range vals {
			valPtrs[i] = &vals[i]
		}
		if err := rows.Scan(valPtrs...); err != nil {
			return nil, err
		}
		rowMap := make(map[string]interface{})
		for i, colName := range cols {
			val := vals[i]

			if val == nil {
				rowMap[colName] = nil
				continue
			}
			switch v := val.(type) {
			case []byte:
				rowMap[colName] = string(v)
			case int64:
				rowMap[colName] = strconv.FormatInt(v, 10)
			case uint64:
				rowMap[colName] = strconv.FormatUint(v, 10)
			default:
				rowMap[colName] = fmt.Sprintf("%v", v)
			}
		}
		results = append(results, rowMap)
	}

	return &QueryResult{
		Columns: cols,
		Rows:    results,
	}, nil
}

func knownHostsCallback() (ssh.HostKeyCallback, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, errors.New("Failed to determine home directory for known_hosts: " + err.Error())
	}

	khPath := filepath.Join(home, ".ssh", "known_hosts")
	callback, err := knownhosts.New(khPath)
	if err != nil {
		return nil, errors.New("Failed to load known_hosts file (" + khPath + "): " + err.Error() +
			". Connect to the host once with the ssh command to add its key, then retry.")
	}
	return callback, nil
}

func expandHome(path string) string {
	if path == "" {
		return ""
	}

	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
