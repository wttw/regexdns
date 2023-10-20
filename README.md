# regexdns
A PowerDNS pipe backend to serve responses based on regular expression matches.

## Why?

Sometimes you want to create a lot of DNS records in a way that's a bit more flexible
than DNS wildcards, but you don't want to have to pregenerate huge zone files every time
some data changes.

The particular case I was thinking of was an ESP providing customer-specific hostnames for
customers too small to own their own domains, and needing a way of providing authentication
records for all of them, as I [discussed on our blog](https://wordtothewise.com/2023/10/customer-subdomain-authentication/).

## Usage

Add the `pipe` backend to the `launch=` PowerDNS parameter, after the more normal backends, e.g. `launch=bind,pipe`

Set `zone-cache-refresh-interval=0` - the pipe backend doesn't support listing all supported zones. Older versions
of PowerDNS may not support this setting.

Set `pipe-command=/path/to/regexdns --config /path/to/regexdns/zone/file`

## Configuration

regexdns takes a file similar to a bind zone file as input.

Blank lines are ignored, as are any lines starting with a semicolon.

Remaining lines must contain space-separated fields:

1. Query name
2. Class (this should always be "IN")
3. TTL, the ttl in seconds of the DNS response
4. Query type, any DNS type understood by PowerDNS, e.g. "MX" or "A" or "CNAME"
5. The answer to this query, in [PowerDNS format](https://doc.powerdns.com/authoritative/appendices/types.html)

The query name can be a literal hostname, such as "example.com", or it can be a 
[Go style regular expression](https://github.com/google/re2/wiki/Syntax).

If the query name is a regular expression and has capturing groups, then the answer to the query is treated
as a template, and any template variables such as `$1` or `${name}` will be replaced by the corresponding
value captured from the query name.

For example, if you were using this configuration file

```
example.com 3600 IN SOA ns1.example.com. zonmaster.example.com. 1 10800 3600 2592000 600
www.example.com 3600 IN A 10.11.12.13
([^.]+)\.example\.com 3600 IN CNAME $1.wesendmail.com
(?P<selector>[^.]+)\._domainkey\.(?P<user>[^.]+)\.example\.com 3600 IN CNAME ${selector}.dkim.wesendmail.com
gopher.example.com 3600 IN A 10.11.12.13
```

A query for the address of www.example.com would return 10.11.12.13.

A query for the MX record of www.example.com would return no records.

A query for hello.example.com would return a CNAME to hello.wesendmail.com.

A query for the address of gopher.example.com would return a CNAME to gopher.example.com. The first matching query
name in the file is used, so this query would match line 3, not line 5.

A query for a TXT record for k1._domainkey.steve.example.com would return a CNAME to k1.dkim.wesendmail.com

The configuration file must contain an SOA record for any zone regexdns is expected to serve.

## Future

The code could be extended fairly easily to support serving the pipe on a unix socket, or to support the
PowerDNS remote backends in addition to the pipe backends.

Given there's no concept of all the hostnames that are in a zone I don't think that it can support DNSSEC.
