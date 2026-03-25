<script>
  import { t } from 'svelte-i18n';
  import { draftTheme, selectedTokenId, updateToken } from '../../../../lib/store/themeEditor';
  import ColorEditor from './ColorEditor.svelte';
  import GradientEditor from './GradientEditor.svelte';
  import NumberEditor from './NumberEditor.svelte';
  import ShadowEditor from './ShadowEditor.svelte';

  const NUMBER_LIKE_TYPES = new Set(['number', 'spacing', 'radius']);
  const DERIVED_OPS = ['alpha', 'lighten', 'darken'];

  let selectedState = 'default';
  let literalJsonText = '';
  let literalJsonError = '';
  let literalJsonSyncKey = '';

  $: tokens = Array.isArray($draftTheme?.tokens) ? $draftTheme.tokens : [];
  $: token = $selectedTokenId ? tokens.find((t) => t.id === $selectedTokenId) : null;
  $: tokenMap = new Map(tokens.map((entry) => [entry.id, entry]));
  $: tokenStates = getTokenStates(token);
  $: if (tokenStates.length > 0 && !tokenStates.includes(selectedState)) {
    selectedState = tokenStates[0];
  }
  $: currentValue = getTokenValue(token, selectedState);
  $: currentKind = getTokenValueKind(currentValue);
  $: currentColor = token?.type === 'color' ? resolveColorValue(currentValue, tokenMap) : null;
  $: isNumberLikeType = token ? NUMBER_LIKE_TYPES.has(token.type) : false;
  $: currentNumber = isNumberLikeType ? resolveNumberValue(currentValue, tokenMap) : null;
  $: currentShadow = token?.type === 'shadow' ? currentValue?.literal?.shadow || [] : null;
  $: currentGradient = token?.type === 'gradient' ? currentValue?.literal?.gradient || null : null;
  $: valueKind = describeValueKind(currentValue);

  $: compatibleRefIDs = getCompatibleRefIDs(token, tokens);
  $: colorRefIDs = getColorRefIDs(token, tokens);
  $: activeRefTargetID = currentKind === 'ref' ? String(currentValue?.ref?.id || '') : '';
  $: derivedOp = getDerivedOp(currentValue);
  $: derivedFromRefID = getDerivedFromRefID(currentValue);
  $: derivedAmount = getDerivedAmount(currentValue);

  $: literalJsonSource = `${token?.id || ''}:${selectedState}:${currentKind}:${JSON.stringify(
    currentValue?.literal || null
  )}`;
  $: if (literalJsonSource !== literalJsonSyncKey) {
    literalJsonSyncKey = literalJsonSource;
    literalJsonText = formatLiteralJSON(currentValue?.literal, token?.type);
    literalJsonError = '';
  }

  function hasConcreteValue(value) {
    if (value?.literal) return true;
    if (value?.ref && String(value.ref.id || '').trim() !== '') return true;
    if (value?.derived) return true;
    return false;
  }

  function getTokenValueKind(value) {
    if (value?.literal) return 'literal';
    if (value?.ref?.id) return 'ref';
    if (value?.derived) return 'derived';
    return 'none';
  }

  function describeValueKind(value) {
    if (value?.literal) return 'literal';
    if (value?.ref?.id) return `ref:${value.ref.id}`;
    if (value?.derived?.op) return `derived:${value.derived.op}`;
    return 'none';
  }

  function getTokenStates(currentToken) {
    if (!currentToken) return [];

    const states = ['default'];
    for (const stateName of Object.keys(currentToken.states || {}).sort((a, b) =>
      a.localeCompare(b)
    )) {
      if (stateName !== 'default') {
        states.push(stateName);
      }
    }

    return states;
  }

  function getTokenValue(currentToken, stateName) {
    if (!currentToken) return null;

    if (stateName === 'default') {
      if (hasConcreteValue(currentToken.states?.default)) {
        return currentToken.states.default;
      }
      if (hasConcreteValue(currentToken.value)) {
        return currentToken.value;
      }
      return null;
    }

    return hasConcreteValue(currentToken.states?.[stateName])
      ? currentToken.states[stateName]
      : null;
  }

  function getFallbackTokenValue(currentToken) {
    if (!currentToken) return null;
    if (hasConcreteValue(currentToken.value)) return currentToken.value;
    if (hasConcreteValue(currentToken.states?.default)) return currentToken.states.default;

    const stateNames = Object.keys(currentToken.states || {});
    for (const stateName of stateNames) {
      const candidate = currentToken.states?.[stateName];
      if (hasConcreteValue(candidate)) {
        return candidate;
      }
    }

    return null;
  }

  function toFinite(value, fallback) {
    const n = Number(value);
    return Number.isFinite(n) ? n : fallback;
  }

  function clamp(value, min, max) {
    if (value < min) return min;
    if (value > max) return max;
    return value;
  }

  function normalizeColorLiteral(color) {
    return {
      space: String(color?.space || 'srgb'),
      r: clamp(toFinite(color?.r, 255), 0, 255),
      g: clamp(toFinite(color?.g, 255), 0, 255),
      b: clamp(toFinite(color?.b, 255), 0, 255),
      a: clamp(toFinite(color?.a, 1), 0, 1)
    };
  }

  function normalizeNumberLiteral(number, type) {
    const defaultUnit = type === 'number' ? '' : 'px';
    return {
      value: toFinite(number?.value, 0),
      unit: typeof number?.unit === 'string' ? number.unit : defaultUnit
    };
  }

  function buildDefaultColorLiteral() {
    return {
      color: {
        space: 'srgb',
        r: 255,
        g: 255,
        b: 255,
        a: 1
      }
    };
  }

  function buildDefaultNumberLiteral(type) {
    return {
      number: normalizeNumberLiteral(null, type)
    };
  }

  function buildDefaultShadowLiteral() {
    return {
      shadow: [
        {
          x: 0,
          y: 2,
          blur: 8,
          spread: 0,
          inset: false,
          color: {
            literal: {
              color: {
                space: 'srgb',
                r: 0,
                g: 0,
                b: 0,
                a: 0.35
              }
            }
          }
        }
      ]
    };
  }

  function buildDefaultGradientLiteral() {
    return {
      gradient: {
        kind: 'linear',
        angle: 180,
        stops: [
          {
            pos: 0,
            color: {
              literal: {
                color: {
                  space: 'srgb',
                  r: 255,
                  g: 255,
                  b: 255,
                  a: 1
                }
              }
            }
          },
          {
            pos: 1,
            color: {
              literal: {
                color: {
                  space: 'srgb',
                  r: 0,
                  g: 0,
                  b: 0,
                  a: 1
                }
              }
            }
          }
        ]
      }
    };
  }

  function buildDefaultLiteralForType(type) {
    if (type === 'color') return buildDefaultColorLiteral();
    if (NUMBER_LIKE_TYPES.has(type)) return buildDefaultNumberLiteral(type);
    if (type === 'shadow') return buildDefaultShadowLiteral();
    if (type === 'gradient') return buildDefaultGradientLiteral();
    return {};
  }

  function deepClone(value) {
    if (value == null) return value;
    if (typeof structuredClone === 'function') {
      return structuredClone(value);
    }
    return JSON.parse(JSON.stringify(value));
  }

  function resolveColorValue(value, map, visited = new Set()) {
    if (!value) return null;

    if (value.literal?.color) {
      return value.literal.color;
    }

    if (value.ref?.id) {
      const refID = String(value.ref.id || '').trim();
      if (!refID || visited.has(refID)) return null;

      const refToken = map.get(refID);
      if (!refToken) return null;

      const nextVisited = new Set(visited);
      nextVisited.add(refID);

      const refValue = getTokenValue(refToken, 'default') || getFallbackTokenValue(refToken);
      return resolveColorValue(refValue, map, nextVisited);
    }

    if (value.derived) {
      const base = resolveColorValue(value.derived.from, map, visited);
      if (!base) return null;

      const amount = Number(value.derived.amount || 0);
      const next = { ...base };

      switch (value.derived.op) {
        case 'alpha':
          next.a = clamp(amount, 0, 1);
          return next;
        case 'lighten':
          next.r = clamp(next.r + amount, 0, 255);
          next.g = clamp(next.g + amount, 0, 255);
          next.b = clamp(next.b + amount, 0, 255);
          return next;
        case 'darken':
          next.r = clamp(next.r - amount, 0, 255);
          next.g = clamp(next.g - amount, 0, 255);
          next.b = clamp(next.b - amount, 0, 255);
          return next;
        default:
          return null;
      }
    }

    return null;
  }

  function resolveNumberValue(value, map, visited = new Set()) {
    if (!value) return null;

    if (value.literal?.number) {
      return value.literal.number;
    }

    if (value.ref?.id) {
      const refID = String(value.ref.id || '').trim();
      if (!refID || visited.has(refID)) return null;

      const refToken = map.get(refID);
      if (!refToken) return null;

      const nextVisited = new Set(visited);
      nextVisited.add(refID);

      const refValue = getTokenValue(refToken, 'default') || getFallbackTokenValue(refToken);
      return resolveNumberValue(refValue, map, nextVisited);
    }

    return null;
  }

  function isCompatibleType(sourceType, targetType) {
    if (sourceType === targetType) return true;
    if (NUMBER_LIKE_TYPES.has(sourceType) && NUMBER_LIKE_TYPES.has(targetType)) return true;
    return false;
  }

  function getCompatibleRefIDs(currentToken, allTokens) {
    if (!currentToken) return [];

    return allTokens
      .filter((candidate) => {
        if (!candidate?.id || candidate.id === currentToken.id) return false;
        return isCompatibleType(currentToken.type, candidate.type);
      })
      .map((entry) => entry.id)
      .sort((a, b) => a.localeCompare(b));
  }

  function getColorRefIDs(currentToken, allTokens) {
    if (!currentToken) return [];

    return allTokens
      .filter((candidate) => {
        if (!candidate?.id || candidate.id === currentToken.id) return false;
        return candidate.type === 'color';
      })
      .map((entry) => entry.id)
      .sort((a, b) => a.localeCompare(b));
  }

  function setTokenValue(currentToken, stateName, nextValue) {
    if (!currentToken) return currentToken;

    const nextToken = { ...currentToken };
    const clonedValue = deepClone(nextValue);

    if (stateName === 'default') {
      if (hasConcreteValue(nextToken.states?.default)) {
        nextToken.states = {
          ...(nextToken.states || {}),
          default: clonedValue
        };
        return nextToken;
      }

      nextToken.value = clonedValue;
      return nextToken;
    }

    nextToken.states = {
      ...(nextToken.states || {}),
      [stateName]: clonedValue
    };

    return nextToken;
  }

  function updateCurrentValue(nextValue) {
    if (!$selectedTokenId) return;
    updateToken($selectedTokenId, (currentToken) =>
      setTokenValue(currentToken, selectedState, nextValue)
    );
  }

  function switchToLiteral() {
    if (!token) return;

    if (token.type === 'color') {
      const color = normalizeColorLiteral(currentColor || currentValue?.literal?.color || {});
      updateCurrentValue({ literal: { color } });
      return;
    }

    if (NUMBER_LIKE_TYPES.has(token.type)) {
      const number = normalizeNumberLiteral(
        currentNumber || currentValue?.literal?.number,
        token.type
      );
      updateCurrentValue({ literal: { number } });
      return;
    }

    const fallback = buildDefaultLiteralForType(token.type);
    const literal =
      currentKind === 'literal' && currentValue?.literal
        ? deepClone(currentValue.literal)
        : deepClone(fallback);
    updateCurrentValue({ literal });
  }

  function switchToRef() {
    if (!compatibleRefIDs.length) return;

    const currentRefID = String(currentValue?.ref?.id || '').trim();
    const fallbackRefID = compatibleRefIDs[0];
    const nextRefID = compatibleRefIDs.includes(currentRefID) ? currentRefID : fallbackRefID;
    if (!nextRefID) return;
    updateCurrentValue({ ref: { id: nextRefID } });
  }

  function defaultAmountForOp(op) {
    return op === 'alpha' ? 1 : 10;
  }

  function normalizeDerivedAmount(op, amount) {
    const parsed = toFinite(amount, defaultAmountForOp(op));
    return op === 'alpha' ? clamp(parsed, 0, 1) : parsed;
  }

  function getDerivedOp(value) {
    const op = String(value?.derived?.op || '').trim();
    return DERIVED_OPS.includes(op) ? op : 'alpha';
  }

  function getDerivedFromRefID(value) {
    return String(value?.derived?.from?.ref?.id || '').trim();
  }

  function getDerivedAmount(value) {
    const op = getDerivedOp(value);
    return normalizeDerivedAmount(op, value?.derived?.amount);
  }

  function buildDefaultDerived() {
    const fromID = colorRefIDs[0];
    return {
      op: 'alpha',
      from: fromID
        ? { ref: { id: fromID } }
        : {
            literal: {
              color: normalizeColorLiteral(currentColor || currentValue?.literal?.color || {})
            }
          },
      amount: 1
    };
  }

  function switchToDerived() {
    if (token?.type !== 'color') return;

    const source = currentValue?.derived ? deepClone(currentValue.derived) : buildDefaultDerived();
    const op = DERIVED_OPS.includes(source.op) ? source.op : 'alpha';
    const from = hasConcreteValue(source.from) ? source.from : buildDefaultDerived().from;
    const amount = normalizeDerivedAmount(op, source.amount);

    updateCurrentValue({
      derived: {
        op,
        from,
        amount
      }
    });
  }

  function patchDerived(patch) {
    if (token?.type !== 'color') return;

    const base =
      currentKind === 'derived' && currentValue?.derived
        ? deepClone(currentValue.derived)
        : buildDefaultDerived();
    const next = {
      ...base,
      ...patch
    };

    const op = DERIVED_OPS.includes(next.op) ? next.op : 'alpha';
    const from = hasConcreteValue(next.from) ? next.from : buildDefaultDerived().from;
    const amount = normalizeDerivedAmount(op, next.amount);

    updateCurrentValue({
      derived: {
        op,
        from,
        amount
      }
    });
  }

  function handleRefChange(event) {
    const nextRefID = String(event.currentTarget?.value || '').trim();
    if (!nextRefID) return;
    updateCurrentValue({ ref: { id: nextRefID } });
  }

  function handleDerivedOpChange(event) {
    patchDerived({ op: String(event.currentTarget?.value || 'alpha') });
  }

  function handleDerivedFromChange(event) {
    const nextRefID = String(event.currentTarget?.value || '').trim();
    if (!nextRefID) return;
    patchDerived({ from: { ref: { id: nextRefID } } });
  }

  function handleDerivedAmountChange(event) {
    patchDerived({ amount: event.currentTarget?.value });
  }

  function updateColor(color) {
    if (!$selectedTokenId) return;
    updateCurrentValue({ literal: { color: normalizeColorLiteral(color) } });
  }

  function updateNumber(number) {
    if (!$selectedTokenId) return;
    updateCurrentValue({ literal: { number: normalizeNumberLiteral(number, token?.type) } });
  }

  function updateShadow(shadow) {
    if (!$selectedTokenId) return;
    const normalized = Array.isArray(shadow) ? shadow : [];
    updateCurrentValue({ literal: { shadow: normalized } });
  }

  function updateGradient(gradient) {
    if (!$selectedTokenId) return;
    if (!gradient || typeof gradient !== 'object') return;
    updateCurrentValue({ literal: { gradient } });
  }

  function handleStateChange(event) {
    selectedState = String(event.currentTarget?.value || 'default');
  }

  function formatLiteralJSON(literal, type) {
    try {
      if (literal && typeof literal === 'object' && !Array.isArray(literal)) {
        return JSON.stringify(literal, null, 2);
      }
      return JSON.stringify(buildDefaultLiteralForType(type), null, 2);
    } catch {
      return '{}';
    }
  }

  function handleLiteralJsonInput(event) {
    literalJsonText = String(event.currentTarget?.value || '');
    literalJsonError = '';
  }

  function applyLiteralJson() {
    try {
      const parsed = JSON.parse(literalJsonText);
      if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
        throw new Error($t('SETTINGS.THEME_EDITOR.literal_json_object_required'));
      }
      updateCurrentValue({ literal: parsed });
      literalJsonError = '';
    } catch (error) {
      literalJsonError = error instanceof Error ? error.message : String(error);
    }
  }
</script>

<div class="editor">
  {#if token}
    <h3>{token.id}</h3>
    <div class="meta">
      <span class="badge">{$t('COMMON.type')}: {token.type}</span>
      <span class="badge">{$t('COMMON.value')}: {valueKind}</span>

      {#if tokenStates.length > 1}
        <label class="state-select">
          {$t('COMMON.state')}
          <select value={selectedState} on:change={handleStateChange}>
            {#each tokenStates as stateName}
              <option value={stateName}>{stateName}</option>
            {/each}
          </select>
        </label>
      {:else}
        <span class="badge">{$t('COMMON.state')}: {tokenStates[0] || 'default'}</span>
      {/if}
    </div>

    <div class="source-modes">
      <button class:active={currentKind === 'literal'} on:click={switchToLiteral}
        >{$t('SETTINGS.THEME_EDITOR.literal')}</button
      >
      <button
        class:active={currentKind === 'ref'}
        on:click={switchToRef}
        disabled={!compatibleRefIDs.length}
      >
        {$t('SETTINGS.THEME_EDITOR.ref')}
      </button>
      {#if token.type === 'color'}
        <button class:active={currentKind === 'derived'} on:click={switchToDerived}
          >{$t('SETTINGS.THEME_EDITOR.derived')}</button
        >
      {/if}
    </div>

    {#if currentKind === 'ref'}
      <div class="panel">
        {#if compatibleRefIDs.length > 0}
          <label>
            {$t('SETTINGS.THEME_EDITOR.reference_token')}
            <select value={activeRefTargetID || compatibleRefIDs[0]} on:change={handleRefChange}>
              {#each compatibleRefIDs as refID}
                <option value={refID}>{refID}</option>
              {/each}
            </select>
          </label>
        {:else}
          <div class="placeholder">{$t('SETTINGS.THEME_EDITOR.no_compatible_ref_tokens')}</div>
        {/if}
      </div>
    {:else if currentKind === 'derived'}
      {#if token.type === 'color'}
        <div class="panel derived">
          <label>
            {$t('SETTINGS.THEME_EDITOR.operation')}
            <select value={derivedOp} on:change={handleDerivedOpChange}>
              {#each DERIVED_OPS as op}
                <option value={op}>{op}</option>
              {/each}
            </select>
          </label>

          <label>
            {$t('SETTINGS.THEME_EDITOR.from_token')}
            <select
              value={derivedFromRefID || colorRefIDs[0] || ''}
              on:change={handleDerivedFromChange}
              disabled={!colorRefIDs.length}
            >
              {#if !colorRefIDs.length}
                <option value="">{$t('SETTINGS.THEME_EDITOR.no_color_tokens')}</option>
              {:else}
                {#each colorRefIDs as refID}
                  <option value={refID}>{refID}</option>
                {/each}
              {/if}
            </select>
          </label>

          <label>
            {$t('SETTINGS.THEME_EDITOR.amount')}
            <input
              type="number"
              value={derivedAmount}
              step={derivedOp === 'alpha' ? '0.01' : '1'}
              min={derivedOp === 'alpha' ? '0' : undefined}
              max={derivedOp === 'alpha' ? '1' : undefined}
              on:change={handleDerivedAmountChange}
            />
          </label>

          <div class="hint">{$t('SETTINGS.THEME_EDITOR.derived_hint')}</div>
        </div>
      {:else}
        <div class="placeholder">{$t('SETTINGS.THEME_EDITOR.derived_only_for_color')}</div>
      {/if}
    {:else if token.type === 'color' && currentColor}
      <ColorEditor color={currentColor} on:change={(e) => updateColor(e.detail)} />
    {:else if isNumberLikeType && currentNumber}
      <NumberEditor number={currentNumber} on:change={(e) => updateNumber(e.detail)} />
    {:else if token.type === 'shadow'}
      <ShadowEditor layers={currentShadow} on:change={(e) => updateShadow(e.detail)} />
    {:else if token.type === 'gradient'}
      <GradientEditor gradient={currentGradient} on:change={(e) => updateGradient(e.detail)} />
    {:else if token.type === 'color'}
      <div class="placeholder">{$t('SETTINGS.THEME_EDITOR.unresolved_color')}</div>
    {:else if isNumberLikeType}
      <div class="placeholder">{$t('SETTINGS.THEME_EDITOR.unresolved_number')}</div>
    {:else}
      <div class="panel">
        <label class="json-label">
          {$t('SETTINGS.THEME_EDITOR.literal_json')}
          <textarea value={literalJsonText} on:input={handleLiteralJsonInput}></textarea>
        </label>
        <div class="json-actions">
          <button on:click={applyLiteralJson}>{$t('SETTINGS.THEME_EDITOR.apply_json')}</button>
          {#if literalJsonError}
            <span class="json-error">{literalJsonError}</span>
          {/if}
        </div>
      </div>
    {/if}
  {:else}
    <div class="placeholder">{$t('SETTINGS.THEME_EDITOR.select_token')}</div>
  {/if}
</div>

<style>
  .editor {
    padding: 16px;
    min-height: 0;
    overflow-y: auto;
    overscroll-behavior: contain;
  }

  h3 {
    margin-top: 0;
    font-family: monospace;
  }

  .meta {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 8px;
    margin-bottom: 14px;
  }

  .badge {
    padding: 4px 8px;
    border-radius: 999px;
    background: var(--surface-elevated, #2a2a2a);
    border: 1px solid var(--border-color, #444);
    font-size: 11px;
    color: var(--text-color-secondary, #ccc);
  }

  .state-select {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    font-size: 12px;
    color: var(--text-color-muted, #999);
  }

  .state-select select {
    min-width: 120px;
    padding: 4px 8px;
    border-radius: 6px;
    border: 1px solid var(--border-color, #444);
    background: var(--surface-base, #222);
    color: var(--text-color-primary, #fff);
  }

  .placeholder {
    color: var(--text-color-muted, #777);
  }

  .source-modes {
    display: flex;
    gap: 8px;
    margin-bottom: 12px;
  }

  .source-modes button {
    border: 1px solid var(--border-color, #444);
    background: var(--surface-base, #232323);
    color: var(--text-color-primary, #fff);
    border-radius: 6px;
    padding: 5px 10px;
    cursor: pointer;
    font-size: 12px;
  }

  .source-modes button.active {
    border-color: var(--action-primary-bg, #3ba475);
    background: rgba(59, 164, 117, 0.2);
  }

  .source-modes button:disabled {
    opacity: 0.5;
    cursor: default;
  }

  .panel {
    display: flex;
    flex-direction: column;
    gap: 10px;
    max-width: 460px;
  }

  .panel label {
    display: flex;
    flex-direction: column;
    gap: 6px;
    font-size: 12px;
    color: var(--text-color-muted, #999);
  }

  .panel select,
  .panel input,
  .panel textarea {
    width: 100%;
    padding: 8px 10px;
    border-radius: 6px;
    border: 1px solid var(--border-color, #444);
    background: var(--surface-base, #222);
    color: var(--text-color-primary, #fff);
    font-size: 13px;
  }

  .panel textarea {
    min-height: 180px;
    resize: vertical;
    font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, 'Liberation Mono', monospace;
  }

  .panel select:focus,
  .panel input:focus,
  .panel textarea:focus {
    outline: none;
    border-color: var(--action-primary-bg, #3ba475);
  }

  .derived {
    max-width: 560px;
  }

  .hint {
    font-size: 11px;
    color: var(--text-color-muted, #888);
  }

  .json-actions {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .json-actions button {
    border: 1px solid var(--border-color, #444);
    background: var(--surface-elevated, #2b2b2b);
    color: var(--text-color-primary, #fff);
    border-radius: 6px;
    padding: 6px 12px;
    cursor: pointer;
    font-size: 12px;
  }

  .json-actions button:hover {
    opacity: 0.9;
  }

  .json-error {
    font-size: 12px;
    color: var(--status-error-text, #fecaca);
  }
</style>
