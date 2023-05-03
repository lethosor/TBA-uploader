const Reports = {};
window.Reports = Reports;

Reports.Type = Object.freeze({
    QUAL_SCHEDULE: 'ScheduleReportQualification',
    PLAYOFF_SCHEDULE: 'ScheduleReportPlayoff',
    TEAM_LIST: 'TeamListActiveEvent',
    QUAL_CYCLE_TIMES: 'CycleTimeReportQualification',
    PLAYOFF_CYCLE_TIMES: 'CycleTimeReportPlayoff',
});

let parseCellText = function(cell) {
    if (!cell.ItemModel) {
        return '';
    }
    return (cell.ItemModel.Paragraphval || []).map(paragraph =>
        (paragraph.Runs || []).map(run => run.RunText).filter(Boolean).join(''),
    ).filter(Boolean).join('');
};

Reports.convertToCellsWithPages = function(page_models) {
    return page_models.map(page_model =>
        page_model.reportPageModel.PageData.map(page_data =>
            page_data.PageModel.map(page =>
                (page.CellModels || []).map(row =>
                    row.map(parseCellText),
                ),
            ),
        ).flat(), // flatten PageModel
    ).flat(); // flatten PageData
};

Reports.convertToCells = function(page_models) {
    return Reports.convertToCellsWithPages(page_models).flat();
};

export default Object.freeze(Reports);
