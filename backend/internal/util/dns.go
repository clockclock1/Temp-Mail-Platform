package util

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"
	"time"
)

const dnsLookupTimeout = 3 * time.Second

var (
	defaultDNSResolvers = []string{"1.1.1.1:53", "8.8.8.8:53", "9.9.9.9:53"}
	dnsResolverMu       sync.RWMutex
	dnsResolvers        = append([]string(nil), defaultDNSResolvers...)
)

type mxResolver interface {
	LookupMX(ctx context.Context, name string) ([]*net.MX, error)
}

type systemMXResolver struct{}

func (systemMXResolver) LookupMX(ctx context.Context, name string) ([]*net.MX, error) {
	return (&net.Resolver{}).LookupMX(ctx, name)
}

type dnsServerMXResolver struct {
	server string
}

func (r dnsServerMXResolver) LookupMX(ctx context.Context, name string) ([]*net.MX, error) {
	dialer := &net.Dialer{Timeout: dnsLookupTimeout}
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, r.server)
		},
	}
	return resolver.LookupMX(ctx, name)
}

func SetDNSResolvers(resolvers []string) {
	dnsResolverMu.Lock()
	dnsResolvers = normalizeResolverList(resolvers)
	dnsResolverMu.Unlock()
}

func DNSResolvers() []string {
	dnsResolverMu.RLock()
	defer dnsResolverMu.RUnlock()
	return append([]string(nil), dnsResolvers...)
}

func LookupMXRecords(domainName string) ([]string, error) {
	domainName = strings.ToLower(strings.TrimSpace(domainName))
	if domainName == "" {
		return nil, fmt.Errorf("domain name cannot be empty")
	}

	resolvers := []mxResolver{systemMXResolver{}}
	for _, server := range DNSResolvers() {
		resolvers = append(resolvers, dnsServerMXResolver{server: server})
	}
	return lookupMXRecords(domainName, resolvers)
}

func lookupMXRecords(domainName string, resolvers []mxResolver) ([]string, error) {
	var attempts []string

	for idx, resolver := range resolvers {
		ctx, cancel := context.WithTimeout(context.Background(), dnsLookupTimeout)
		records, err := resolver.LookupMX(ctx, domainName)
		cancel()
		if err != nil {
			attempts = append(attempts, fmt.Sprintf("resolver %d: %v", idx+1, err))
			continue
		}
		values, err := normalizeMXRecords(domainName, records)
		if err != nil {
			attempts = append(attempts, fmt.Sprintf("resolver %d: %v", idx+1, err))
			continue
		}
		return values, nil
	}

	if len(attempts) == 0 {
		return nil, fmt.Errorf("lookup mx for %s failed", domainName)
	}
	return nil, fmt.Errorf("lookup mx for %s failed: %s", domainName, strings.Join(attempts, "; "))
}

func normalizeMXRecords(domainName string, records []*net.MX) ([]string, error) {
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

func normalizeResolverList(resolvers []string) []string {
	if len(resolvers) == 0 {
		return append([]string(nil), defaultDNSResolvers...)
	}

	normalized := make([]string, 0, len(resolvers))
	seen := map[string]struct{}{}
	for _, resolver := range resolvers {
		resolver = strings.TrimSpace(resolver)
		if resolver == "" {
			continue
		}
		if _, _, err := net.SplitHostPort(resolver); err != nil {
			resolver = net.JoinHostPort(resolver, "53")
		}
		if _, ok := seen[resolver]; ok {
			continue
		}
		seen[resolver] = struct{}{}
		normalized = append(normalized, resolver)
	}
	if len(normalized) == 0 {
		return append([]string(nil), defaultDNSResolvers...)
	}
	return normalized
}
