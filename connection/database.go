package connection

import (
        "neoai/globals"
        "neoai/utils"
        "crypto/tls"
        "crypto/x509"
        "database/sql"
        "fmt"
        "os"
        "strings"

        "github.com/go-sql-driver/mysql"
        _ "github.com/go-sql-driver/mysql"
        _ "github.com/mattn/go-sqlite3"
        "github.com/spf13/viper"
)

var DB *sql.DB

func InitMySQLSafe() *sql.DB {
        ConnectDatabase()

        // using DB as a global variable to point to the latest db connection
        MysqlWorker(DB)
        return DB
}

func getConn() *sql.DB {
        if viper.GetString("mysql.host") == "" {
                globals.SqliteEngine = true
                globals.Warn("[connection] mysql host is not set, using sqlite (~/db/chatnio.db)")
                db, err := sql.Open("sqlite3", utils.FileSafe("./db/chatnio.db"))
                if err != nil {
                        panic(err)
                }

                return db
        }

        mysqlUrl := fmt.Sprintf(
                "%s:%s@tcp(%s:%d)/%s",
                viper.GetString("mysql.user"),
                viper.GetString("mysql.password"),
                viper.GetString("mysql.host"),
                viper.GetInt("mysql.port"),
                utils.GetStringConfs("mysql.database", "mysql.db"),
        )

        // NeoAI: enhanced TLS support.
        //
        // Config keys:
        //   mysql.tls              (bool)   — enable TLS
        //   mysql.tls_skip_verify  (bool)   — skip cert verification (self-signed)
        //   mysql.tls_ca           (string) — path to CA cert file (PEM)
        //   mysql.tls_server_name  (string) — override SNI / ServerName
        //
        // Common scenarios:
        //   1. Cloud MySQL with public CA (RDS, PlanetScale, etc.):
        //        mysql.tls: true
        //   2. Self-signed / internal CA:
        //        mysql.tls: true
        //        mysql.tls_ca: /path/to/ca.pem
        //   3. Skip verification (not recommended for production):
        //        mysql.tls: true
        //        mysql.tls_skip_verify: true
        if viper.GetBool("mysql.tls") {
                tlsConfig := &tls.Config{
                        MinVersion: tls.VersionTLS12,
                        ServerName: viper.GetString("mysql.tls_server_name"),
                }
                if tlsConfig.ServerName == "" {
                        tlsConfig.ServerName = viper.GetString("mysql.host")
                }
                if viper.GetBool("mysql.tls_skip_verify") {
                        tlsConfig.InsecureSkipVerify = true
                }
                if caPath := viper.GetString("mysql.tls_ca"); caPath != "" {
                        caCert, err := os.ReadFile(caPath)
                        if err != nil {
                                globals.Warn(fmt.Sprintf("[connection] failed to read mysql TLS CA file %s: %s", caPath, err.Error()))
                        } else {
                                caPool := x509.NewCertPool()
                                if caPool.AppendCertsFromPEM(caCert) {
                                        tlsConfig.RootCAs = caPool
                                } else {
                                        globals.Warn(fmt.Sprintf("[connection] mysql TLS CA file %s is not a valid PEM", caPath))
                                }
                        }
                }
                tlsName := "neoai-mysql-tls"
                if err := mysql.RegisterTLSConfig(tlsName, tlsConfig); err != nil {
                        // "tls config already registered" — fine, reuse it.
                        _ = err
                }
                mysqlUrl += "?tls=" + tlsName
        }

        // also support common connection-string params via mysql.params
        if params := viper.GetString("mysql.params"); params != "" {
                if strings.Contains(mysqlUrl, "?") {
                        mysqlUrl += "&" + strings.TrimPrefix(params, "?")
                } else {
                        mysqlUrl += "?" + strings.TrimPrefix(params, "?")
                }
        }

        // connect to MySQL
        db, err := sql.Open("mysql", mysqlUrl)

        if pingErr := db.Ping(); err != nil || pingErr != nil {
                errMsg := utils.Multi[string](err != nil, utils.GetError(err), utils.GetError(pingErr)) // err.Error() may contain nil pointer
                globals.Warn(
                        fmt.Sprintf("[connection] failed to connect to mysql server: %s (message: %s), will retry in 5 seconds",
                                viper.GetString("mysql.host"), errMsg,
                        ),
                )

                utils.Sleep(5000)
                db.Close()

                return getConn()
        }

        globals.Debug(fmt.Sprintf("[connection] connected to mysql server (host: %s)", viper.GetString("mysql.host")))
        return db
}

func ConnectDatabase() *sql.DB {
        db := getConn()

        db.SetMaxOpenConns(512)
        db.SetMaxIdleConns(64)

        CreateUserTable(db)
        CreateConversationTable(db)
        CreateMaskTable(db)
        CreateSharingTable(db)
        CreatePackageTable(db)
        CreateQuotaTable(db)
        CreateSubscriptionTable(db)
        CreateApiKeyTable(db)
        CreateInvitationTable(db)
        CreateRedeemTable(db)
        CreateBroadcastTable(db)

        // NeoAI: registration code table — created by auth package on first
        // call to ConsumeRegistrationCode / GenerateRegistrationCodes, since
        // the auth package can't be imported here (would cause a cycle).
        // auth.EnsureRegistrationCodeTable(db) is invoked from main.go after
        // middleware setup.

        if err := doMigration(db); err != nil {
                fmt.Println(fmt.Sprintf("migration error: %s", err))
        }

        DB = db

        return db
}

func InitRootUser(db *sql.DB) {
        // create root user if totally empty
        var count int
        err := globals.QueryRowDb(db, "SELECT COUNT(*) FROM auth").Scan(&count)
        if err != nil {
                globals.Warn(fmt.Sprintf("[service] failed to query user count: %s", err.Error()))
                return
        }

        if count == 0 {
                globals.Debug("[service] no user found, creating root user (username: root, password: chatnio123456, email: root@example.com)")
                _, err := globals.ExecDb(db, `
                        INSERT INTO auth (username, password, email, is_admin, bind_id, token)
                        VALUES (?, ?, ?, ?, ?, ?)
                `, "root", utils.Sha2Encrypt("chatnio123456"), "root@example.com", true, 0, "root")
                if err != nil {
                        globals.Warn(fmt.Sprintf("[service] failed to create root user: %s", err.Error()))
                }
        } else {
                globals.Debug(fmt.Sprintf("[service] %d user(s) found, skip creating root user", count))
        }
}

func CreateUserTable(db *sql.DB) {
        _, err := globals.ExecDb(db, `
                CREATE TABLE IF NOT EXISTS auth (
                  id INT PRIMARY KEY AUTO_INCREMENT,
                  bind_id INT UNIQUE,
                  username VARCHAR(24) UNIQUE,
                  token VARCHAR(255) NOT NULL,
                  email VARCHAR(255) UNIQUE,
                  password VARCHAR(64) NOT NULL,
                  is_admin BOOLEAN DEFAULT FALSE,
                  is_banned BOOLEAN DEFAULT FALSE
                );
        `)
        if err != nil {
                fmt.Println(err)
        }

        InitRootUser(db)
}

func CreatePackageTable(db *sql.DB) {
        _, err := globals.ExecDb(db, `
                CREATE TABLE IF NOT EXISTS package (
                  id INT PRIMARY KEY AUTO_INCREMENT,
                  user_id INT,
                  type VARCHAR(255),
                  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                  FOREIGN KEY (user_id) REFERENCES auth(id),
                  UNIQUE KEY (user_id, type)
                );
        `)
        if err != nil {
                fmt.Println(err)
        }
}

func CreateQuotaTable(db *sql.DB) {
        _, err := globals.ExecDb(db, `
                CREATE TABLE IF NOT EXISTS quota (
                  id INT PRIMARY KEY AUTO_INCREMENT,
                  user_id INT UNIQUE,
                  quota DECIMAL(24, 6),
                  used DECIMAL(24, 6),
                  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                  FOREIGN KEY (user_id) REFERENCES auth(id)
                );
        `)
        if err != nil {
                fmt.Println(err)
        }
}

func CreateConversationTable(db *sql.DB) {
        _, err := globals.ExecDb(db, `
                CREATE TABLE IF NOT EXISTS conversation (
                  id INT PRIMARY KEY AUTO_INCREMENT,
                  user_id INT,
                  conversation_id INT,
                  conversation_name VARCHAR(255),
                  data MEDIUMTEXT,
                  model VARCHAR(255) NOT NULL DEFAULT 'gpt-3.5-turbo-0613',
                  task_id VARCHAR(255) NULL,
                  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                  UNIQUE KEY (user_id, conversation_id)
                );
        `)
        if err != nil {
                fmt.Println(err)
        }
}

func CreateMaskTable(db *sql.DB) {
        _, err := globals.ExecDb(db, `
                CREATE TABLE IF NOT EXISTS mask (
                  id INT PRIMARY KEY AUTO_INCREMENT,
                  user_id INT,
                  avatar VARCHAR(255),
                  name VARCHAR(255),
                  description TEXT,
                  context TEXT,
                  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                  FOREIGN KEY (user_id) REFERENCES auth(id)
                );
        `)
        if err != nil {
                fmt.Println(err)
        }
}

func CreateSharingTable(db *sql.DB) {
        // refs is an array of message id, separated by comma (-1 means all messages)
        _, err := globals.ExecDb(db, `
                CREATE TABLE IF NOT EXISTS sharing (
                  id INT PRIMARY KEY AUTO_INCREMENT,
                  hash CHAR(32) UNIQUE,
                  user_id INT,
                  conversation_id INT,
                  refs TEXT,
                  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                  FOREIGN KEY (user_id) REFERENCES auth(id)
                );
        `)
        if err != nil {
                fmt.Println(err)
        }
}

func CreateSubscriptionTable(db *sql.DB) {
        _, err := globals.ExecDb(db, `
                CREATE TABLE IF NOT EXISTS subscription (
                  id INT PRIMARY KEY AUTO_INCREMENT,
                  level INT DEFAULT 1,
                  user_id INT UNIQUE,
                  expired_at DATETIME,
                  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                  total_month INT DEFAULT 0,
                  enterprise BOOLEAN DEFAULT FALSE,
                  FOREIGN KEY (user_id) REFERENCES auth(id)
                );
        `)
        if err != nil {
                fmt.Println(err)
        }
}

func CreateApiKeyTable(db *sql.DB) {
        _, err := globals.ExecDb(db, `
                CREATE TABLE IF NOT EXISTS apikey (
                  id INT PRIMARY KEY AUTO_INCREMENT,
                  user_id INT UNIQUE,
                  api_key VARCHAR(255) UNIQUE,
                  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                  FOREIGN KEY (user_id) REFERENCES auth(id)
                );
        `)
        if err != nil {
                fmt.Println(err)
        }
}

func CreateInvitationTable(db *sql.DB) {
        _, err := globals.ExecDb(db, `
                CREATE TABLE IF NOT EXISTS invitation (
                  id INT PRIMARY KEY AUTO_INCREMENT,
                  code VARCHAR(255) UNIQUE,
                  quota DECIMAL(16, 4),
                  type VARCHAR(255),
                  used BOOLEAN DEFAULT FALSE,
                  used_id INT,
                  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                  UNIQUE KEY (used_id, type),
                  FOREIGN KEY (used_id) REFERENCES auth(id)
                );
        `)
        if err != nil {
                fmt.Println(err)
        }
}

func CreateRedeemTable(db *sql.DB) {
        _, err := globals.ExecDb(db, `
                CREATE TABLE IF NOT EXISTS redeem (
                  id INT PRIMARY KEY AUTO_INCREMENT,
                  code VARCHAR(255) UNIQUE,
                  quota DECIMAL(16, 4),
                  used BOOLEAN DEFAULT FALSE,
                  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
                );
        `)
        if err != nil {
                fmt.Println(err)
        }
}

func CreateBroadcastTable(db *sql.DB) {
        _, err := globals.ExecDb(db, `
                CREATE TABLE IF NOT EXISTS broadcast (
                  id INT PRIMARY KEY AUTO_INCREMENT,
                  poster_id INT,
                  content TEXT,
                  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                  FOREIGN KEY (poster_id) REFERENCES auth(id)
                );
        `)
        if err != nil {
                fmt.Println(err)
        }
}
