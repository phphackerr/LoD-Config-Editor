<script>
  import { createEventDispatcher } from 'svelte';
  import { colord } from 'colord';
  import { t } from 'svelte-i18n';

  export let layers = [];

  const dispatch = createEventDispatcher();

  let localLayers = [];
  let syncKey = '';

  $: nextSyncKey = JSON.stringify(Array.isArray(layers) ? layers : []);
  $: if (nextSyncKey !== syncKey) {
    syncKey = nextSyncKey;
    localLayers = normalizeLayers(layers);
  }

  function clamp(value, min, max) {
    if (value < min) return min;
    if (value > max) return max;
    return value;
  }

  function toFinite(value, fallback) {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : fallback;
  }

  function defaultColor() {
    return {
      space: 'srgb',
      r: 0,
      g: 0,
      b: 0,
      a: 0.35
    };
  }

  function defaultLayer() {
    return {
      x: 0,
      y: 2,
      blur: 8,
      spread: 0,
      inset: false,
      color: defaultColor()
    };
  }

  function normalizeColor(raw) {
    return {
      space: 'srgb',
      r: clamp(toFinite(raw?.r, 0), 0, 255),
      g: clamp(toFinite(raw?.g, 0), 0, 255),
      b: clamp(toFinite(raw?.b, 0), 0, 255),
      a: clamp(toFinite(raw?.a, 1), 0, 1)
    };
  }

  function normalizeLayer(raw) {
    const color = raw?.color?.literal?.color || raw?.color || defaultColor();
    return {
      x: toFinite(raw?.x, 0),
      y: toFinite(raw?.y, 0),
      blur: Math.max(0, toFinite(raw?.blur, 0)),
      spread: toFinite(raw?.spread, 0),
      inset: Boolean(raw?.inset),
      color: normalizeColor(color)
    };
  }

  function normalizeLayers(input) {
    if (!Array.isArray(input) || input.length === 0) {
      return [defaultLayer()];
    }

    return input.map((entry) => normalizeLayer(entry));
  }

  function toPayloadLayer(layer) {
    return {
      x: toFinite(layer?.x, 0),
      y: toFinite(layer?.y, 0),
      blur: Math.max(0, toFinite(layer?.blur, 0)),
      spread: toFinite(layer?.spread, 0),
      inset: Boolean(layer?.inset),
      color: {
        literal: {
          color: normalizeColor(layer?.color)
        }
      }
    };
  }

  function emitChange() {
    dispatch(
      'change',
      localLayers.map((layer) => toPayloadLayer(layer))
    );
  }

  function updateNumeric(index, field, value) {
    const next = [...localLayers];
    const base = next[index];
    if (!base) return;

    const parsed = toFinite(value, 0);
    next[index] = {
      ...base,
      [field]: field === 'blur' ? Math.max(0, parsed) : parsed
    };
    localLayers = next;
    emitChange();
  }

  function updateInset(index, checked) {
    const next = [...localLayers];
    const base = next[index];
    if (!base) return;

    next[index] = {
      ...base,
      inset: Boolean(checked)
    };
    localLayers = next;
    emitChange();
  }

  function updateColorHex(index, hex) {
    const rgb = colord(hex).toRgb();
    if (!colord(hex).isValid()) return;

    const next = [...localLayers];
    const base = next[index];
    if (!base) return;

    next[index] = {
      ...base,
      color: {
        ...base.color,
        r: rgb.r,
        g: rgb.g,
        b: rgb.b
      }
    };
    localLayers = next;
    emitChange();
  }

  function updateColorAlpha(index, alpha) {
    const next = [...localLayers];
    const base = next[index];
    if (!base) return;

    next[index] = {
      ...base,
      color: {
        ...base.color,
        a: clamp(toFinite(alpha, 1), 0, 1)
      }
    };
    localLayers = next;
    emitChange();
  }

  function addLayer() {
    localLayers = [...localLayers, defaultLayer()];
    emitChange();
  }

  function removeLayer(index) {
    if (localLayers.length <= 1) return;
    localLayers = localLayers.filter((_, i) => i !== index);
    emitChange();
  }

  function colorHex(color) {
    return colord({
      r: clamp(toFinite(color?.r, 0), 0, 255),
      g: clamp(toFinite(color?.g, 0), 0, 255),
      b: clamp(toFinite(color?.b, 0), 0, 255),
      a: 1
    }).toHex();
  }

  function shadowPreview(layer) {
    const c = normalizeColor(layer?.color);
    const inset = layer?.inset ? 'inset ' : '';
    return `${inset}${toFinite(layer?.x, 0)}px ${toFinite(layer?.y, 0)}px ${Math.max(
      0,
      toFinite(layer?.blur, 0)
    )}px ${toFinite(layer?.spread, 0)}px rgba(${Math.round(c.r)}, ${Math.round(c.g)}, ${Math.round(c.b)}, ${c.a})`;
  }
</script>

<div class="shadow-editor">
  {#each localLayers as layer, index}
    <section class="layer">
      <div class="layer-head">
        <strong>{$t('SETTINGS.THEME_EDITOR.layer')} {index + 1}</strong>
        <button on:click={() => removeLayer(index)} disabled={localLayers.length <= 1}
          >{$t('COMMON.remove')}</button
        >
      </div>

      <div class="grid">
        <label>
          X
          <input
            type="number"
            value={layer.x}
            step="0.1"
            on:change={(e) => updateNumeric(index, 'x', e.currentTarget?.value)}
          />
        </label>
        <label>
          Y
          <input
            type="number"
            value={layer.y}
            step="0.1"
            on:change={(e) => updateNumeric(index, 'y', e.currentTarget?.value)}
          />
        </label>
        <label>
          {$t('SETTINGS.THEME_EDITOR.blur')}
          <input
            type="number"
            min="0"
            value={layer.blur}
            step="0.1"
            on:change={(e) => updateNumeric(index, 'blur', e.currentTarget?.value)}
          />
        </label>
        <label>
          {$t('SETTINGS.THEME_EDITOR.spread')}
          <input
            type="number"
            value={layer.spread}
            step="0.1"
            on:change={(e) => updateNumeric(index, 'spread', e.currentTarget?.value)}
          />
        </label>
      </div>

      <div class="color-row">
        <label>
          {$t('SETTINGS.THEME_EDITOR.color')}
          <input
            type="color"
            value={colorHex(layer.color)}
            on:input={(e) => updateColorHex(index, e.currentTarget?.value)}
          />
        </label>

        <label>
          {$t('SETTINGS.THEME_EDITOR.alpha')}
          <input
            type="number"
            min="0"
            max="1"
            step="0.01"
            value={layer.color.a}
            on:change={(e) => updateColorAlpha(index, e.currentTarget?.value)}
          />
        </label>

        <label class="inset">
          <input
            type="checkbox"
            checked={layer.inset}
            on:change={(e) => updateInset(index, e.currentTarget?.checked)}
          />
          {$t('SETTINGS.THEME_EDITOR.inset')}
        </label>
      </div>

      <div class="preview-row">
        <div class="preview" style="box-shadow: {shadowPreview(layer)}"></div>
        <code>{shadowPreview(layer)}</code>
      </div>
    </section>
  {/each}

  <button class="add" on:click={addLayer}>{$t('SETTINGS.THEME_EDITOR.add_layer')}</button>
</div>

<style>
  .shadow-editor {
    display: flex;
    flex-direction: column;
    gap: 12px;
    max-width: 720px;
  }

  .layer {
    border: 1px solid var(--border-color, #444);
    border-radius: 10px;
    padding: 10px;
    background: var(--surface-base, #1f1f1f);
  }

  .layer-head {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 10px;
  }

  .layer-head strong {
    font-size: 13px;
  }

  .grid {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 8px;
  }

  label {
    display: flex;
    flex-direction: column;
    gap: 6px;
    font-size: 12px;
    color: var(--text-color-muted, #999);
  }

  input[type='number'],
  input[type='color'] {
    width: 100%;
    padding: 6px 8px;
    border-radius: 6px;
    border: 1px solid var(--border-color, #444);
    background: var(--surface-elevated, #2a2a2a);
    color: var(--text-color-primary, #fff);
  }

  .color-row {
    display: grid;
    grid-template-columns: 1fr 1fr auto;
    gap: 8px;
    margin-top: 10px;
    align-items: end;
  }

  .inset {
    align-self: center;
    display: inline-flex;
    flex-direction: row;
    gap: 8px;
    align-items: center;
  }

  .preview-row {
    margin-top: 10px;
    display: grid;
    grid-template-columns: 96px 1fr;
    gap: 10px;
    align-items: center;
  }

  .preview {
    height: 56px;
    border-radius: 8px;
    background: linear-gradient(135deg, #2e2e2e, #1a1a1a);
    border: 1px solid var(--border-color, #444);
  }

  code {
    display: block;
    overflow-wrap: anywhere;
    font-size: 11px;
    color: var(--text-color-secondary, #ccc);
  }

  button {
    border: 1px solid var(--border-color, #444);
    background: var(--surface-elevated, #2b2b2b);
    color: var(--text-color-primary, #fff);
    border-radius: 6px;
    padding: 6px 10px;
    cursor: pointer;
    font-size: 12px;
  }

  button:disabled {
    opacity: 0.5;
    cursor: default;
  }

  .add {
    align-self: flex-start;
  }
</style>
