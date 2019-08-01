package tba

import (
    "bytes"
    "crypto/md5"
    "fmt"
    "net/http"
)

type EventParams struct {
    Event string
    Auth string
    Secret string
}

type MatchCode struct {
    Level string `json:"comp_level"`
    Set int `json:"set_number"`
    Match int `json:"match_number"`
}

func GetPlayoffCode(match_id int) MatchCode {
    if (match_id <= 12) {
        return MatchCode{
            Level: "qf",
            Set: ((match_id - 1) % 4) + 1,
            Match: ((match_id - 1) / 4) + 1,
        }
    } else if (match_id <= 18) {
        return MatchCode{
            Level: "sf",
            Set: ((match_id - 1) % 2) + 1,
            Match: ((match_id - 1) / 2) - 5,
        }
    } else {
        return MatchCode{
            Level: "f",
            Set: 1,
            Match: match_id - 18,
        }
    }
}

func SendRequest(url string, body []byte, params *EventParams) (*http.Response, error) {
    url = fmt.Sprintf("/api/trusted/v1/event/%s/%s", params.Event, url)
    sig := fmt.Sprintf("%x", md5.Sum(append([]byte(params.Secret + url), body...)))

    url = "https://www.thebluealliance.com" + url
    request, err := http.NewRequest("POST", url, bytes.NewReader(body))
    if err != nil {
        return nil, err
    }
    request.Header.Add("X-TBA-Auth-Id", params.Auth);
    request.Header.Add("X-TBA-Auth-Sig", sig);
    client := http.Client{}
    return client.Do(request)
}
