<script>
  import { onDestroy, tick } from 'svelte';
  import Portal from 'svelte-portal';
  import { t } from 'svelte-i18n';
  import { tt } from '../../lib/tooltip';

  import { configStore, getConfigValue } from '../../lib/store/config';
  import { persistControlValue } from './lib/controlPipeline';
  import Base from './Base.svelte';

  export let label = '';
  export let section = '';
  export let option = '';
  export let ttKey = '';
  export let ttImage = '';
  export let ttPlace = '';
  export let visible = true;
  export let onUpdate = null;

  let hex = '#ffffff';
  let draftHex = '#ffffff';
  let hexInput = 'FFFFFF';
  let hsv = { h: 0, s: 0, v: 100 };

  let prevConfigData = null;
  let configAvailable = false;
  let wrapperClass = 'color-picker-wrapper';

  let isOpen = false;
  let dragMode = null;

  let triggerElement;
  let popoverElement;
  let svAreaElement;

  let popoverLeft = 0;
  let popoverTop = 0;

  $: {
    const storeValue = $configStore;
    configAvailable = !!storeValue?.path && storeValue.error === null;
    wrapperClass = `color-picker-wrapper ${!configAvailable ? 'disabled' : ''}`;
  }

  function clamp(value, min, max) {
    return Math.min(Math.max(value, min), max);
  }

  function normalizeHex(value) {
    const cleaned = String(value || '')
      .replace(/[^0-9a-fA-F]/g, '')
      .slice(-6);
    if (!cleaned) return 'FFFFFF';
    return cleaned.toUpperCase();
  }

  function hexToRgb(hexValue) {
    const normalized = normalizeHex(hexValue);
    return {
      r: parseInt(normalized.slice(0, 2), 16),
      g: parseInt(normalized.slice(2, 4), 16),
      b: parseInt(normalized.slice(4, 6), 16)
    };
  }

  function rgbToHex(r, g, b) {
    return [r, g, b]
      .map((ch) => clamp(Math.round(ch), 0, 255).toString(16).padStart(2, '0'))
      .join('')
      .toUpperCase();
  }

  function rgbToHsv(r, g, b) {
    const rn = r / 255;
    const gn = g / 255;
    const bn = b / 255;

    const max = Math.max(rn, gn, bn);
    const min = Math.min(rn, gn, bn);
    const d = max - min;

    let h = 0;
    if (d !== 0) {
      if (max === rn) {
        h = ((gn - bn) / d) % 6;
      } else if (max === gn) {
        h = (bn - rn) / d + 2;
      } else {
        h = (rn - gn) / d + 4;
      }
      h *= 60;
      if (h < 0) h += 360;
    }

    const s = max === 0 ? 0 : (d / max) * 100;
    const v = max * 100;

    return { h, s, v };
  }

  function hsvToRgb(h, s, v) {
    const sat = clamp(s, 0, 100) / 100;
    const val = clamp(v, 0, 100) / 100;

    const c = val * sat;
    const hh = (h % 360) / 60;
    const x = c * (1 - Math.abs((hh % 2) - 1));

    let rn = 0;
    let gn = 0;
    let bn = 0;

    if (hh >= 0 && hh < 1) {
      rn = c;
      gn = x;
    } else if (hh < 2) {
      rn = x;
      gn = c;
    } else if (hh < 3) {
      gn = c;
      bn = x;
    } else if (hh < 4) {
      gn = x;
      bn = c;
    } else if (hh < 5) {
      rn = x;
      bn = c;
    } else {
      rn = c;
      bn = x;
    }

    const m = val - c;

    return {
      r: Math.round((rn + m) * 255),
      g: Math.round((gn + m) * 255),
      b: Math.round((bn + m) * 255)
    };
  }

  function syncDraftFromHex(value) {
    const normalized = normalizeHex(value);
    const rgb = hexToRgb(normalized);

    draftHex = `#${normalized.toLowerCase()}`;
    hexInput = normalized;
    hsv = rgbToHsv(rgb.r, rgb.g, rgb.b);
  }

  function setLocalHex(value) {
    const normalized = normalizeHex(value);
    hex = `#${normalized.toLowerCase()}`;

    if (!isOpen) {
      syncDraftFromHex(normalized);
    }
  }

  function syncDraftFromHSV() {
    const rgb = hsvToRgb(hsv.h, hsv.s, hsv.v);
    const next = rgbToHex(rgb.r, rgb.g, rgb.b);
    draftHex = `#${next.toLowerCase()}`;
    hexInput = next;
  }

  async function saveHex(newHex) {
    if (!configAvailable) return;

    let colorToSave = normalizeHex(newHex);
    if (section === 'DOTA2HPBARSOPTIONS') {
      colorToSave = 'FF' + colorToSave;
    }

    const result = await persistControlValue(
      section,
      option,
      colorToSave,
      $t('ERRORS.controls.save_color')
    );
    if (!result.ok) {
      await loadValue();
      return;
    }

    const savedHex = colorToSave.slice(-6);
    setLocalHex(savedHex);
    onUpdate?.(`#${savedHex.toLowerCase()}`);
  }

  async function loadValue() {
    if (!configAvailable) {
      setLocalHex('FFFFFF');
      return;
    }

    const value = await getConfigValue(section, option);
    const stored = section === 'DOTA2HPBARSOPTIONS' ? String(value || '').slice(-6) : value;
    setLocalHex(stored || 'FFFFFF');
  }

  function cleanupGlobalListeners() {
    window.removeEventListener('pointerdown', handleGlobalPointerDown, true);
    window.removeEventListener('keydown', handleGlobalKeyDown);
    window.removeEventListener('resize', updatePopoverPosition);
    window.removeEventListener('scroll', updatePopoverPosition, true);
  }

  function stopPointerDrag() {
    dragMode = null;
    window.removeEventListener('pointermove', handlePointerMove);
    window.removeEventListener('pointerup', stopPointerDrag);
  }

  function closePicker({ resetDraft = true } = {}) {
    isOpen = false;
    cleanupGlobalListeners();
    stopPointerDrag();

    if (resetDraft) {
      syncDraftFromHex(hex);
    }
  }

  async function applyPicker() {
    await saveHex(draftHex);
    closePicker({ resetDraft: false });
  }

  async function openPicker() {
    if (!configAvailable || isOpen) return;

    syncDraftFromHex(hex);
    isOpen = true;

    await tick();
    updatePopoverPosition();

    window.addEventListener('pointerdown', handleGlobalPointerDown, true);
    window.addEventListener('keydown', handleGlobalKeyDown);
    window.addEventListener('resize', updatePopoverPosition);
    window.addEventListener('scroll', updatePopoverPosition, true);
  }

  function togglePicker() {
    if (isOpen) {
      closePicker();
      return;
    }

    void openPicker();
  }

  function handleGlobalPointerDown(event) {
    if (!isOpen) return;

    const target = event.target;
    if (popoverElement?.contains(target) || triggerElement?.contains(target)) {
      return;
    }

    closePicker();
  }

  function handleGlobalKeyDown(event) {
    if (!isOpen) return;

    if (event.key === 'Escape') {
      event.preventDefault();
      closePicker();
    }
  }

  function updatePopoverPosition() {
    if (!triggerElement) return;

    const rect = triggerElement.getBoundingClientRect();
    const popWidth = popoverElement?.offsetWidth ?? 296;
    const popHeight = popoverElement?.offsetHeight ?? 336;

    let left = rect.left;
    let top = rect.bottom + 8;

    if (left + popWidth > window.innerWidth - 8) {
      left = Math.max(8, window.innerWidth - popWidth - 8);
    }

    if (top + popHeight > window.innerHeight - 8) {
      top = Math.max(8, rect.top - popHeight - 8);
    }

    popoverLeft = left;
    popoverTop = top;
  }

  function updateFromSVClient(clientX, clientY) {
    if (!svAreaElement) return;

    const rect = svAreaElement.getBoundingClientRect();
    const x = clamp(clientX - rect.left, 0, rect.width);
    const y = clamp(clientY - rect.top, 0, rect.height);

    hsv = {
      ...hsv,
      s: (x / rect.width) * 100,
      v: 100 - (y / rect.height) * 100
    };

    syncDraftFromHSV();
  }

  function startSVDrag(event) {
    if (!configAvailable) return;

    dragMode = 'sv';
    updateFromSVClient(event.clientX, event.clientY);

    window.addEventListener('pointermove', handlePointerMove);
    window.addEventListener('pointerup', stopPointerDrag);
  }

  function handlePointerMove(event) {
    if (dragMode === 'sv') {
      updateFromSVClient(event.clientX, event.clientY);
    }
  }

  function handleHueInput(event) {
    const nextHue = Number(event.currentTarget?.value ?? hsv.h);
    hsv = { ...hsv, h: clamp(nextHue, 0, 360) };
    syncDraftFromHSV();
  }

  function handleHexInput(event) {
    const raw = String(event.currentTarget?.value || '');
    const cleaned = raw
      .replace(/[^0-9a-fA-F]/g, '')
      .slice(0, 6)
      .toUpperCase();

    hexInput = cleaned;
    if (cleaned.length === 6) {
      syncDraftFromHex(cleaned);
    }
  }

  function handleHexKeyDown(event) {
    if (event.key === 'Enter') {
      event.preventDefault();
      void applyPicker();
    }
  }

  $: currentConfigData = $configStore?.data ?? null;
  $: if (currentConfigData !== prevConfigData) {
    prevConfigData = currentConfigData;
    void loadValue();
  }

  $: displayHex = isOpen ? draftHex : hex;
  $: hasChanges = normalizeHex(draftHex) !== normalizeHex(hex);

  $: svBackground = `background-color: hsl(${hsv.h}, 100%, 50%);`;
  $: svHandleStyle = `left: ${hsv.s}%; top: ${100 - hsv.v}%;`;

  onDestroy(() => {
    cleanupGlobalListeners();
    stopPointerDrag();
  });
</script>

<Base
  {label}
  {section}
  {option}
  {ttKey}
  {ttImage}
  {ttPlace}
  className={wrapperClass}
  style={visible ? '' : 'display: none;'}
>
  <button
    type="button"
    class="cp-trigger"
    bind:this={triggerElement}
    disabled={!configAvailable}
    on:click={togglePicker}
    aria-haspopup="dialog"
    aria-expanded={isOpen}
    aria-label={$t(label)}
  >
    <span
      class="cp-swatch"
      style="--current-color: {displayHex};"
      use:tt={{ content: displayHex.toUpperCase() }}
    ></span>
    <span class="cp-name">{$t(label)}</span>
    <span class="cp-value">{displayHex.toUpperCase()}</span>
  </button>
</Base>

{#if isOpen}
  <Portal target="body">
    <div
      class="cp-popover"
      bind:this={popoverElement}
      role="dialog"
      aria-label={$t(label)}
      style="left: {popoverLeft}px; top: {popoverTop}px;"
    >
      <div class="cp-header">
        <span>{$t(label)}</span>
        <span class="cp-header-hex">{draftHex.toUpperCase()}</span>
      </div>

      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="cp-sv"
        bind:this={svAreaElement}
        style={svBackground}
        on:pointerdown={startSVDrag}
      >
        <div class="cp-sv-white"></div>
        <div class="cp-sv-black"></div>
        <div class="cp-sv-handle" style={svHandleStyle}></div>
      </div>

      <label class="cp-field cp-field-hue">
        <span>{$t('CONTROLS.hue')}</span>
        <input
          type="range"
          min="0"
          max="360"
          step="1"
          value={hsv.h}
          class="cp-hue"
          on:input={handleHueInput}
        />
      </label>

      <div class="cp-row">
        <label class="cp-field cp-field-hex">
          <span>{$t('CONTROLS.hex')}</span>
          <input
            type="text"
            maxlength="6"
            value={hexInput}
            class="cp-input"
            on:input={handleHexInput}
            on:keydown={handleHexKeyDown}
          />
        </label>

        <span
          class="cp-preview"
          style="--current-color: {draftHex};"
          use:tt={{ content: draftHex.toUpperCase() }}
        ></span>
      </div>

      <div class="cp-actions">
        <button type="button" class="cp-btn" on:click={() => closePicker()}>
          {$t('SETTINGS.LANGUAGE.cancel')}
        </button>
        <button
          type="button"
          class="cp-btn cp-btn-primary"
          disabled={!hasChanges}
          on:click={applyPicker}
        >
          {$t('SETTINGS.LANGUAGE.save')}
        </button>
      </div>
    </div>
  </Portal>
{/if}

<style>
  :global(.color-picker-wrapper) {
    display: inline-flex;
    align-items: center;
    width: fit-content;
  }

  .cp-trigger {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    cursor: pointer;
    padding: 5px 10px;
    background: var(--element-bg-color);
    border-radius: 4px;
    border: 1px solid transparent;
    transition:
      background-color 0.2s,
      border-color 0.2s;
    color: var(--text-color-primary);
    font: inherit;
  }

  .cp-trigger:hover {
    background: var(--element-bg-hover-color);
  }

  .cp-trigger:focus-visible {
    outline: none;
    border-color: var(--control-focus-border-color, var(--accent-color, #ffd700));
    box-shadow: 0 0 0 2px var(--control-focus-ring-color, rgba(255, 215, 0, 0.2));
  }

  .cp-swatch,
  .cp-preview {
    position: relative;
    width: 24px;
    height: 24px;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    overflow: hidden;
    background: var(--current-color);
    flex: 0 0 auto;
  }

  .cp-name {
    font-size: 13px;
    color: var(--text-color-primary);
    line-height: 1.2;
  }

  .cp-value {
    color: var(--text-color-muted, #aaa);
    font-family: monospace;
    font-size: 12px;
  }

  :global(.color-picker-wrapper.disabled) {
    pointer-events: none;
    cursor: not-allowed;
    opacity: var(--control-disabled-opacity, 0.5);
    filter: grayscale(100%);
  }

  .cp-popover {
    position: fixed;
    z-index: 2400;
    width: 296px;
    padding: 12px;
    border-radius: 10px;
    border: 1px solid var(--cp-border-color, #fff);
    background: var(--cp-bg-color, #333);
    color: var(--cp-text-color, #fff);
    box-shadow: var(--cp-shadow, 0 10px 28px rgba(0, 0, 0, 0.45));
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .cp-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    font-size: 13px;
    font-weight: 600;
  }

  .cp-header-hex {
    font-family: monospace;
    color: var(--cp-text-color, #fff);
    opacity: 0.9;
  }

  .cp-sv {
    position: relative;
    width: 100%;
    height: 160px;
    border-radius: 8px;
    cursor: crosshair;
    overflow: hidden;
  }

  .cp-sv-white,
  .cp-sv-black {
    position: absolute;
    inset: 0;
    pointer-events: none;
  }

  .cp-sv-white {
    background: linear-gradient(to right, #fff, rgba(255, 255, 255, 0));
  }

  .cp-sv-black {
    background: linear-gradient(to top, #000, rgba(0, 0, 0, 0));
  }

  .cp-sv-handle {
    position: absolute;
    width: 14px;
    height: 14px;
    border-radius: 50%;
    border: 2px solid #fff;
    box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.45);
    transform: translate(-50%, -50%);
    pointer-events: none;
  }

  .cp-field {
    display: flex;
    flex-direction: column;
    gap: 6px;
    font-size: 12px;
    color: var(--cp-text-color, #fff);
  }

  .cp-field-hue span,
  .cp-field-hex span {
    opacity: 0.8;
    font-weight: 600;
    letter-spacing: 0.02em;
  }

  .cp-hue {
    width: 100%;
    appearance: none;
    height: 8px;
    border-radius: 999px;
    outline: none;
    border: 1px solid var(--cp-border-color, #fff);
    background: linear-gradient(
      to right,
      #f00 0%,
      #ff0 17%,
      #0f0 33%,
      #0ff 50%,
      #00f 67%,
      #f0f 83%,
      #f00 100%
    );
  }

  .cp-hue::-webkit-slider-thumb {
    appearance: none;
    width: 14px;
    height: 14px;
    border-radius: 50%;
    border: 2px solid #fff;
    background: var(--cp-input-color, #555);
    cursor: pointer;
    box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.45);
  }

  .cp-hue::-moz-range-thumb {
    width: 14px;
    height: 14px;
    border-radius: 50%;
    border: 2px solid #fff;
    background: var(--cp-input-color, #555);
    cursor: pointer;
    box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.45);
  }

  .cp-row {
    display: grid;
    grid-template-columns: 1fr auto;
    gap: 8px;
    align-items: end;
  }

  .cp-input {
    height: 32px;
    border-radius: 6px;
    border: 1px solid var(--cp-border-color, #fff);
    background: var(--cp-input-color, #555);
    color: var(--cp-text-color, #fff);
    padding: 0 8px;
    font-family: monospace;
    text-transform: uppercase;
  }

  .cp-input:focus {
    outline: none;
    border-color: var(--accent-color, #ffd700);
    box-shadow: 0 0 0 2px rgba(255, 215, 0, 0.2);
  }

  .cp-actions {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
  }

  .cp-btn {
    border: 1px solid var(--cp-border-color, #fff);
    background: var(--cp-input-color, #555);
    color: var(--cp-text-color, #fff);
    border-radius: 6px;
    padding: 6px 10px;
    font-size: 12px;
    cursor: pointer;
  }

  .cp-btn:hover {
    background: var(--cp-button-hover-color, #777);
  }

  .cp-btn-primary {
    background: var(--accent-color, #ffd700);
    color: var(--control-active-text-color, #000);
    border-color: var(--accent-color, #ffd700);
    font-weight: 700;
  }

  .cp-btn-primary:hover {
    filter: brightness(1.08);
  }

  .cp-btn:disabled,
  .cp-btn-primary:disabled {
    opacity: 0.55;
    cursor: default;
    filter: none;
  }
</style>
