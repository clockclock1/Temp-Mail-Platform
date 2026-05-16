package util

import (
	"context"
	"errors"
	"net"
	"reflect"
	"testing"
)

type stubMXResolver struct {
	records []*net.MX
	err     error
}

func (r stubMXResolver) LookupMX(context.Context, string) ([]*net.MX, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.records, nil
}

func TestLookupMXRecordsFallsBackToNextResolver(t *testing.T) {
	values, err := lookupMXRecords("example.com", []mxResolver{
		stubMXResolver{err: errors.New("lookup example.com on 127.0.0.11:53: no such host")},
		stubMXResolver{records: []*net.MX{
			{Pref: 20, Host: "mx2.example.com."},
			{Pref: 10, Host: "mx1.example.com"},
		}},
	})
	if err != nil {
		t.Fatalf("lookupMXRecords returned error: %v", err)
	}

	expected := []string{"10 mx1.example.com", "20 mx2.example.com"}
	if !reflect.DeepEqual(values, expected) {
		t.Fatalf("expected %v, got %v", expected, values)
	}
}

func TestLookupMXRecordsFailsWhenEveryResolverFails(t *testing.T) {
	_, err := lookupMXRecords("example.com", []mxResolver{
		stubMXResolver{err: errors.New("resolver 1 failed")},
		stubMXResolver{records: []*net.MX{}},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
