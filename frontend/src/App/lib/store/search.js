import { get, writable } from 'svelte/store';
import { tr } from './storeUtils';

export const searchQuery = writable('');
export const activeTab = writable('');

const SEARCH_OP_SET_QUERY = 'search:set-query';
const SEARCH_OP_SET_ACTIVE_TAB = 'search:set-active-tab';
const SEARCH_OP_REGISTER_ITEM = 'search:register-item';
const SEARCH_OP_UNREGISTER_ITEM = 'search:unregister-item';
const SEARCH_OP_UPDATE_ITEM = 'search:update-item';
const SEARCH_OP_CLEAR_ITEMS = 'search:clear-items';

const initialSearchState = {
  loading: false,
  success: false,
  error: null,
  operation: '',
  lastResult: null,
  query: '',
  activeTab: '',
  itemsCount: 0
};

export const searchState = writable({
  ...initialSearchState
});

let operationSeq = 0;

function patchSearchState(patch) {
  searchState.update((state) => ({ ...state, ...patch }));
}

function createSearchResult(operation, { ok, data = null, error = null, rolledBack = false }) {
  return {
    ok: Boolean(ok),
    operation,
    data,
    error: error ? String(error) : null,
    rolledBack: Boolean(rolledBack)
  };
}

function startSearchOperation(operation, { loading = false } = {}) {
  operationSeq += 1;
  const token = operationSeq;
  patchSearchState({
    loading,
    success: false,
    error: null,
    operation
  });
  return token;
}

function finishSearchOperation(token, result, patch = {}) {
  if (token !== operationSeq) {
    return result;
  }

  patchSearchState({
    loading: false,
    success: result.ok,
    error: result.error,
    operation: result.operation,
    lastResult: result,
    query: get(searchQuery),
    activeTab: get(activeTab),
    ...patch
  });
  return result;
}

function normalizeId(value) {
  if (typeof value === 'symbol') return value;

  const id = String(value ?? '').trim();
  return id || null;
}

function normalizeItem(item) {
  const id = normalizeId(item?.id);
  const label = String(item?.label || '').trim();

  if (!id || !label) return null;

  return {
    ...item,
    id,
    label
  };
}

function normalizeText(value) {
  return String(value ?? '').trim();
}

function syncSearchSnapshot(itemsCount = null) {
  patchSearchState({
    query: get(searchQuery),
    activeTab: get(activeTab),
    ...(itemsCount == null ? {} : { itemsCount: Math.max(0, Number(itemsCount) || 0) })
  });
}

export function setSearchQuery(value) {
  const token = startSearchOperation(SEARCH_OP_SET_QUERY, { loading: false });
  const query = normalizeText(value);
  searchQuery.set(query);
  return finishSearchOperation(
    token,
    createSearchResult(SEARCH_OP_SET_QUERY, {
      ok: true,
      data: { query }
    })
  );
}

export function clearSearchQuery() {
  return setSearchQuery('');
}

export function setActiveTab(tabId) {
  const token = startSearchOperation(SEARCH_OP_SET_ACTIVE_TAB, { loading: false });
  const nextTabID = normalizeText(tabId);
  activeTab.set(nextTabID);
  return finishSearchOperation(
    token,
    createSearchResult(SEARCH_OP_SET_ACTIVE_TAB, {
      ok: true,
      data: { tabId: nextTabID }
    })
  );
}

export function clearActiveTab() {
  return setActiveTab('');
}

function createSearchableItemsStore() {
  const itemsStore = writable([]);
  const { subscribe, update, set } = itemsStore;

  return {
    subscribe,
    register(item) {
      const token = startSearchOperation(SEARCH_OP_REGISTER_ITEM, { loading: false });
      const nextItem = normalizeItem(item);
      if (!nextItem) {
        return finishSearchOperation(
          token,
          createSearchResult(SEARCH_OP_REGISTER_ITEM, {
            ok: false,
            error: tr('ERRORS.search.invalid_item')
          })
        );
      }

      let itemsCount = 0;
      update((items) => {
        const existingIndex = items.findIndex((entry) => entry.id === nextItem.id);
        let nextItems = items;
        if (existingIndex === -1) {
          nextItems = [...items, nextItem];
        } else {
          nextItems = [...items];
          nextItems[existingIndex] = { ...nextItems[existingIndex], ...nextItem };
        }
        itemsCount = nextItems.length;
        return nextItems;
      });

      return finishSearchOperation(
        token,
        createSearchResult(SEARCH_OP_REGISTER_ITEM, {
          ok: true,
          data: nextItem
        }),
        { itemsCount }
      );
    },
    unregister(id) {
      const token = startSearchOperation(SEARCH_OP_UNREGISTER_ITEM, { loading: false });
      const targetId = normalizeId(id);
      if (!targetId) {
        return finishSearchOperation(
          token,
          createSearchResult(SEARCH_OP_UNREGISTER_ITEM, {
            ok: false,
            error: tr('ERRORS.search.invalid_item_id')
          })
        );
      }

      let removed = false;
      let itemsCount = 0;
      update((items) => {
        const nextItems = items.filter((item) => {
          const keep = item.id !== targetId;
          if (!keep) removed = true;
          return keep;
        });
        itemsCount = nextItems.length;
        return nextItems;
      });

      return finishSearchOperation(
        token,
        createSearchResult(SEARCH_OP_UNREGISTER_ITEM, {
          ok: true,
          data: {
            id: targetId,
            removed
          }
        }),
        { itemsCount }
      );
    },
    update(id, patch) {
      const token = startSearchOperation(SEARCH_OP_UPDATE_ITEM, { loading: false });
      const targetId = normalizeId(id);
      if (!targetId || !patch) {
        return finishSearchOperation(
          token,
          createSearchResult(SEARCH_OP_UPDATE_ITEM, {
            ok: false,
            error: tr('ERRORS.search.invalid_item_update')
          })
        );
      }

      let updated = false;
      let itemsCount = 0;
      update((items) => {
        const nextItems = items.map((item) => {
          if (item.id !== targetId) return item;
          updated = true;
          return { ...item, ...patch };
        });
        itemsCount = nextItems.length;
        return nextItems;
      });

      return finishSearchOperation(
        token,
        createSearchResult(SEARCH_OP_UPDATE_ITEM, {
          ok: true,
          data: {
            id: targetId,
            updated
          }
        }),
        { itemsCount }
      );
    },
    clear() {
      const token = startSearchOperation(SEARCH_OP_CLEAR_ITEMS, { loading: false });
      set([]);
      return finishSearchOperation(
        token,
        createSearchResult(SEARCH_OP_CLEAR_ITEMS, {
          ok: true,
          data: []
        }),
        { itemsCount: 0 }
      );
    }
  };
}

export const searchableItems = createSearchableItemsStore();

const unsubscribeQuery = searchQuery.subscribe(() => {
  syncSearchSnapshot();
});
const unsubscribeActiveTab = activeTab.subscribe(() => {
  syncSearchSnapshot();
});

if (import.meta.hot) {
  import.meta.hot.dispose(() => {
    unsubscribeQuery();
    unsubscribeActiveTab();
  });
}
