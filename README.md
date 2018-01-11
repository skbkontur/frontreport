[![Build Status](https://travis-ci.org/skbkontur/frontreport.svg?branch=master)](https://travis-ci.org/skbkontur/frontreport) [![Go Report Card](https://goreportcard.com/badge/github.com/skbkontur/frontreport)](https://goreportcard.com/report/github.com/skbkontur/frontreport) [![Join the chat at https://gitter.im/frontreport/frontreport](https://badges.gitter.im/frontreport/frontreport.svg)](https://gitter.im/frontreport/frontreport?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)


## What is this tool for?

Frontreport is useful only if you have an existing infrastructure for backend log aggregation. For example, we use ELK stack with RabbitMQ as a broker on top. So, your logging infrastructure may look like this:

Backend application → Logstash → RabbitMQ → Elastic RabbitMQ River → Elastic → Kibana

You may want to reuse this infrastructure to collect frontend logs from browsers of your visitors. So, you need to replace Logstash in the above scheme with something fast that can validate incoming JSON and manage high load by batching documents.

Frontreport does all that. Resulting architecture is something like the following:

Browser → Frontreport → RabbitMQ → Elastic RabbitMQ River → Elastic → Kibana

See code for details or ask us on [Gitter][].


## Usage

```
Usage:
  frontreport [OPTIONS]

Application Options:
  -p, --port=                port to listen (default: 8888) [$FRONTREPORT_PORT]
  -a, --amqp=                AMQP connection string (default: amqp://guest:guest@localhost:5672/) [$FRONTREPORT_AMQP]
  -s, --service-whitelist=   list of services to accept reports from (all are allowed if not specified) [$FRONTREPORT_SERVICE_WHITELIST]
  -d, --domain-whitelist=    list of domains to accept CORS requests from (all are allowed if not specified) [$FRONTREPORT_DOMAIN_WHITELIST]
  -t, --sourcemap-whitelist= trusted sourcemap pattern (default: ^(http|https)://localhost/) [$FRONTREPORT_SOURCEMAP_WHITELIST]
  -l, --logfile=             log file name (writes to stdout if not specified) [$FRONTREPORT_LOGFILE]
  -g, --graphite=            Graphite connection string for internal metrics [$FRONTREPORT_GRAPHITE]
  -r, --graphite-prefix=     prefix for Graphite metrics [$FRONTREPORT_GRAPHITE_PREFIX]
  -v, --version              print version and exit

Help Options:
  -h, --help             Show this help message
```


## What can you collect from browsers?

1. CSP violation reports. CSP stands for [Content Security Policy][]. Send reports to `/csp`, `/csp/`, `/_reports/csp` or basically any URL that contains substring `csp`.
2. HPKP violation reports. HPKP stands for [HTTP Public Key Pinning][]. URL must contain substring `pkp`.
3. StacktraceJS reports. [StacktraceJS][] is a JS library that collects unified stacktrace reports from any browser. URL must contain substring `stacktracejs`.


[Content Security Policy]: http://en.wikipedia.org/wiki/Content_Security_Policy
[HTTP Public Key Pinning]: https://en.wikipedia.org/wiki/HTTP_Public_Key_Pinning
[StacktraceJS]:            https://www.stacktracejs.com
[Gitter]:                  https://gitter.im/frontreport/frontreport
