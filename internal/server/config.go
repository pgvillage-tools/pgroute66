package server

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	v1 "github.com/pgvillage-tools/pgroute66/api/v1"
	"github.com/pgvillage-tools/pgroute66/internal/version"
	"gopkg.in/yaml.v2"
)

/*
 * This module reads the config file and returns a config object with all entries from the config yaml file.
 */

const (
	envConfName     = "PGROUTE66CONFIG"
	defaultConfFile = "/etc/pgroute66/config.yaml"
	debugLoglevel   = "debug"
)

// Config defines all config for the api
type Config struct {
	Hosts    v1.HostsConfig `yaml:"hosts"`
	Groups   v1.HostGroups  `yaml:"groups"`
	Bind     string         `yaml:"bind"`
	Port     int            `yaml:"port"`
	Ssl      v1.SSLConfig   `yaml:"ssl"`
	LogLevel string         `yaml:"loglevel"`
	LogFile  string         `yaml:"logfile"`
}

// NewConfig initializes and returns a route config
func NewConfig() (config Config, err error) {
	var debug bool

	var showVersion bool

	var configFile string

	flag.BoolVar(&debug, "d", false, "Add debugging output")
	flag.BoolVar(&showVersion, "v", false, "Show version information")

	flag.StringVar(&configFile, "c", os.Getenv(envConfName), "Path to configfile")

	flag.Parse()

	if showVersion {
		fmt.Println(version.Version)
		os.Exit(0)
	}

	if configFile == "" {
		configFile = defaultConfFile
	}

	configFile, err = filepath.EvalSymlinks(configFile)
	if err != nil {
		return config, err
	}

	// This only parsed as yaml, nothing else
	// #nosec
	yamlConfig, err := os.ReadFile(configFile)
	if err != nil {
		return config, err
	}

	if err = yaml.Unmarshal(yamlConfig, &config); err != nil {
		return Config{}, err
	} else if debug {
		config.LogLevel = debugLoglevel
	} else {
		config.LogLevel = strings.ToLower(config.LogLevel)
	}

	return config, nil
}

// GroupHosts returns a list of hosts that are part of a group as defined in rc.HostGroups.
// HostGroup "all" is a special placeholder for all hosts defined in rc.Hosts.
func (rc Config) GroupHosts(groupName string) v1.HostGroup {
	if groupName == "all" {
		var rhg v1.HostGroup
		for host := range rc.Hosts {
			rhg = append(rhg, host)
		}

		return rhg
	}

	groupHosts, ok := rc.Groups[groupName]
	if !ok {
		globalHandler.log.Errorf("hostgroup %s is not defined", groupName)

		return v1.HostGroup{}
	}
	return groupHosts
}

// BindTo returns the string of the host/port to bind to
func (rc Config) BindTo() string {
	port := rc.Port
	if port == 0 {
		if rc.Ssl.Enabled() {
			port = defaultSSLPort
		} else {
			port = defaultNoSSLPort
		}
	}

	if rc.Bind == "" {
		return fmt.Sprintf("localhost:%d", port)
	}

	return fmt.Sprintf("%s:%d", rc.Bind, port)
}

// Debug returns the debug level of this route
func (rc Config) Debug() bool {
	return rc.LogLevel == debugLoglevel
}
