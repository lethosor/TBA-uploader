#!/usr/bin/env node

// usage: generate-rankings-csv.js FMS_RANKINGS_FILENAME.json [YEAR]

const fs = require('fs');

const tba = require('../web/src/tba.js');

const rankings_path = process.argv[2];
const year = Number(process.argv[3]) || (new Date().getYear() + 1900);

const rankings_raw = JSON.parse(fs.readFileSync(rankings_path));
const rankings = rankings_raw.qualRanks.map(tba.convertToTBARankings[year]);

const columns = [
  ['Rank', 'rank'],
  ['Team', 'team_key'],
  ['Played', 'played'],
  ['Wins', 'wins'],
  ['Losses', 'losses'],
  ['Ties', 'ties'],
];
for (let item of tba.RANKING_NAMES[year]) {
  columns.push([item, item]);
}

console.log(columns.map(c => c[0]).join(','));
for (let row of rankings) {
  row.team_key = row.team_key.replace(/frc/, '');
  console.log(columns.map(c => row[c[1]]).join(','));
}

