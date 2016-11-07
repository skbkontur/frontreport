[![Build Status](https://travis-ci.org/skbkontur/frontreport.svg?branch=master)](https://travis-ci.org/skbkontur/frontreport) [![Go Report Card](https://goreportcard.com/badge/github.com/skbkontur/frontreport)](https://goreportcard.com/report/github.com/skbkontur/frontreport)

- CSP stands for Content Security Policy: http://en.wikipedia.org/wiki/Content_Security_Policy
- HPKP stands for HTTP Public Key Pinning: https://en.wikipedia.org/wiki/HTTP_Public_Key_Pinning

To use CSP and/or HPKP in report mode, you need a tool that gathers violation reports and puts them into a decent storage.

Intended architecture: Nginx -> THIS TOOL -> RabbitMQ -> Elastic RabbitMQ River -> Elastic -> Kibana.
