package util

import "strings"

func IsDomainOrSubdomain(domainName, baseDomain string) bool {
	domainName = strings.ToLower(strings.TrimSpace(domainName))
	baseDomain = strings.ToLower(strings.TrimSpace(baseDomain))
	if domainName == "" || baseDomain == "" {
		return false
	}
	return domainName == baseDomain || strings.HasSuffix(domainName, "."+baseDomain)
}

func DomainDepth(domainName string) int {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(domainName)), ".")
	count := 0
	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			count++
		}
	}
	return count
}
