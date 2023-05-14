type NavLink = {
    readonly path: string,
    readonly name: string,
    readonly hideFromSidebar?: boolean,
    readonly requiresEvent?: boolean,
};

export const ALL_PAGES : NavLink[] = [
    {path: '/setup/', name: 'Event Setup'},
    {path: '/awards/', name: 'Awards', requiresEvent: true},
    {path: '/settings/', name: 'Settings', hideFromSidebar: true},
    {path: '/help/', name: 'Help', hideFromSidebar: true},
];

export const SIDEBAR_PAGES = ALL_PAGES.filter(p => !p.hideFromSidebar);
export const HEADER_PAGES = ALL_PAGES.filter(p => p.hideFromSidebar);

export const PAGES_BY_PATH : {string: NavLink} = Object.fromEntries(ALL_PAGES.map(p => [p.path, p]));

export function pageNameFromPath(path: string): string {
    return PAGES_BY_PATH[path] ? PAGES_BY_PATH[path].name : '';
};
