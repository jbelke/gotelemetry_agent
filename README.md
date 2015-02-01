# Telemetry Agent

The Telemetry Agent simplifies the process of creating daemon processes that feed data into one or more [Telemetry](http://telemetryapp.com) flows.

Typical use-case scenarios include:

  - Feeding data from existing infrastructure (e.g.: a MySQL database, Excel sheet, custom script written in your language of choice) to one or more Telemetry data flows
  - Automatically creating boards for your customers
  - Interfacing third-party APIs with Telemetry

The Agent is written in Go and runs fine on most Linux distros, OS X, and Windows. It is designed to run on your infrastructure, and its only requirement is that it be able to reach the Telemetry API endpoint (https://api.telemetryapp.com) on port 443 via HTTPS. It can therefore happily live behind firewalls without posing a security risk.

Full documentation is available on the [Telemetry Documentation website](http://telemetry.readme.io/v1.0/docs/agents).
