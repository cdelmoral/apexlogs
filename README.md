<p align="center">
  <img src="images/apexlogs-logo.png" alt="Logo" width="150">

  <h3 align="center">A simple terminal UI for viewing Salesforce logs</h3>
</p>

## What is apexlogs?

Do you find debugging Salesforce Apex code annoying?

Apexlogs is a minimal TUI[^1] application for viewing Salesforce Apex logs from
your terminal.

It automatically creates the required debug level and trace flag records for you
and allows you to easily fetch and view logs.

![Demo](images/demo.gif)

## Installation

Download and install go[^2]. Once installed, open a terminal and run:

```sh
go install github.com/cdelmoral/apexlogs
```

## Usage

If you haven't already install the Salesfoce CLI[^3].

Open a terminal and navigate to your Salesforce project directory.

If a default scratch org is not set already configure it by running
`sf config set target-org my-scratch-org-alias`.

Open the application by running `apexlogs` in your terminal.

[^1]: <https://en.wikipedia.org/wiki/Text-based_user_interface>
[^2]: <https://go.dev/dl/>
[^3]: <https://developer.salesforce.com/tools/salesforcecli>
