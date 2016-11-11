[![Build Status](https://travis-ci.org/skbkontur/frontreport.svg?branch=master)](https://travis-ci.org/skbkontur/frontreport) [![Go Report Card](https://goreportcard.com/badge/github.com/skbkontur/frontreport)](https://goreportcard.com/report/github.com/skbkontur/frontreport) [![Join the chat at https://gitter.im/frontreport/frontreport](https://badges.gitter.im/frontreport/frontreport.svg)](https://gitter.im/frontreport/frontreport?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

## What is this tool for?

Frontreport is useful only if you have an existing infrastructure for backend log aggregation. For example, we use ELK stack with RabbitMQ as a broker on top. So, your logging infrastructure may look like this:

Backend application → Logstash → RabbitMQ → Elastic RabbitMQ River → Elastic → Kibana

You may want to reuse this infrastructure to collect frontend logs from browsers of your visitors. So, you need to replace Logstash in the above scheme with something fast that can validate incoming JSON and manage high load gracefully.

Frontreport does all that. See code for details or ask us on [Gitter][].


## What can you collect from browsers?

1. CSP violation reports. CSP stands for [Content Security Policy][].
2. HPKP violation reports. HPKP stands for [HTTP Public Key Pinning][].
3. StacktraceJS reports. [StacktraceJS][] is a JS library that collects unified stacktrace reports from any browser.


[Content Security Policy]: http://en.wikipedia.org/wiki/Content_Security_Policy
[HTTP Public Key Pinning]: https://en.wikipedia.org/wiki/HTTP_Public_Key_Pinning
[StacktraceJS]:            https://www.stacktracejs.com
[Gitter]:                  https://gitter.im/frontreport/frontreport
