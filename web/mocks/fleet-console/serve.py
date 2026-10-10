#!/usr/bin/env python3
"""Static file server for the Vyzorix UI mock (stdlib only)."""
import http.server
import os
import socketserver

PORT = int(os.environ.get("PORT", "12000"))
DIR = os.path.dirname(os.path.abspath(__file__))


class Handler(http.server.SimpleHTTPRequestHandler):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, directory=DIR, **kwargs)

    def end_headers(self):
        self.send_header("Cache-Control", "no-store")
        super().end_headers()

    def log_message(self, fmt, *args):
        print("%s - %s" % (self.address_string(), fmt % args), flush=True)


class ReusableTCPServer(socketserver.ThreadingTCPServer):
    allow_reuse_address = True


if __name__ == "__main__":
    with ReusableTCPServer(("0.0.0.0", PORT), Handler) as httpd:
        print(f"Vyzorix UI mock serving {DIR} on http://0.0.0.0:{PORT}", flush=True)
        httpd.serve_forever()
