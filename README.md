CSP stands for Content Security Policy: http://en.wikipedia.org/wiki/Content_Security_Policy

To use CSP in report-only mode, you need a tool that gathers CSP violation reports and puts them into a decent storage.

Intended architecture: Nginx -> THIS TOOL -> RabbitMQ -> Elastic RabbitMQ River -> Elastic -> Kibana.
