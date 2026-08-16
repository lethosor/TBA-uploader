#!/usr/bin/env python3

import argparse
import json
import os
import time
import pathlib

import requests

parser = argparse.ArgumentParser()
parser.add_argument('event')
parser.add_argument('level', type=int, choices=(0, 1, 2, 3))
args = parser.parse_args()

FMS_URL = os.environ.get('FMS_URL', 'http://10.0.100.5')

outdir = pathlib.Path(f'{args.event}-level{args.level}-' + time.strftime("%Y%m%d-%H%M%S"))
if outdir.exists():
    raise RuntimeError("output exists: " + outdir)
os.mkdir(outdir)

matches_json = requests.get(f'{FMS_URL}/fieldMonitor/MatchesPartialByLevel', {'levelParam': args.level}).json()
with open(outdir / 'matches.json', 'w') as f:
    json.dump(matches_json, f)

for match in matches_json['matches']:
    print(match['description'], match['fmsMatchId'])
    match_json = requests.get(f'{FMS_URL}/GameSpecific/api/v1.0/fieldmonitor_gs/GetScoreDetails?matchId={match["fmsMatchId"]}').json()
    with open(outdir / f'{match["matchNumber"]}-{match["playNumber"]}-{match["fmsMatchId"]}', 'w') as f:
        json.dump(match_json, f)

print(f"wrote to: {outdir}")
