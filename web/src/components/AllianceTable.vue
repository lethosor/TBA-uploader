<template>
    <table class="table">
        <thead>
            <tr>
                <th>Alliance</th>
                <th
                    v-for="i in allianceSize"
                    :key="i"
                >
                    Team {{ i }}
                </th>
            </tr>
        </thead>
        <tbody>
            <tr
                v-for="a in allianceCount"
                :key="a"
            >
                <td>Alliance {{ a }}</td>
                <td
                    v-for="t in allianceSize"
                    :key="t"
                >
                    <b-form-input
                        :value="getTeam(a - 1, t - 1)"
                        :tabindex="getTabIndex(a - 1, t - 1)"
                        :class="getCellClass(a - 1, t - 1)"
                        @change="setTeam(a - 1, t - 1, $event)"
                        @blur="emitValue"
                    />
                </td>
            </tr>
        </tbody>
    </table>
</template>

<script>
import {
    BFormInput,
} from 'bootstrap-vue';

export default {
    name: 'AllianceTable',
    components: {
        BFormInput,
    },
    props: {
        allianceCount: {
            type: Number,
            required: true,
        },
        allianceSize: {
            type: Number,
            required: true,
        },
        value: {
            type: Array,
            default: () => [],
        },
        fmsTabOrder: {
            type: Boolean,
            default: true,
        },
    },
    data: () => ({
        teams: [],
    }),
    computed: {

    },
    watch: {
        value(value) {
            this._importTeams(value);
        },
    },
    mounted() {
        this._importTeams(this.value);
    },
    methods: {
        getTeam(allianceIndex, teamIndex) {
            this._resizeStorage();
            return this.teams[allianceIndex][teamIndex];
        },
        setTeam(allianceIndex, teamIndex, value) {
            this._resizeStorage();
            this.$set(this.teams[allianceIndex], teamIndex, value);
            this.emitValue();
        },
        emitValue() {
            this.$emit('input', this._exportTeams());
        },
        getTabIndex(allianceIndex, teamIndex) {
            const offset = 10;  // if we start at 0, we may overlap with another field
            if (this.fmsTabOrder) {
                if (teamIndex < 2) {
                    return offset + (2 * allianceIndex) + teamIndex;
                }

                const numAlliances = this.teams.length;
                let rowIndex = allianceIndex;
                if (teamIndex == 2) {
                    // reverse order
                    rowIndex = numAlliances - rowIndex - 1;
                }
                return offset + (numAlliances * teamIndex) + rowIndex;
            }
            else {
                return offset + (allianceIndex * this.allianceSize) + teamIndex;
            }
        },
        getCellClass(allianceIndex, teamIndex) {
            const value = (this.teams[allianceIndex] || [])[teamIndex];
            if (!value) {
                return '';
            }
            for (let a = 0; a < this.allianceCount; a++) {
                for (let t = 0; t < this.allianceSize; t++) {
                    if (a == allianceIndex && t == teamIndex) {
                        continue;
                    }
                    if (this.teams[a][t] == value) {
                        return 'is-invalid';
                    }
                }
            }
            return '';
        },

        _resizeStorage() {
            while (this.teams.length < this.allianceCount) {
                this.teams.push([]);
            }
            for (let alliance of this.teams) {
                while (alliance.length < this.allianceSize) {
                    alliance.push('');
                }
            }
        },
        _exportTeams() {
            let out = [];
            for (let a = 0; a < this.allianceCount; a++) {
                out[a] = [];
                for (let t = 0; t < this.allianceSize; t++) {
                    let team = Number(this.teams[a][t]);
                    if (team && !isNaN(team)) {
                        out[a][t] = team;
                    }
                }
            }
            return out;
        },
        _importTeams(keys) {
            for (let a = 0; a < this.allianceCount; a++) {
                for (let t = 0; t < this.allianceSize; t++) {
                    let team = String((keys[a] || [])[t] || '').replace('frc', '');
                    this.$set(this.teams[a], t, team);
                }
            }
        },
    },
};
</script>
