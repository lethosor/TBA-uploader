export function getApiBaseUrl() {
    return import.meta.env.DEV ? 'http://localhost:8808/api' : '/api';
}
