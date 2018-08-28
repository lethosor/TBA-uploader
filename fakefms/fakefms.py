import os
import sys

import flask

file_dir = os.path.abspath(sys.argv[1])
if not os.path.isdir(file_dir):
    raise RuntimeError('Folder not found: %r' % file_dir)

app = flask.Flask(__name__)

@app.route('/FieldMonitor/MatchesPartialByLevel')
def match_list():
    level = int(flask.request.args.get('levelParam'))
    return flask.send_from_directory(file_dir, 'list%i.html' % level)

if __name__ == '__main__':
    app.run()
