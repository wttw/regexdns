package main

func expandResponses(rows []RR, qname string, record Record, rrs []RR, matches []int) []RR {
	for _, rr := range rrs {
		rows = append(rows, RR{
			QType:   rr.QType,
			QName:   qname,
			Content: string(record.Query.ExpandString(nil, rr.Content, qname, matches)),
			Ttl:     rr.Ttl,
		})
	}
	return rows
}

func Respond(c Config, qname string, qtype string, id string) []RR {
	for _, record := range c.Records {
		matches := record.Query.FindStringSubmatchIndex(qname)
		if len(matches) == 0 {
			continue
		}
		var rows []RR
		if qtype == "ANY" {
			for _, rrs := range record.Response {
				rows = expandResponses(rows, qname, record, rrs, matches)
			}
		} else {
			rrs, ok := record.Response[qtype]
			if ok {
				rows = expandResponses(rows, qname, record, rrs, matches)
			}
		}
		return rows
	}
	return nil
}
