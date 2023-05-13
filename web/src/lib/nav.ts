type NavLink = {
    readonly path: string,
    readonly name: string,
};

export const ALL_PAGES : NavLink[] = [
    {path: '/setup', name: 'Event Setup'},
    {path: '/awards', name: 'Awards'},
];

export const PAGES_BY_PATH : {string: NavLink} = Object.fromEntries(ALL_PAGES.map(p => [p.path, p]));

export function pageNameFromPath(path: string): string {
    return PAGES_BY_PATH[path] ? PAGES_BY_PATH[path].name : '';
};
