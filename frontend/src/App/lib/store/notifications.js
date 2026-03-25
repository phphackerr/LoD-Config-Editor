import { writable } from 'svelte/store';

const MAX_NOTIFICATIONS = 5;
const DEFAULT_DURATION = 5000;

const listStore = writable([]);
const timers = new Map();
let seq = 0;

function nextID() {
  seq += 1;
  return `ntf-${Date.now()}-${seq}`;
}

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

function toMessageText(value) {
  if (typeof value === 'string') return value;
  if (value instanceof Error) return value.message || '';

  if (value && typeof value === 'object') {
    if (typeof value.message === 'string') return value.message;
    if (typeof value.error === 'string') return value.error;
    if (typeof value.error?.message === 'string') return value.error.message;

    try {
      return JSON.stringify(value);
    } catch {
      return '';
    }
  }

  return String(value ?? '');
}

function isPlaceholderText(value) {
  return /^(?:\{\}|\[\]|null|undefined|\[object Object\]|"")$/i.test(String(value || '').trim());
}

function normalizeMessage(value) {
  const source = toMessageText(value);
  const runtimeMessage = extractRuntimeMessage(source);
  const text = String(runtimeMessage || source || '')
    .replace(/[\u0000-\u001f\u007f-\u009f]/g, ' ')
    .replace(/[\u200b-\u200f\u202a-\u202e\u2060-\u206f\ufeff]/g, '')
    .replace(/\s+/g, ' ')
    .trim();

  if (!text || isPlaceholderText(text)) {
    return '';
  }

  return text;
}

function clearTimer(id) {
  const timer = timers.get(id);
  if (!timer) return;
  clearTimeout(timer);
  timers.delete(id);
}

export function dismissNotification(id) {
  clearTimer(id);
  listStore.update((items) => items.filter((item) => item.id !== id));
}

export function clearNotifications() {
  for (const id of timers.keys()) {
    clearTimer(id);
  }
  listStore.set([]);
}

export function pushNotification({
  type = 'info',
  title = '',
  message = '',
  duration = DEFAULT_DURATION
} = {}) {
  const text = normalizeMessage(message);
  if (!text) return null;

  const id = nextID();
  const item = {
    id,
    type: String(type || 'info'),
    title: String(title || '').trim(),
    message: text,
    createdAt: Date.now()
  };

  listStore.update((items) => {
    const next = [...items, item];
    const overflow = next.length - MAX_NOTIFICATIONS;
    if (overflow <= 0) return next;

    const removed = next.slice(0, overflow);
    for (const entry of removed) {
      clearTimer(entry.id);
    }
    return next.slice(overflow);
  });

  const timeout = Number(duration);
  if (Number.isFinite(timeout) && timeout > 0) {
    timers.set(
      id,
      setTimeout(() => {
        dismissNotification(id);
      }, timeout)
    );
  }

  return id;
}

export function notifySuccess(message, title = '') {
  return pushNotification({ type: 'success', message, title });
}

export function notifyError(message, title = '') {
  return pushNotification({ type: 'error', message, title, duration: 7000 });
}

export function notifyInfo(message, title = '') {
  return pushNotification({ type: 'info', message, title });
}

export function notifyWarning(message, title = '') {
  return pushNotification({ type: 'warning', message, title, duration: 6500 });
}

export const notifications = {
  subscribe: listStore.subscribe
};
