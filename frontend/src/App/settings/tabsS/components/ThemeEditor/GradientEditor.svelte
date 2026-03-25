<script>
  import { createEventDispatcher } from 'svelte';
  import { colord } from 'colord';
  import { t } from 'svelte-i18n';

  export let gradient = null;

  const dispatch = createEventDispatcher();
  const SUPPORTED_KINDS = ['linear'];

  let model = null;
  let syncKey = '';

  $: nextSyncKey = JSON.stringify(gradient || null);
  $: if (nextSyncKey !== syncKey) {
    syncKey = nextSyncKey;
    model = normalizeGradient(gradient);
  }

  $: preview = gradientPreview(model);

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
      r: 255,
      g: 255,
      b: 255,
      a: 1
    };
  }

  function defaultStop(pos, color) {
    return {
      pos: clamp(toFinite(pos, 0), 0, 1),
      color: normalizeColor(color || defaultColor())
    };
  }

  function normalizeColor(raw) {
    return {
      space: 'srgb',
      r: clamp(toFinite(raw?.r, 255), 0, 255),
      g: clamp(toFinite(raw?.g, 255), 0, 255),
      b: clamp(toFinite(raw?.b, 255), 0, 255),
      a: clamp(toFinite(raw?.a, 1), 0, 1)
    };
  }

  function normalizeStop(raw, fallbackPos = 0) {
    const color = raw?.color?.literal?.color || raw?.color || defaultColor();
    return {
      pos: clamp(toFinite(raw?.pos, fallbackPos), 0, 1),
      color: normalizeColor(color)
    };
  }

  function normalizeGradient(raw) {
    const kind = SUPPORTED_KINDS.includes(raw?.kind) ? raw.kind : 'linear';
    const angle = toFinite(raw?.angle, 180);
    const incomingStops = Array.isArray(raw?.stops) ? raw.stops : [];
    const stops =
      incomingStops.length >= 2
        ? incomingStops.map((stop, index) =>
            normalizeStop(stop, index / Math.max(incomingStops.length - 1, 1))
          )
        : [
            defaultStop(0, { r: 255, g: 255, b: 255, a: 1 }),
            defaultStop(1, { r: 0, g: 0, b: 0, a: 1 })
          ];

    return {
      kind,
      angle,
      stops: stops.sort((a, b) => a.pos - b.pos)
    };
  }

  function toPayload(modelValue) {
    return {
      kind: modelValue.kind,
      angle: toFinite(modelValue.angle, 180),
      stops: modelValue.stops
        .map((stop) => ({
          pos: clamp(toFinite(stop.pos, 0), 0, 1),
          color: {
            literal: {
              color: normalizeColor(stop.color)
            }
          }
        }))
        .sort((a, b) => a.pos - b.pos)
    };
  }

  function emitChange(nextModel = model) {
    dispatch('change', toPayload(nextModel));
  }

  function updateKind(kind) {
    model = {
      ...model,
      kind: SUPPORTED_KINDS.includes(kind) ? kind : 'linear'
    };
    emitChange();
  }

  function updateAngle(value) {
    model = {
      ...model,
      angle: toFinite(value, 180)
    };
    emitChange();
  }

  function updateStopPos(index, value) {
    const stops = [...model.stops];
    const current = stops[index];
    if (!current) return;

    stops[index] = {
      ...current,
      pos: clamp(toFinite(value, current.pos), 0, 1)
    };
    model = {
      ...model,
      stops: stops.sort((a, b) => a.pos - b.pos)
    };
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

  function updateStopColor(index, hex) {
    const c = colord(hex);
    if (!c.isValid()) return;
    const rgb = c.toRgb();
    const stops = [...model.stops];
    const current = stops[index];
    if (!current) return;

    stops[index] = {
      ...current,
      color: {
        ...current.color,
        r: rgb.r,
        g: rgb.g,
        b: rgb.b
      }
    };
    model = { ...model, stops };
    emitChange();
  }

  function updateStopAlpha(index, alpha) {
    const stops = [...model.stops];
    const current = stops[index];
    if (!current) return;

    stops[index] = {
      ...current,
      color: {
        ...current.color,
        a: clamp(toFinite(alpha, current.color.a), 0, 1)
      }
    };
    model = { ...model, stops };
    emitChange();
  }

  function addStop() {
    const stops = [...model.stops];
    const midpoint = stops.length > 0 ? clamp((stops[stops.length - 1].pos + 1) / 2, 0, 1) : 0.5;
    stops.push(defaultStop(midpoint, defaultColor()));
    model = {
      ...model,
      stops: stops.sort((a, b) => a.pos - b.pos)
    };
    emitChange();
  }

  function removeStop(index) {
    if (model.stops.length <= 2) return;
    model = {
      ...model,
      stops: model.stops.filter((_, i) => i !== index)
    };
    emitChange();
  }

  function gradientPreview(modelValue) {
    if (!modelValue) return 'linear-gradient(180deg, #fff 0%, #000 100%)';
    const stops = modelValue.stops
      .map((stop) => {
        const c = normalizeColor(stop.color);
        const pos = Math.round(clamp(toFinite(stop.pos, 0), 0, 1) * 100);
        return `rgba(${Math.round(c.r)}, ${Math.round(c.g)}, ${Math.round(c.b)}, ${c.a}) ${pos}%`;
      })
      .join(', ');
    return `linear-gradient(${toFinite(modelValue.angle, 180)}deg, ${stops})`;
  }
</script>

<div class="gradient-editor">
  <div class="controls">
    <label>
      {$t('SETTINGS.THEME_EDITOR.kind')}
      <select value={model.kind} on:change={(e) => updateKind(e.currentTarget?.value)}>
        {#each SUPPORTED_KINDS as kind}
          <option value={kind}>{kind}</option>
        {/each}
      </select>
    </label>

    <label>
      {$t('SETTINGS.THEME_EDITOR.angle')}
      <input
        type="number"
        value={model.angle}
        step="1"
        on:change={(e) => updateAngle(e.currentTarget?.value)}
      />
    </label>
  </div>

  <div class="preview" style="background: {preview}"></div>

  <div class="stops">
    {#each model.stops as stop, index}
      <section class="stop">
        <div class="stop-head">
          <strong>{$t('SETTINGS.THEME_EDITOR.stop')} {index + 1}</strong>
          <button on:click={() => removeStop(index)} disabled={model.stops.length <= 2}
            >{$t('COMMON.remove')}</button
          >
        </div>

        <div class="stop-grid">
          <label>
            {$t('SETTINGS.THEME_EDITOR.position')}
            <input
              type="number"
              min="0"
              max="1"
              step="0.01"
              value={stop.pos}
              on:change={(e) => updateStopPos(index, e.currentTarget?.value)}
            />
          </label>

          <label>
            {$t('SETTINGS.THEME_EDITOR.color')}
            <input
              type="color"
              value={colorHex(stop.color)}
              on:input={(e) => updateStopColor(index, e.currentTarget?.value)}
            />
          </label>

          <label>
            {$t('SETTINGS.THEME_EDITOR.alpha')}
            <input
              type="number"
              min="0"
              max="1"
              step="0.01"
              value={stop.color.a}
              on:change={(e) => updateStopAlpha(index, e.currentTarget?.value)}
            />
          </label>
        </div>
      </section>
    {/each}
  </div>

  <button class="add" on:click={addStop}>{$t('SETTINGS.THEME_EDITOR.add_stop')}</button>
</div>

<style>
  .gradient-editor {
    display: flex;
    flex-direction: column;
    gap: 12px;
    max-width: 720px;
  }

  .controls {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 8px;
  }

  label {
    display: flex;
    flex-direction: column;
    gap: 6px;
    font-size: 12px;
    color: var(--text-color-muted, #999);
  }

  select,
  input[type='number'],
  input[type='color'] {
    width: 100%;
    padding: 6px 8px;
    border-radius: 6px;
    border: 1px solid var(--border-color, #444);
    background: var(--surface-elevated, #2a2a2a);
    color: var(--text-color-primary, #fff);
  }

  .preview {
    height: 80px;
    border-radius: 10px;
    border: 1px solid var(--border-color, #444);
  }

  .stops {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .stop {
    border: 1px solid var(--border-color, #444);
    border-radius: 10px;
    padding: 10px;
    background: var(--surface-base, #1f1f1f);
  }

  .stop-head {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 8px;
  }

  .stop-grid {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 8px;
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
