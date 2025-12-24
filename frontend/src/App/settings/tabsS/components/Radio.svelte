<script>
  import { updateGamePath } from '../../../lib/store/appSettings';
  import { DeleteIc, FolderIc } from '../../../lib/icons';
  import { tt } from '../../../lib/tooltip';
  import { OpenFolderInExplorer } from '/bindings/lce/backend/utils/utils';

  export let options = [];
  export let selectedValue;
  export let name = 'custom-radio';
  export let onDelete = (value) => {};

  async function handleSelectionChange(newValue) {
    if (selectedValue == newValue) {
      return;
    }
    selectedValue = newValue;
    await updateGamePath(newValue);
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
      <button
        class="folder button"
        on:click|stopPropagation={() => OpenFolderInExplorer(option.value)}
      >
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
    background: rgba(0, 0, 0, 0.2);
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
    background-color: rgba(255, 255, 255, 0.2);
    transition:
      background-color 0.3s ease,
      transform 0.3s ease,
      box-shadow 0.3s ease;
    font-size: 16px;
    color: #333333;
    user-select: none;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
  }
  .custom-radio-container:hover {
    background-color: rgba(255, 255, 255, 0.3);
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
    border: 2px solid #ffffff;
    border-radius: 50%;
    background-color: rgba(0, 0, 0, 0.3);
    transition:
      background-color 0.4s ease,
      transform 0.4s ease;
    margin-right: 12px;
    display: inline-block;
    vertical-align: middle;
    box-shadow: 0 2px 6px rgba(0, 0, 0, 0.4);
  }
  .custom-radio-container input[type='radio']:checked + .custom-radio-checkmark {
    background-color: #ffffff;
    border-color: #007bff;
    box-shadow: 0 0 0 8px rgba(0, 123, 255, 0.2);
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
    background: #007bff;
    transform: translate(-50%, -50%);
  }

  /* NEW: Стили для текста метки и кнопки удаления */
  .radio-label-text {
    flex-grow: 1; /* Позволяет тексту занимать все доступное пространство */
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis; /* Добавляет многоточие, если текст слишком длинный */
    color: #d4cbcb;
  }

  .button {
    width: 35px;
    height: 35px;
    background: none;
    border: none;
    color: #256aaf;
    font-size: 1.2em;
    cursor: pointer;
    margin-left: 10px; /* Отступ от текста */
    padding: 0;
    line-height: 1; /* Убираем лишний отступ */
    transition: color 0.2s ease;
  }

  .button:hover {
    color: #318ce7;
    transform: scale(1.1);
  }

  .delete {
    color: #ff4d4d; /* Красный цвет для кнопки удаления */
  }

  .delete:hover {
    color: #cc0000; /* Темно-красный при наведении */
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
