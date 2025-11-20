# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- JSON response examples in generated markdown documentation
- Support for OpenAPI `example` and `examples` fields in response media types
- Automatic example generation from response schemas using openapi-proto.go
- Validation that response schemas use $ref (rejects inline schemas)
- Three-tier priority system for example selection (explicit → named → generated)
- Golden file test for regression protection
- Comprehensive test suite for response example handling

### Changed
- `renderResponses()` now generates JSON code blocks for responses with content
- `generateMarkdown()` signature includes examples map parameter

### Dependencies
- Added: github.com/duh-rpc/openapi-proto.go v0.5.0 for example generation
