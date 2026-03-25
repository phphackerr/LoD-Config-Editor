<script>
  import { updateGamePath } from '../../../lib/store/appSettings';
  import { DeleteIc, FolderIc } from '../../../lib/icons';
  import { notifyError } from '../../../lib/store/notifications';
  import { toErrorMessage } from '../../../lib/store/storeUtils';
  import { tt } from '../../../lib/tooltip';
  import { OpenFolderInExplorer } from '/bindings/lce/backend/utils/utils';
  import { t } from 'svelte-i18n';

  export let options = [];
  export let selectedValue;
  export let name = 'custom-radio';
  export let onDelete = (value) => {};

  async function handleSelectionChange(newValue) {
    if (selectedValue == newValue) {
      return;
    }
    const previousValue = selectedValue;
    selectedValue = newValue;

    const result = await updateGamePath(newValue);
    if (!result.ok) {
      selectedValue = previousValue;
      notifyError(result.error || $t('ERRORS.paths.update_game_path'));
    }
  }

  async function handleOpenFolder(path) {
    try {
      await OpenFolderInExplorer(path);
    } catch (error) {
      notifyError(toErrorMessage(error, $t('ERRORS.paths.open_folder')));
    }
  }

  function handleKeyDown(event, option) {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      handleSelectionChange(option.value);
    }
  }
</script>

<div class="custom-radio-group" role="radiogroup">
  {#each options as option (option.value)}
    <div
      class="custom-radio-container"
      use:tt={{ content: option.label, placement: 'top' }}
      on:keydown={(e) => handleKeyDown(e, option)}
      on:click={() => handleSelectionChange(option.value)}
      tabindex="0"
      role="radio"
      aria-checked={selectedValue === option.value}
    >
      <input
        type="radio"
        {name}
        value={option.value}
        checked={selectedValue === option.value}
        tabindex="-1"
      />
      <span class="custom-radio-checkmark"></span>
      <span class="radio-label-text">{option.label}</span>
      <button class="folder button" on:click|stopPropagation={() => handleOpenFolder(option.value)}>
        <FolderIc />
      </button>
      <button class="delete button" on:click|stopPropagation={() => onDelete(option.value)}>
        <DeleteIc />
      </button>
    </div>
  {/each}
</div>

<style>
  .custom-radio-group {
    display: flex;
    flex-direction: column;
    gap: 12px;
    width: calc(100% - 30px);
    border-radius: 12px;
    background: var(--surface-panel-muted, rgba(0, 0, 0, 0.3));
    padding: 16px;
    box-shadow: 0 6px 12px rgba(0, 0, 0, 0.3);
  }
  .custom-radio-container {
    position: relative;
    display: flex;
    align-items: center;
    cursor: pointer;
    padding: 12px 20px;
    border-radius: 8px;
    background-color: var(--element-bg-color, rgba(255, 255, 255, 0.2));
    transition:
      background-color 0.3s ease,
      transform 0.3s ease,
      box-shadow 0.3s ease;
    font-size: 16px;
    color: var(--text-color, #f6f6f6);
    user-select: none;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
  }
  .custom-radio-container:hover {
    background-color: var(--element-bg-hover-color, rgba(255, 255, 255, 0.3));
    transform: scale(1.03);
    box-shadow: 0 6px 12px rgba(0, 0, 0, 0.4);
  }
  .custom-radio-container input[type='radio'] {
    opacity: 0;
    position: absolute;
  }
  .custom-radio-checkmark {
    position: relative;
    height: 24px;
    width: 24px;
    border: 2px solid var(--text-color-primary, #fff);
    border-radius: 50%;
    background-color: var(--surface-panel-muted, rgba(0, 0, 0, 0.3));
    transition:
      background-color 0.4s ease,
      transform 0.4s ease;
    margin-right: 12px;
    display: inline-block;
    vertical-align: middle;
    box-shadow: 0 2px 6px rgba(0, 0, 0, 0.4);
  }
  .custom-radio-container input[type='radio']:checked + .custom-radio-checkmark {
    background-color: var(--text-color-primary, #fff);
    border-color: var(--accent-color, #ffd700);
    box-shadow: 0 0 0 8px var(--status-warning-bg, rgba(245, 158, 11, 0.12));
    transform: scale(1.2);
    animation: pulse 0.6s forwards;
  }
  .custom-radio-checkmark::after {
    content: '';
    position: absolute;
    display: none;
  }
  .custom-radio-container input[type='radio']:checked + .custom-radio-checkmark::after {
    display: block;
    left: 50%;
    top: 50%;
    width: 14px;
    height: 14px;
    border-radius: 50%;
    background: var(--accent-color, #ffd700);
    transform: translate(-50%, -50%);
  }

  /* NEW: Стили для текста метки и кнопки удаления */
  .radio-label-text {
    flex-grow: 1; /* Позволяет тексту занимать все доступное пространство */
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis; /* Добавляет многоточие, если текст слишком длинный */
    color: var(--text-color-secondary, #d4cbcb);
  }

  .button {
    width: 35px;
    height: 35px;
    background: none;
    border: none;
    color: var(--status-info-text, #bae6fd);
    font-size: 1.2em;
    cursor: pointer;
    margin-left: 10px; /* Отступ от текста */
    padding: 0;
    line-height: 1; /* Убираем лишний отступ */
    transition: color 0.2s ease;
  }

  .button:hover {
    color: var(--accent-color, #ffd700);
    transform: scale(1.1);
  }

  .delete {
    color: var(--status-error-text, #fecaca); /* Красный цвет для кнопки удаления */
  }

  .delete:hover {
    color: var(--action-danger-bg, #ca3333); /* Темно-красный при наведении */
  }

  @keyframes pulse {
    0% {
      transform: scale(1.2);
    }
    50% {
      transform: scale(1.4);
    }
    100% {
      transform: scale(1.2);
    }
  }
</style>
