<template>
    <div>
        <div class="float-right">{{ version }}</div>
        <h2>
            TBA uploader<span v-if="eventSelected"> -
                <a
                    :href="'https://www.thebluealliance.com/event/' + selectedEvent"
                    target="_blank"
                >{{ selectedEvent }}</a>
            </span>
        </h2>

        <b-tabs v-model="selectedTab">
            <b-tab title="Event setup">
                <div
                    v-if="!inAddEvent"
                    class="form-inline"
                >
                    <label>
                        Event:
                        <b-form-select
                            v-model="selectedEvent"
                            class="ml-2"
                        >
                            <option
                                value=""
                                disabled
                                selected
                            >Select an event</option>
                            <option
                                v-for="event in events"
                                :key="event"
                                :value="event"
                            >{{ event }}</option>
                            <option value="_add">Add an event...</option>
                        </b-form-select>
                    </label>
                    <b-button
                        variant="success"
                        @click="selectedEvent='_add'"
                    >
                        Add an event
                    </b-button>
                    <b-button
                        variant="info"
                        :disabled="!eventSelected"
                        @click="editSelectedEvent"
                    >
                        Edit this event
                    </b-button>
                    <b-button
                        variant="danger"
                        :disabled="!eventSelected"
                        @click="deleteSelectedEvent"
                    >
                        Delete this event
                    </b-button>
                    <b-button
                        variant="info"
                        @click="syncEvents"
                    >
                        Sync
                    </b-button>
                </div>
                <div v-if="!inAddEvent">
                    <div v-if="tbaEventData.name">Event name: {{ tbaEventData.name }} ({{ tbaEventData.year }})</div>
                    <hr>
                    <alert
                        v-model="tbaReadError"
                        type="danger"
                    />
                    <label class="row col">
                        Read API key (<a
                            href="https://www.thebluealliance.com/account"
                            target="_blank"
                        >create a key</a>&nbsp;if needed)
                        <b-form-input
                            v-model="readApiKey"
                            :type="authInputType"
                        />
                    </label>
                    <div>
                        <b-form-checkbox v-model="addEventUI.showAuth">
                            Show key
                        </b-form-checkbox>
                    </div>
                    <div>
                        <b-button
                            variant="success"
                            @click="fetchEventData"
                        >
                            Fetch event data from TBA
                        </b-button>
                    </div>
                </div>
                <div v-if="inAddEvent">
                    <h3>Add event</h3>
                    <label>
                        Event code:
                        <b-form-input v-model="addEventUI.event" />
                        <span
                            v-if="addEventUI.event.length >= 4 && !addEventIsValidYear"
                            class="text-danger"
                        >Invalid year</span>
                    </label>
                    <label class="row col">
                        Auth ID:
                        <b-form-input
                            v-model="addEventUI.auth"
                            :type="authInputType"
                        />
                    </label>
                    <label class="row col">
                        Auth secret:
                        <b-form-input
                            v-model="addEventUI.secret"
                            :type="authInputType"
                        />
                    </label>
                    <div>
                        <b-form-checkbox v-model="addEventUI.showAuth">
                            Show auth parameters
                        </b-form-checkbox>
                    </div>
                    <div>
                        <b-button
                            variant="danger"
                            @click="cancelAddEvent"
                        >
                            Cancel
                        </b-button>
                        <b-button
                            variant="success"
                            :disabled="!canAddEvent"
                            @click="addEvent"
                        >
                            Add
                        </b-button>
                    </div>
                </div>
                <div v-if="isEventSelected">
                    <hr>

                    <div>
                        <div>Note: any changes below may not be reflected quickly in data that is already uploaded to TBA (e.g. match schedules) unless they are uploaded again.</div>
                        <h3 class="mt-2">Team remappings</h3>
                        <alert
                            v-model="remapError"
                            type="danger"
                        />

                        <ul>
                            <li
                                v-for="(remap, i) in eventExtras[selectedEvent].remap_teams"
                                :key="i"
                            >
                                <form class="form-inline">
                                    <b-form-input v-model="remap.fms" />
                                    &rarr;
                                    <b-form-input v-model="remap.tba" />
                                    <b-button-close @click="removeTeamRemap(i)" />
                                </form>
                            </li>
                        </ul>

                        <b-button
                            variant="success"
                            @click="addTeamRemap"
                        >
                            Add
                        </b-button>
                        <b-button
                            v-if="eventExtras[selectedEvent].remap_teams.length"
                            variant="success"
                            @click="uploadTeamRemap"
                        >
                            Upload
                        </b-button>
                        <b-button
                            v-else
                            variant="warning"
                            @click="uploadTeamRemap"
                        >
                            Remove all remappings from TBA
                        </b-button>

                        <h3 class="mt-2">Playoff type</h3>
                        <div class="input-group">
                            <b-form-select
                                v-model.number="eventExtras[selectedEvent].playoff_type"
                                class="col-md-4 col-sm-6"
                                :options="BRACKET_TYPES"
                                :disabled="eventExtras[selectedEvent].playoff_type === undefined"
                            />
                            <b-button
                                variant="danger"
                                class="ml-2"
                                :disabled="eventExtras[selectedEvent].playoff_type === undefined || eventExtras[selectedEvent].playoff_type === tbaEventData.playoff_type"
                                @click="updatePlayoffType"
                            >
                                Change playoff type
                            </b-button>
                        </div>
                    </div>
                </div>
            </b-tab>

            <b-tab
                v-if="eventSelected"
                title="Schedule"
            >
                <alert
                    v-model="scheduleError"
                    type="danger"
                />
                <b-alert
                    v-model="scheduleUploaded"
                    variant="success"
                    dismissible
                >
                    The schedule was uploaded successfully.
                </b-alert>
                <h3>1. Select a schedule</h3>
                <p>This needs to be a CSV qualification or playoff <strong>schedule report</strong> generated by FMS.</p>
                <dropzone
                    ref="scheduleUpload"
                    title="Upload a schedule"
                    accept="text/csv"
                    @upload="onScheduleUpload"
                />
                <div v-if="scheduleStats.length">
                    <div class="mb-4">
                        <b-button
                            variant="danger"
                            @click="scheduleReset(false)"
                        >
                            Reset
                        </b-button>
                    </div>
                    <h3>2. Verify and upload to TBA</h3>
                    <p>Verify that the following information is correct:</p>
                    <ul>
                        <li
                            v-for="(s, i) in scheduleStats"
                            :key="i"
                        >
                            {{ s }}
                        </li>
                    </ul>
                    <div v-if="schedulePendingMatches.length">
                        <p>If this schedule looks right, <strong>click "Upload" below</strong>.</p>
                        <table class="table">
                            <thead>
                                <tr>
                                    <th>ID</th>
                                    <th>Key</th>
                                    <th>Time</th>
                                    <th>Red 1</th>
                                    <th>Red 2</th>
                                    <th>Red 3</th>
                                    <th>Blue 1</th>
                                    <th>Blue 2</th>
                                    <th>Blue 3</th>
                                </tr>
                            </thead>
                            <tr
                                v-for="(match, i) in schedulePendingMatchCells"
                                :key="i"
                            >
                                <td
                                    v-for="(cell, j) in match"
                                    :key="j"
                                    :class="cell.cls"
                                >
                                    <span :class="cell.cls">{{ cell.text }}</span>
                                </td>
                            </tr>
                        </table>
                        <alert
                            v-model="scheduleError"
                            type="danger"
                        />
                        <p>
                            <span class="warning">Warning:</span> this will overwrite any match data on TBA for these matches.
                            Double-check the event page if you have any doubt that these matches have not already been uploaded.
                        </p>
                        <b-button
                            variant="success"
                            :disabled="inScheduleRequest || !schedulePendingMatches.length"
                            @click="postSchedule"
                        >
                            Upload these matches to TBA
                        </b-button>
                    </div>
                    <b-alert
                        v-else-if="!inScheduleRequest"
                        variant="warning"
                        show
                    >
                        No new competition levels detected; nothing to upload.
                    </b-alert>
                </div>
            </b-tab>

            <b-tab
                v-if="eventSelected"
                title="Match play"
            >
                <div class="form-inline">
                    <b-form-select
                        v-model="matchLevel"
                    >
                        <option
                            v-if="uiOptions.showAllLevels"
                            value="0"
                        >
                            Test
                        </option>
                        <option
                            v-if="uiOptions.showAllLevels"
                            value="1"
                        >
                            Practice
                        </option>
                        <option value="2">Qualification</option>
                        <option value="3">Playoff</option>
                    </b-form-select>
                    <b-button
                        variant="success"
                        data-accesskey="f"
                        :disabled="inMatchRequest"
                        @click="fetchMatches(false)"
                    >
                        Fetch new matches
                    </b-button>
                    <b-button
                        variant="danger"
                        data-accesskey="a"
                        @click="inMatchAdvanced = !inMatchAdvanced"
                    >
                        Advanced options
                    </b-button>
                    <b-button
                        v-if="isQual"
                        variant="info"
                        class="ml-auto"
                        data-accesskey="r"
                        :disabled="inUploadRankings"
                        @click="uploadRankings"
                    >
                        Upload rankings
                    </b-button>
                </div>
                <p>
                    <span class="warning">Warning:</span> do not click any buttons on this page while a match is running.
                    Be sure to only fetch (or re-fetch) matches <strong>after</strong> scores have been posted in FMS.
                    <span v-if="isQual">Rankings can be updated at any time if necessary, but will also be updated after posting scores.</span>
                </p>
                <alert
                    v-model="rankingsError"
                    type="danger"
                    prefix="Rankings:"
                />
                <div v-if="inMatchAdvanced">
                    <hr>
                    <h3>
                        Advanced options <b-button
                            variant="outline-danger"
                            size="sm"
                            @click.prevent="inMatchAdvanced = false"
                        >
                            Close
                        </b-button>
                    </h3>
                    <div class="form-inline mb-2">
                        <label>
                            Match ID (Match-Play):
                            <b-form-input
                                v-model="advSelectedMatch"
                                placeholder="M-P"
                                style="width: 5em;"
                            />
                        </label>
                        <b-button
                            variant="warning"
                            :disabled="inMatchRequest"
                            @click="purgeAdvSelectedMatch"
                        >
                            Purge
                        </b-button>
                        <b-button
                            variant="warning"
                            :disabled="inMatchRequest"
                            @click="markAdvSelectedMatchUploaded"
                        >
                            Mark as uploaded
                        </b-button>
                        <span class="warning">{{ advMatchError }}</span>
                    </div>
                    <div>
                        <b-button
                            variant="danger"
                            :disabled="inMatchRequest"
                            @click="fetchMatches(true)"
                        >
                            Purge and re-fetch all matches
                        </b-button>
                    </div>
                    <hr>
                </div>
                <alert
                    v-model="matchError"
                    type="danger"
                />
                <div v-if="matchSummaries.length">
                    <h3>Matches to upload ({{ matchSummaries.length }})</h3>
                    <b-button
                        variant="success"
                        class="mb-2"
                        data-accesskey="u"
                        accesskey="u"
                        title="[u]"
                        :disabled="inMatchRequest"
                        @click="uploadMatches"
                    >
                        Upload scores
                    </b-button>
                    <b-button
                        variant="warning"
                        class="mb-2"
                        data-accesskey="e"
                        accesskey="e"
                        title="[e]"
                        :disabled="inMatchRequest"
                        @click="refetchMatches"
                    >
                        Re-fetch scores
                    </b-button>
                    <b-alert
                        variant="danger"
                        :show="fetchedScorelessMatches"
                    >
                        One or more matches below appear to have been fetched before scores were posted. Click "Re-fetch scores" to try again (wait for scores to be posted from FMS first). If the scores below are correct, click "Upload scores".
                    </b-alert>
                    <b-alert
                        variant="danger"
                        :show="unhandledBreakdowns.length > 0"
                    >
                        Some breakdowns were not handled: {{ unhandledBreakdowns.join(", ") }}. Any affected matches will need to be manually edited.
                    </b-alert>
                    <div>
                        Click "Upload scores" to upload these scores to TBA<span v-if="isQual"> and update rankings</span>. If a match needs to be edited, click on it below. Reasons for this include:
                        <ul>
                            <li>Any extra rankings points awarded by the head referee (this accompanies score changes in FMS)</li>
                            <li>Red cards</li>
                            <li>Surrogate teams</li>
                        </ul>
                    </div>
                    <div class="row">
                        <score-summary
                            v-for="match in matchSummaries"
                            :key="match.id"
                            :match="match"
                            @click="showEditMatch(match)"
                        />
                    </div> <!-- row -->
                </div>
            </b-tab>

            <b-tab
                v-if="eventSelected"
                title="Match videos"
            >
                <div>
                    <b-button
                        variant="success"
                        :disabled="inVideoRequest"
                        @click="fetchVideos"
                    >
                        Fetch data from TBA
                    </b-button>
                    <b-button
                        variant="success"
                        :disabled="inVideoRequest"
                        @click="uploadVideos"
                    >
                        Upload data to TBA
                    </b-button>
                </div>
                <ul>
                    <li><span class="warning">Warning:</span> Anything entered in this tab will not be saved locally. You will have to upload all data here to TBA before closing this window.</li>
                    <li>Videos cannot be removed from TBA once they are uploaded.</li>
                </ul>
                <alert
                    v-model="videoError"
                    type="danger"
                />

                <b-form-checkbox v-model="showExistingVideos">
                    Show matches that already have videos
                </b-form-checkbox>

                <ul>
                    <li
                        v-for="[key, video] in getSortedVideos()"
                        :key="key"
                    >
                        <div class="form-inline">
                            <label>
                                {{ key }}:
                                <b-form-input
                                    v-model="video.current"
                                    @blur="cleanVideoUrls"
                                />
                            </label>
                        </div>
                    </li>
                </ul>
            </b-tab>

            <b-tab
                v-if="eventSelected"
                title="Awards"
            >
                <table class="table">
                    <thead>
                        <tr>
                            <th>Award Name</th>
                            <th>Team (recommended)</th>
                            <th>Person (optional)</th>
                            <th>Options</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr
                            v-for="award in awards[selectedEvent]"
                            :key="award.id"
                        >
                            <td>
                                <b-form-input
                                    v-model="award.name"
                                    title="Award Name"
                                    placeholder="Award Name"
                                    @blur="saveAwards"
                                />
                            </td>
                            <td>
                                <b-form-input
                                    v-model="award.team"
                                    type="number"
                                    title="Team"
                                    placeholder="Team"
                                    @blur="saveAwards"
                                />
                            </td>
                            <td>
                                <b-form-input
                                    v-model="award.person"
                                    title="Person"
                                    placeholder="Person"
                                    @blur="saveAwards"
                                />
                            </td>
                            <td>
                                <b-button
                                    variant="info"
                                    title="Create a new award with the same name"
                                    @click="duplicateAward(award)"
                                >
                                    Duplicate
                                </b-button>
                                <b-button
                                    variant="danger"
                                    @click="clearAward(award)"
                                >
                                    Clear
                                </b-button>
                                <b-button
                                    variant="danger"
                                    @click="deleteAward(award)"
                                >
                                    Delete
                                </b-button>
                            </td>
                        </tr>
                    </tbody>
                </table>
                <b-button
                    variant="success"
                    @click="addAward"
                >
                    Add award
                </b-button>
                <b-button
                    variant="success"
                    @click="fetchAutomaticAwards"
                >
                    Auto-detect winners/finalists
                </b-button>
                <hr>
                <b-button
                    variant="success"
                    :disabled="inAwardRequest"
                    @click="uploadAwards"
                >
                    Upload all awards
                </b-button>
                <span>{{ awardStatus }}</span>
            </b-tab>

            <b-tab
                title="Options"
                title-item-class="ml-auto"
            >
                <h2>UI options</h2>
                <p>All options in this section are saved automatically.</p>
                <div class="row">
                    <div class="col-sm-12">
                        <b-form-checkbox v-model="uiOptions.showAllLevels">
                            Show hidden tournament levels (Practice/Test)
                        </b-form-checkbox>
                    </div>
                </div>
                <hr>
                <h2>FMS options</h2>
                <p>Options in this section need to be saved by clicking "Save" below. Also note that these can be specified on the command line as well, which is more useful for development.</p>
                <div class="row mb-2">
                    <label class="col-sm-12 col-md-8">
                        Server (default: <code>http://10.0.100.5</code>):
                        <b-form-input v-model="fmsConfig.server" />
                    </label>
                    <label class="col-sm-12 col-md-8">
                        Data folder:
                        <b-form-input v-model="fmsConfig.data_folder" />
                    </label>
                    <div class="col-sm-12">
                        <b-button
                            variant="success"
                            @click="saveFMSConfig"
                        >
                            Save
                        </b-button>
                        <b-button
                            variant="danger"
                            @click="resetFMSConfig"
                        >
                            Reset
                        </b-button>
                    </div>
                </div>
                <alert
                    v-model="fmsConfigError"
                    type="danger"
                />
            </b-tab>

            <!-- eslint-disable vue/no-v-html -->
            <b-tab
                title="Help"
                v-html="helpHTML"
            />
        </b-tabs>

        <b-modal
            ref="matchEditModal"
            :title="matchEditing && `Edit match: ${matchEditing.key} (${matchEditing.id})`"
            return-focus="body"
        >
            <div v-if="inEditMatch">
                <div>
                    <b-form-checkbox v-model="matchEditOverrideCode">
                        Match code override:
                    </b-form-checkbox>
                </div>
                <div
                    v-if="matchEditOverrideCode"
                    class="form-inline mb-2"
                >
                    <b-form-select
                        v-model="matchEditing.code.comp_level"
                    >
                        <option>qm</option>
                        <option>ef</option>
                        <option>qf</option>
                        <option>sf</option>
                        <option>f</option>
                    </b-form-select>
                    <b-form-input
                        v-if="matchEditing.code.comp_level != 'qm'"
                        v-model.number="matchEditing.code.set_number"
                        type="number"
                        min="1"
                        max="999"
                        step="1"
                    />
                    match
                    <b-form-input
                        v-model.number="matchEditing.code.match_number"
                        type="number"
                        min="1"
                        max="999"
                        step="1"
                    />
                </div>
                <alert
                    v-model="matchEditError"
                    type="danger"
                />
                <table class="table match-edit">
                    <thead>
                        <tr>
                            <th>Red</th>
                            <th>Blue</th>
                        </tr>
                    </thead>
                    <tbody v-if="!isPlayoff">
                        <tr
                            v-for="i in 3"
                            :key="i"
                        >
                            <td class="red">
                                <h6>{{ matchEditData.teams.red[i-1].team }}</h6>
                                <b-form-checkbox v-model="matchEditData.teams.red[i-1].dq">
                                    DQ (Red Card)
                                </b-form-checkbox>
                                <b-form-checkbox v-model="matchEditData.teams.red[i-1].surrogate">
                                    Surrogate
                                </b-form-checkbox>
                            </td>
                            <td class="blue">
                                <h6>{{ matchEditData.teams.blue[i-1].team }}</h6>
                                <b-form-checkbox v-model="matchEditData.teams.blue[i-1].dq">
                                    DQ (Red Card)
                                </b-form-checkbox>
                                <b-form-checkbox v-model="matchEditData.teams.blue[i-1].surrogate">
                                    Surrogate
                                </b-form-checkbox>
                            </td>
                        </tr>
                        <tr v-if="eventYear == 2018">
                            <td class="red">
                                <b-form-checkbox v-model="matchEditData.flags.red.invert_auto">
                                    Force auto RP {{ matchEditData.text.red.auto_rp }}
                                </b-form-checkbox>
                            </td>
                            <td class="blue">
                                <b-form-checkbox v-model="matchEditData.flags.blue.invert_auto">
                                    Force auto RP {{ matchEditData.text.blue.auto_rp }}
                                </b-form-checkbox>
                            </td>
                        </tr>
                        <tr v-if="eventYear == 2019">
                            <td class="red">
                                <b-form-checkbox v-model="matchEditData.flags.red.add_rp_rocket">
                                    Give rocket RP
                                </b-form-checkbox>
                                <b-form-checkbox v-model="matchEditData.flags.red.add_rp_hab_climb">
                                    Give HAB climb RP
                                </b-form-checkbox>
                            </td>
                            <td class="blue">
                                <b-form-checkbox v-model="matchEditData.flags.blue.add_rp_rocket">
                                    Give rocket RP
                                </b-form-checkbox>
                                <b-form-checkbox v-model="matchEditData.flags.blue.add_rp_hab_climb">
                                    Give HAB climb RP
                                </b-form-checkbox>
                            </td>
                        </tr>
                    </tbody>

                    <tbody v-if="isPlayoff">
                        <tr>
                            <td class="red">
                                <h6>{{ matchEditData.teams.red[0].team }} &bull; {{ matchEditData.teams.red[1].team }} &bull; {{ matchEditData.teams.red[2].team }}</h6>
                                <b-form-checkbox v-model="matchEditData.flags.red.dq">
                                    DQ (Red Card)
                                </b-form-checkbox>
                            </td>
                            <td class="blue">
                                <h6>{{ matchEditData.teams.blue[0].team }} &bull; {{ matchEditData.teams.blue[1].team }} &bull; {{ matchEditData.teams.blue[2].team }}</h6>
                                <b-form-checkbox v-model="matchEditData.flags.blue.dq">
                                    DQ (Red Card)
                                </b-form-checkbox>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>

            <b-button
                variant="warning"
                @click="editMatchMarkUploaded"
            >
                Mark this match as uploaded
            </b-button>

            <template v-slot:modal-footer>
                <b-button
                    variant="secondary"
                    @click="hideEditMatch"
                >
                    Cancel
                </b-button>
                <b-button
                    variant="primary"
                    @click="saveEditMatch"
                >
                    Save changes
                </b-button>
            </template>
        </b-modal>
    </div>
</template>

<script>
import {
    BAlert,
    BButton,
    BButtonClose,
    BFormCheckbox,
    BFormInput,
    BFormSelect,
    BModal,
    BTab,
    BTabs,
} from 'bootstrap-vue';
import 'regenerator-runtime';
import showdown from 'showdown';
import Vue from 'vue';

import api from 'src/api.js';
import {
    BRACKET_NAME,
    BRACKET_TYPE,
    MATCH_LEVEL,
} from 'src/consts.js';
import Schedule from 'src/schedule.js';
import tba from 'src/tba.js';
import utils from 'src/utils.js';

import Alert from 'components/Alert.vue';
import Dropzone from 'components/Dropzone.vue';
import ScoreSummary from 'components/ScoreSummary.vue';

import 'bootstrap/dist/css/bootstrap.css';
import 'bootstrap-vue/dist/bootstrap-vue.css';
import 'src/app.css';

const STORED_EVENTS = utils.safeParseLocalStorageObject('storedEvents');
const STORED_AWARDS = utils.safeParseLocalStorageObject('awards');

function sendApiRequest(url, event, body) {
    return $.ajax({
        type: 'POST',
        url: url,
        contentType: 'application/json',
        data: JSON.stringify(body),
        headers: {
            'X-Event': event,
            'X-Auth': STORED_EVENTS[event].auth,
            'X-Secret': STORED_EVENTS[event].secret,
        },
    });
}
window.sendApiRequest = sendApiRequest;

function tbaApiEventRequest(event, route) {
    var url = 'https://www.thebluealliance.com/api/v3/event/' + event;
    if (route) {
        url += '/' + route;
    }
    return $.ajax({
        type: 'GET',
        url: url,
        headers: {
            'X-TBA-Auth-Key': localStorage.getItem('readApiKey'),
        },
        cache: false,
    });
}
window.tbaApiEventRequest = tbaApiEventRequest;

function confirmPurge() {
    return confirm('Are you sure? This may replace old match results and re-send notifications when these match(es) are uploaded again.');
}

function makeAddEventUI() {
    return {
        event: '',
        auth: '',
        secret: '',
        showAuth: false,
    };
}

function makeAward(data) {
    return $.extend({}, {
        name: '',
        team: '',
        person: '',
    }, data || {});
}

const EXTRA_FIELDS = {
    2018: {
        invert_auto: false,
    },
    2019: {
        add_rp_rocket: false,
        add_rp_hab_climb: false,
    },
};

export default {
    components: {
        Alert,
        BAlert,
        BButton,
        BButtonClose,
        BFormCheckbox,
        BFormInput,
        BFormSelect,
        BModal,
        BTab,
        BTabs,
        Dropzone,
        ScoreSummary,
    },
    data: () => ({
        version: window.VERSION || 'missing version',
        helpHTML: '',
        fmsConfig: window.FMS_CONFIG || {},
        fmsConfigError: '',
        selectedTab: utils.safeParseLocalStorageInteger('lastTab', 0),
        events: Object.keys(STORED_EVENTS).sort(),
        selectedEvent: localStorage.getItem('selectedEvent') || '',
        addEventUI: makeAddEventUI(),
        readApiKey: localStorage.getItem('readApiKey') || '',
        tbaEventData: {},
        tbaReadError: '',

        uiOptions: $.extend({
            showAllLevels: false,
        }, utils.safeParseLocalStorageObject('uiOptions')),
        eventExtras: utils.safeParseLocalStorageObject('eventExtras'),
        remapError: '',

        inScheduleRequest: false,
        scheduleUploaded: false,
        scheduleError: '',
        scheduleStats: [],
        schedulePendingMatches: [],

        matchLevel: utils.safeParseLocalStorageInteger('matchLevel', MATCH_LEVEL.QUAL),
        showAllLevels: false,
        inMatchRequest: false,
        matchError: '',
        // pendingMatches: [], // not set yet to avoid Vue binding to this
        matchSummaries: [],
        fetchedScorelessMatches: false,
        unhandledBreakdowns: [],
        inMatchAdvanced: false,
        advSelectedMatch: '',
        advMatchError: '',

        inEditMatch: false,
        matchEditing: null,
        matchEditData: null,
        matchEditError: '',
        matchEditOverrideCode: false,

        inUploadRankings: false,
        rankingsError: '',

        videos: {},
        inVideoRequest: false,
        videoError: '',
        showExistingVideos: false,

        awards: STORED_AWARDS,
        awardStatus: '',
        inAwardRequest: false,
    }),
    computed: {
        BRACKET_TYPES: function() {
            return Object.fromEntries(Object.keys(BRACKET_TYPE).map((key) => [
                BRACKET_TYPE[key], BRACKET_NAME[key],
            ]));
        },
        eventSelected: function() {
            return !!this.selectedEvent && !this.inAddEvent;
        },
        inAddEvent: function() {
            return this.selectedEvent == '_add';
        },
        canAddEvent: function() {
            return this.addEventUI.event && this.addEventUI.auth && this.addEventUI.secret &&
                tba.isValidYear(this.addEventUI.event);
        },
        addEventIsValidYear: function() {
            return tba.isValidYear(this.addEventUI.event);
        },
        eventRequestHeaders: function() {
            if (!STORED_EVENTS[this.selectedEvent]) {
                return {};
            }
            return {
                'X-Event': this.selectedEvent,
                'X-Auth': STORED_EVENTS[this.selectedEvent].auth,
                'X-Secret': STORED_EVENTS[this.selectedEvent].secret,
            };
        },
        authInputType: function() {
            return this.addEventUI.showAuth ? 'text' : 'password';
        },
        isEventSelected: function() {
            return tba.isValidEventCode(this.selectedEvent);
        },
        eventYear: function() {
            var year = parseInt(this.selectedEvent);
            if (isNaN(year)) {
                return new Date().getFullYear();
            }
            return year;
        },
        isQual: function() {
            return this.matchLevel == MATCH_LEVEL.QUAL;
        },
        isPlayoff: function() {
            return this.matchLevel == MATCH_LEVEL.PLAYOFF;
        },
        schedulePendingMatchCells: function() {
            var addTeamCell = function(cells, match, color, i) {
                var cls = {};
                cls[color] = true;
                cls['surrogate'] = match.alliances[color].surrogates.indexOf(match.alliances[color].teams[i]) >= 0;
                cells.push({
                    text: match.alliances[color].teams[i].replace('frc', ''),
                    cls: cls,
                });
            };
            return this.schedulePendingMatches.map(function(match) {
                var cells = [
                    {text: match._id},
                    {text: match._key},
                    {text: match.time_string},
                ];
                ['red', 'blue'].forEach(function(color) {
                    for (var i = 0; i < 3; i++) {
                        addTeamCell(cells, match, color, i);
                    }
                });
                return cells;
            });
        },
    },
    watch: {
        selectedTab: function(tab) {
            localStorage.setItem('lastTab', tab);
        },
        readApiKey: function(key) {
            localStorage.setItem('readApiKey', key);
        },
        selectedEvent: function(event) {
            localStorage.setItem('selectedEvent', event);
            this.initEvent(event);
            this.fetchEventData();
            this.scheduleReset(false);
        },
        matchLevel: function() {
            localStorage.setItem('matchLevel', this.matchLevel);
        },
        uiOptions: {
            handler: function() {
                localStorage.setItem('uiOptions', JSON.stringify(this.uiOptions));
            },
            deep: true,
        },
        eventExtras: {
            handler: function() {
                localStorage.setItem('eventExtras', JSON.stringify(this.eventExtras));
            },
            deep: true,
        },
    },
    mounted: function() {
        if (this.selectedEvent) {
            this.initEvent(this.selectedEvent);
            this.fetchEventData();
        }

        $.get('/README.md', function(readme) {
            // remove first line (header)
            readme = readme.substr(readme.indexOf('\n'));
            this.helpHTML = new showdown.Converter().makeHtml(readme);
        }.bind(this));

        $(function() {
            $(this.$el).removeClass('hidden');
        }.bind(this));
    },
    methods: {
        saveFMSConfig: function() {
            this.fmsConfigError = '';
            $.ajax({
                type: 'POST',
                url: '/api/fms_config/set',
                contentType: 'application/json',
                data: JSON.stringify(this.fmsConfig),
            }).always(function(data) {
                data = JSON.parse(data);
                if (!data.ok) {
                    this.fmsConfigError = 'Failed to save options: ' + data.error;
                }
            }.bind(this));
        },
        resetFMSConfig: function() {
            $.getJSON('/api/fms_config/get', function(data) {
                this.fmsConfig = data;
            }.bind(this));
        },

        addEvent: function() {
            var event = this.addEventUI.event;
            STORED_EVENTS[event] = {
                auth: this.addEventUI.auth,
                secret: this.addEventUI.secret,
            };
            this.selectedEvent = event;
            if (this.events.indexOf(event) == -1) {
                this.events.push(event);
                this.events.sort();
                this.initEvent(event);
            }
            localStorage.setItem('storedEvents', JSON.stringify(STORED_EVENTS));
            this.addEventUI = makeAddEventUI();
            this.syncEvents();
        },
        cancelAddEvent: function() {
            this.selectedEvent = '';
        },
        editSelectedEvent: function() {
            this.addEventUI.event = this.selectedEvent;
            this.addEventUI.auth = STORED_EVENTS[this.selectedEvent].auth;
            this.addEventUI.secret = STORED_EVENTS[this.selectedEvent].secret;
            this.selectedEvent = '_add';
        },
        deleteSelectedEvent: function() {
            var oldEvent = this.selectedEvent;
            if (!confirm('Are you sure you want to delete the event "' + oldEvent + '"?')) {
                return;
            }
            this.selectedEvent = '';
            STORED_EVENTS[oldEvent] = undefined;
            this.events = this.events.filter(function(event) {
                return event != oldEvent;
            }.bind(this));
            localStorage.setItem('storedEvents', JSON.stringify(STORED_EVENTS));
        },
        fetchEventData: function() {
            this.tbaReadError = '';
            this.$set(this, 'tbaEventData', {});
            if (!tba.isValidEventCode(this.selectedEvent)) {
                return;
            }
            if (!this.readApiKey) {
                this.tbaReadError = 'No TBA Read API key is present, so event data cannot be retrieved from TBA.';
                return;
            }
            tbaApiEventRequest(this.selectedEvent).then(function(data) {
                this.$set(this, 'tbaEventData', data);
                this.eventExtras[this.selectedEvent].playoff_type = data.playoff_type;
            }.bind(this))
            .fail(function(error) {
                this.tbaReadError = utils.parseErrorJSON(error);
            }.bind(this));
        },
        initEvent: function(event) {
            if (!tba.isValidEventCode(event)) {
                return;
            }

            this.$set(this.eventExtras, event, $.extend({}, {
                remap_teams: [],
            }, this.eventExtras[event]));

            if (!this.awards[event] || !this.awards[event].length) {
                this.$set(this.awards, event, [makeAward()]);
                this.saveAwards();
            }
        },
        syncEvents: async function() {
            var events = await api.postJson({url: '/api/keys/fetch'});
            try {
                events = JSON.parse(events);
            }
            catch (e) {
                events = {};
            }
            if (events._read_api_key && !this.readApiKey) {
                this.readApiKey = events._read_api_key;
            }
            for (const k of Object.keys(events)) {
                if (!STORED_EVENTS[k] && !k.startsWith('_')) {
                    STORED_EVENTS[k] = events[k];
                    this.events.push(k);
                }
            }
            localStorage.setItem('storedEvents', JSON.stringify(STORED_EVENTS));
            await api.postJson({
                url: '/api/keys/update',
                body: {
                    ...STORED_EVENTS,
                    _read_api_key: events._read_api_key || this.readApiKey,
                },
            });
        },

        addTeamRemap: function() {
            this.eventExtras[this.selectedEvent].remap_teams.push({
                fms: '',
                tba: '',
            });
        },
        removeTeamRemap: function(i) {
            this.eventExtras[this.selectedEvent].remap_teams.splice(i, 1);
        },
        uploadTeamRemap: function() {
            this.remapError = '';
            var remapList = this.eventExtras[this.selectedEvent].remap_teams;
            var remapMap = {};
            var validate = function(team, isTba) {
                var match = team
                    .toUpperCase()
                    .trim()
                    .replace(/^FRC/, '')
                    .match(isTba ? /^\d+[B-Z]$/ : /^\d+$/);
                if (!match) {
                    this.remapError += 'Invalid ' + (isTba ? 'TBA' : 'FMS') + ' team number: ' + team +
                        ': expected format: 1234' + (isTba ? 'B (or 1234C, etc.)' : '') + '\n';
                }
                return 'frc' + match;
            }.bind(this);
            remapList.forEach(function(r) {
                var fms_team = validate(r.fms);
                var tba_team = validate(r.tba, true);
                if (fms_team && tba_team) {
                    remapMap[fms_team] = tba_team;
                }
            });
            if (this.remapError) {
                return;
            }
            sendApiRequest('/api/info/upload', this.selectedEvent, {
                remap_teams: remapMap,
            }).fail(function(error) {
                this.remapError = utils.parseErrorText(error);
            }.bind(this));
        },

        updatePlayoffType: function() {
            const playoff_type = this.eventExtras[this.selectedEvent].playoff_type;
            sendApiRequest('/api/info/upload', this.selectedEvent, {
                playoff_type,
            }).then(function() {
                this.tbaEventData.playoff_type = playoff_type;
            }.bind(this)).fail(function(error) {
                this.remapError = utils.parseErrorText(error);
            }.bind(this));
        },

        scheduleReset: function(keepFile) {
            this.inScheduleRequest = false;
            this.scheduleUploaded = false;
            this.scheduleError = '';
            this.scheduleStats = [];
            this.schedulePendingMatches = [];

            if (!keepFile && this.$refs.scheduleUpload) {
                this.$refs.scheduleUpload.reset();
            }
        },
        onScheduleUpload: function(event) {
            this.scheduleReset(true);
            try {
                var schedule = Schedule.parse(event.body);
            }
            catch (error) {
                if (typeof error == 'string') {
                    this.scheduleError = error;
                    return;
                }
                else {
                    throw error;
                }
            }
            this.scheduleStats.push(schedule.length + ' match(es) found');
            var numSurrogates = schedule.map(function(match) {
                return match.alliances.red.surrogates.length + match.alliances.blue.surrogates.length;
            }).reduce(function(a, b) {
                return a + b;
            }, 0);
            this.scheduleStats.push(numSurrogates + ' surrogate team(s)');

            this.scheduleStats.push('Checking against TBA schedule...');
            this.inScheduleRequest = true;
            tbaApiEventRequest(this.selectedEvent, 'matches').always(function() {
                this.inScheduleRequest = false;
                this.scheduleStats.pop();
            }.bind(this)).then(function(tbaMatches) {
                if (!tbaMatches) {
                    tbaMatches = [];
                }
                var newLevels = Schedule.findAllCompLevels(schedule);
                var tbaLevels = Schedule.findAllCompLevels(tbaMatches);
                this.scheduleStats.push('TBA has level(s): ' + tbaLevels.join(', '));
                this.scheduleStats.push('The FMS report has level(s): ' + newLevels.join(', '));
                newLevels = newLevels.filter(function(level) {
                    return tbaLevels.indexOf(level) < 0;
                });
                if (!newLevels.length) {
                    this.scheduleStats.push('No new levels are present in the FMS report.');
                    return;
                }
                this.scheduleStats.push('Level(s) to be added from the FMS report: ' + newLevels.join(', '));
                this.schedulePendingMatches = schedule.filter(function(match) {
                    return newLevels.indexOf(match.comp_level) >= 0;
                });
            }.bind(this)).fail(function(error) {
                this.scheduleError = utils.parseErrorJSON(error);
            }.bind(this));
        },
        postSchedule: function() {
            this.scheduleError = '';
            this.inScheduleRequest = true;
            sendApiRequest('/api/matches/upload', this.selectedEvent, this.schedulePendingMatches).always(function() {
                this.inScheduleRequest = false;
            }.bind(this)).then(function() {
                this.scheduleReset(false);
                this.scheduleUploaded = true;
            }.bind(this)).fail(function(res) {
                this.scheduleError = res.responseText;
            }.bind(this));
        },

        fetchMatches: function(all) {
            if (all && !confirmPurge()) {
                return;
            }
            this.inMatchRequest = true;
            this.fetchedScorelessMatches = false;
            this.matchError = '';
            $.get('/api/matches/fetch', {
                event: this.selectedEvent,
                level: this.matchLevel,
                all: all ? '1' : '',
            }).always(function() {
                this.inMatchRequest = false;
            }.bind(this)).then(function(data) {
                this.pendingMatches = JSON.parse(data);
                this.pendingMatches.sort(function(a, b) {
                    return Number(a._fms_id.split('-')[0]) - Number(b._fms_id.split('-')[0]);
                });
                this.matchSummaries = this.generateMatchSummaries(this.pendingMatches);
                this.fetchedScorelessMatches = this.checkScorelessMatches(this.pendingMatches);
                this.unhandledBreakdowns = this.findUnhandledBreakdowns(this.pendingMatches);
            }.bind(this)).fail(function(res) {
                this.matchError = res.responseText;
            }.bind(this));
        },
        refetchMatches: function() {
            var match_ids = this.pendingMatches.map(function(match) {
                return match._fms_id;
            });
            this.inMatchRequest = true;
            this.matchError = '';
            sendApiRequest('/api/matches/purge?level=' + this.matchLevel, this.selectedEvent, match_ids)
            .then(function() {
                this.fetchMatches(false);
            }.bind(this))
            .fail(function(res) {
                this.matchError = 'Purge: ' + res.responseText;
                this.inMatchRequest = false;
            }.bind(this));
        },
        generateMatchSummaries: function(matches) {
            var rmFRC = function(team) {
                return team.replace('frc', '');
            };
            var formatScoreSummary = function(match, breakdown, color) {
                var s = '' + match.alliances[color].score;
                if (match.comp_level == 'qm') {
                    s += ' (' + breakdown[color].rp + ')';
                }
                return s;
            };
            var genClasses = function(match, team_key, color) {
                var classes = [color];
                if (match.alliances[color].dqs.indexOf(team_key) != -1) {
                    classes.push('dq');
                }
                if (match.alliances[color].surrogates.indexOf(team_key) != -1) {
                    classes.push('surrogate');
                }
                return classes;
            };

            return matches.map(function(match) {
                var classes = {};
                match.alliances.blue.teams.forEach(function(team_key) {
                    classes[rmFRC(team_key)] = genClasses(match, team_key, 'blue');
                });
                match.alliances.red.teams.forEach(function(team_key) {
                    classes[rmFRC(team_key)] = genClasses(match, team_key, 'red');
                });
                return {
                    id: match._fms_id,
                    key: Schedule.getTBAMatchKey(match),
                    code: {
                        comp_level: match.comp_level,
                        set_number: match.set_number,
                        match_number: match.match_number,
                    },
                    teams: {
                        blue: match.alliances.blue.teams.map(rmFRC),
                        red: match.alliances.red.teams.map(rmFRC),
                    },
                    score_summary: {
                        blue: formatScoreSummary(match, match.score_breakdown, 'blue'),
                        red: formatScoreSummary(match, match.score_breakdown, 'red'),
                    },
                    classes: classes,
                };
            });
        },
        cleanMatches: function(matches) {
            return matches.map(function(match) {
                match = Object.assign({}, match);
                delete match._fms_id;
                return match;
            });
        },
        uploadMatches: function() {
            this.matchError = '';
            this.inMatchRequest = true;
            var matches = this.cleanMatches(this.pendingMatches);
            var match_ids = this.pendingMatches.map(function(match) {
                return match._fms_id;
            });
            sendApiRequest('/api/matches/upload', this.selectedEvent, matches).always(function() {
                this.inMatchRequest = false;
            }.bind(this)).then(function() {
                this.pendingMatches = [];
                this.matchSummaries = [];
                if (this.isQual) {
                    this.uploadRankings();
                }
                sendApiRequest('/api/matches/mark_uploaded?level=' + this.matchLevel,
                    this.selectedEvent, match_ids
                ).fail(function(res) {
                    this.matchError += '\nReceipt generation failed: ' + res.responseText;
                }.bind(this));
            }.bind(this)).fail(function(res) {
                this.matchError = res.responseText;
            }.bind(this));
        },
        checkScorelessMatches: function(matches) {
            return matches.filter(function(match) {
                return match.alliances.blue.score == -1 || match.alliances.blue.teams[0] == '';
            }).length > 0;
        },
        findUnhandledBreakdowns: function(matches) {
            var unhandled = new Set();
            for (const match of matches) {
                for (const breakdown of Object.values(match.score_breakdown)) {
                    for (const field of Object.keys(breakdown)) {
                        if (field.startsWith('!')) {
                            unhandled.add(field.replace(/!/g, ''));
                        }
                    }
                }
            }
            return [...unhandled];
        },
        _checkAdvSelectedMatch: function() {
            var parts = this.advSelectedMatch.split('-');
            if (parts.length == 1) {
                parts.push('1');
            }
            this.advSelectedMatch = parts.join('-');
            this.advMatchError = '';
            if (!this.advSelectedMatch.match(/^\d+-\d+$/)) {
                this.advMatchError = 'Invalid match ID format';
                return false;
            }
            return true;
        },
        purgeAdvSelectedMatch: async function() {
            if (!this._checkAdvSelectedMatch() || !confirmPurge()) {
                return;
            }
            this.inMatchRequest = true;
            this.advMatchError = '';
            try {
                await api.postJson({
                    url: '/api/matches/purge?level=' + this.matchLevel,
                    headers: this.eventRequestHeaders,
                    body: [this.advSelectedMatch],
                });
                this.fetchMatches(false);
            }
            catch (error) {
                this.advMatchError = error;
            }
            finally {
                this.inMatchRequest = false;
            }
        },
        markAdvSelectedMatchUploaded: function() {
            if (!this._checkAdvSelectedMatch()) {
                return;
            }
            this.inMatchRequest = true;
            this.advMatchError = '';
            sendApiRequest('/api/matches/mark_uploaded?level=' + this.matchLevel,
                this.selectedEvent, [this.advSelectedMatch])
            .always(function() {
                this.inMatchRequest = false;
            }.bind(this))
            .then(function() {
                this.fetchMatches(false);
            }.bind(this))
            .fail(function(res) {
                this.advMatchError = 'Receipt generation failed: ' + res.responseText;
            }.bind(this));
        },

        showEditMatch: function(match) {
            if (this.inMatchRequest)
                return;
            this.inMatchRequest = true;
            this.matchEditing = match;
            var score_breakdown = this.pendingMatches.filter(function(m) {
                return m._fms_id == match.id;
            })[0].score_breakdown;
            sendApiRequest('/api/matches/extra?id=' + this.matchEditing.id + '&level=' + this.matchLevel, this.selectedEvent)
            .then(function(raw) {
                this.inEditMatch = true;
                var data = JSON.parse(raw);
                this.matchEditData = {
                    teams: {},
                    flags: {},
                    text: {},
                };
                this.matchEditOverrideCode = Boolean(data.match_code_override);
                ['blue', 'red'].forEach(function(color) {
                    this.matchEditData.teams[color] = this.matchEditing.teams[color].map(function(team) {
                        return {
                            team: team,
                            dq: data[color].dqs.indexOf('frc' + team) != -1,
                            surrogate: data[color].surrogates.indexOf('frc' + team) != -1,
                        };
                    });
                    if (this.isPlayoff) {
                        this.matchEditData.flags[color] = {
                            dq: data[color].dqs.length > 0,
                        };
                    }
                    else {
                        var editData = {};
                        Object.keys(EXTRA_FIELDS[this.eventYear]).forEach(function(field) {
                            editData[field] = data[color][field];
                        });
                        this.matchEditData.flags[color] = editData;
                        if (this.eventYear == 2018) {
                            this.matchEditData.text[color] = {
                                auto_rp: score_breakdown[color].autoQuestRankingPoint ^ data[color].invert_auto ?
                                    'missed (FMS returned scored)' :
                                    'scored (FMS returned missed)',
                            };
                        }
                    }
                }.bind(this));
                this.$refs.matchEditModal.show();
            }.bind(this))
            .always(function() {
                this.inMatchRequest = false;
            }.bind(this));
        },
        hideEditMatch: function() {
            this.inEditMatch = false;
            this.matchEditing = null;
            this.$refs.matchEditModal.hide();
        },
        saveEditMatch: function() {
            if (this.inMatchRequest)
                return;
            this.matchEditError = '';
            this.inMatchRequest = true;

            var findTeamKeysByFlag = function(color, flag) {
                return this.matchEditData.teams[color].filter(function(t) {
                    return t[flag];
                }).map(function(t) {
                    return 'frc' + t.team;
                });
            }.bind(this);
            var genExtraData = function(color) {
                if (this.isPlayoff) {
                    return Object.assign({
                        dqs: this.matchEditData.flags[color].dq ?
                            this.matchEditData.teams[color].map(function(t) {
                                return 'frc' + t.team;
                            }) :
                            [],
                        surrogates: [],
                    }, EXTRA_FIELDS[this.eventYear]);
                }
                return Object.assign({
                    dqs: findTeamKeysByFlag(color, 'dq'),
                    surrogates: findTeamKeysByFlag(color, 'surrogate'),
                }, this.matchEditData.flags[color]);
            }.bind(this);
            var data = {
                blue: genExtraData('blue'),
                red: genExtraData('red'),
            };
            if (this.matchEditOverrideCode) {
                data.match_code_override = this.matchEditing.code;
            }

            sendApiRequest('/api/matches/extra/save?id=' + this.matchEditing.id + '&level=' + this.matchLevel,
                this.selectedEvent, data)
            .always(function() {
                this.inMatchRequest = false;
            }.bind(this))
            .then(this.hideEditMatch.bind(this))
            .then(this.refetchMatches.bind(this))
            .fail(function(res) {
                this.matchEditError = res.responseText;
            }.bind(this));
        },
        editMatchMarkUploaded: function() {
            this.advSelectedMatch = this.matchEditing.id;
            this.markAdvSelectedMatchUploaded();
            this.hideEditMatch();
        },

        uploadRankings: function() {
            this.rankingsError = '';
            this.inUploadRankings = true;
            $.getJSON('/api/rankings/fetch', function(data) {
                var rankings = ((data && data.qualRanks) || []).map(tba.convertToTBARankings[this.eventYear]);
                if (!rankings || !rankings.length) {
                    this.rankingsError = 'No rankings available from FMS';
                    this.inUploadRankings = false;
                    return;
                }

                sendApiRequest('/api/rankings/upload', this.selectedEvent, {
                    breakdowns: tba.RANKING_NAMES[this.eventYear],
                    rankings: rankings,
                }).fail(function(res) {
                    this.rankingsError = res.responseText;
                }.bind(this)).always(function() {
                    this.inUploadRankings = false;
                }.bind(this));
            }.bind(this)).fail(function(res) {
                this.rankingsError = 'fetch failed: ' + res.responseText;
                this.inUploadRankings = false;
            }.bind(this));
        },

        fetchVideos: function() {
            this.inVideoRequest = true;
            this.videoError = '';
            tbaApiEventRequest(this.selectedEvent, 'matches')
            .always(function() {
                this.inVideoRequest = false;
            }.bind(this))
            .then(function(matches) {
                matches.forEach(function(match) {
                    var key = match.key.split('_')[1];
                    if (match.alliances && match.alliances.blue && match.alliances.blue.score != -1) {
                        var v = this.videos[key] || {};
                        v.tba = match.videos.filter(function(cv) {
                            return cv.type == 'youtube';
                        }).map(function(cv) {
                            return cv.key;
                        })[0] || '';
                        v.current = v.current || v.tba || '';
                        v.uploaded = Boolean(v.uploaded);
                        Vue.set(this.videos, key, v);
                    }
                }.bind(this));
            }.bind(this))
            .fail(function(error) {
                this.videoError = utils.parseErrorJSON(error);
            }.bind(this));
        },
        uploadVideos: function() {
            this.cleanVideoUrls();
            var videos = this.getChangedVideos();
            if (!Object.keys(videos).length) {
                this.videoError = 'No videos have changed; not uploading anything.';
                return;
            }
            var invalidVideos = Object.values(videos).filter(function(value) {
                return !value.match(/^[A-Za-z0-9_-]{11}$/);
            });
            if (invalidVideos.length) {
                this.videoError = 'The following IDs are not valid Youtube video IDs. Please submit only the 11-character video ID.\n' +
                    invalidVideos.join('\n');
                return;
            }

            this.inVideoRequest = true;
            this.videoError = '';
            sendApiRequest('/api/videos/upload', this.selectedEvent, videos)
            .always(function() {
                this.inVideoRequest = false;
            }.bind(this))
            .then(function() {
                Object.keys(videos).forEach(function(key) {
                    this.videos[key].uploaded = true;
                }.bind(this));
                this.fetchVideos();
            }.bind(this))
            .fail(function(error) {
                this.videoError = utils.parseErrorJSON(error);
            }.bind(this));
        },
        getSortedVideos: function() {
            return Object.entries(this.videos).sort(function(a, b) {
                return Number(a[0].replace(/[^\d]/g, '')) - Number(b[0].replace(/[^\d]/g, ''));
            }).filter(function(v) {
                return this.showExistingVideos || (!v[1].uploaded && !v[1].tba);
            }.bind(this));
        },
        getChangedVideos: function() {
            var videos = {};
            Object.entries(this.videos).forEach(function(v) {
                if (v[1].current && utils.cleanYoutubeUrl(v[1].current) != utils.cleanYoutubeUrl(v[1].tba)) {
                    videos[v[0]] = v[1].current;
                }
            });
            return videos;
        },
        cleanVideoUrls: function() {
            Object.values(this.videos).forEach(function(v) {
                v.current = utils.cleanYoutubeUrl(v.current);
            });
        },

        addAward: function() {
            this.awards[this.selectedEvent].push(makeAward());
            this.saveAwards();
        },
        duplicateAward: function(award) {
            var newAward = makeAward();
            newAward.name = award.name;
            this.awards[this.selectedEvent].splice(this.awards[this.selectedEvent].indexOf(award) + 1, 0, newAward);
            this.saveAwards();
        },
        clearAward: function(award) {
            award.name = award.team = award.person = '';
            this.saveAwards();
        },
        deleteAward: function(award) {
            var index = this.awards[this.selectedEvent].indexOf(award);
            if (index >= 0) {
                this.awards[this.selectedEvent].splice(index, 1);
            }
            this.saveAwards();
        },
        fetchAutomaticAwards: function() {
            var cleanedAwards = this.awards[this.selectedEvent].filter(function(award) {
                return ['winner', 'finalist'].indexOf(award.name.toLowerCase().trim()) == -1;
            });
            if (cleanedAwards.length != this.awards[this.selectedEvent].length &&
                    !confirm('This will remove all current Winner/Finalist awards from this event. Continue?')) {
                return;
            }
            this.awards[this.selectedEvent] = cleanedAwards;

            this.inAwardRequest = true;
            tbaApiEventRequest(this.selectedEvent, 'alliances')
            .always(function() {
                this.inAwardRequest = false;
            }.bind(this))
            .then(function(data) {
                var alliances = data || [];
                var winnerAwards = [];
                var finalistAwards = [];
                alliances.forEach(function(alliance) {
                    var status = alliance.status || {};
                    if (status.level == 'f') {
                        var awardName;
                        var awardList;
                        if (status.status == 'won') {
                            awardName = 'Winner';
                            awardList = winnerAwards;
                        }
                        else if (status.status == 'eliminated') {
                            awardName = 'Finalist';
                            awardList = finalistAwards;
                        }
                        if (awardName) {
                            alliance.picks.forEach(function(team) {
                                awardList.push(makeAward({
                                    team: team.replace('frc', ''),
                                    name: awardName,
                                }));
                            });
                        }
                    }
                });
                if (!winnerAwards.length && !finalistAwards.length) {
                    this.awardStatus = 'No winners or finalists were detected. Make sure finals have been uploaded ' +
                        'and the TBA event page is up to date.';
                }
                this.awards[this.selectedEvent] = [].concat(winnerAwards, finalistAwards, this.awards[this.selectedEvent]);
                this.saveAwards();
            }.bind(this))
            .fail(function(error) {
                this.awardStatus = utils.parseErrorJSON(error);
            }.bind(this));
        },
        saveAwards: function() {
            if (typeof this.awards != 'object' || Array.isArray(this.awards)) {
                throw new TypeError('awards is not a map');
            }
            if (tba.isValidEventCode(this.selectedEvent) && !Array.isArray(this.awards[this.selectedEvent])) {
                throw new TypeError('awards[' + this.selectedEvent + '] is not an array');
            }
            localStorage.setItem('awards', JSON.stringify(this.awards));
        },
        uploadAwards: function() {
            var json = this.awards[this.selectedEvent].map(function(award) {
                return {
                    name_str: award.name,
                    team_key: award.team ? 'frc' + award.team : null,
                    awardee: award.person || null,
                };
            });
            if (json.filter(function(award) { return !award.name_str; }).length) {
                this.awardStatus = 'One or more awards have an empty name. Please correct this and try again.';
                return;
            }
            this.inAwardRequest = true;
            this.awardStatus = 'Uploading...';
            var request = sendApiRequest('/api/awards/upload', this.selectedEvent, json);
            request.always(function() {
                this.inAwardRequest = false;
            }.bind(this));
            request.then(function() {
                this.awardStatus = 'Upload succeeded.';
            }.bind(this));
            request.fail(function(res) {
                this.awardStatus = 'Error: ' + res.responseText;
            }.bind(this));
        },
    },
};
</script>
