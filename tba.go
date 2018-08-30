package main

import (
    "bytes"
    "crypto/md5"
    "fmt"
    "net/http"
)

type eventParams struct {
    event string
    auth string
    secret string
}

var playoff_codes = map[int]string {
    1: "qf1m1",
    2: "qf2m1",
    3: "qf3m1",
    4: "qf4m1",
    5: "qf1m2",
    6: "qf2m2",
    7: "qf3m2",
    8: "qf4m2",
    9: "qf1m3",
    10: "qf2m3",
    11: "qf3m3",
    12: "qf4m3",

    13: "sf1m1",
    14: "sf2m1",
    15: "sf1m2",
    16: "sf2m2",
    17: "sf1m3",
    18: "sf2m3",

    19: "f1m1",
    20: "f1m2",
    21: "f1m3",
}

func sendTBARequest(url string, body []byte, params *eventParams) (*http.Response, error) {
    url = fmt.Sprintf("/api/trusted/v1/event/%s/%s", params.event, url)
    sig := fmt.Sprintf("%x", md5.Sum(append([]byte(params.secret + url), body...)))

    url = "https://www.thebluealliance.com" + url
    request, err := http.NewRequest("POST", url, bytes.NewReader(body))
    if err != nil {
        return nil, err
    }
    request.Header.Add("X-TBA-Auth-Id", params.auth);
    request.Header.Add("X-TBA-Auth-Sig", sig);
    client := http.Client{}
    return client.Do(request)
}
