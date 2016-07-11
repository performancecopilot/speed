package speed

import (
	"os"
	"testing"
)

func TestRootPath(t *testing.T) {
	if rootPath == "" {
		t.Errorf("RootPath is invalid")
		return
	}

	_, err := os.Stat(rootPath)
	if err != nil {
		t.Errorf("RootPath err: %s", err)
	}
}

func TestConfPath(t *testing.T) {
	if confPath == "" {
		t.Errorf("ConfPath is invalid")
		return
	}

	fi, err := os.Stat(confPath)
	if err != nil {
		return
	}

	if !fi.Mode().IsRegular() {
		t.Errorf("%s should be a regular file", confPath)
		return
	}
}

var keysToTest = []string{
	"PCP_VERSION",
	"PCP_USER",
	"PCP_GROUP",
	"PCP_PLATFORM",
	"PCP_PLATFORM_PATHS",
	"PCP_ETC_DIR",
	"PCP_SYSCONF_DIR",
	"PCP_SYSCONFIG_DIR",
	"PCP_RC_DIR",
	"PCP_BIN_DIR",
	"PCP_BINADM_DIR",
	"PCP_LIB_DIR",
	"PCP_LIB32_DIR",
	"PCP_SHARE_DIR",
	"PCP_INC_DIR",
	"PCP_MAN_DIR",
	"PCP_PMCDCONF_PATH",
	"PCP_PMCDOPTIONS_PATH",
	"PCP_PMCDRCLOCAL_PATH",
	"PCP_PMPROXYOPTIONS_PATH",
	"PCP_PMWEBDOPTIONS_PATH",
	"PCP_PMMGROPTIONS_PATH",
	"PCP_PMIECONTROL_PATH",
	"PCP_PMSNAPCONTROL_PATH",
	"PCP_PMLOGGERCONTROL_PATH",
	"PCP_PMDAS_DIR",
	"PCP_RUN_DIR",
	"PCP_PMDAS_DIR",
	"PCP_LOG_DIR",
	"PCP_TMP_DIR",
	"PCP_TMPFILE_DIR",
	"PCP_DOC_DIR",
	"PCP_DEMOS_DIR",
}

func TestConfig(t *testing.T) {
	if config == nil {
		return
	}

	for _, key := range keysToTest {
		_, ok := config[key]
		if !ok {
			t.Errorf("key %s not present in Config", key)
		}
	}
}
