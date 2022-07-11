package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/lethosor/TBA-uploader/fms_parser"
	"github.com/lethosor/TBA-uploader/tba"
)

// @todo hack
var keys_json []byte

type APIError struct {
	code    int
	message string
}

func apiPanicCode(code int, message string, args ...interface{}) {
	logger.Printf("API Error %d: "+message, append([]interface{}{code}, args...)...)
	panic(APIError{
		code:    code,
		message: fmt.Sprintf(message, args...),
	})
}

func apiPanicBadRequest(message string, args ...interface{}) {
	apiPanicCode(http.StatusBadRequest, message, args...)
}

func apiPanicInternal(message string, args ...interface{}) {
	apiPanicCode(http.StatusInternalServerError, message, args...)
}

func getRequestEventParams(r *http.Request) (*tba.EventParams, bool) {
	if len(r.Header.Get("X-Event")) > 0 && len(r.Header.Get("X-Auth")) > 0 && len(r.Header.Get("X-Secret")) > 0 {
		return &tba.EventParams{
			Event:  r.Header.Get("X-Event"),
			Auth:   r.Header.Get("X-Auth"),
			Secret: r.Header.Get("X-Secret"),
		}, true
	}
	return nil, false
}

func checkRequestEventParams(r *http.Request) *tba.EventParams {
	params, ok := getRequestEventParams(r)
	if params == nil || !ok {
		apiPanicBadRequest("missing event/auth parameters")
	}
	return params
}

func getRequestLevel(r *http.Request) (int, error) {
	level, err := strconv.Atoi(r.URL.Query().Get("level"))
	if err != nil {
		return -1, err
	} else if level < 0 || level > 3 {
		return -1, fmt.Errorf("Invalid level: %d", level)
	} else {
		return level, nil
	}
}

func checkRequestLevel(r *http.Request) int {
	level, err := getRequestLevel(r)
	if err != nil {
		apiPanicBadRequest("bad level parameter: %v", err)
	}
	return level
}

func checkRequestQueryParam(r *http.Request, param string) string {
	res := r.URL.Query().Get(param)
	if res == "" {
		apiPanicBadRequest("missing parameter: %s", param)
	}
	return res
}

func parseEventYear(event string) int {
	var year int
	fmt.Sscanf(event, "%d", &year)
	return year
}

func apiTBARequest(path string, w http.ResponseWriter, r *http.Request) {
	params := checkRequestEventParams(r)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		apiPanicInternal("read failed: %s", err)
	}

	res, err := tba.SendRequest(FMSConfig.TbaUrl, path, body, params)
	if err != nil {
		apiPanicInternal("TBA request failed: %s", err)
	}

	if res.StatusCode != http.StatusOK {
		res_body, _ := ioutil.ReadAll(res.Body)
		apiPanicInternal("TBA error %d: %s", res.StatusCode, res_body)
	}

	w.Write([]byte("ok"))
}

func marshalFMSConfig(w http.ResponseWriter) ([]byte, error) {
	out, err := json.Marshal(FMSConfig)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Printf("JSON.Marshal(FMSConfig): %s\n", err)
		return nil, err
	}
	return out, nil
}

func sendJson(w http.ResponseWriter, val any) {
	out, err := json.Marshal(val)
	if err != nil {
		apiPanicInternal("json encode failed: %s", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func jsVersion(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf("window.VERSION=\"%s\";", Version)))
}

func jsFMSConfig(w http.ResponseWriter, r *http.Request) {
	out, err := marshalFMSConfig(w)
	if err == nil {
		w.Write([]byte("window.FMS_CONFIG="))
		w.Write(out)
		w.Write([]byte(";"))
	}
}

func jsBrackets(w http.ResponseWriter, r *http.Request) {
	brackets := make(map[int]tba.Bracket)
	// todo: figure out proper bounds
	for i := 0; i < 100; i++ {
		bracket := tba.GetBracket(i)
		if bracket != nil {
			brackets[i] = bracket
		}
	}

	out, err := json.Marshal(brackets)
	if err != nil {
		apiPanicInternal("%s", err)
	}
	w.Write([]byte("window.BRACKETS=Object.freeze("))
	w.Write(out)
	w.Write([]byte(");"))
}

func apiGetFMSConfig(w http.ResponseWriter, r *http.Request) {
	out, err := marshalFMSConfig(w)
	if err == nil {
		w.Write(out)
	}
}

func apiSetFMSConfig(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &FMSConfig)
	resp := make(map[string]interface{})
	resp["ok"] = (err == nil)
	if err != nil {
		resp["error"] = err.Error()
	}
	resp["config"] = FMSConfig
	out, err := json.Marshal(resp)
	logger.Printf("Changed FMS config: %s\n", out)
	w.Write(out)
	if err != nil {
		logger.Printf("apiSetFMSConfig: Marshal failed: %s\n", err)
	}
}

func apiKeysFetch(w http.ResponseWriter, r *http.Request) {
	w.Write(keys_json)
}

func apiKeysUpdate(w http.ResponseWriter, r *http.Request) {
	var err error
	keys_json, err = ioutil.ReadAll(r.Body)
	if err != nil {
		apiPanicBadRequest("failed to read body: %v", err)
	}
}

func apiUploadEventInfo(w http.ResponseWriter, r *http.Request) {
	apiTBARequest("info/update", w, r)
}

func apiUploadTeams(w http.ResponseWriter, r *http.Request) {
	apiTBARequest("team_list/update", w, r)
}

func apiUploadAwards(w http.ResponseWriter, r *http.Request) {
	apiTBARequest("awards/update", w, r)
}

func apiUploadMatches(w http.ResponseWriter, r *http.Request) {
	apiTBARequest("matches/update", w, r)
}

func apiFetchMatches(w http.ResponseWriter, r *http.Request) {
	download_all := (r.URL.Query().Get("all") != "")
	level := checkRequestLevel(r)
	var event_year = parseEventYear(r.URL.Query().Get("event"))
	var match_folder = getMatchDownloadPath(level, r.URL.Query().Get("event"))
	var files []string
	var err error
	if download_all {
		files, err = downloadAllMatches(level, r.URL.Query().Get("event"))
	} else {
		files, err = downloadNewMatches(level, r.URL.Query().Get("event"))
	}
	if err != nil {
		apiPanicInternal("match downloaded failed: %s", err)
	}

	if files != nil {
		for i := 0; i < len(files); i++ {
			logger.Printf("Downloaded %s\n", files[i])
			fname := filepath.Base(files[i])
			match_number, err := strconv.Atoi(strings.Split(fname, "-")[0])
			if err != nil {
				apiPanicInternal("%s: failed to parse match ID", fname)
			}
			folder := filepath.Dir(files[i])

			match_extra_path := path.Join(folder, replaceExtension(fname, "extrajson"))
			var extra_info fms_parser.ExtraMatchInfo
			if isFile(match_extra_path) {
				raw, _ := ioutil.ReadFile(match_extra_path)
				err := json.Unmarshal(raw, &extra_info)
				if err != nil {
					apiPanicInternal("failed to parse %s: %v", match_extra_path, err)
				}
			}

			is_playoff := (level == MATCH_LEVEL_PLAYOFF)
			if extra_info.MatchCodeOverride != nil {
				is_playoff = (extra_info.MatchCodeOverride.Level != "qm")
			}

			match_info, err := fms_parser.ParseHTMLtoJSON(event_year, files[i], is_playoff)
			if err != nil {
				apiPanicInternal("failed to parse %s: %s", fname, err)
			}

			if extra_info.MatchCodeOverride != nil {
				match_info["comp_level"] = extra_info.MatchCodeOverride.Level
				match_info["set_number"] = extra_info.MatchCodeOverride.Set
				match_info["match_number"] = extra_info.MatchCodeOverride.Match
			} else if level == MATCH_LEVEL_PLAYOFF {
				// playoffs
				code := tba.GetPlayoffCode(BRACKET_TYPE_BRACKET_8_TEAM, match_number)
				match_info["comp_level"] = code.Level
				match_info["set_number"] = code.Set
				match_info["match_number"] = code.Match
			} else {
				match_info["comp_level"] = "qm"
				match_info["set_number"] = 1
				match_info["match_number"] = match_number
			}

			match_json, err := json.Marshal(match_info)
			if err != nil {
				apiPanicInternal("%s: JSON serialization failed %s", fname, err)
			}

			fname_json := replaceExtension(fname, "json")
			ioutil.WriteFile(path.Join(folder, fname_json), match_json, os.ModePerm)

			// remove any receipts for newly-downloaded files
			fname_receipt := replaceExtension(fname, "receipt")
			os.Remove(path.Join(folder, fname_receipt))
		}
	}

	match_json_list := make([]map[string]interface{}, 0)
	json_files, err := listFilesWithExtension(match_folder, "json")
	if err != nil {
		apiPanicInternal("download folder %s scan failed: %s", match_folder, err)
	}

	for _, json_file := range json_files {
		json_path := path.Join(match_folder, json_file.Name())
		receipt_path := replaceExtension(json_path, "receipt")
		if fileExists(receipt_path) {
			// receipt exists, match was already uploaded to TBA
			continue
		}

		match_info := make(map[string]interface{})
		contents, err := ioutil.ReadFile(json_path)
		err = json.Unmarshal(contents, &match_info)
		if err != nil {
			apiPanicInternal("failed to parse %s: %s", json_path, err)
		}

		match_info["_fms_id"] = strings.Split(json_file.Name(), ".")[0]
		match_json_list = append(match_json_list, match_info)
	}

	output, err := json.Marshal(match_json_list)
	if err != nil {
		apiPanicInternal("json encode failed: %s", err)
	}

	w.Write(output)
}

func apiMarkMatchesUploaded(w http.ResponseWriter, r *http.Request) {
	params := checkRequestEventParams(r)
	level := checkRequestLevel(r)
	var match_folder = getMatchDownloadPath(level, params.Event)
	match_ids := make([]string, 0)
	body, err := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &match_ids)
	if err != nil {
		apiPanicBadRequest("failed to parse match ID list: %s", err)
	}
	for _, match_id := range match_ids {
		ioutil.WriteFile(path.Join(match_folder, match_id+".receipt"), []byte(match_id), os.ModePerm)
	}
}

func apiPurgeMatches(w http.ResponseWriter, r *http.Request) {
	params := checkRequestEventParams(r)
	level := checkRequestLevel(r)
	match_folder := getMatchDownloadPath(level, params.Event)
	all := (r.URL.Query().Get("all") != "")
	match_ids := make(map[string]bool)
	if !all {
		match_id_list := make([]string, 0)
		body, err := ioutil.ReadAll(r.Body)
		err = json.Unmarshal(body, &match_id_list)
		if err != nil {
			apiPanicBadRequest("failed to parse match ID list: %s", err)
		}
		for _, mid := range match_id_list {
			match_ids[mid] = true
		}
	}

	match_files, err := ioutil.ReadDir(match_folder)
	if err != nil {
		apiPanicInternal("download folder %s scan failed: %s", match_folder, err)
	}
	for _, file := range match_files {
		if _, in_match_ids := match_ids[strings.Split(file.Name(), ".")[0]]; in_match_ids || all {
			ext := filepath.Ext(file.Name())
			if ext == ".html" || ext == ".json" || ext == ".receipt" {
				err := os.Remove(path.Join(match_folder, file.Name()))
				if err != nil {
					logger.Printf("purge: failed to delete %s: %v\n", file.Name(), err)
				}
			}
		}
	}
}

func apiMatchLoadExtra(w http.ResponseWriter, r *http.Request) {
	params := checkRequestEventParams(r)
	event_year := parseEventYear(params.Event)
	level := checkRequestLevel(r)
	id := checkRequestQueryParam(r, "id")

	extra_filename := path.Join(getMatchDownloadPath(level, params.Event), id+".extrajson")
	extra_json, err := ioutil.ReadFile(extra_filename)
	if err != nil {
		tmp, err := fms_parser.MakeExtraMatchInfo(event_year)
		if err != nil {
			apiPanicInternal("MakeExtraMatchInfo: %v", err)
		}
		extra_json, _ = json.Marshal(tmp)
	}
	w.Write(extra_json)
}

func apiMatchSaveExtra(w http.ResponseWriter, r *http.Request) {
	params := checkRequestEventParams(r)
	level := checkRequestLevel(r)
	id := checkRequestQueryParam(r, "id")

	extra_filename := path.Join(getMatchDownloadPath(level, params.Event), id+".extrajson")
	var tmp fms_parser.ExtraMatchInfo
	body, _ := ioutil.ReadAll(r.Body)
	if json.Unmarshal(body, &tmp) != nil {
		apiPanicBadRequest("invalid json")
	}
	ioutil.WriteFile(extra_filename, body, os.ModePerm)
}

func apiDeleteMatches(w http.ResponseWriter, r *http.Request) {
	apiTBARequest("matches/delete", w, r)
}

func apiFetchRankings(w http.ResponseWriter, r *http.Request) {
	event := r.URL.Query().Get("event")
	level := checkRequestLevel(r)
	out, err := downloadRankings(level, event)
	if err != nil {
		apiPanicInternal("ranking fetch failed: %s", err)
	}
	w.Write(out)
}

func apiUploadRankings(w http.ResponseWriter, r *http.Request) {
	apiTBARequest("rankings/update", w, r)
}

func apiUploadVideos(w http.ResponseWriter, r *http.Request) {
	apiTBARequest("match_videos/add", w, r)
}

func apiUploadMedia(w http.ResponseWriter, r *http.Request) {
	apiTBARequest("media/add", w, r)
}

func apiFetchReport(w http.ResponseWriter, r *http.Request) {
	report_type := r.URL.Query().Get("report_type")
	if report_type == "" {
		apiPanicInternal("report_type param is required")
	}

	out, err := downloadReport(report_type)
	if err != nil {
		apiPanicInternal("failed to download report %s: %s", report_type, err)
	}

	sendJson(w, out)
}

func handleFuncWrapper(r *mux.Router, route string, handler func(w http.ResponseWriter, r *http.Request)) {
	r.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				switch err := r.(type) {
				case APIError:
					w.WriteHeader(err.code)
					w.Write([]byte(err.message))
				default:
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(fmt.Sprintf("Unknown internal error: %v", err)))
				}
			}
		}()
		handler(w, r)
	})
}

//go:embed web/dist/*
var embeddedFS embed.FS

func RunWebServer(port int, web_folder string) {
	r := mux.NewRouter()
	var web_files http.FileSystem
	if web_folder != "" {
		web_files = http.Dir(web_folder)
	} else {
		subfs, _ := fs.Sub(embeddedFS, "web/dist")
		web_files = http.FS(subfs)
	}
	handleFuncWrapper(r, "/js/version.js", jsVersion)
	handleFuncWrapper(r, "/js/fms_config.js", jsFMSConfig)
	handleFuncWrapper(r, "/js/brackets.js", jsBrackets)
	handleFuncWrapper(r, "/api/fms_config/get", apiGetFMSConfig)
	handleFuncWrapper(r, "/api/fms_config/set", apiSetFMSConfig)
	handleFuncWrapper(r, "/api/keys/fetch", apiKeysFetch)
	handleFuncWrapper(r, "/api/keys/update", apiKeysUpdate)
	handleFuncWrapper(r, "/api/info/upload", apiUploadEventInfo)
	handleFuncWrapper(r, "/api/teams/upload", apiUploadTeams)
	handleFuncWrapper(r, "/api/awards/upload", apiUploadAwards)
	handleFuncWrapper(r, "/api/matches/fetch", apiFetchMatches)
	handleFuncWrapper(r, "/api/matches/upload", apiUploadMatches)
	handleFuncWrapper(r, "/api/matches/mark_uploaded", apiMarkMatchesUploaded)
	handleFuncWrapper(r, "/api/matches/purge", apiPurgeMatches)
	handleFuncWrapper(r, "/api/matches/extra", apiMatchLoadExtra)
	handleFuncWrapper(r, "/api/matches/extra/save", apiMatchSaveExtra)
	handleFuncWrapper(r, "/api/matches/delete", apiDeleteMatches)
	handleFuncWrapper(r, "/api/rankings/fetch", apiFetchRankings)
	handleFuncWrapper(r, "/api/rankings/upload", apiUploadRankings)
	handleFuncWrapper(r, "/api/videos/upload", apiUploadVideos)
	handleFuncWrapper(r, "/api/media/upload", apiUploadMedia)
	handleFuncWrapper(r, "/api/report/fetch", apiFetchReport)
	r.PathPrefix("/").Handler(http.FileServer(web_files))
	addr := fmt.Sprintf(":%d", port)
	logger.Printf("Serving on %s\n", addr)
	err := http.ListenAndServe(addr, r)
	if err != nil {
		logger.Fatalf("Could not start server: %s\n", err)
	}
}
