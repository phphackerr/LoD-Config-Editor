import { get } from 'svelte/store';
import { _ } from 'svelte-i18n';

function extractRuntimeMessage(value) {
  const text = String(value || '').trim();
  if (!text) return '';

  const jsonStart = text.indexOf('{');
  if (jsonStart < 0) return '';

  try {
    const parsed = JSON.parse(text.slice(jsonStart));
    if (typeof parsed?.message === 'string' && parsed.message.trim()) {
      return parsed.message.trim();
    }
  } catch {
    return '';
  }

  return '';
}

function isPlaceholderErrorText(value) {
  const text = String(value || '').trim();
  if (!text) return true;

  return /^(?:\{\}|\[\]|null|undefined|\[object Object\])$/i.test(text);
}

function isJsonPayload(value) {
  const text = String(value || '').trim();
  if (!text) return false;

  const jsonStart = text.indexOf('{');
  if (jsonStart < 0) return false;

  try {
    JSON.parse(text.slice(jsonStart));
    return true;
  } catch {
    return false;
  }
}

export function toErrorMessage(error, fallback = tr('ERRORS.unexpected', {}, 'Unexpected error')) {
  if (!error) return fallback;

  if (typeof error === 'string') {
    const runtimeMessage = extractRuntimeMessage(error);
    if (runtimeMessage && !isPlaceholderErrorText(runtimeMessage)) {
      return runtimeMessage;
    }

    if (isJsonPayload(error) || isPlaceholderErrorText(error)) {
      return fallback;
    }

    return String(error).trim();
  }

  if (error instanceof Error && error.message) {
    const runtimeMessage = extractRuntimeMessage(error.message);
    if (runtimeMessage && !isPlaceholderErrorText(runtimeMessage)) {
      return runtimeMessage;
    }

    if (isJsonPayload(error.message) || isPlaceholderErrorText(error.message)) {
      return fallback;
    }

    return String(error.message).trim();
  }

  if (typeof error.message === 'string' && error.message) {
    const runtimeMessage = extractRuntimeMessage(error.message);
    if (runtimeMessage && !isPlaceholderErrorText(runtimeMessage)) {
      return runtimeMessage;
    }

    if (isJsonPayload(error.message) || isPlaceholderErrorText(error.message)) {
      return fallback;
    }

    return String(error.message).trim();
  }

  return fallback;
}

export function tr(key, values = {}, fallback = '') {
  const messageKey = String(key || '').trim();
  if (!messageKey) return String(fallback || '');

  try {
    const translate = get(_);
    if (typeof translate === 'function') {
      const options =
        values && typeof values === 'object' && !Array.isArray(values) && 'values' in values
          ? values
          : { values: values && typeof values === 'object' ? values : {} };
      const translated = translate(messageKey, options);
      if (typeof translated === 'string' && translated.trim()) {
        if (!fallback || translated !== messageKey) {
          return translated;
        }
      }
    }
  } catch {
    // i18n might not be initialised yet, fallback is used below.
  }

  return String(fallback || messageKey);
}

export function toLocalizedError(
  error,
  key,
  fallback = tr('ERRORS.unexpected', {}, 'Unexpected error'),
  values = {}
) {
  return toErrorMessage(error, tr(key, values, fallback));
}

export function normalizePath(path) {
  if (!path) return '';
  return String(path).replace(/\\/g, '/').replace(/\/+/g, '/').trim().toLowerCase();
}

export function samePath(a, b) {
  return normalizePath(a) === normalizePath(b);
}

export function uniquePaths(paths) {
  const seen = new Set();
  const result = [];

  for (const rawPath of paths || []) {
    const path = String(rawPath || '').trim();
    if (!path) continue;

    const key = normalizePath(path);
    if (seen.has(key)) continue;
    seen.add(key);
    result.push(path);
  }

  return result;
}
