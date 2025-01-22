# WordPress ICS Generator

## Description

This project is a Go-based application that fetches event data from a specified
WordPress URL (wp-json REST API), processes it, and generates an iCalendar
(ICS) file. The application is designed to run periodically, fetching the
latest event data and updating the ICS file accordingly.

The calendar data is served on the `./events.ics` endpoint.

## Features

- Fetches event data from a specified URL.
- Parses and processes event data.
- Generates an iCalendar (ICS) file.
- Configurable via environment variables.

## Requirements

- Go 1.18 or later
- Environment variables for configuration

## Installation

1. Clone the repository:
    ```bash
    git clone https://github.com/yourusername/yourproject.git
    cd yourproject
    ```
2. Install dependencies:
    ```bash
    go mod tidy
    ```