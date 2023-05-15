import type { Writable } from 'svelte/store';
import { localStorageStore } from '@skeletonlabs/skeleton';

type APISettings = {
    tbaUrl: string;
    useProxy: boolean;
};

export const API_SETTINGS_DEFAULT: APISettings = Object.freeze({
    tbaUrl: 'https://www.thebluealliance.com',
    useProxy: false,
});

export const apiSettingsStore: Writable<APISettings> = localStorageStore('apiSettings', {...API_SETTINGS_DEFAULT});
let apiSettingsLocal: APISettings;
apiSettingsStore.subscribe(state => {
    apiSettingsLocal = {...state};
});

export function getApiBaseUrl(): string {
    return import.meta.env.DEV ? 'http://localhost:8808/api' : '/api';
}

export function getTbaBaseUrl(): string {
    return apiSettingsLocal.tbaUrl;
}

type TBAReadParams = {
    event?: string;
    key: string;
    route?: string;
};

export async function tbaRead(params: TBAReadParams) {
    let url = `${getTbaBaseUrl()}/api/v3`;
    if (params.event) {
        url += `/event/${params.event}`;
    }
    if (params.route) {
        url += `/${params.route}`;
    }
    const response = await fetch(url, {
        method: 'GET',
        cache: 'no-cache',
        headers: {
            'X-TBA-Auth-Key': params.key,
        },
    });
    const body = await response.json();
    if (!response.ok) {
        throw new Error(`TBA error ${response.status}: ${body.Error}`)
    }
    return body;
}

type TBAWriteBody = Object | any[];

type TBAWriteParams = {
    event: string;
    authId: string;
    authSecret: string;
    body: TBAWriteBody;
};

export async function tbaWrite(params: TBAWriteParams) {
    // todo
}

type TBAClientParams = {
    event: string;
    readKey: string;
    authId: string;
    authSecret: string;
};

export function makeTbaApiClient(params: TBAClientParams) {
    params = {...params};

    return {
        async read(route?: string) {
            return tbaRead({event: params.event, key: params.readKey, route});
        },
        async write(route: string, body: TBAWriteBody) {
            return tbaWrite({event: params.event, authId: params.authId, authSecret: params.authSecret, route, body});
        },
    };
}
