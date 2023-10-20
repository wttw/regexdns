package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func ServePipe(config Config, r io.Reader, w io.Writer) error {
	s := bufio.NewScanner(r)
	if !s.Scan() {
		return errors.New("failed to read line")
	}
	helo := s.Text()
	if !strings.HasPrefix(helo, "HELO\t") {
		_, _ = io.WriteString(w, "FAIL\n")
		return fmt.Errorf("bad HELO: %s", helo)
	}
	protocolVersion, err := strconv.Atoi(helo[5:])
	if err != nil {
		return fmt.Errorf("bad HELO: %s: %w", helo, err)
	}
	var expectedQueryLength int
	var responsePrefix = "DATA\t"
	switch protocolVersion {
	case 1:
		expectedQueryLength = 6
	case 2:
		expectedQueryLength = 7
	case 3, 4, 5:
		expectedQueryLength = 8
		responsePrefix = "DATA\t0\t1\t"
	default:
		_, _ = io.WriteString(w, "FAIL\n")
		return fmt.Errorf("unexpected powerdns protocol version %d", protocolVersion)
	}
	fmt.Fprintf(w, "OK\tregexdns starting with %s for %s\n", configFile, strings.Join(config.Zones, ", "))

	for s.Scan() {
		req := s.Text()
		if strings.HasPrefix(req, "Q\t") {
			fields := strings.Split(req, "\t")
			if len(fields) != expectedQueryLength {
				fmt.Fprintf(os.Stderr, "unparsable query\n")
				_, err = io.WriteString(w, "LOG\tPowerDNS sent unparseable query\nFAIL\n")
				if err != nil {
					return err
				}
				continue
			}
			qname := fields[1]
			qclass := fields[2]
			qtype := fields[3]
			id := fields[4]
			if qclass != "IN" {
				fmt.Fprintf(os.Stderr, "unexpected class\n")
				_, err = fmt.Fprintf(w, "LOG\tUnexpected class '%s'\nFAIL\n", qclass)
				if err != nil {
					return err
				}
				continue
			}
			responses := Respond(config, qname, qtype, id)
			for _, response := range responses {
				_, err = io.WriteString(w, responsePrefix+response.PipeResponse(id)+"\n")
				if err != nil {
					return err
				}
			}
			_, err = io.WriteString(w, "END\n")
			if err != nil {
				return err
			}
			continue
		}
		if strings.HasPrefix(req, "AXFR\t") {
			// We don't do AXFR
			_, err = io.WriteString(w, "FAIL\n")
			if err != nil {
				return err
			}
			continue
		}
		if strings.HasPrefix(req, "CMD\t") {
			// Commands!
			if req != "CMD\treload" {
				_, err = io.WriteString(w, "'reload' is the only valid command\nEND\n")
				if err != nil {
					return err
				}
				continue
			}
			newConfig, err := LoadConfig(configFile)
			if err != nil {
				_, _ = fmt.Fprintf(w, "Failed to read configuration: %s\nEND\n", err)
				continue
			}
			config = newConfig
			_, err = io.WriteString(w, "Configuration reloaded\nEND\n")
			if err != nil {
				return err
			}
			continue
		}
		_, err = io.WriteString(w, "LOG\tunexpected command from PowerDNS\nFAIL\n")
		if err != nil {
			return err
		}
		continue
	}
	return s.Err()
}
