<script>
  import { colord } from 'colord';
  import { createEventDispatcher, onMount } from 'svelte';
  import { t } from 'svelte-i18n';

  export let color;
  // { space: 'srgb', r, g, b, a }

  const dispatch = createEventDispatcher();

  // HSV state
  let h = 0;
  let s = 100;
  let v = 100;
  let a = 1;

  let textInput = '';
  let textInputError = false;

  // Hex input
  let hexInput = '#ff0000';

  // Refs
  let satValBox;
  let isDraggingSatVal = false;
  let isDraggingHue = false;
  let isDraggingAlpha = false;

  // Track if we're currently editing to prevent prop sync
  $: isEditing = isDraggingSatVal || isDraggingHue || isDraggingAlpha;

  // Sync from prop to internal state (only when not editing)
  $: if (color && !isEditing) {
    const c = colord({
      r: color.r,
      g: color.g,
      b: color.b,
      a: color.a ?? 1
    });

    if (c.isValid()) {
      const hsv = c.toHsv();

      h = hsv.h;
      s = hsv.s;
      v = hsv.v;
      a = hsv.a;

      textInput = c.toRgbString();
      hexInput = c.toHex();
      textInputError = false;
    }
  }

  // Pure hue color for saturation-value gradient
  $: hueColor = colord({ h, s: 100, v: 100, a: 1 }).toHex();

  // Current RGB from HSV
  $: currentRgb = colord({ h, s, v, a }).toRgb();
  $: previewStyle = `rgba(${currentRgb.r}, ${currentRgb.g}, ${currentRgb.b}, ${a})`;

  function emitChange() {
    dispatch('change', {
      space: 'srgb',
      r: currentRgb.r,
      g: currentRgb.g,
      b: currentRgb.b,
      a: a
    });
  }

  // Saturation-Value box handlers
  function handleSatValStart(e) {
    isDraggingSatVal = true;
    updateSatVal(e);
  }

  function updateSatVal(e) {
    if (!satValBox) return;
    const rect = satValBox.getBoundingClientRect();
    const x = Math.max(0, Math.min(e.clientX - rect.left, rect.width));
    const y = Math.max(0, Math.min(e.clientY - rect.top, rect.height));
    s = (x / rect.width) * 100;
    v = 100 - (y / rect.height) * 100;
  }

  function handleSatValMove(e) {
    if (isDraggingSatVal) {
      updateSatVal(e);
    }
  }

  function handleSatValEnd() {
    if (isDraggingSatVal) {
      isDraggingSatVal = false;
      emitChange();
    }
  }

  // Hue slider handlers
  function handleHueStart(e) {
    isDraggingHue = true;
    updateHue(e);
  }

  function updateHue(e) {
    const slider = e.target.closest('.hue-slider');
    if (!slider) return;
    const rect = slider.getBoundingClientRect();
    const x = Math.max(0, Math.min(e.clientX - rect.left, rect.width));
    h = (x / rect.width) * 360;
  }

  function handleHueMove(e) {
    if (isDraggingHue) {
      const rect = e.target.closest('.hue-slider').getBoundingClientRect();
      const x = Math.max(0, Math.min(e.clientX - rect.left, rect.width));
      h = (x / rect.width) * 360;
    }
  }

  function handleHueEnd() {
    if (isDraggingHue) {
      isDraggingHue = false;
      emitChange();
    }
  }

  // Alpha slider handlers
  function handleAlphaStart(e) {
    isDraggingAlpha = true;
    updateAlpha(e);
  }

  function updateAlpha(e) {
    const slider = e.target.closest('.alpha-slider');
    if (!slider) return;
    const rect = slider.getBoundingClientRect();
    const x = Math.max(0, Math.min(e.clientX - rect.left, rect.width));
    a = x / rect.width;
  }

  function handleAlphaMove(e) {
    if (isDraggingAlpha) {
      const rect = e.target.closest('.alpha-slider').getBoundingClientRect();
      const x = Math.max(0, Math.min(e.clientX - rect.left, rect.width));
      a = x / rect.width;
    }
  }

  function handleAlphaEnd() {
    if (isDraggingAlpha) {
      isDraggingAlpha = false;
      emitChange();
    }
  }

  // Global mouse tracking
  function handleGlobalMove(e) {
    handleSatValMove(e);
    handleHueMove(e);
    handleAlphaMove(e);
  }

  function handleGlobalUp() {
    handleSatValEnd();
    handleHueEnd();
    handleAlphaEnd();
  }

  function normalizeColorInput(input) {
    if (!input) return input;

    let v = input.trim();

    // Если просто числа: "0,0,0" или "0, 0, 0, 0.5"
    if (/^\(?\s*\d+(\s*,\s*\d+){2,3}\s*\)?$/.test(v)) {
      v = v.replace(/[()]/g, '');
      const parts = v.split(',').map((p) => p.trim());

      if (parts.length === 3) {
        return `rgb(${parts.join(',')})`;
      }

      if (parts.length === 4) {
        return `rgba(${parts.join(',')})`;
      }
    }

    // "255 255 255" или "255 255 255 0.5"
    if (/^\d+(\s+\d+){2,3}$/.test(v)) {
      const parts = v.split(/\s+/);

      if (parts.length === 3) {
        return `rgb(${parts.join(',')})`;
      }

      if (parts.length === 4) {
        return `rgba(${parts.join(',')})`;
      }
    }

    // Уже rgb(...) / rgba(...) / hex / hsl / etc
    return v;
  }

  function handleTextInput(e) {
    const raw = e.target.value;
    textInput = raw;

    const normalized = normalizeColorInput(raw);
    const c = colord(normalized);

    if (!c.isValid()) {
      textInputError = true;
      return;
    }

    textInputError = false;

    const hsv = c.toHsv();

    h = hsv.h;
    s = hsv.s;
    v = hsv.v;
    a = hsv.a;

    emitChange();
  }

  function handleHexInput(e) {
    const raw = String(e.target.value || '').trim();
    hexInput = raw;

    const c = colord(raw);
    if (!c.isValid()) {
      textInputError = true;
      return;
    }

    textInputError = false;
    const hsv = c.toHsv();
    h = hsv.h;
    s = hsv.s;
    v = hsv.v;
    emitChange();
  }

  function handleNativeColorInput(e) {
    const raw = String(e.target.value || '').trim();
    hexInput = raw;

    const c = colord(raw);
    if (!c.isValid()) return;

    const hsv = c.toHsv();
    h = hsv.h;
    s = hsv.s;
    v = hsv.v;
    emitChange();
  }

  function handleAlphaInput(e) {
    const value = Number(e.target.value);
    if (!Number.isFinite(value)) return;
    a = Math.max(0, Math.min(value, 1));
    emitChange();
  }

  onMount(() => {
    window.addEventListener('mousemove', handleGlobalMove);
    window.addEventListener('mouseup', handleGlobalUp);

    return () => {
      window.removeEventListener('mousemove', handleGlobalMove);
      window.removeEventListener('mouseup', handleGlobalUp);
    };
  });
</script>

<div class="color-picker">
  <!-- Saturation-Value Box -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    class="sat-val-box"
    bind:this={satValBox}
    style="--hue-color: {hueColor}"
    on:mousedown={handleSatValStart}
  >
    <div class="sat-val-cursor" style="left: {s}%; top: {100 - v}%"></div>
  </div>

  <!-- Sliders -->
  <div class="sliders">
    <!-- Hue Slider -->
    <div class="slider-row">
      <span class="slider-label">{$t('SETTINGS.THEME_EDITOR.hue_short')}</span>
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <div class="hue-slider" on:mousedown={handleHueStart}>
        <div class="hue-cursor" style="left: {(h / 360) * 100}%"></div>
      </div>
    </div>

    <!-- Alpha Slider -->
    <div class="slider-row">
      <span class="slider-label">{$t('SETTINGS.THEME_EDITOR.alpha_short')}</span>
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <div class="alpha-slider" on:mousedown={handleAlphaStart}>
        <div class="alpha-gradient" style="--current-color: {previewStyle}"></div>
        <div class="alpha-cursor" style="left: {a * 100}%"></div>
      </div>
    </div>
  </div>

  <!-- Preview & Inputs -->
  <div class="inputs-row">
    <div class="preview" style="background-color: {previewStyle}"></div>

    <div class="inputs-col">
      <label class="text-input">
        {$t('SETTINGS.THEME_EDITOR.color')}
        <input
          type="text"
          class:error={textInputError}
          value={textInput}
          on:input={handleTextInput}
          placeholder={$t('SETTINGS.THEME_EDITOR.color_input_placeholder')}
        />
      </label>

      <div class="quick-row">
        <label class="mini-input">
          {$t('CONTROLS.hex')}
          <input
            type="text"
            value={hexInput}
            on:input={handleHexInput}
            placeholder={$t('SETTINGS.THEME_EDITOR.hex_input_placeholder')}
          />
        </label>

        <label class="mini-input">
          {$t('SETTINGS.THEME_EDITOR.pick')}
          <input type="color" value={hexInput} on:input={handleNativeColorInput} />
        </label>

        <label class="mini-input alpha">
          {$t('SETTINGS.THEME_EDITOR.alpha')}
          <input type="number" min="0" max="1" step="0.01" value={a} on:change={handleAlphaInput} />
        </label>
      </div>
    </div>
  </div>
</div>

<style>
  .color-picker {
    display: flex;
    flex-direction: column;
    gap: 16px;
    width: 100%;
    max-width: 400px;
    user-select: none;
  }

  /* Saturation-Value Box */
  .sat-val-box {
    position: relative;
    width: 100%;
    height: 260px;
    border-radius: 8px;
    cursor: crosshair;
    background:
      linear-gradient(to top, #000, transparent), linear-gradient(to right, #fff, transparent),
      var(--hue-color);
  }

  .sat-val-cursor {
    position: absolute;
    width: 18px;
    height: 18px;
    border: 2px solid var(--text-color-primary, #fff);
    border-radius: 50%;
    box-shadow: 0 0 4px rgba(0, 0, 0, 0.5);
    transform: translate(-50%, -50%);
    pointer-events: none;
  }

  /* Sliders */
  .sliders {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .slider-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .slider-label {
    width: 14px;
    font-size: 11px;
    color: var(--text-color-muted, #999);
    font-weight: 600;
  }

  .hue-slider {
    position: relative;
    flex: 1;
    height: 20px;
    border-radius: 10px;
    cursor: pointer;
    background: linear-gradient(to right, #f00, #ff0, #0f0, #0ff, #00f, #f0f, #f00);
  }

  .hue-cursor {
    position: absolute;
    top: 50%;
    width: 18px;
    height: 18px;
    background: var(--text-color-primary, #fff);
    border-radius: 50%;
    box-shadow: 0 0 4px rgba(0, 0, 0, 0.4);
    transform: translate(-50%, -50%);
    pointer-events: none;
  }

  .alpha-slider {
    position: relative;
    flex: 1;
    height: 20px;
    border-radius: 10px;
    cursor: pointer;
    background-image:
      linear-gradient(45deg, var(--border-color, #444) 25%, transparent 25%),
      linear-gradient(-45deg, var(--border-color, #444) 25%, transparent 25%),
      linear-gradient(45deg, transparent 75%, var(--border-color, #444) 75%),
      linear-gradient(-45deg, transparent 75%, var(--border-color, #444) 75%);
    background-size: 10px 10px;
    background-position:
      0 0,
      0 5px,
      5px -5px,
      -5px 0px;
    background-color: var(--bg-color-dark, #222);
    overflow: hidden;
  }

  .alpha-gradient {
    position: absolute;
    inset: 0;
    background: linear-gradient(to right, transparent, var(--current-color));
    border-radius: 10px;
  }

  .alpha-cursor {
    position: absolute;
    top: 50%;
    width: 18px;
    height: 18px;
    background: var(--text-color-primary, #fff);
    border-radius: 50%;
    box-shadow: 0 0 4px rgba(0, 0, 0, 0.4);
    transform: translate(-50%, -50%);
    pointer-events: none;
  }

  /* Inputs Row */
  .inputs-row {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .preview {
    width: 48px;
    height: 48px;
    border-radius: 8px;
    border: 2px solid var(--border-color, #444);
    flex-shrink: 0;
  }

  .text-input {
    display: flex;
    flex-direction: column;
    gap: 4px;
    font-size: 10px;
    color: var(--text-color-muted, #888);
  }

  .inputs-col {
    display: flex;
    flex-direction: column;
    gap: 8px;
    min-width: 0;
    flex: 1;
  }

  .quick-row {
    display: grid;
    grid-template-columns: 1fr auto auto;
    gap: 8px;
    align-items: end;
  }

  .mini-input {
    display: flex;
    flex-direction: column;
    gap: 4px;
    font-size: 10px;
    color: var(--text-color-muted, #888);
  }

  .mini-input input[type='color'] {
    width: 48px;
    height: 34px;
    padding: 2px;
    border-radius: 6px;
    border: 1px solid var(--border-color, #444);
    background: var(--bg-color-medium, #2a2a2a);
  }

  .mini-input.alpha {
    width: 90px;
  }

  .text-input input {
    padding: 6px 8px;
    background: var(--bg-color-medium, #2a2a2a);
    border: 1px solid var(--border-color, #444);
    border-radius: 4px;
    color: var(--text-color-primary, #fff);
    font-size: 12px;
    font-family: monospace;
  }

  .text-input input.error {
    border-color: var(--status-error-border, #c93838);
  }

  input:focus {
    outline: none;
    border-color: var(--action-primary-bg, #3ba475);
  }
</style>
