import re
import os
import sys

import flask

file_dir = os.path.abspath(sys.argv[1])
if not os.path.isdir(file_dir):
    raise RuntimeError('Folder not found: %r' % file_dir)

def build_level_index(level):
    index = []
    lpath = os.path.join(file_dir, 'list%i.html' % level)
    with open(lpath) as f:
        list_html = f.read()
    link_matches = re.findall(r'<a.*?href=([\'"]).*?matchId=(.*?)\1', list_html)
    for match in link_matches:
        index.append(match[1])
    return index

# [match level] -> [list of match IDs]
level_index = {}
for i in range(1, 3 + 1):
    if os.path.isfile(os.path.join(file_dir, 'list%i.html' % i)):
        level_index[i] = build_level_index(i)

app = flask.Flask(__name__)

@app.route('/FieldMonitor/MatchesPartialByLevel')
def match_list():
    level = int(flask.request.args.get('levelParam'))
    return flask.send_from_directory(file_dir, 'list%i.html' % level)

@app.route('/FieldMonitor/Matches/Score')
def match_score():
    match_id = flask.request.args.get('matchId')
    for level, index in level_index.items():
        if match_id in index:
            return flask.send_from_directory(file_dir, 'level%i/raw%i.html' % (level, index.index(match_id)))

if __name__ == '__main__':
    app.run(port=5555, host='0.0.0.0')
