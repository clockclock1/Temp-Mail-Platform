package util

import "strings"

func DomainLevelFromName(name string) int {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(name)), ".")
	level := 0
	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			level++
		}
	}
	if level <= 0 {
		level = 2
	}
	if level > 10 {
		level = 10
	}
	return level
}

func NormalizeDomainLevel(level int) int {
	if level <= 0 {
		return 2
	}
	if level < 1 {
		return 1
	}
	if level > 10 {
		return 10
	}
	return level
}

func NormalizeRandomDomainLevelRange(minLevel, maxLevel int) (int, int) {
	if minLevel <= 0 {
		minLevel = 1
	}
	if maxLevel <= 0 {
		maxLevel = 7
	}
	if minLevel < 1 {
		minLevel = 1
	}
	if maxLevel > 7 {
		maxLevel = 7
	}
	if minLevel > maxLevel {
		minLevel, maxLevel = maxLevel, minLevel
	}
	return minLevel, maxLevel
}
