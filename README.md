# TBA-uploader

This is a tool to make uploading certain FRC off-season event data to [The Blue
Alliance](https://www.thebluealliance.com/) easier. It currently supports
uploading:

* Match scores, including full score breakdowns
* Qualification rankings
* Awards

This tool is intended to be used alongside the [TBA event
wizard](https://www.thebluealliance.com/eventwizard). The features provided by
this tool are either easier to use than the event wizard (match scores and
rankings) or missing from the event wizard altogether (awards).

## Installation

If you're accessing this help page on localhost:8808, you can skip this section
and "running" (although you can also double-check to make sure you have the
latest release).

1. Download the latest release from
   [GitHub](https://github.com/lethosor/TBA-uploader/releases). On Windows, use
   the "win64" (64-bit) version unless it doesn't work.
2. Extract the zip file anywhere you like. The Downloads folder should generally
   work. Note that this tool creates files, so you need to install it somewhere
   where it is able to write files.

## Running

Simply double-click the TBA-uploader executable to start it. On most systems,
this should open a command prompt of some kind. Be sure to keep this window open
(although you can minimize it), because closing it will quit TBA-uploader.

## Usage

You need to [obtain a TBA write
key](https://www.thebluealliance.com/request/apiwrite) for your event. If you
are unable to obtain one, it should still be possible to use TBA-uploader to
save match results (see "Backups" below) and upload these at a later time.
Once you have a write key, you will need to set up the event in TBA-uploader.

There are several tabs available once you have set up and selected an event:

### Event setup

The "Event setup" tab allows you to create or select an event. The event code
must match the code used on TBA exactly (e.g.
"[2018marc](https://www.thebluealliance.com/event/2018marc)").

### Match play

**Note:** you should always make sure that the most recent match schedule has
been imported in the TBA event wizard. This is important when a new round of
matches (qualifications, quarterfinals, etc.) starts. For example, before
starting semifinals, import a new schedule in the event wizard.

This tab is only visible when an event is selected. Ensure that the correct
competition level (qualification or playoff) is selected. Then, after scores are
posted to the audience display for each match, click "fetch new matches". An
overview of the match score should appear. Click "upload scores" to upload this
score to TBA. "Upload all" can be safely clicked at any time, when it is
visible.

*Before* clicking "upload all", if red cards were issued in a match, or there
were surrogate teams, click on the match scores to enter them. You can also
override the "auto quest" ranking point in this screen by toggling the "Invert
Auto RP" checkbox if it was misdetected (see limitations below).

If you see all 0s in a score, you probably fetched the scores before they were
posted in FMS (see limitations below). You should see a warning in this case,
but if not, click "Re-fetch scores" after scores have been posted in FMS. Do not
upload scores before doing this. If you accidentally upload scores before making
the necessary changes, see "advanced options" below.

Click "upload rankings" at any time to upload qualification rankings. This
button is only available when "qualification" is selected.

If you forget to fetch a match before the next match starts, avoid clicking
"fetch new match(es)" before scores for the next match have been posted. After
scores for the next match have been posted, any un-uploaded matches will be
fetched when you click "fetch matches".

### Awards

This tab is only visible when an event is selected. Each award can have an
associated team, person, or both.

The "duplicate award" button will create a new award with the same name as the
chosen award. The "upload awards" button will overwrite all awards online for
the event, so be sure to enter all awards on the same device.

Note that TBA restricts the names of awards to a predefined list. A crude list
of allowed keywords can be found
[here](https://github.com/the-blue-alliance/the-blue-alliance/blob/master/helpers/award_helper.py)
(under `AWARD_MATCHING_STRINGS`, all keywords are in quotes).

### Options

This tab allows changing the backup location and the FMS address. It is not very
user-friendly and should be avoided if at all possible.

### Help

This tab displays this help page.

## Backups

TBA-uploader will back up all of its data to the `fms_data` folder in the same
folder as itself (e.g. if the TBA-uploader executable is in Downloads, the
backup folder will be Downloads/fms_data). This will be created automatically.

It is recommended to make a backup copy of this folder periodically (or at least
when the event concludes), in case manual modifications need to be made later to
posted match results (for example, surrogates or DQs). A backup can also help
identify and fix any bugs in TBA-uploader.

If needed, match-related files can be found in `fms_data/EVENT/levelX/matches`,
where `EVENT` is the event code and `X` is the competition level (2 for
qualifications and 3 for playoffs). The filename format is
``match-play.extension`` (``play`` is 1 except in the case of replays or aborted
matches).

## Known issues and limitations
* If you click the "fetch matches" button before scores have been committed,
  TBA-uploader may fetch a score of 0-0. Avoid doing this - always wait until
  scores have been committed before clicking any "fetch" buttons. If this
  happens, a warning should be displayed, and clicking "Re-fetch scores" after
  posting scores in FMS should resolve the issue.
* 2018: the "auto quest" ranking point cannot be reliably determined due to FMS
  limitations. Sometimes alliances will be credited with this ranking point when
  they didn't actually earn it. Note that this only applies to match scores -
  rankings are unaffected. This can be toggled in the match edit dialog by
  clicking on the match results before uploading scores.
* Editing properties of a specific play of a match that has already been
  uploaded to TBA is rather convoluted. See "advanced options" below. Note that
  a *replay* of a match for any reason counts as a separate play in FMS and can
  be treated normally like any other match (TBA-uploader will fetch and upload
  it separately, overwriting the earlier play(s) if necessary).

## Advanced options

These options allow some manual management of previously-uploaded matches. They
aren't easy to use, so use them only if you need to recover from a mistake with
already-uploaded scores.

Note that these all apply only to the current competition level
(qualification/playoff). Make sure you have selected the right one!

Individual matches can be adjusted by entering the raw match ID (in the form
``match number-play number``, or just ``match number`` for the first play of a
match) and clicking one of the yellow buttons. "Purge" will delete most files in
fms_data related to the match, which will cause it to be fetched again and
uploaded in the next batch of uploads. "Mark as uploaded" will partially undo
"Purge" - it will prevent the match from being uploaded in the next batch of
uploads.

One common use case: if a match play was uploaded with errors (i.e. the auton
RP, red cards, or surrogates were wrong, or the score was all 0s), this can be
changed with the following procedure:

1. Enter the match ID in the "Match ID" box
2. Click "Purge", then "OK" in the resulting dialog
3. Click on the match under "Matches to upload"
4. Make the necessary modifications and save them
5. Click "upload scores" again (making sure that all displayed matches are still
   correct)

The "Purge and re-fetch all matches" button purges all matches. This will likely
trigger a *lot* of notifications for anyone with TBA notifications enabled, so
avoid it.
