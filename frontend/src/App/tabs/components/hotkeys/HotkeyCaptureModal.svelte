<script>
  // @ts-nocheck
  import { t } from 'svelte-i18n';
  import { setConfigValue } from '../../../lib/store/config';
  import { encodeKey, normalizeKey } from './keyFuncs';
  import { CODE_TO_CANONICAL_KEY } from './keyCodes';
  import Portal from 'svelte-portal';
  import { onMount } from 'svelte';
  import { isInternalChange } from '../../../lib/store/internalChange';

  export let section = '';
  export let option = '';
  export let onclose = null;

  let isCapturing = false;
  let displayText = '';
  let captureAreaElement;

  let activeModifiers = { ctrl: false, alt: false, shift: false };
  let hasNonModifierBeenPressed = false; // Флаг: была ли нажата не-модификаторная клавиша
  let lastRecordedModifierDisplay = ''; // Сохраняет отображаемую строку модификаторов до keyup

  function startCapture() {
    isCapturing = true;
    displayText = $t('press_any_key');
    // Сбрасываем состояние при каждом новом захвате
    activeModifiers = { ctrl: false, alt: false, shift: false };
    hasNonModifierBeenPressed = false;
    lastRecordedModifierDisplay = '';
  }

  function stopCapture() {
    isCapturing = false;
    displayText = $t('press_any_key');
    onclose?.();
  }

  async function clearHotkey() {
    isInternalChange.mark();
    await setConfigValue(section, option, '');
    displayText = $t('press_any_key');
    stopCapture();
  }

  function getModifierDisplayString() {
    const parts = [];
    if (activeModifiers.ctrl) parts.push('Ctrl');
    if (activeModifiers.alt) parts.push('Alt');
    if (activeModifiers.shift) parts.push('Shift');
    return parts.join(' + ');
  }

  async function handleKeyDown(event) {
    if (!isCapturing) return;

    event.preventDefault();
    event.stopPropagation();

    const canonicalKey = CODE_TO_CANONICAL_KEY[event.code] || event.code.toLowerCase();

    if (canonicalKey === 'backspace') {
      clearHotkey();
      return;
    }
    if (canonicalKey === 'esc') {
      stopCapture();
      return;
    }

    // Обновляем состояние зажатых модификаторов
    activeModifiers.ctrl = event.ctrlKey;
    activeModifiers.alt = event.altKey;
    activeModifiers.shift = event.shiftKey;

    const isCurrentKeyModifier =
      canonicalKey === 'control' || canonicalKey === 'alt' || canonicalKey === 'shift';

    if (!isCurrentKeyModifier) {
      // Была нажата не-модификаторная клавиша
      hasNonModifierBeenPressed = true;

      const parts = [];
      if (activeModifiers.ctrl) parts.push('Ctrl');
      if (activeModifiers.alt) parts.push('Alt');
      if (activeModifiers.shift) parts.push('Shift');
      parts.push(canonicalKey === 'space' ? 'Space' : canonicalKey.toUpperCase());

      const display = parts.join(' + ');
      const encoded = encodeKey(display);
      displayText = display;
      isInternalChange.mark();
      await setConfigValue(section, option, encoded);
      stopCapture(); // Завершаем захват сразу
    } else {
      // Нажата только модификаторная клавиша (или модификатор, который уже был зажат)
      // Обновляем только отображаемый текст
      lastRecordedModifierDisplay = getModifierDisplayString(); // Сохраняем для возможного последующего keyup
      displayText = lastRecordedModifierDisplay || $t('press_any_key');
    }
  }

  async function handleKeyUp(event) {
    if (!isCapturing) return;

    const releasedCanonicalKey = CODE_TO_CANONICAL_KEY[event.code] || event.code.toLowerCase();
    const isReleasedKeyModifier =
      releasedCanonicalKey === 'control' ||
      releasedCanonicalKey === 'alt' ||
      releasedCanonicalKey === 'shift';

    // Если не-модификаторная клавиша уже была нажата и комбинация сохранена,
    // то игнорируем последующие keyup
    if (hasNonModifierBeenPressed) {
      return;
    }

    // Обновляем состояние activeModifiers на основе отпущенной клавиши
    if (releasedCanonicalKey === 'control') activeModifiers.ctrl = false;
    if (releasedCanonicalKey === 'alt') activeModifiers.alt = false;
    if (releasedCanonicalKey === 'shift') activeModifiers.shift = false;

    // Получаем текущую строку отображения модификаторов ПОСЛЕ отпускания
    const currentModifierDisplay = getModifierDisplayString();

    // Проверяем, все ли модификаторы теперь отпущены
    const allModifiersReleased =
      !activeModifiers.ctrl && !activeModifiers.alt && !activeModifiers.shift;

    if (isReleasedKeyModifier && allModifiersReleased && lastRecordedModifierDisplay) {
      // Если отпущенная клавиша была модификатором,
      // и теперь все модификаторы отпущены,
      // и у нас есть последняя записанная комбинация модификаторов
      // (это означает, что была нажата чистая комбинация модификаторов)
      const encoded = encodeKey(lastRecordedModifierDisplay);
      displayText = lastRecordedModifierDisplay; // Отображаем комбинацию, которую сохраняем
      isInternalChange.mark();
      await setConfigValue(section, option, encoded);
      stopCapture(); // Сохраняем и закрываем модальное окно
      // Сбрасываем флаги
      hasNonModifierBeenPressed = false;
      lastRecordedModifierDisplay = '';
    } else if (isReleasedKeyModifier) {
      // Если модификатор отпущен, но другие модификаторы еще зажаты,
      // или это была одиночная модификаторная клавиша, которая была отпущена,
      // и нам просто нужно обновить отображение
      lastRecordedModifierDisplay = currentModifierDisplay; // Обновляем для возможного следующего keyup
      displayText = currentModifierDisplay || $t('press_any_key');
    }
    // Если это keyup не-модификатора (и hasNonModifierBeenPressed == false),
    // это значит, что этот не-модификатор не был обработан в keydown (что странно)
    // или это какая-то другая клавиша, которую мы не отслеживаем.
    // В этом случае просто игнорируем.
  }

  onMount(() => {
    startCapture();
    // Устанавливаем фокус на capture-area после монтирования компонента
    if (captureAreaElement) {
      captureAreaElement.focus();
    }
  });
</script>

<Portal target="body">
  <div
    class="modal-backdrop"
    role="button"
    tabindex="0"
    on:keydown={handleKeyDown}
    on:keyup={handleKeyUp}
    on:click={stopCapture}
  >
    <div class="modal-content" on:click|stopPropagation role="presentation">
      <div class="modal-header">
        <h3>{$t('capture_modal_label')}</h3>
      </div>
      <div class="modal-body">
        <div
          class="capture-area"
          class:capturing={isCapturing}
          tabindex="-1"
          bind:this={captureAreaElement}
        >
          {displayText}
        </div>
      </div>
      <div class="modal-footer">
        <button class="clear-button" on:click={clearHotkey}>{$t('clear_button')}</button>
        <button on:click={stopCapture}>{$t('cancel_button')}</button>
      </div>
    </div>
  </div>
</Portal>

<style>
  .modal-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 1000;
  }
  .modal-content {
    background: #2a2a2a;
    border-radius: 8px;
    padding: 20px;
    min-width: 300px;
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.3);
  }
  .modal-header h3 {
    margin: 0;
    color: #fff;
  }
  .modal-body {
    margin: 20px 0;
  }
  .capture-area {
    padding: 20px;
    border: 2px dashed #666;
    border-radius: 4px;
    text-align: center;
    color: #fff;
    min-height: 60px;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .capturing {
    border-color: #4caf50;
    background: rgba(76, 175, 80, 0.1);
  }
  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
  }
  button {
    padding: 8px 16px;
    border: none;
    border-radius: 4px;
    background: #4caf50;
    color: white;
    cursor: pointer;
  }
  button:hover {
    background: #45a049;
  }
  .clear-button {
    background: #f44336;
  }
  .clear-button:hover {
    background: #d32f2f;
  }
</style>
