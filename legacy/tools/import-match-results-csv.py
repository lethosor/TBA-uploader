import argparse
import csv
import itertools
import json
import os

parser = argparse.ArgumentParser()
parser.add_argument('input_file', help='.csv file to read from')
parser.add_argument('-o', '--output-dir', help='fms_data "matches" subfolder to write json results to', required=True)
parser.add_argument('-v', '--verbose', action='store_true')
parser.add_argument('--skip-existing', action='store_true', help='skip matches already written (default: error)')
args = parser.parse_args()

if os.listdir(args.output_dir) and not args.skip_existing:
    raise RuntimeError('output folder not empty and --skip-existing not set')

REQUIRED_HEADERS = [
    'fms_id', 'comp_level', 'set_number', 'match_number',
    'blue 1', 'blue 2', 'blue 3', 'blue score',
    'red 1', 'red 2', 'red 3', 'red score',

]
# todo: share with go
DEFAULT_BREAKDOWN_VALUES_2022 = {
    "adjustPoints":            0,
    "autoCargoLowerBlue":      0,
    "autoCargoLowerFar":       0,
    "autoCargoLowerNear":      0,
    "autoCargoLowerRed":       0,
    "autoCargoPoints":         0,
    "autoCargoTotal":          0,
    "autoCargoUpperBlue":      0,
    "autoCargoUpperFar":       0,
    "autoCargoUpperNear":      0,
    "autoCargoUpperRed":       0,
    "autoPoints":              0,
    "autoTaxiPoints":          0,
    "cargoBonusRankingPoint":  False,
    "endgamePoints":           0,
    "endgameRobot1":           "None",
    "endgameRobot2":           "None",
    "endgameRobot3":           "None",
    "foulCount":               0,
    "foulPoints":              0,
    "hangarBonusRankingPoint": False,
    "matchCargoTotal":         0,
    "quintetAchieved":         False,
    "rp":                      0,
    "taxiRobot1":              "No",
    "taxiRobot2":              "No",
    "taxiRobot3":              "No",
    "techFoulCount":           0,
    "teleopCargoLowerBlue":    0,
    "teleopCargoLowerFar":     0,
    "teleopCargoLowerNear":    0,
    "teleopCargoLowerRed":     0,
    "teleopCargoPoints":       0,
    "teleopCargoTotal":        0,
    "teleopCargoUpperBlue":    0,
    "teleopCargoUpperFar":     0,
    "teleopCargoUpperNear":    0,
    "teleopCargoUpperRed":     0,
    "teleopPoints":            0,
    "totalPoints":             0,
}
BREAKDOWN_TYPES_2022 = {k: type(v) for k, v in DEFAULT_BREAKDOWN_VALUES_2022.items()}

def make_match_result():
    return {
        "comp_level": "",
        "match_number": 0,
        "set_number": 0,
        "alliances": {
            "blue": {
                "dqs": [],
                "score": 0,
                "surrogates": [],
                "teams": [],
            },
            "red": {
                "dqs": [],
                "score": 0,
                "surrogates": [],
                "teams": [],
            },
        },
        "score_breakdown": {
            "blue": {},
            "red": {},
        },
    }

def print_verbose(*print_args, **print_kwargs):
    if args.verbose:
        print(*print_args, **print_kwargs)

def normalize_value(key, value):
    try:
        return int(value)
    except ValueError:
        pass
    return value

def normalize_row(row):
    out = {}
    for key in row.keys():
        # lowercase only first segment before '.'
        key_parts = key.split('.')
        key_parts[0] = key_parts[0].lower()
        key_normalized = '.'.join(key_parts)
        if key_normalized in out:
            raise ValueError('Conflicting headers: %r, %r' % (key, key_normalized))
        out[key_normalized] = normalize_value(key_normalized, row[key])
    return out

def iter_alliance_teams():
    for alliance in ('red', 'blue'):
        for team_id in range(1, 4):
            yield (alliance, team_id, '%s %i' % (alliance, team_id))

def team_to_tba_key(team):
    team_key = str(team)
    if not team_key.startswith('frc'):
        team_key = 'frc' + team_key
    return team_key

def assign_teams(row, match_result):
    for alliance, team_id, field in iter_alliance_teams():
        team = str(row[field]).lower()

        if '*' in team:
            team = team.replace('*', '')
            match_result['alliances'][alliance]['surrogates'].append(team_to_tba_key(team))
        if 'd' in team:
            team = team.replace('d', '')
            match_result['alliances'][alliance]['dqs'].append(team_to_tba_key(team))

        match_result['alliances'][alliance]['teams'].append(team_to_tba_key(team))

def assign_breakdown(row, match_result):
    for key, value in row.items():
        key_parts = key.split('.')
        if key_parts[0] in {'red', 'blue'} and key_parts[1] in BREAKDOWN_TYPES_2022:
            match_result['score_breakdown'][key_parts[0]][key_parts[1]] = BREAKDOWN_TYPES_2022[key_parts[1]](value)


VALID_HANGAR_RESULTS = set(map(sum, itertools.product((0, 4, 6, 10, 15), repeat=3)))

def validate_match_result(match_result):
    RP_REQUIRED_FIELDS = ('rp', 'cargoBonusRankingPoint', 'hangarBonusRankingPoint', 'totalPoints')
    for alliance, other_alliance in itertools.permutations(('red', 'blue'), 2):
        breakdown = match_result['score_breakdown'][alliance]
        if all(field in breakdown for field in RP_REQUIRED_FIELDS):
            expected_rp = 0
            score_diff = breakdown['totalPoints'] - match_result['score_breakdown'][other_alliance]['totalPoints']
            if score_diff > 0:
                expected_rp += 2
            elif score_diff == 0:
                expected_rp += 1
            if breakdown['cargoBonusRankingPoint']:
                expected_rp += 1
            if breakdown['hangarBonusRankingPoint']:
                expected_rp += 1
            if match_result['comp_level'] != 'qm':
                expected_rp = 0  # match FMS
            if breakdown['rp'] != expected_rp:
                raise ValueError('%s: expected rp = %r, got rp = %r' % (alliance, expected_rp, breakdown['rp']))

        if 'endgamePoints' in breakdown:
            if breakdown['endgamePoints'] not in VALID_HANGAR_RESULTS:
                raise ValueError('%s: invalid endgamePoints: %r' % (alliance, breakdown['endgamePoints']))


all_match_results = {}

with open(args.input_file, newline='') as input_file:
    reader = csv.DictReader(input_file)
    for i, row in enumerate(reader):
        row_number = i + 2
        row = normalize_row(row)

        missing_headers = [h for h in REQUIRED_HEADERS if h not in row]
        if missing_headers:
            raise ValueError('row %i missing headers: %r' % (row_number, missing_headers))

        missing_values = [h for h in REQUIRED_HEADERS if row[h] == '']
        if missing_values:
            print_verbose('skipping row %i: missing %r' % (row_number, missing_values))
            continue

        fms_id = row['fms_id']
        if fms_id in all_match_results:
            raise ValueError('duplicate fms_id: %r' % fms_id)

        match_result = make_match_result()

        for field in ('comp_level', 'set_number', 'match_number'):
            match_result[field] = row[field]

        for alliance in ('red', 'blue'):
            score = row['%s score' % alliance]
            match_result['alliances'][alliance]['score'] = score
            match_result['score_breakdown'][alliance]['totalPoints'] = score

        try:
            assign_teams(row, match_result)
            assign_breakdown(row, match_result)

            validate_match_result(match_result)
        except ValueError as e:
            raise ValueError('In row %i: %s' % (row_number, e))

        all_match_results[fms_id] = match_result

for fms_id, match_result in all_match_results.items():
    html_path = os.path.join(args.output_dir, '%s.html' % fms_id)
    if os.path.exists(html_path):
        print_verbose('skipping match %s: %r exists' % (fms_id, html_path))
        continue

    json_path = os.path.join(args.output_dir, '%s.json' % fms_id)
    if os.path.exists(json_path):
        print_verbose('skipping match %s: %r exists' % (fms_id, json_path))
        continue

    with open(html_path, 'w') as f:
        pass  # leave html file empty, only needed for listing

    with open(json_path, 'w') as f:
        json.dump(match_result, f, indent=2)
        print_verbose('wrote %r' % json_path)
