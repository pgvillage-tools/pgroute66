package internal

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

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

// RouteConfig defines all config for the api
type RouteConfig struct {
	Hosts    RouteHostsConfig `yaml:"hosts"`
	Groups   RouteHostGroups  `yaml:"groups"`
	Bind     string           `yaml:"bind"`
	Port     int              `yaml:"port"`
	Ssl      RouteSSLConfig   `yaml:"ssl"`
	LogLevel string           `yaml:"loglevel"`
	LogFile  string           `yaml:"logfile"`
}

// NewConfig initializes and returns a route config
func NewConfig() (config RouteConfig, err error) {
	var debug bool

	var version bool

	var configFile string

	flag.BoolVar(&debug, "d", false, "Add debugging output")
	flag.BoolVar(&version, "v", false, "Show version information")

	flag.StringVar(&configFile, "c", os.Getenv(envConfName), "Path to configfile")

	flag.Parse()

	if version {
		fmt.Println(appVersion)
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
	yamlConfig, err := ioutil.ReadFile(configFile)
	if err != nil {
		return config, err
	}

	if err = yaml.Unmarshal(yamlConfig, &config); err != nil {
		return RouteConfig{}, err
	} else if debug {
		config.LogLevel = debugLoglevel
	} else {
		config.LogLevel = strings.ToLower(config.LogLevel)
	}

	return config, nil
}

// GroupHosts returns a list of hosts that are part of a group as defined in rc.HostGroups.
// HostGroup "all" is a special placeholder for all hosts defined in rc.Hosts.
func (rc RouteConfig) GroupHosts(groupName string) RouteHostGroup {
	if groupName == "all" {
		var rhg RouteHostGroup
		for host := range rc.Hosts {
			rhg = append(rhg, host)
		}

		return rhg
	}

	groupHosts, ok := rc.Groups[groupName]
	if !ok {
		globalHandler.log.Errorf("hostgroup %s is not defined", groupName)

		return RouteHostGroup{}
	}
	return groupHosts
}

// BindTo returns the string of the host/port to bind to
func (rc RouteConfig) BindTo() string {
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
func (rc RouteConfig) Debug() bool {
	return rc.LogLevel == debugLoglevel
}
