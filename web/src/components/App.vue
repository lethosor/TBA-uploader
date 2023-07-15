<template>
    <div>
        <div class="float-right">{{ version }}</div>
        <h2>
            TBA uploader<span v-if="eventSelected"> -
                <a
                    :href="fmsConfig.tba_url + '/event/' + selectedEvent"
                    target="_blank"
                >{{ selectedEvent }}</a>
            </span>
        </h2>
        <b-card
            v-if="uiOptions.showFieldState"
            no-body
            :class="'text-center text-light mb-1 ' + fieldStateClass"
        >
            {{ fieldStateMessage }}
        </b-card>

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
                    <div
                        v-if="tbaEventData.name"
                        class="mt-2"
                    >
                        Event name: {{ tbaEventData.name }} ({{ tbaEventData.year }})
                    </div>
                    <hr>
                    <alert
                        v-model="tbaReadError"
                        variant="danger"
                    />
                    <label class="row col">
                        Read API key (<a
                            :href="fmsConfig.tba_url + '/account'"
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
                            variant="danger"
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
                                :disabled="!canChangePlayoffType()[0]"
                                :title="canChangePlayoffType()[1]"
                                @click="updatePlayoffType"
                            >
                                Change playoff type
                            </b-button>
                        </div>
                        <b-alert
                            v-if="selectedEvent && !isPlayoffTypeSupported(eventPlayoffType)"
                            variant="danger"
                            class="mt-2"
                            show
                        >
                            This playoff type is not supported. Schedule and match uploads will fail.
                        </b-alert>

                        <h3 class="mt-2">Extra Ranking Points</h3>
                        <form
                            v-for="(_, i) in eventExtras[selectedEvent].enabled_extra_rps"
                            class="form-inline"
                        >
                            <b-form-checkbox v-model="eventExtras[selectedEvent].enabled_extra_rps[i]">Enable extra RP {{ i + 1 }}</b-form-checkbox>
                        </form>
                    </div>
                </div>
            </b-tab>

            <b-tab
                v-if="eventSelected"
                title="Teams"
            >
                <b-row class="mb-2">
                    <b-col
                        class="my-auto text-center"
                        sm="auto"
                    >
                        <b-button
                            variant="success"
                            :disabled="inTeamsRequest || isMatchRunning"
                            @click="fetchTeamsReport"
                        >
                            Fetch FMS report
                        </b-button>
                    </b-col>
                    <b-col
                        class="my-auto text-center"
                        sm="auto"
                    >
                        <i>-or-</i>
                    </b-col>
                    <b-col class="my-auto">
                        <dropzone
                            ref="teamListUploadDropzone"
                            title="Upload a team list"
                            accept="text/csv"
                            @upload="onTeamListUpload"
                        />
                    </b-col>
                </b-row>

                <alert
                    v-model="teamListError"
                    variant="danger"
                />

                <b-row
                    v-if="teamList.length"
                    class="mb-2"
                >
                    <b-col>{{ teamList.length }} teams</b-col>
                    <b-col sm="auto">
                        <b-button
                            variant="success"
                            :disabled="inTeamsRequest"
                            @click="uploadTeamList"
                        >
                            Upload to TBA
                        </b-button>

                        <b-button
                            variant="danger"
                            @click="resetTeamList"
                        >
                            Cancel
                        </b-button>
                    </b-col>
                </b-row>

                <b-table
                    striped
                    :items="teamListTable"
                />
            </b-tab>

            <b-tab
                v-if="eventSelected"
                title="Schedule"
            >
                <alert
                    v-model="scheduleError"
                    variant="danger"
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
                <b-row class="mb-2">
                    <b-col
                        class="my-auto text-center"
                        sm="auto"
                    >
                        <b-form-select
                            v-model="selectedFMSScheduleType"
                            class="mb-1"
                        >
                            <option
                                v-for="option of SCHEDULE_FMS_OPTIONS"
                                :key="option.type"
                                :value="option.type"
                            >
                                {{ option.name }}
                            </option>
                        </b-form-select>
                        <br>
                        <b-button
                            variant="success"
                            :disabled="inScheduleRequest || isMatchRunning"
                            @click="fetchScheduleReport"
                        >
                            Fetch FMS report
                        </b-button>
                    </b-col>
                    <b-col
                        class="my-auto text-center"
                        sm="auto"
                    >
                        <i>-or-</i>
                    </b-col>
                    <b-col class="my-auto">
                        <dropzone
                            ref="scheduleUpload"
                            title="Upload a schedule"
                            accept="text/csv"
                            @upload="onScheduleUpload"
                        />
                    </b-col>
                </b-row>
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
                            variant="danger"
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
                        No new matches detected; nothing to upload.
                    </b-alert>
                </div>
            </b-tab>

            <b-tab
                v-if="eventSelected"
                ref="matchPlayTab"
                title="Match play"
            >
                <div class="form-inline">
                    <b-form-select
                        v-model="matchLevel"
                    >
                        <option
                            v-if="uiOptions.showAllLevels"
                            :value="consts.MATCH_LEVEL.MANUAL"
                        >
                            Manual
                        </option>
                        <option
                            v-if="uiOptions.showAllLevels"
                            :value="consts.MATCH_LEVEL.TEST"
                        >
                            Test
                        </option>
                        <option
                            v-if="uiOptions.showAllLevels"
                            :value="consts.MATCH_LEVEL.PRACTICE"
                        >
                            Practice
                        </option>
                        <option :value="consts.MATCH_LEVEL.QUAL">Qualification</option>
                        <option :value="consts.MATCH_LEVEL.PLAYOFF">Playoff</option>
                    </b-form-select>
                    <b-button
                        v-if="matchLevel == consts.MATCH_LEVEL.MANUAL"
                        variant="success"
                        data-accesskey="c"
                        :disabled="inMatchRequest"
                        @click="createManualMatch"
                    >
                        Create new match
                    </b-button>
                    <b-button
                        variant="success"
                        data-accesskey="f"
                        :disabled="inMatchRequest || isMatchRunning"
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
                        :disabled="inUploadRankings || isMatchRunning"
                        @click="uploadRankings"
                    >
                        <fragment v-if="anyEnabledExtraRps">Generate and </fragment>Upload rankings
                    </b-button>
                </div>
                <p>
                    <span class="warning">Warning:</span> do not click any buttons on this page while a match is running.
                    Be sure to only fetch (or re-fetch) matches <strong>after</strong> scores have been posted in FMS.
                    <span v-if="isQual">Rankings can be updated at any time if necessary, but will also be updated after posting scores.</span>
                </p>
                <div v-if="isQual || isPlayoff">
                    <b-form-checkbox
                        v-model="autoUploadMatches"
                        name="check-button"
                        switch
                    >
                        Auto upload
                    </b-form-checkbox>
                </div>
                <alert
                    v-model="rankingsError"
                    variant="danger"
                    prefix="Rankings:"
                />
                <alert
                    v-model="rankingsGeneratedMessageHtml"
                    allow-html
                    variant="success"
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
                    <div class="mb-2">
                        <b-button
                            variant="danger"
                            :disabled="inMatchRequest || isMatchRunning"
                            @click="fetchMatches(true)"
                        >
                            Purge and re-fetch all matches
                        </b-button>
                        <b-button
                            variant="danger"
                            :disabled="inMatchRequest || isMatchRunning || !matchSummaries.length"
                            @click="markAdvPendingMatchesUploaded"
                        >
                            Mark all pending matches uploaded
                        </b-button>
                    </div>

                    <h4>Rankings Upload</h4>

                    <b-row class="mb-2">
                        <b-col md="auto">
                            <b-button
                                class="mr-2"
                                :variant="anyEnabledExtraRps ? 'danger' : 'info'"
                                :disabled="inUploadRankings || isMatchRunning"
                                @click="uploadRankingsFromFMS"
                            >
                                Upload rankings (pit display)
                            </b-button>
                        </b-col>
                        <b-col>
                            This is the normal ranking upload flow - unlikely to work unless the active tournament level in FMS is "Qualification".
                            <strong
                                v-if="anyEnabledExtraRps"
                                class="warning"
                            >This is inaccurate with extra ranking points enabled.</strong>
                        </b-col>
                    </b-row>

                    <b-row class="mb-2">
                        <b-col md="auto">
                            <b-button
                                class="mr-2"
                                :variant="anyEnabledExtraRps ? 'info' : 'warning'"
                                :disabled="inUploadRankings || isMatchRunning"
                                @click="uploadRankingsFromTBA"
                            >
                                Upload rankings (generate from TBA match results)
                            </b-button>
                        </b-col>
                        <b-col>
                            This generates rankings from match results on TBA, which may be delayed by up to a minute for active events.
                        </b-col>
                    </b-row>

                    <b-row class="mb-2">
                        <b-col md="auto">
                            <b-button
                                variant="success"
                                :disabled="inUploadRankings"
                                @click="generateRankingsReportFromTBA"
                            >
                                Generate rankings report from TBA
                            </b-button>
                        </b-col>
                        <b-col>
                            This allows you to inspect rankings before uploading them.
                        </b-col>
                    </b-row>

                    <dropzone
                        ref="rankingsUploadDropzone"
                        title="Upload a rankings report (not yet tested with a complete report)"
                        accept="text/csv"
                        @upload="onRankingsReportUpload"
                    />

                    <b-table
                        striped
                        :items="rankingsReportTable"
                    />

                    <div v-if="rankingsReportData.length">
                        <b-button
                            variant="success"
                            :disabled="inUploadRankings"
                            @click="uploadRankingsReport(); rankingsGeneratedMessageHtml=''"
                        >
                            Upload rankings report
                        </b-button>
                        <b-button
                            variant="danger"
                            @click="resetRankingsReport"
                        >
                            Cancel
                        </b-button>
                    </div>

                    <alert
                        v-model="rankingsError"
                        variant="danger"
                        prefix="Rankings:"
                    />

                    <hr>
                </div>
                <alert
                    v-model="matchError"
                    variant="danger"
                />
                <alert
                    v-model="matchMessage"
                    variant="success"
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
                        :disabled="inMatchRequest || isMatchRunning"
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
                    variant="danger"
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
                title="Alliances"
            >
                <div class="form-inline mb-2">
                    <b-button
                        variant="success"
                        :disabled="inAllianceRequest || isMatchRunning"
                        @click="fetchAllianceReport"
                    >
                        Fetch from FMS
                    </b-button>
                </div>
                <div class="form-inline mb-2">
                    <label># Alliances:
                        <b-form-select
                            v-model="eventExtras[selectedEvent].alliance_count"
                            class="ml-1 mr-4"
                        >
                            <option
                                v-for="option in 20"
                                :key="option"
                                :value="option"
                            >
                                {{ option }}
                            </option>
                        </b-form-select>
                    </label>
                    <label># Teams per alliance:
                        <b-form-select
                            v-model="eventExtras[selectedEvent].alliance_size"
                            class="ml-1 mr-4"
                        >
                            <option
                                v-for="option in [3, 4]"
                                :key="option"
                                :value="option"
                            >
                                {{ option }}
                            </option>
                        </b-form-select>
                    </label>
                    <div class="ml-auto">
                        <b-button
                            variant="warning"
                            :disabled="inAllianceRequest"
                            @click="fetchAlliances"
                        >
                            Fetch from TBA
                        </b-button>
                        <b-button
                            variant="danger"
                            :disabled="inAllianceRequest"
                            @click="clearAlliances"
                        >
                            Clear
                        </b-button>
                    </div>
                </div>
                <div class="form-inline">
                    Tab order:
                    <b-form-group class="ml-2">
                        <b-form-radio
                            v-model="alliancesFmsTabOrder"
                            inline
                            name="alliance-tab-order"
                            :value="true"
                        >
                            FMS
                        </b-form-radio>
                        <b-form-radio
                            v-model="alliancesFmsTabOrder"
                            inline
                            name="alliance-tab-order"
                            :value="false"
                        >
                            Table
                        </b-form-radio>
                    </b-form-group>
                </div>

                <alert
                    v-model="allianceError"
                    variant="danger"
                />

                <alliance-table
                    :alliance-count="eventExtras[selectedEvent].alliance_count"
                    :alliance-size="eventExtras[selectedEvent].alliance_size"
                    :value="alliances[selectedEvent]"
                    :fms-tab-order="alliancesFmsTabOrder"
                    @input="onAllianceChange"
                />

                <b-button
                    variant="warning"
                    :disabled="inAllianceRequest"
                    @click="postAlliances"
                >
                    Post to TBA
                </b-button>
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
                <div class="row">
                    <div class="col-sm-12">
                        <b-form-checkbox v-model="uiOptions.showFieldState">
                            Show field status (requires integration)
                        </b-form-checkbox>
                    </div>
                </div>
                <div class="row">
                    <div class="col-sm-12">
                        <b-form-checkbox v-model="uiOptions.useProxy">
                            Use server-side proxy for all requests
                        </b-form-checkbox>
                    </div>
                </div>
                <div class="row mb-2 form-inline">
                    <b-button
                        variant="danger"
                        @click="lastFieldState=null"
                    >
                        Reset field status
                    </b-button> if buttons are disabled due to status being stuck in a match
                </div>
                <hr>
                <h2>FMS options</h2>
                <p>Options in this section need to be saved by clicking "Save" below. Also note that these can be specified on the command line as well, which is more useful for development.</p>
                <div class="row mb-2">
                    <label class="col-sm-12 col-md-8">
                        FMS URL (default: <code>http://10.0.100.5</code>):
                        <b-form-input v-model="fmsConfig.fms_url" />
                    </label>
                    <label class="col-sm-12 col-md-8">
                        Data folder:
                        <b-form-input v-model="fmsConfig.data_folder" />
                    </label>
                    <label class="col-sm-12 col-md-8">
                        TBA URL (default: <code>https://www.thebluealliance.com</code>):
                        <b-form-input v-model="fmsConfig.tba_url" />
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
                    variant="danger"
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
                    variant="danger"
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

                    <tbody v-if="anyEnabledExtraRps">
                        <tr>
                            <td
                                v-for="color in ['red', 'blue']"
                                :key="color"
                                :class="color"
                            >
                                <b-form-checkbox
                                    v-for="(is_enabled, i) in enabledExtraRps"
                                    v-if="is_enabled"
                                    :key="i"
                                    v-model="matchEditData.extra_rps[color][i]"
                                >
                                    Extra RP {{ i + 1 }}
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
    BCard,
    BCol,
    BFormCheckbox,
    BFormGroup,
    BFormInput,
    BFormRadio,
    BFormSelect,
    BModal,
    BRow,
    BTab,
    BTable,
    BTabs,
} from 'bootstrap-vue';
import 'regenerator-runtime';
import showdown from 'showdown';
import Vue from 'vue';

import api from 'src/api.js';
import {
    BRACKET_NAME,
    BRACKET_TYPE,
    FIELD_STATE,
    MATCH_LEVEL,
} from 'src/consts.js';
import Reports from 'src/reports.js';
import Schedule from 'src/schedule.js';
import SocketConnection from 'src/SocketConnection.js';
import tba from 'src/tba.js';
import utils from 'src/utils.js';

import Alert from 'components/Alert.vue';
import AllianceTable from 'components/AllianceTable.vue';
import Dropzone from 'components/Dropzone.vue';
import ScoreSummary from 'components/ScoreSummary.vue';

import 'bootstrap/dist/css/bootstrap.css';
import 'bootstrap-vue/dist/bootstrap-vue.css';
import 'src/app.css';

const STORED_EVENTS = utils.safeParseLocalStorageObject('storedEvents');
const STORED_ALLIANCES = utils.safeParseLocalStorageObject('alliances');
const STORED_AWARDS = utils.safeParseLocalStorageObject('awards');

const DEFAULT_ENABLED_EXTRA_RPS = Object.freeze([false, false]);

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

function tbaApiEventRequest(event, route, useProxy) {
    var url = FMS_CONFIG.tba_url + '/api/v3/event/' + event;
    if (route) {
        url += '/' + route;
    }
    return utils.makeProxiedAjaxRequest({
        type: 'GET',
        url: url,
        headers: {
            'X-TBA-Auth-Key': localStorage.getItem('readApiKey'),
        },
        cache: false,
    }, useProxy ? '/api/proxy' : undefined);
}
window.tbaApiEventRequest = tbaApiEventRequest;

function fetchReport(report_type) {
    return $.getJSON('/api/report/fetch', {report_type});
}

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
    2022: {},
    2023: {},
};

export default {
    components: {
        Alert,
        AllianceTable,
        BAlert,
        BButton,
        BButtonClose,
        BCard,
        BCol,
        BFormCheckbox,
        BFormGroup,
        BFormInput,
        BFormRadio,
        BFormSelect,
        BModal,
        BRow,
        BTab,
        BTable,
        BTabs,
        Dropzone,
        ScoreSummary,
    },
    data: () => ({
        version: window.VERSION || 'missing version',
        consts: {MATCH_LEVEL},
        helpHTML: '',
        fmsConfig: window.FMS_CONFIG || {},
        fmsConfigError: '',
        selectedTab: utils.safeParseLocalStorageInteger('lastTab', 0),

        sock: SocketConnection(),
        lastFieldState: null,

        events: Object.keys(STORED_EVENTS).sort(),
        selectedEvent: localStorage.getItem('selectedEvent') || '',
        addEventUI: makeAddEventUI(),
        readApiKey: localStorage.getItem('readApiKey') || '',
        tbaEventData: {},
        tbaReadError: '',

        uiOptions: $.extend({
            showAllLevels: false,
            showFieldState: true,
            useProxy: true,
        }, utils.safeParseLocalStorageObject('uiOptions')),
        eventExtras: utils.safeParseLocalStorageObject('eventExtras'),
        remapError: '',

        inTeamsRequest: false,
        teamListTable: [],
        teamList: [],
        teamListError: '',

        inScheduleRequest: false,
        scheduleUploaded: false,
        scheduleError: '',
        scheduleStats: [],
        schedulePendingMatches: [],
        selectedFMSScheduleType: Reports.Type.QUAL_SCHEDULE,
        SCHEDULE_FMS_OPTIONS: [
            {name: 'Quals', type: Reports.Type.QUAL_SCHEDULE},
            {name: 'Playoffs', type: Reports.Type.PLAYOFF_SCHEDULE},
        ],

        matchLevel: utils.safeParseLocalStorageInteger('matchLevel', MATCH_LEVEL.QUAL),
        showAllLevels: false,
        inMatchRequest: false,
        matchError: '',
        matchMessage: '',
        // pendingMatches: [], // not set yet to avoid Vue binding to this
        matchSummaries: [],
        fetchedScorelessMatches: false,
        unhandledBreakdowns: [],
        inMatchAdvanced: false,
        advSelectedMatch: '',
        advMatchError: '',
        autoUploadMatches: false,

        inEditMatch: false,
        matchEditing: null,
        matchEditData: null,
        matchEditError: '',
        matchEditOverrideCode: false,

        inUploadRankings: false,
        rankingsError: '',
        rankingsReportData: [],
        rankingsReportTable: [],
        rankingsGeneratedMessageHtml: '',

        videos: {},
        inVideoRequest: false,
        videoError: '',
        showExistingVideos: false,

        alliances: STORED_ALLIANCES,
        alliancesFmsTabOrder: true,
        inAllianceRequest: false,
        allianceError: '',

        awards: STORED_AWARDS,
        awardStatus: '',
        inAwardRequest: false,
    }),
    computed: {
        BRACKET_TYPES: function() {
            return Object.keys(BRACKET_TYPE).map((key) => ({
                value: BRACKET_TYPE[key],
                text: BRACKET_NAME[key],
            })).sort((a, b) => a.text.localeCompare(b.text));
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
        eventPlayoffType: function() {
            const playoff_type = this.eventExtras[this.selectedEvent].playoff_type;
            return Number.isFinite(playoff_type) ? playoff_type : BRACKET_TYPE.BRACKET_8_TEAM;
        },
        enabledExtraRps: function() {
            return this.eventExtras[this.selectedEvent].enabled_extra_rps || DEFAULT_ENABLED_EXTRA_RPS;
        },
        anyEnabledExtraRps: function() {
            return this.enabledExtraRps.find(Boolean);
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
        fieldStateClass() {
            if (this.lastFieldState) {
                if (this.lastFieldState == FIELD_STATE.MatchCancelled) {
                    return 'bg-danger';
                }
                return (utils.isFieldStateInMatch(this.lastFieldState)) ? 'bg-success' : 'bg-primary';
            }
            return 'bg-secondary';
        },
        fieldStateMessage() {
            return this.lastFieldState ? utils.describeFieldState(this.lastFieldState) : 'No Field Status Available';
        },
        isMatchRunning() {
            return this.lastFieldState && utils.isFieldStateInMatch(this.lastFieldState);
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

        this.sock.setUrl('ws://' + location.host + '/ws/state/subscribe');
        this.sock.on('message', (event) => {
            const data = JSON.parse(event.data);
            if (data.field_state !== undefined) {
                this.onFieldStateUpdate(data.field_state);
            }
        });

        $.get('/README.md', function(readme) {
            // remove first line (header)
            readme = readme.substr(readme.indexOf('\n'));
            this.helpHTML = new showdown.Converter({
                simplifiedAutoLink: true,
            }).makeHtml(readme);
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
        tbaApiCurrentEventRequest: function(route) {
            return tbaApiEventRequest(this.selectedEvent, route, this.uiOptions.useProxy);
        },

        onFieldStateUpdate: async function(fieldState) {
            if (fieldState != this.lastFieldState) {
                this.handleMatchesFromFieldStateChange(fieldState);
            }
            this.lastFieldState = fieldState;
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
            this.tbaApiCurrentEventRequest().then(function(data) {
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
                playoff_type: null,
                alliance_count: 8,
                alliance_size: 3,
                enabled_extra_rps: DEFAULT_ENABLED_EXTRA_RPS.slice(),
            }, this.eventExtras[event]));

            if (!this.alliances[event]) {
                this.$set(this.alliances, event, []);
                this.saveAlliances();
            }

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
        _mapTeamKey: function(teamKey, idField, outField) {
            let prefix = '';
            if (teamKey.startsWith('frc')) {
                prefix = 'frc';
                teamKey = teamKey.replace(/^frc/, '');
            }
            for (const mapping of this.eventExtras[this.selectedEvent].remap_teams) {
                if (mapping[idField] == teamKey && mapping[outField]) {
                    return prefix + mapping[outField];
                }
            }
            return prefix + teamKey;
        },
        mapTeamFMStoTBA: function(teamKey) {
            return this._mapTeamKey(teamKey, 'fms', 'tba');
        },
        mapTeamTBAtoFMS: function(teamKey) {
            return this._mapTeamKey(teamKey, 'tba', 'fms');
        },
        convertMatchTeamKeysTBAtoFMS: function(matches) {
            for (let match of matches) {
                for (const alliance of ['red', 'blue']) {
                    for (const field of ['team_keys', 'surrogate_team_keys', 'dq_team_keys']) {
                        match.alliances[alliance][field] = match.alliances[alliance][field].map(this.mapTeamTBAtoFMS.bind(this));
                    }
                }
            }
            return matches;
        },

        canChangePlayoffType: function() {
            if (this.eventExtras[this.selectedEvent].playoff_type === undefined) {
                return [false, 'No event is loaded'];
            }
            if (this.eventExtras[this.selectedEvent].playoff_type === this.tbaEventData.playoff_type) {
                return [false, 'Playoff type is already set to this'];
            }
            return [true];
        },

        isPlayoffTypeSupported: function(type) {
            return Boolean(BRACKETS[type]);
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

        fetchTeamsReport: async function() {
            this.inTeamsRequest = true;
            try {
                const response = await fetchReport(Reports.Type.TEAM_LIST);
                const cells = Reports.convertToCells(response);
                this.processTeamList(cells);
            }
            catch (e) {
                this.teamListError = utils.parseErrorText(e);
            }
            finally {
                this.inTeamsRequest = false;
            }
        },

        onTeamListUpload: function(event) {
            try {
                this.processTeamList(utils.parseCSVRaw(event.body));
            }
            catch (e) {
                this.teamListError = utils.parseErrorText(e);
            }
        },

        processTeamList: function(cells) {
            const headerRowIndex = cells.findIndex(row => row.includes('#'));
            if (!cells[headerRowIndex]) {
                throw 'could not find header row containing "#" header';
            }
            this.teamListTable = utils.parseCSVObjects(cells, headerRowIndex).map(team => ({
                Team: team['#'],
                Name: team['Short Name'],
                Location: team['Location'],
            })).filter(team => Boolean(team.Team) && !isNaN(Number(team.Team)));
            this.teamList = this.teamListTable.map(team => Number(team.Team));
        },

        uploadTeamList: async function() {
            this.inTeamsRequest = true;
            try {
                await sendApiRequest('/api/teams/upload', this.selectedEvent, this.teamList.map(team => 'frc' + team));
                this.resetTeamList();
            }
            catch (e) {
                this.teamListError = utils.parseErrorText(e);
            }
            finally {
                this.inTeamsRequest = false;
            }
        },

        resetTeamList: function() {
            this.teamList = [];
            this.teamListTable = [];
            this.teamListError = '';
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
            this.processSchedule(utils.parseCSVRaw(event.body));
        },
        fetchScheduleReport: async function() {
            this.scheduleReset(false);
            this.inScheduleRequest = true;
            try {
                let response = await fetchReport(this.selectedFMSScheduleType);
                let cells = Reports.convertToCells(response);
                // add missing header
                cells.unshift(['match schedule']);
                this.processSchedule(cells);
            }
            catch (e) {
                this.scheduleError = utils.parseErrorText(e);
            }
            finally {
                this.inScheduleRequest = false;
            }
        },
        processSchedule: async function(cells) {
            try {
                var schedule = Schedule.parse(cells, this.eventPlayoffType);
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

            const progressIndex = this.scheduleStats.push('Checking against TBA schedule...');
            this.inScheduleRequest = true;

            let tbaMatches = [];
            try {
                tbaMatches = (await this.tbaApiCurrentEventRequest('matches')) || [];
            }
            catch (error) {
                console.error(error);  // eslint-disable-line no-console
                this.scheduleError = utils.parseErrorJSON(error);
                return;
            }
            finally {
                this.inScheduleRequest = false;
                this.scheduleStats.splice(progressIndex - 1, 1); // remove progress message
            }

            let newLevels = Schedule.findAllCompLevels(schedule);
            const tbaLevels = Schedule.findAllCompLevels(tbaMatches);
            this.scheduleStats.push('TBA has level(s): ' + tbaLevels.join(', '));
            this.scheduleStats.push('The FMS report has level(s): ' + newLevels.join(', '));
            const tbaMatchKeys = tbaMatches.map(Schedule.getTBAMatchKey);

            newLevels = newLevels.filter(function(level) {
                return tbaLevels.indexOf(level) < 0;
            });
            if (!newLevels.length) {
                this.scheduleStats.push('No new levels are present in the FMS report.');
            }
            else {
                this.scheduleStats.push('Level(s) to be added from the FMS report: ' + newLevels.join(', '));
            }

            this.schedulePendingMatches = schedule.filter(function(match) {
                return !tbaMatchKeys.includes(match._key);
            });
            this.scheduleStats.push(this.schedulePendingMatches.length + ' match(es) will be added');
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

        fetchMatches: async function(all) {
            if (all && !confirmPurge()) {
                return;
            }
            this.inMatchRequest = true;
            this.fetchedScorelessMatches = false;
            this.matchError = '';
            try {
                let data = await $.get('/api/matches/fetch', {
                    event: this.selectedEvent,
                    level: this.matchLevel,
                    playoff_type: this.eventPlayoffType,
                    enabled_extra_rps: this.enabledExtraRps.join(','),
                    all: all ? '1' : '',
                });
                this.pendingMatches = JSON.parse(data);
                this.pendingMatches.sort(function(a, b) {
                    return Number(a._fms_id.split('-')[0]) - Number(b._fms_id.split('-')[0]);
                });
                this.matchSummaries = this.generateMatchSummaries(this.pendingMatches);
                this.fetchedScorelessMatches = this.checkScorelessMatches(this.pendingMatches);
                this.unhandledBreakdowns = this.findUnhandledBreakdowns(this.pendingMatches);
            }
            catch (e) {
                this.matchError = utils.parseErrorText(e);
            }
            finally {
                this.inMatchRequest = false;
            }
        },
        refetchMatches: async function() {
            if (this.matchLevel == MATCH_LEVEL.MANUAL) {
                this.fetchMatches(false);
                return;
            }

            var match_ids = this.pendingMatches.map(function(match) {
                return match._fms_id;
            });
            this.inMatchRequest = true;
            this.matchError = '';
            try {
                await sendApiRequest('/api/matches/purge?level=' + this.matchLevel, this.selectedEvent, match_ids);
                await this.fetchMatches(false);
            }
            catch (e) {
                this.matchError = 'Purge: ' + utils.parseErrorText(e);
            }
            finally {
                this.inMatchRequest = false;
            }
        },
        handleMatchesFromFieldStateChange: async function(fieldState) {
            if (!this.selectedEvent) {
                return;
            }
            if (!this.$refs.matchPlayTab.tabClasses[0].active) {
                return;
            }

            if (fieldState == FIELD_STATE.WaitingForPostResults) {
                await this.fetchMatches();
            } else if (
                fieldState == FIELD_STATE.TournamentLevelComplete ||
                (fieldState == FIELD_STATE.WaitingForPrestart && this.lastFieldState == FIELD_STATE.WaitingForPostResults)
            ) {
                if (!this.autoUploadMatches) {
                    return;
                }
                if (!(this.isQual || this.isPlayoff)) {
                    // requires manual match code override
                    return;
                }
                if (this.fetchedScorelessMatches) {
                    return;
                }

                if (this.pendingMatches.length) {
                    this.uploadMatches();
                }
            }
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
                    this.selectedEvent, match_ids,
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
        markAdvPendingMatchesUploaded: async function() {
            if (!this.pendingMatches || !this.pendingMatches.length) {
                return;
            }
            const match_ids = this.pendingMatches.map(match => match._fms_id);
            this.inMatchRequest = true;
            this.advMatchError = '';
            try {
                await sendApiRequest('/api/matches/mark_uploaded?level=' + this.matchLevel,
                    this.selectedEvent, match_ids);
            }
            catch (e) {
                this.advMatchError = 'Receipt generation failed: ' + utils.parseErrorText(e);
                return;
            }
            finally {
                this.inMatchRequest = false;
            }
            await this.fetchMatches();
        },
        createManualMatch: async function() {
            this.inMatchRequest = true;
            this.matchError = '';
            this.matchMessage = '';
            try {
                const res = await sendApiRequest('/api/matches/create?level=' + this.matchLevel, this.selectedEvent);
                this.matchMessage = res;
            }
            catch (error) {
                this.matchError = utils.parseErrorText(error);
            }
            finally {
                this.inMatchRequest = false;
            }
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
                    extra_rps: {},
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
                    this.matchEditData.extra_rps[color] = data[color].extra_rps || DEFAULT_ENABLED_EXTRA_RPS.slice();
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
                }, this.matchEditData.flags[color],
                this.anyEnabledExtraRps ? {extra_rps: this.matchEditData.extra_rps[color]} : null);
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

        uploadRankingsFromFMS: function() {
            this.rankingsError = '';
            this.inUploadRankings = true;
            const params = {
                event: this.selectedEvent,
                level: this.matchLevel,
            };
            $.getJSON('/api/rankings/fetch', params, function(data) {
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
        uploadRankingsFromTBA: async function() {
            this.rankingsError = '';
            await this.generateRankingsReportFromTBA();
            if (this.rankingsError) {
                return;
            }
            if (!this.rankingsReportData.length) {
                this.rankingsError = 'No rankings were generated from TBA match results';
                return;
            }
            await this.uploadRankingsReport();
        },
        uploadRankings: async function() {
            if (this.anyEnabledExtraRps) {
                return this.uploadRankingsFromTBA();
            }
            else {
                return this.uploadRankingsFromFMS();
            }
        },

        onRankingsReportUpload: function(event) {
            this.resetRankingsReport();
            try {
                const cells = utils.parseCSVRaw(event.body.toLowerCase());
                const headerRowIndex = cells.findIndex(row => row.includes('team'));
                if (!cells[headerRowIndex]) {
                    throw 'could not find header row containing "Team" header';
                }
                this.rankingsReportData = utils.parseCSVObjects(cells, headerRowIndex)
                    .map(tba.convertToTBARankings[this.eventYear]);
                this.rankingsReportTable = this.rankingsReportData.map(team => ({
                    Team: team.team_key.replace('frc', ''),
                    Rank: team.rank,
                }));
            }
            catch (e) {
                console.error(e);   // eslint-disable-line no-console
                this.rankingsError = utils.parseErrorText(e);
            }
        },
        generateRankingsReportFromTBA: async function() {
            this.resetRankingsReport();
            this.inUploadRankings = true;
            this.rankingsGeneratedMessageHtml = '';
            try {
                const matchResults = await this.tbaApiCurrentEventRequest('matches');
                this.convertMatchTeamKeysTBAtoFMS(matchResults);
                this.rankingsReportData = this.rankingsReportTable = tba.generateRankingsFromMatchResults(matchResults, this.eventYear);
                this.rankingsGeneratedMessageHtml = 'Rankings generated from <strong>' + matchResults.length + '</strong> matches';
            }
            catch (e) {
                console.error(e);   // eslint-disable-line no-console
                this.rankingsError = utils.parseErrorText(e);
            }
            finally {
                this.inUploadRankings = false;
            }
        },
        uploadRankingsReport: async function() {
            if (!this.rankingsReportData.length) {
                this.rankingsError = 'No rankings to upload';
                return;
            }
            this.inUploadRankings = true;
            this.rankingsError = '';
            try {
                await sendApiRequest('/api/rankings/upload', this.selectedEvent, {
                    breakdowns: tba.RANKING_NAMES[this.eventYear],
                    rankings: this.rankingsReportData,
                });
                this.resetRankingsReport();
            }
            catch (e) {
                this.rankingsError = utils.parseErrorText(e);
            }
            finally {
                this.inUploadRankings = false;
            }
        },
        resetRankingsReport: function() {
            this.rankingsReportData = [];
            this.rankingsReportTable = [];
            this.rankingsError = '';
        },

        fetchVideos: function() {
            this.inVideoRequest = true;
            this.videoError = '';
            this.tbaApiCurrentEventRequest('matches')
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

        onAllianceChange: function(newAlliances) {
            this.$set(this.alliances, this.selectedEvent, newAlliances);
            this.saveAlliances();
        },
        fetchAllianceReport: async function() {
            this.inAllianceRequest = true;
            this.allianceError = '';
            try {
                const response = await fetchReport(Reports.Type.PLAYOFF_RANKINGS);
                const cells = Reports.convertToCells(response);
                const headerRowIndex = cells.findIndex(row => row.includes('Alliance'));
                if (!cells[headerRowIndex]) {
                    throw 'could not find header row containing "Alliance" header';
                }
                const rows = utils.parseCSVObjects(cells, headerRowIndex);
                const newAlliances = [];
                for (const row of rows) {
                    if (!row.Teams) {
                        continue;
                    }
                    const allianceNumber = row.Alliance.match(/A(\d+)/)?.[1];
                    if (!allianceNumber) {
                        throw 'invalid alliance number: ' + row.Alliance;
                    }
                    const teams = Array.from(row.Teams.matchAll(/\d+/g)).map(match => match[0]);
                    newAlliances[allianceNumber - 1] = teams;
                }
                this._setAlliances(newAlliances);
            }
            catch (e) {
                this.allianceError = utils.parseErrorJSON(e);
            }
            finally {
                this.inAllianceRequest = false;
            }
        },
        fetchAlliances: async function() {
            this.inAllianceRequest = true;
            this.allianceError = '';
            try {
                const response = await this.tbaApiCurrentEventRequest('alliances');
                const newAlliances = response.map(a => a.picks.map(t => Number(t.replace('frc', ''))));
                this._setAlliances(newAlliances);
            }
            catch (e) {
                this.allianceError = utils.parseErrorJSON(e);
            }
            finally {
                this.inAllianceRequest = false;
            }
        },
        _setAlliances: function(newAlliances) {
            this.eventExtras[this.selectedEvent].alliance_count = newAlliances.length;
            this.eventExtras[this.selectedEvent].alliance_size = Math.max(...newAlliances.map(a => a.length));
            this.$set(this.alliances, this.selectedEvent, newAlliances);
        },
        clearAlliances: function() {
            this.alliances[this.selectedEvent] = [];
        },
        postAlliances: async function() {
            this.saveAlliances();
            this.inAllianceRequest = true;
            this.allianceError = '';
            let payload = this.alliances[this.selectedEvent].map(a => a.map(t => 'frc' + t));
            try {
                await sendApiRequest('/api/alliances/upload', this.selectedEvent, payload);
            }
            catch (e) {
                this.allianceError = utils.parseErrorJSON(e);
            }
            finally {
                this.inAllianceRequest = false;
            }
        },
        saveAlliances: function() {
            if (typeof this.alliances != 'object' || Array.isArray(this.alliances)) {
                throw new TypeError('alliances is not a map');
            }
            if (tba.isValidEventCode(this.selectedEvent) && !Array.isArray(this.alliances[this.selectedEvent])) {
                throw new TypeError('alliances[' + this.selectedEvent + '] is not an array');
            }
            localStorage.setItem('alliances', JSON.stringify(this.alliances));
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
            this.tbaApiCurrentEventRequest('alliances')
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
                        let addAward = function(teamKey) {
                            awardList.push(makeAward({
                                team: teamKey.replace('frc', ''),
                                name: awardName,
                            }));
                        };
                        if (awardName) {
                            alliance.picks.forEach(addAward);
                            if (alliance.backup && alliance.backup.in) {
                                addAward(alliance.backup.in);
                            }
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
