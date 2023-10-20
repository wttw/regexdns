package main

import (
	"github.com/google/go-cmp/cmp"
	"sort"
	"testing"
)

func TestRespond(t *testing.T) {
	tests := map[string]struct {
		qname string
		qtype string
		want  []string
	}{
		"regular_nomatch":       {"wurst1_com", "TXT", nil},
		"regular_match":         {"wurst1.com", "TXT", []string{"wurst1.com\tIN\tTXT\t3600\t1\t\"hello\""}},
		"regular_wrongtype":     {"wurst1.com", "MX", nil},
		"static_expansion":      {"foo.wurst2.com", "MX", []string{"foo.wurst2.com\tIN\tMX\t3600\t1\t10\treplies.sausagemail.com"}},
		"unnamed_expansion":     {"bar.wurst3.com", "MX", []string{"bar.wurst3.com\tIN\tMX\t3600\t1\t20\tbar.sausagemail.com"}},
		"named_expansion":       {"baz.wurst4.com", "MX", []string{"baz.wurst4.com\tIN\tMX\t3600\t1\t30\tbaz.sausagemail.com"}},
		"multiple_matches":      {"foo._domainkey.bar.wurst5.com", "CNAME", []string{"foo._domainkey.bar.wurst5.com\tIN\tCNAME\t3600\t1\tfoo.dkim.sausagemail.com"}},
		"multiple_replacements": {"foo.bar.wurst6.com", "CNAME", []string{"foo.bar.wurst6.com\tIN\tCNAME\t3600\t1\tbar.foo.sausagemail.com"}},
		"just_txt":              {"foo.wurst7.com", "TXT", []string{"foo.wurst7.com\tIN\tTXT\t3600\t1\t\"Hello foo\""}},
		"just_mx":               {"foo.wurst7.com", "MX", []string{"foo.wurst7.com\tIN\tMX\t3600\t1\t40\tfoo.sausagemail.com"}},
		"any":                   {"foo.wurst7.com", "ANY", []string{"foo.wurst7.com\tIN\tTXT\t3600\t1\t\"Hello foo\"", "foo.wurst7.com\tIN\tMX\t3600\t1\t40\tfoo.sausagemail.com"}},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			config, err := LoadConfig("testdata/sausagemail.zone")
			if err != nil {
				t.Fatal(err)
			}
			var got []string
			responses := Respond(config, test.qname, test.qtype, "1")
			for _, r := range responses {
				got = append(got, r.PipeResponse("1"))
			}
			sort.Strings(got)
			sort.Strings(test.want)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("result mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
