package speed

import (
	"bufio"
	"os"
	"path"
	"regexp"
)

// RootPath stores path to the pcp root installation
var RootPath string

// ConfPath stores path to pcp.conf
var ConfPath string

// Config stores the configuration as defined in current PCP environment
var Config map[string]string

// pat stores a valid key-value pattern line
var pat = "([A-Z0-9_]+)=(.*)"

// initConfig initializes the config constants
func initConfig() error {
	re, _ := regexp.Compile(pat)

	rootPath, ok := os.LookupEnv("PCP_DIR")
	if !ok {
		rootPath = "/"
	}
	RootPath = rootPath

	confPath, ok := os.LookupEnv("PCP_CONF")
	if !ok {
		confPath = path.Join(RootPath, "etc", "pcp.conf")
	}
	ConfPath = confPath

	f, err := os.Open(ConfPath)
	if err != nil {
		return err
	}

	// if we reach at this point, it means we have a valid config
	// that can be read, so we can make the map non-nil
	Config = make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		t := scanner.Text()
		if re.MatchString(t) {
			matches := re.FindStringSubmatch(t)
			Config[matches[1]] = matches[2]
		}
	}

	return nil
}
