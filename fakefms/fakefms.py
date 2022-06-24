import argparse
import json
import re
import os
import sys
import uuid

import flask

parser = argparse.ArgumentParser(add_help=False)
parser.add_argument('folder', help='Root event folder to search in fms_data')
parser.add_argument('-?', '--help', action='help')
parser.add_argument('-h', '--host', help='Host to listen on', default='127.0.0.1')
parser.add_argument('-p', '--port', help='Port to listen on', type=int, default=5555)
args = parser.parse_args()

args.folder = os.path.abspath(args.folder)
if not os.path.isdir(args.folder):
    raise RuntimeError('Folder not found: %r' % args.folder)

def match_list_filename(level):
    return 'level%i/match_list.html' % level

def match_filename(level, match_name):
    return 'level%i/matches/%s.html' % (level, match_name)

def build_level_index(level):
    index = {}
    lpath = os.path.join(args.folder, match_list_filename(level))
    if not os.path.exists(lpath):
        raise IOError('level %i not found in %r' % (level, lpath))
    with open(lpath) as f:
        list_html = f.read()
    link_matches = re.findall(r'<a.*?href=([\'"]).*?matchId=(.*?)\1', list_html)
    button_matches = re.findall(r'<button.*?btn-success.*?<b>(.*?)</b>', list_html)
    button_matches = [s.replace(' ', '').replace('/', '-') for s in button_matches]
    if len(link_matches) != len(button_matches):
        raise ValueError('')
    for i in range(len((link_matches))):
        match_uuid = link_matches[i][1]
        button = button_matches[i]
        print(match_uuid, button)
        if match_uuid in index:
            raise ValueError('Duplicate match %r' % match_uuid)
        index[match_uuid] = button
    return index

# [match level] -> [{UUID: name}, ...]
level_index = {}
for i in range(0, 3 + 1):
    try:
        level_index[i] = build_level_index(i)
    except IOError as e:
        print(e)

def validate_dict(desc, actual, expected_values):
    def abort(message):
        flask.abort(flask.make_response('%s: %s' % (desc, message), 400))
    for k, v in expected_values.items():
        if k not in actual:
            abort('missing key: %s' % k)
        if actual[k] != expected_values[k]:
            abort('invalid %s: expected %r, got %r' % (k, expected_values[k], actual[k]))

app = flask.Flask(__name__)

@app.route('/FieldMonitor/MatchesPartialByLevel')
def match_list():
    level = int(flask.request.args.get('levelParam'))
    return flask.send_from_directory(args.folder, match_list_filename(level))

@app.route('/FieldMonitor/Matches/Score')
def match_score():
    match_uuid = flask.request.args.get('matchId')
    for level, index in level_index.items():
        if match_uuid in index:
            return flask.send_from_directory(args.folder, match_filename(level, index[match_uuid]))
    return 'invalid match ID', 404

@app.route('/Pit/GetData')
def rankings():
    if not flask.request.headers.get('Referer', '').endswith('/Pit/Qual'):
        return flask.jsonify(None)
    return flask.send_from_directory(args.folder, 'rankings.json')

report_uuids = set()

@app.route('/Reports/PostReportAction', methods=['POST'])
def get_report():
    body = flask.request.get_json()
    report_type = json.loads(body['CustomData'])[0]['reportType']
    validate_dict('headers', flask.request.headers, {
        'Content-Type': 'application/json; charset=UTF-8',
        'Origin': 'http://10.0.100.5',
        'Referer': 'http://10.0.100.5/Reports/' + report_type,
    })
    validate_dict('body', body, {
        'controlId': 'sfreportviewer',
        'reportPath': '',
        'reportServerUrl': '',
        'processingMode': 'local',
        'locale': 'en-US',
    })
    if body['reportAction'] == 'ReportLoad':
        new_uuid = str(uuid.uuid4())
        report_uuids.add(new_uuid)
        return flask.jsonify({
            'reportViewerID': new_uuid,
            'reportViewerToken': new_uuid,
        })
    elif body['reportAction'] == 'GetPageModel':
        validate_dict('body', body, {
            'dataRefresh': True,
            'dataSources': None,
            'isPrint': True,
            'pageInit': True,
            'parameters': None,
            'refresh': False,
        })
        page_id = int(body['pageindex'])
        if page_id < 1:
            return 'invalid page: %i' % page_id, 400
        if body['reportViewerClientId'] != body['reportViewerToken']:
            return 'tokens do not match', 400
        if body['reportViewerToken'] not in report_uuids:
            return 'invalid token', 400
        report_folder = os.path.join(args.folder, 'reports', 'json')
        report_filename = '%s.%i.json' % (report_type, page_id)
        print('sending %s/%s' % (report_folder, report_filename))
        return flask.send_from_directory(report_folder, report_filename)
    else:
        return 'invalid reportAction: %r' % body['reportAction'], 400

@app.route('/api/index/')
def api_index():
    return flask.jsonify(level_index)

if __name__ == '__main__':
    app.run(port=args.port, host=args.host, debug=True)
