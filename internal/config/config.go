package config

import (
	"flag"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"

	apiv1 "github.com/pgvillage-tools/pgroute66/api/v1"
	"github.com/pgvillage-tools/pgroute66/internal/logging"
	"github.com/pgvillage-tools/pgroute66/internal/version"
	"gopkg.in/yaml.v2"
)

const (
	defaultSSLPort   = 8443
	defaultNoSSLPort = 8080

	// DefaultHostGroup is a special placeholder for all hosts defined in rc.Hosts.
	DefaultHostGroup = "default"
)

/*
 * This module reads the config file and returns a config object with all entries from the config yaml file.
 */

const (
	envConfName     = "PGROUTE66CONFIG"
	defaultConfFile = "/etc/pgroute66/config.yaml"
)

// Config defines all config for the api
type Config struct {
	Hosts    apiv1.HostsConfig `yaml:"hosts"`
	Groups   apiv1.HostGroups  `yaml:"groups"`
	Bind     string            `yaml:"bind"`
	Port     int               `yaml:"port"`
	Ssl      apiv1.SSLConfig   `yaml:"ssl"`
	LogLevel string            `yaml:"loglevel"`
	LogFile  string            `yaml:"logfile"`
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
		fmt.Println(version.GetAppVersion())
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
	}
	if debug {
		config.LogLevel = "debug"
	}
	logging.SetStaticLevel(strings.ToLower(config.LogLevel))

	return config, nil
}

// GetHostGroups returns a combination of the host groups and the hosts
func (rc Config) GetHostGroups() apiv1.HostGroups {
	hg := apiv1.HostGroups{DefaultHostGroup: rc.Hosts}
	maps.Copy(hg, rc.Groups)
	return hg
}

// GroupHosts returns a list of hosts that are part of a group as defined in rc.HostGroups.
func (rc Config) GroupHosts(groupName string) apiv1.HostsConfig {
	groupHosts, ok := rc.Groups[groupName]
	if !ok {
		return rc.Hosts
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
