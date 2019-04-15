# tftp-http-proxy

[![Build Status](https://api.travis-ci.org/bwalex/tftp-http-proxy.svg?branch=master)](https://travis-ci.org/bwalex/tftp-http-proxy)

tftp-http-proxy is a simple TFTP server that proxies all read requests to a backing HTTP server, and serves the response.

## Get it

Release binaries for linux amd64 platforms are built by default and can be downloaded from the [Releases page](https://github.com/bwalex/tftp-http-proxy/releases).

For other platforms or to build from source, clone the repository and just run `make`.

## Usage

    Usage of dist/tftp-http-proxy:
      -http-base-url string
                HTTP base URL (default "http://127.0.0.1/tftp")
      -http-append-path
                Append path to url (add slash and then requested path), warning
                no path sanitization is done, so it may contain /../ or other
                wild characters.
      -tftp-timeout duration
                TFTP timeout (default 5s)
      -tftp-bind-address address
                UDP address to bind to

## Details

When tftp-http-proxy is started, it will listen on (UDP) port 69 as a normal TFTP server. Whenever a new TFTP read request is received, the request will be forwarded as an HTTP request to the configured HTTP URL (`-http-base-url` flag). The HTTP request will have some additional HTTP headers containing information about the TFTP request:

 - `X-TFTP-IP`: The IP of the requesting TFTP client
 - `X-TFTP-Port`: The port used by the requesting TFTP client
 - `X-TFTP-File`: The filename requested by the TFTP client

If the HTTP request returns a status 200 response, the contents of the response will be sent as the file contents for the TFTP read request. The HTTP response should contain an accurate ContentLength header, as it will be used to set the TFTP TSize option on the read response.

If the HTTP response status is not 200, an error response will be sent to the TFTP client instead.

## License

tftp-http-proxy is released under the [MIT License](http://www.opensource.org/licenses/MIT).
