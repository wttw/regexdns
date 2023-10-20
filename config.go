package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

type Config struct {
	Records []Record
	Zones   []string
}

type RR struct {
	QType   string `json:"qtype"`
	QName   string `json:"qname"`
	Content string `json:"content"`
	Ttl     int    `json:"ttl"`
}

func (r RR) PipeResponse(id string) string {
	var content string
	switch r.QType {
	case "MX", "SRV":
		content = strings.Join(strings.Fields(r.Content), "\t")
	default:
		content = r.Content
	}
	return fmt.Sprintf("%s\tIN\t%s\t%d\t%s\t%s", r.QName, r.QType, r.Ttl, id, content)
}

type Record struct {
	Query    *regexp.Regexp
	Response map[string][]RR
}

func NewConfig(r io.Reader) (Config, error) {
	var hostPatterns []string
	records := map[string]map[string][]RR{}
	fieldsRe := regexp.MustCompile(`^(\S+)\s+(\d+)\s+IN\s+([A-Z]+)\s*(.+)$`)
	notHostnameRe := regexp.MustCompile(`[^a-zA-Z0-9_./-]`)
	hasCaseRe := regexp.MustCompile(`^\(\?[imsU]+\)`)
	scanner := bufio.NewScanner(r)
	var zones []string
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, ";") {
			continue
		}
		if line == "" {
			continue
		}
		matches := fieldsRe.FindStringSubmatch(line)
		if matches == nil {
			return Config{}, fmt.Errorf("failed to read line %d: %s", lineNum, line)
		}
		var qPattern string
		qname := strings.TrimSuffix(strings.TrimSuffix(matches[1], "\\."), ".")
		ttl, err := strconv.Atoi(matches[2])
		if err != nil {
			return Config{}, fmt.Errorf("invalid TTL '%s' on line %d: %w", matches[2], lineNum, err)
		}
		qtype := strings.ToUpper(matches[3])
		content := matches[4]
		if !notHostnameRe.MatchString(qname) {
			// it's not a regex
			qPattern = "(?i)^" + regexp.QuoteMeta(qname) + "$"
			if qtype == "SOA" {
				zones = append(zones, strings.ToLower(qname))
			}
		} else {
			qPattern = qname
			if !hasCaseRe.MatchString(qPattern) {
				qPattern = "(?i)^" + qPattern + "$"
			}
			_, err := regexp.Compile(qPattern)
			if err != nil {
				return Config{}, fmt.Errorf("bad regexp '%s' on line %d: %w", qPattern, lineNum, err)
			}
		}

		rec, ok := records[qPattern]
		if !ok {
			hostPatterns = append(hostPatterns, qPattern)
			rec = map[string][]RR{}
		}
		rec[qtype] = append(rec[qtype], RR{
			QType:   qtype,
			QName:   qname,
			Content: content,
			Ttl:     ttl,
		})
		records[qPattern] = rec
	}
	if scanner.Err() != nil {
		return Config{}, scanner.Err()
	}
	slices.Sort(zones)
	conf := Config{
		Zones: slices.Compact(zones),
	}
	for _, pattern := range hostPatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return Config{}, fmt.Errorf("internal error, failed to compile '%s': %w", pattern, err)
		}
		conf.Records = append(conf.Records, Record{
			Query:    re,
			Response: records[pattern],
		})
	}
	return conf, nil
}

func LoadConfig(configFile string) (Config, error) {
	configReader, err := os.Open(configFile)
	if err != nil {
		return Config{}, err
	}
	return NewConfig(configReader)
}
