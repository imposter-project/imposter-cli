# SOAP 1.1/1.2 Proxy Support

The imposter-cli proxy command supports both SOAP 1.1 and SOAP 1.2 services with the `--soap1.1` flag, enabling action-aware recording and mock generation.

## Usage

```bash
imposter proxy --soap1.1 [URL] [flags]
```

### Flags

- `--soap1.1` - Enable SOAP 1.1/1.2 aware mode for capturing requests/responses with action-based differentiation
- `--insecure` - Skip TLS certificate verification for HTTPS upstream servers (useful for self-signed certificates)

### Example

```bash
imposter proxy --soap1.1 http://soap-service.example.com/service
```

For HTTPS services with self-signed certificates:

```bash
imposter proxy --soap1.1 --insecure https://soap-service.example.com/service
```

## Features

When `--soap1.1` is enabled:

1. **SOAP Action Detection** - Supports both SOAP 1.1 and SOAP 1.2 action specifications:
   - **SOAP 1.1**: SOAPAction header (`SOAPAction: "http://example.com/GetUser"`)
   - **SOAP 1.2**: Content-Type action parameter (`Content-Type: application/soap+xml;action="http://example.com/GetUser"`)

2. **Operation-Specific File Naming** - Files include action in names for same-endpoint differentiation

3. **Automatic Configuration** - Generated configs include proper header matching based on SOAP version

## Action Detection

### SOAP 1.1 Style (SOAPAction Header)
```http
POST /service HTTP/1.1
Content-Type: text/xml; charset=utf-8
SOAPAction: "http://example.com/GetUser"
```

### SOAP 1.2 Style (Content-Type Action Parameter)
```http
POST /service HTTP/1.1
Content-Type: application/soap+xml;charset=UTF-8;action="http://example.com/GetUser"
```

## File Naming

- Standard: `POST-endpoint.xml`
- SOAP 1.1/1.2: `POST-endpoint_http___example_com_GetUser.xml` (for action "http://example.com/GetUser")

## Generated Configuration

### SOAP 1.1 Configuration
```yaml
plugin: rest
path: /service
resources:
  - method: POST
    requestHeaders:
      SOAPAction: "http://example.com/GetUser"
    response:
      file: POST-endpoint_http___example_com_GetUser.xml
```

### SOAP 1.2 Configuration
```yaml
plugin: rest
path: /service
resources:
  - method: POST
    requestHeaders:
      Content-Type: "application/soap+xml;charset=UTF-8;action=\"http://example.com/GetUser\""
    response:
      file: POST-endpoint_http___example_com_GetUser.xml
```

## Docker Usage

Run the proxy in a Docker container:

```bash
# Basic usage
docker run -d --name imposter-soap-proxy -p 8080:8080 -v $PWD:/output \
  nexus.bcn.crealogix.net:18080/imposter-cli-soap11:latest \
  proxy --soap1.1 --capture-request-body --capture-request-headers --output-dir /output https://soap-service.example.com/service

# For HTTPS with self-signed certificates
docker run -d --name imposter-soap-proxy -p 8080:8080 -v $PWD:/output \
  nexus.bcn.crealogix.net:18080/imposter-cli-soap11:latest \
  proxy --soap1.1 --insecure --capture-request-body --capture-request-headers --output-dir /output https://soap-service.example.com/service
```

### Docker Container Notes

The `nexus.bcn.crealogix.net:18080/imposter-cli-soap11:latest` image includes:

- **Full engine support**: Supports `-t docker`, `-t jvm`, and `-t golang` engine types
- **Java Runtime**: OpenJDK 17 for JVM engine operations
- **Docker Client**: For Docker engine operations (requires socket mapping)
- **SOAP 1.1/1.2**: Enhanced proxy functionality for SOAP services

**Engine Usage Examples:**

```bash
# Using JVM engine (recommended)
docker run --rm -v $PWD:/config -p 8080:8080 \
  nexus.bcn.crealogix.net:18080/imposter-cli-soap11:latest \
  up -t jvm /config

# Using Docker engine (requires socket mapping and privileges)
docker run --rm --privileged -v /var/run/docker.sock:/var/run/docker.sock \
  -v $PWD:/config -p 8080:8080 \
  nexus.bcn.crealogix.net:18080/imposter-cli-soap11:latest \
  up -t docker /config
```

## Compatibility

- **SOAP 1.1**: Full support via SOAPAction header
- **SOAP 1.2**: Full support via Content-Type action parameter  
- **Mixed environments**: Automatically detects and handles both formats
- **Backward compatibility**: Maintains full compatibility with existing proxy functionality
- **Enterprise HTTPS**: Supports self-signed certificates with `--insecure` flag
