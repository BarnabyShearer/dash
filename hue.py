#! /usr/bin/env python3
"""Call Hive V6 API"""

import os
import socket
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
    parser.add_argument('-r','--register', help="Register with the hub (you must push the hub's button less then 30s before running this", action="store_true")
    return parser

def main(nodes, register):
    """CLI"""
    hue = Session().get("https://www.meethue.com/api/nupnp").json()[0]["internalipaddress"] #SSPD just dosn't work for me
    if register:
        print(http.post("http://%s/api/" % hue, json={"devicetype": "dash#%s" % socket.gethostname()}).text)
        return
    with PrefixSession("http://%s/api/%s/" % (hue, os.environ["HUE_USERNAME"])) as http:
        state = {}
        for id, node in http.get("lights").json().items():
            if node["name"] in nodes:
                state[id] = node["state"]["on"]
                print(node["name"], "turning", "off." if state[id] else "on.")
        for id, current in state.items():
            resp = http.put(
                "lights/%s/state" % id,
                json={
                    "on": not current,
                    "bri": 254
                }
            ).json()
            if "error" in resp[0]:
                print(resp)


if __name__ == "__main__":
    main(**vars(_parser().parse_args()))


