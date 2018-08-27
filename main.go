package main

import (
    "flag"
)

func main() {
    var port int
    flag.IntVar(&port, "port", 8808, "web server port")
    flag.Parse()
    RunWebServer(port);
}
