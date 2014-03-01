package lamp

import (
	"github.com/codegangsta/martini"
)

// NewApp instantiates the application
func NewApp(configPath string) (*martini.ClassicMartini, string, error) {
	// Create app
	m := martini.Classic()

	// Create config service
	config, err := NewConfig(configPath)
	if err != nil {
		return nil, "", err
	}

	// Create database service
	conn, err := NewDatabaseConn(config)
	if err != nil {
		return nil, "", err
	}

	// Map services
	m.Map(config)
	m.Map(conn)

	// Add routes

	return m, config.Port, nil
}