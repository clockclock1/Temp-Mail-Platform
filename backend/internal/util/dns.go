package util

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"
)

func LookupMXRecords(domainName string) ([]string, error) {
	domainName = strings.ToLower(strings.TrimSpace(domainName))
	if domainName == "" {
		return nil, fmt.Errorf("domain name cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resolver := net.Resolver{}
	records, err := resolver.LookupMX(ctx, domainName)
	if err != nil {
		return nil, fmt.Errorf("lookup mx for %s: %w", domainName, err)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("mx record not found for %s", domainName)
	}

	values := make([]string, 0, len(records))
	for _, record := range records {
		host := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(record.Host)), ".")
		if host == "" {
			continue
		}
		values = append(values, fmt.Sprintf("%d %s", record.Pref, host))
	}
	if len(values) == 0 {
		return nil, fmt.Errorf("mx record not found for %s", domainName)
	}
	sort.Strings(values)
	return values, nil
}
