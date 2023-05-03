<template>
    <div class="col-md-6 col-lg-4">
        <table
            class="table table-sm"
            @click="$emit('click')"
        >
            <tbody>
                <tr>
                    <td>{{ match.id }}</td>
                    <td :class="match.classes[match.teams.red[0]]"><span>{{ match.teams.red[0] }}</span></td>
                    <td :class="match.classes[match.teams.red[1]]"><span>{{ match.teams.red[1] }}</span></td>
                    <td :class="match.classes[match.teams.red[2]]"><span>{{ match.teams.red[2] }}</span></td>
                    <td class="red score">{{ match.score_summary.red }}</td>
                </tr>
                <tr>
                    <td>{{ match.key }}</td>
                    <td :class="match.classes[match.teams.blue[0]]"><span>{{ match.teams.blue[0] }}</span></td>
                    <td :class="match.classes[match.teams.blue[1]]"><span>{{ match.teams.blue[1] }}</span></td>
                    <td :class="match.classes[match.teams.blue[2]]"><span>{{ match.teams.blue[2] }}</span></td>
                    <td class="blue score">{{ match.score_summary.blue }}</td>
                </tr>
            </tbody>
        </table>
    </div>
</template>

<style scoped>
table td, table.match-edit td {
    border: 1px solid #ccc;
}

table td {
    cursor: pointer;
}

table:hover {
    box-shadow: 0px 0px 3px #00000066;
}
</style>

<script>
export default {
    name: 'ScoreSummary',
    props: {
        match: {
            type: Object,
            required: true,
        },
    },
};
</script>
