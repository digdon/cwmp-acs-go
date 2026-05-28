package cwmp

import (
	"fmt"
	"strings"
)

type SupportedCwmpVersion int

const (
	UNKNOWN_CWMP_VERSION SupportedCwmpVersion = iota
	CWMP_V1_0
	CWMP_V1_1
	CWMP_V1_2
	CWMP_V1_3
	CWMP_V1_4
)

var AllVersions = []SupportedCwmpVersion{CWMP_V1_0, CWMP_V1_1, CWMP_V1_2, CWMP_V1_3, CWMP_V1_4}

func (v SupportedCwmpVersion) VersionMajor() int {
	return 1
}

func (v SupportedCwmpVersion) VersionMinor() int {
	switch v {
	case CWMP_V1_0:
		return 0
	case CWMP_V1_1:
		return 1
	case CWMP_V1_2:
		return 2
	case CWMP_V1_3:
		return 3
	case CWMP_V1_4:
		return 4
	default:
		return -1 // invalid
	}
}

func (v SupportedCwmpVersion) String() string {
	switch v {
	case CWMP_V1_0:
		return "1.0"
	case CWMP_V1_1:
		return "1.1"
	case CWMP_V1_2:
		return "1.2"
	case CWMP_V1_3:
		return "1.3"
	case CWMP_V1_4:
		return "1.4"
	default:
		return "unknown"
	}
}

func MaxCommonVersion(possibleVersions string) SupportedCwmpVersion {
	if possibleVersions == "" {
		return UNKNOWN_CWMP_VERSION
	}

	cwmpVersion := UNKNOWN_CWMP_VERSION
	maxMajorVersion, maxMinorVersion := -1, -1

	versions := strings.Split(possibleVersions, ",")

	for _, version := range versions {
		version = strings.TrimSpace(version)
		var major, minor int
		_, err := fmt.Sscanf(version, "%d.%d", &major, &minor)
		if err != nil {
			fmt.Println(version, "does not appear to be a proper version number - skipping")
			continue
		}

		for _, available := range AllVersions {
			if major == available.VersionMajor() && minor == available.VersionMinor() {
				if major > maxMajorVersion || (major == maxMajorVersion && minor > maxMinorVersion) {
					maxMajorVersion = major
					maxMinorVersion = minor
					cwmpVersion = available
				}
			}
		}
	}

	return cwmpVersion
}
