if [ "$(uname)" = "Darwin" ]; then
    # need to use system compiler
    unset CC
    unset CXX
fi

run_cmd() {
    echo "=> $@"
    "$@"
}

set -e
