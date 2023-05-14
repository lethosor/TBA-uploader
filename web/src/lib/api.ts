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
})

export function getApiBaseUrl(): string {
    return import.meta.env.DEV ? 'http://localhost:8808/api' : '/api';
}

export function getTbaBaseUrl(): string {
    return apiSettingsLocal.tbaUrl;
}
