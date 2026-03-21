package testdb

import (
	"context"
	"fmt"
	"testing"

	"github.com/dxe/adb/config"
	tc "github.com/testcontainers/testcontainers-go"
	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
)

const (
	testContainerDBUser     = "root"
	testContainerDBPassword = "adbpassword"
)

func StartMySQLContainer(ctx context.Context) (func(), error) {
	container, err := tcmysql.Run(
		ctx,
		"mysql:8.4",
		tcmysql.WithDatabase(config.TestDBName),
		tcmysql.WithUsername(testContainerDBUser),
		tcmysql.WithPassword(testContainerDBPassword),
		tc.WithTmpfs(map[string]string{
			"/var/lib/mysql": "rw",
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("start mysql testcontainer: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("get mysql testcontainer host: %w", err)
	}
	port, err := container.MappedPort(ctx, "3306/tcp")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("get mysql testcontainer port: %w", err)
	}

	config.DBUser = testContainerDBUser
	config.DBPassword = testContainerDBPassword
	config.DBName = config.TestDBName
	config.DBProtocol = fmt.Sprintf("tcp(%s:%s)", host, port.Port())
	config.DataSourceBase = fmt.Sprintf("%s:%s@%s", config.DBUser, config.DBPassword, config.DBProtocol)

	return func() {
		if err := container.Terminate(ctx); err != nil {
			fmt.Printf("warning: failed to terminate mysql testcontainer: %v\n", err)
		}
	}, nil
}

func RunWithMySQLContainer(m *testing.M) int {
	ctx := context.Background()

	cleanup, err := StartMySQLContainer(ctx)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	return m.Run()
}
