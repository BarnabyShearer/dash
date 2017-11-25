#! /usr/bin/env python3
"""Call Hive V6 API"""

import os
import argparse
from requests import Session
from urllib.parse import urljoin

class PrefixSession(Session):
    """Session with a base url"""
    def __init__(self, prefix_url):
        """Start"""
        self.prefix_url = prefix_url
        super().__init__()

    def request(self, method, url, *args, **kwargs):
        """All http methods"""
        return super().request(method, urljoin(self.prefix_url, url), *args, **kwargs)

def _parser():
    """Parser"""
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('-n','--nodes', nargs='+', help='<Required> List of nodes to toggle', required=True)
    return parser

def main(nodes):
    """CLI"""
    with PrefixSession("https://api.prod.bgchprod.info:443/omnia/") as http:
        http.headers.update({
            "Content-Type": "application/vnd.alertme.zoo-6.1+json",
            "Accept": "application/vnd.alertme.zoo-6.1+json",
            "X-Omnia-Client": "Hive Web Dashboard",
        })
        http.headers.update({
            "X-Omnia-Access-Token": http.post(
                "auth/sessions",
                json={
                    "sessions": [
                        {
                            "username": os.environ["HIVE_USERNAME"],
                            "password": os.environ["HIVE_PASSWORD"],
                            "caller": "WEB"
                        }
                    ]
                }
            ).json()["sessions"][0]["sessionId"]
        })
        state = {}
        for node in http.get("nodes").json()["nodes"]:
            if node["name"] in nodes:
                state[node["id"]] = node["attributes"]["state"]["targetValue"] == "ON"
                print(node["name"], "turning", "off." if state[node["id"]] else "on.")
        for id, current in state.items():
            resp = http.put(
                "nodes/%s" % id,
                json={
                    "nodes": [
                        {
                            "attributes": {
                                "state": {
                                    "targetValue": "OFF" if current else "ON"
                                },
                                "brightness": {
                                    "targetValue": 100
                                }
                            }
                        }
                    ]
                }
            ).json()
            if "errors" in resp:
                print(resp)

if __name__ == "__main__":
    main(**vars(_parser().parse_args()))

