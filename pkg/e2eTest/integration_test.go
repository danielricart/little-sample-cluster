package e2eTest

import (
	"context"
	"database/sql"
	"fmt"
	"little-sample-cluster/pkg/api"
	"little-sample-cluster/pkg/metrics"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	mysqlclient "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testDB *sql.DB

type mysqlContainer struct {
	container     testcontainers.Container
	dsn           string
	ctx           *context.Context
	containerPort string
	containerHost string
}

// TestMain sets up and tears down the required dependency environment
func TestMain(m *testing.M) {
	ctx := context.Background()
	logger := log.New()
	logger.SetLevel(log.DebugLevel)

	containerDb, err := initializeTestContainerDb(ctx, logger)
	if err != nil {
		logger.Fatal(err)
	}
	defer containerDb.container.Terminate(context.Background())

	err = connectToDb(containerDb.dsn)
	if err != nil {
		logger.Fatalf("cannot connect to DB: %s, %s", containerDb.dsn, err)
	}

	err = ensureProperTransactionsPresent()
	if err != nil {
		log.Fatalf("failed to ensure transactions present: %v\n", err)
	}

	// Run tests
	code := m.Run()

	// Clean up
	teardownTestDB(ctx)

	os.Exit(code)
}

func TestHealthIntegration(t *testing.T) {
	server := initializeServer()

	cases := []struct {
		name           string
		method         string
		expectedStatus int
		expectedBody   string
	}{
		{"GET returns 200", http.MethodGet, http.StatusOK, "OK"},
		{"PUT returns 405", http.MethodPut, http.StatusMethodNotAllowed, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/health", nil)
			w := httptest.NewRecorder()

			server.HealthHandler(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			if tc.expectedBody != "" {
				assert.Equal(t, tc.expectedBody, w.Body.String())
			}
		})
	}
}

func TestHelloGetIntegration(t *testing.T) {
	server := initializeServer()
	today := time.Now().AddDate(-1, 0, 0).Format("2006-01-02")
	inThreeDays := time.Now().AddDate(-1, 0, +3).Format("2006-01-02")
	cases := []struct {
		testName       string
		username       string
		method         string
		expectedStatus int
		expectedBody   string
		expectedUser   api.DateOfBirth
	}{
		{"valid username, in 3 days", "getexistingthreedays",
			http.MethodGet, http.StatusOK,
			fmt.Sprintf(`{"message":"Hello, %s! Your birthday is in %d day(s)"}`, "getexistingthreedays", 3),
			api.DateOfBirth{
				Username:    "getexistingthreedays",
				DateOfBirth: inThreeDays,
			}},
		{"valid username, today", "getexistingtoday",
			http.MethodGet, http.StatusOK,
			fmt.Sprintf(`{"message":"Hello, %s! Happy birthday!"}`, "getexistingtoday"),
			api.DateOfBirth{
				Username:    "getexistingtoday",
				DateOfBirth: today,
			}},
		{"invalid username pattern", "asd123", http.MethodGet, http.StatusBadRequest, "", api.DateOfBirth{}},
		{"invalid username does not exist", "idontexist", http.MethodGet, http.StatusNotFound, "", api.DateOfBirth{}},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, fmt.Sprintf("/hello/%s", tc.username), nil)
			w := httptest.NewRecorder()

			server.HelloGetHandler(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			if tc.expectedBody != "" {
				assert.Equal(t, tc.expectedBody, w.Body.String())
			}
			if tc.expectedStatus == http.StatusOK {
				count := 0
				err := testDB.QueryRow(`
					SELECT count(*)
					FROM users
					WHERE username = ? AND date_of_birth = ?`, tc.expectedUser.Username, tc.expectedUser.DateOfBirth).Scan(&count)

				assert.NoErrorf(t, err, "Failed to query incidences: %v", err)
				assert.Equal(t, 1, count)
			}
		})
	}
}

func TestHelloPutIntegration(t *testing.T) {
	server := initializeServer()

	pastDate := time.Now().AddDate(-1, 0, -1).Format("2006-01-02")
	futureDate := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	cases := []struct {
		testName             string
		username             string
		requestBody          string
		method               string
		expectedStatus       int
		expectedResponseBody string
	}{
		// All users here use the same date_of_birth. this eases the validations on DB.
		{"valid username missing body", "asdasda", ``, http.MethodPut, http.StatusBadRequest, ""},
		{"valid username, wrong date", "asdasda", fmt.Sprintf(`{"dateOfBirth": "%s"}`, futureDate), http.MethodPut, http.StatusBadRequest, ""},
		{"invalid username pattern", "asd123", fmt.Sprintf(`{"dateOfBirth": "%s"}`, pastDate), http.MethodPut, http.StatusBadRequest, ""},
		{"valid new username", "newuser", fmt.Sprintf(`{"dateOfBirth": "%s"}`, pastDate), http.MethodPut, http.StatusNoContent, ""},
		{"valid username existing same date", "existingusersame", fmt.Sprintf(`{"dateOfBirth": "%s"}`, pastDate), http.MethodPut, http.StatusNoContent, ""},
		{"valid username existing different date", "existinguserdiff", fmt.Sprintf(`{"dateOfBirth": "%s"}`, pastDate), http.MethodPut, http.StatusNoContent, ""},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, fmt.Sprintf("/hello/%s", tc.username), strings.NewReader(tc.requestBody))
			w := httptest.NewRecorder()

			server.HelloPutHandler(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			if tc.expectedResponseBody != "" {
				assert.Equal(t, tc.expectedResponseBody, w.Body.String())
			}

			if tc.expectedStatus == http.StatusNoContent {
				count := 0
				err := testDB.QueryRow(`SELECT COUNT(*) from users WHERE username = ? AND date_of_birth = ?`, tc.username, pastDate).Scan(&count)
				assert.NoErrorf(t, err, "Failed to query user %s: %v", tc.username, err)
				assert.Equal(t, 1, count)
			}
		})
	}
}

func TestMetricsIntegration(t *testing.T) {
	server := initializeServer()
	_, metrHandler := metrics.NewMetrics(server.Logger)
	t.Run("check metrics are exposed", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/metrics", strings.NewReader(""))
		w := httptest.NewRecorder()
		(*metrHandler).ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		body := w.Body.String()
		assert.Contains(t, body, "# HELP birthday_invalid_total")
		assert.Contains(t, body, "# HELP birthday_registered_valid_total")
		assert.Contains(t, body, `# HELP go_info`)
	})
}

func initializeServer() api.Server {
	logger := log.New()
	logger.SetFormatter(&log.JSONFormatter{})
	logger.SetLevel(log.DebugLevel)
	logger.SetReportCaller(true)

	metricsObj, _ := metrics.NewMetrics(logger)
	server := api.Server{
		Logger:      logger,
		Database:    testDB,
		HelloServer: api.NewHelloServer(testDB, logger),
		Metrics:     metricsObj,
	}
	return server
}

// initializeTestContainerDb() (string, error) starts a MySQL container with schema and fixtures mounted
func initializeTestContainerDb(ctx context.Context, logger *log.Logger) (*mysqlContainer, error) {
	dbUser := "testing"
	dbPassword := "password1"
	dbName := "e2etests"
	dbRootPwd := "rootpwd1"
	var dbHost, dbPort string
	testContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Logger: logger,
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "mysql:8.4.8",
			ExposedPorts: []string{"3306/tcp"},
			Env: map[string]string{
				"MYSQL_ROOT_PASSWORD": dbRootPwd,
				"MYSQL_DATABASE":      dbName,
				"MYSQL_USER":          dbUser,
				"MYSQL_PASSWORD":      dbPassword,
			},
			Networks:       []string{"testnet"},
			NetworkAliases: map[string][]string{"testnet": []string{"mysql"}},
			Files:          []testcontainers.ContainerFile{},
			WaitingFor: wait.ForAll(
				wait.ForLog("/usr/sbin/mysqld: ready for connections").WithPollInterval(20*time.Second),
				wait.ForListeningPort("3306/tcp"),
			),
		},
		Started: true,
	})

	var mysqlCtr *mysqlContainer
	mysqlCtr = &mysqlContainer{ctx: &ctx, container: nil}
	if testContainer != nil {
		dbHost, _ = testContainer.Host(ctx)
		portObj, _ := testContainer.MappedPort(ctx, "3306")
		dbPort = portObj.Port()

		generatedConnString := generateDsn(dbUser, dbPassword, dbHost, dbPort, dbName)
		mysqlCtr = &mysqlContainer{container: testContainer, dsn: generatedConnString, containerPort: dbPort, containerHost: dbHost}
	}
	if err != nil {
		return mysqlCtr, err
	}
	containerName, _ := mysqlCtr.container.Name(ctx)
	log.Infof("MySQL container started with Name: %s\n", containerName)
	return mysqlCtr, nil
}

func generateDsn(user, password, host, port, dbName string) string {
	sqlClientCfg := mysqlclient.Config{
		User:                     user,
		Passwd:                   password,
		Net:                      "tcp",
		Addr:                     fmt.Sprintf("%s:%s", host, port),
		DBName:                   dbName,
		Params:                   nil,
		Collation:                "",
		Loc:                      nil,
		Timeout:                  1 * time.Second,
		ReadTimeout:              1 * time.Second,
		WriteTimeout:             1 * time.Second,
		AllowAllFiles:            false,
		AllowCleartextPasswords:  true,
		AllowFallbackToPlaintext: true,
		AllowNativePasswords:     true,
		AllowOldPasswords:        true,
		CheckConnLiveness:        true,
		ClientFoundRows:          false,
		ColumnsWithAlias:         false,
		InterpolateParams:        false,
		MultiStatements:          false,
		ParseTime:                true,
	}
	return sqlClientCfg.FormatDSN()
}

// teardownTestDB closes DB and terminates container if running
func teardownTestDB(ctx context.Context) {
	if testDB != nil {
		_ = testDB.Close()
	}

}

func ensureProperTransactionsPresent() error {
	_, err := testDB.Exec(`
create table users
(
    username    varchar(200) not null
        primary key,
    date_of_birth DATE         not null
);

`)
	pastDate := time.Now().AddDate(-1, 0, -1).Format("2006-01-02")
	pastDate2 := time.Now().AddDate(-1, -1, -1).Format("2006-01-02")
	futureDay := time.Now().AddDate(-1, 0, 3).Format("2006-01-02")
	today := time.Now().AddDate(-1, 0, 0).Format("2006-01-02")
	res, err := testDB.Exec(
		`INSERT INTO users (username, date_of_birth) VALUES
				("getexistingthreedays", ?),
				("getexistingtoday", ?),
				("existingusersame", ?),
				("existinguserdiff", ?);
`, futureDay, today, pastDate, pastDate2)

	if err != nil {
		return fmt.Errorf("failed to insert test data: %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected != 4 {
		return fmt.Errorf("expected 4 rows affected, got %d", rowsAffected)
	}
	return nil
}

// connectExternalDB connects using TEST_DATABASE_URL when provided
func connectToDb(connStrDsn string) error {
	var err error
	fmt.Println("Connecting to database with DSN:", connStrDsn)
	testDB, err = sql.Open("mysql", connStrDsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	if err := testDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	return nil
}
