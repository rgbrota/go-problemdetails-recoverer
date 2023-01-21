# 🛡️ go-problemdetails-recoverer

[![License](http://img.shields.io/badge/license-MIT-brightgreen.svg)](http://opensource.org/licenses/MIT)
[![Build status](https://github.com/rgbrota/go-problemdetails-recoverer/actions/workflows/ci.yml/badge.svg)](https://github.com/rgbrota/go-problemdetails-recoverer/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/rgbrota/go-problemdetails-recoverer)](https://goreportcard.com/report/github.com/rgbrota/go-problemdetails-recoverer)

## About

Problem details specification [RFC-7807] recoverer middleware library written in Go.

The objective of this middleware is to recover from panics gracefully and return a 500 Internal Server Error instead. The error response follows the problem details specification and looks like the following structure:

```json
{
  "type": "https://www.rfc-editor.org/rfc/rfc7231#section-6.6.1",
  "title": "Internal Server Error",
  "status": 500
}
```

For more information on how to be compliant with the specification, please see [RFC-7807](https://www.rfc-editor.org/rfc/rfc7807).

## Installation

```go get github.com/rgbrota/go-problemdetails-recoverer```

## Getting started

This repository contains a recoverer middleware compatible with ```net/http``` which recovers from panics and sends a problem details 500 Internal Server Error response instead. 

In order to use it you only need to register it in the middleware chain, either by using the default configuration or by creating your own.


