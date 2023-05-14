// A context that allows tracking whether a form is being processed
// Used for e.g. disabling buttons

import { getContext } from 'svelte';
import { writable } from 'svelte/store';

export const FORM_CONTEXT_KEY = Symbol('FormContext');

type FormContext = {
    inSubmit: boolean;
};

export function makeFormContext(): FormContext {
    const store = writable({
        inSubmit: false,
        inSubmitDepth: 0,
    });

    store.withSubmit = async (callback) => {
        try {
            store.update(state => ({
                ...state,
                inSubmit: true,
                inSubmitDepth: state.inSubmitDepth + 1,
            }));
            await callback();
        } finally {
            store.update(state => ({
                ...state,
                inSubmit: state.inSubmitDepth > 1,
                inSubmitDepth: state.inSubmitDepth - 1,
            }));
        }
    }

    return store;
}

export function getFormContext() {
    return getContext(FORM_CONTEXT_KEY);
}
