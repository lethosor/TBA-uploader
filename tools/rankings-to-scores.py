import argparse
import contextlib
import csv
import os
import sys

import numpy as np
import scipy

parser = argparse.ArgumentParser()
parser.add_argument('-s', '--schedule', help='CSV schedule file', required=True)
parser.add_argument('-r', '--rankings', help='CSV rankings file', required=True)
parser.add_argument('-o', '--output-results', help='CSV results file (output)', required=False)
parser.add_argument('--output-stats', help='CSV results file for regression statistics (output)', required=False)
parser.add_argument('-y', '--year', type=int, required=True)
parser.add_argument('-v', '--verbose', action='store_true')
args = parser.parse_args()

REQUIRED_HEADERS_SCHEDULE = {'blue 1', 'blue 2', 'blue 3', 'red 1', 'red 2', 'red 3', 'id'}

REQUIRED_HEADERS_RANKINGS = {
    2017: {'rank', 'team', 'totalpts', 'auto', 'rotor', 'takeoff', 'kpa'},
}

def print_verbose(*print_args, **print_kwargs):
    if args.verbose:
        print(*print_args, **print_kwargs)

def normalize_row(row):
    out = {}
    for k, v in row.items():
        k = k.lower()
        try:
            out[k] = int(v)
        except ValueError:
            out[k] = v
    return out

@contextlib.contextmanager
def csv_row_context(i, row, description):
    try:
        yield (i, normalize_row(row))
    except Exception as e:
        raise type(e)('%s: line %i: %s' % (description, i + 2, e))

def iter_alliance_teams():
    for alliance in ('red', 'blue'):
        for team_id in range(1, 4):
            yield (alliance, team_id, '%s %i' % (alliance, team_id))

def iter_csv_file(filename, description, required_headers):
    with open(filename, newline='') as csv_file:
        reader = csv.DictReader(csv_file)
        missing_headers = [h for h in required_headers if h.lower() not in map(str.lower, reader.fieldnames)]
        if missing_headers:
            raise ValueError('%s missing headers: %r' % (description, missing_headers))

        for i, row in enumerate(reader):
            with csv_row_context(i, row, description) as (line_number, row):
                missing_values = [h for h in required_headers if row[h] == '']
                if missing_values:
                    print_verbose('%s: skipping line %i: missing %r' % (description, line_number, missing_values))
                    continue

                yield csv_row_context(i, row, description)

rankings_by_team = {}

for context in iter_csv_file(args.rankings, 'rankings', REQUIRED_HEADERS_RANKINGS[args.year]):
    with context as (_, row):
        rankings_by_team[row['team']] = row

print_verbose('number of ranked teams:', len(rankings_by_team))

match_ids = []
alliance_ids = set()
alliances_by_team = {}

for context in iter_csv_file(args.schedule, 'schedule', REQUIRED_HEADERS_SCHEDULE):
    with context as (_, row):
        match_id = row['id']
        if match_id in match_ids:
            raise ValueError('Duplicate match ID: %r' % match_id)
        match_ids.append(match_id)

        for alliance, _, field in iter_alliance_teams():
            team = row[field]
            alliance_id = f'{match_id}{alliance}'
            alliance_ids.add(alliance_id)
            alliances_by_team.setdefault(team, set()).add(alliance_id)

print_verbose('number of match alliances:', len(alliance_ids))

# for consistent order
team_index = list(sorted(rankings_by_team.keys()))
alliance_ids_index = list(sorted(alliance_ids))
mat_A = np.zeros((len(team_index), len(alliance_ids)))

for i, team in enumerate(team_index):
    for alliance_key in alliances_by_team[team]:
        mat_A[i, alliance_ids_index.index(alliance_key)] = 1

match_results_by_alliance = {}
results_regression_stats_by_field = {}

for ranking_field in sorted(REQUIRED_HEADERS_RANKINGS[args.year]):
    if ranking_field in {'rank', 'team'}:
        continue

    mat_B = np.zeros((len(team_index), 1))

    for team, ranking in rankings_by_team.items():
        mat_B[team_index.index(team), 0] = ranking[ranking_field]

    print_verbose('processing:', ranking_field)

    regression_results = scipy.sparse.linalg.lsmr(mat_A, mat_B)
    results_regression_stats_by_field[ranking_field] = {
        k: regression_results[i + 1]
        for i, k in enumerate(('stop_reason', 'iters', 'norm_r', 'norm_ar', 'norm_a', 'cond_a', 'norm_x'))
    }
    regression_solution = regression_results[0]

    for i, alliance_key in enumerate(alliance_ids_index):
        match_results_by_alliance.setdefault(alliance_key, {})[ranking_field] = regression_solution[i]

def write_results(out_file, match_results_by_alliance, match_ids):
    for value in match_results_by_alliance.values():  # can't use next()
        alliance_ranking_field_names = list(sorted(value.keys()))
        break

    ranking_field_names = ['match']
    for alliance in ('red', 'blue'):
        ranking_field_names += [f'{alliance}.{field}' for field in alliance_ranking_field_names]

    writer = csv.DictWriter(out_file, fieldnames=ranking_field_names)
    writer.writeheader()
    for match_id in match_ids:
        row = {'match': match_id}
        for alliance in ('red', 'blue'):
            row.update({f'{alliance}.{field}': value for field, value in match_results_by_alliance[f'{match_id}{alliance}'].items()})
        writer.writerow(row)

def write_stats(out_file, stats):
    for value in stats.values():
        field_names = list(sorted(value.keys()))
        break

    field_names = ['field'] + field_names

    writer = csv.DictWriter(out_file, fieldnames=field_names)
    writer.writeheader()
    for field, field_stats in stats.items():
        row = {'field': field}
        row.update(field_stats)
        writer.writerow(row)

if args.output_results:
    results_file = open(args.output_results, 'w', newline='')
else:
    results_file = contextlib.nullcontext(sys.stdout)
with results_file as f:
    write_results(f, match_results_by_alliance, match_ids)

if args.output_stats:
    with open(args.output_stats, 'w', newline='') as f:
        write_stats(f, results_regression_stats_by_field)
