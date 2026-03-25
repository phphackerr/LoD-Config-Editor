<script>
  import { Dialogs } from '@wailsio/runtime';
  import { appSettings, updateAllPaths, updateGamePath } from '../../../lib/store/appSettings';
  import { notifyError, notifySuccess, notifyWarning } from '../../../lib/store/notifications';
  import { get } from 'svelte/store';
  import { tt } from '../../../lib/tooltip';
  import { t } from 'svelte-i18n';
  import { toErrorMessage } from '../../../lib/store/storeUtils';

  let errorMessage = '';

  // Функция для обработки нажатия на кнопку "Добавить путь"
  async function handleAddPathClick() {
    errorMessage = '';
    try {
      const selectedDirectory = await Dialogs.OpenFile({
        Title: $t('SETTINGS.PATHS.select_game_folder_title'),
        CanChooseDirectories: true, // Разрешаем выбор директорий
        CanChooseFiles: false, // Запрещаем выбор файлов
        AllowsMultipleSelection: false // Разрешаем выбор только одной директории
      });

      if (selectedDirectory) {
        // OpenFile возвращает строку (путь) или пустую строку, если отменено
        const newPath = selectedDirectory;
        const currentSettings = get(appSettings);
        const currentAllPaths = currentSettings.all_paths;

        // Проверяем, существует ли уже такой путь
        if (!currentAllPaths.includes(newPath)) {
          const updatedPaths = [...currentAllPaths, newPath];
          const updatePathsResult = await updateAllPaths(updatedPaths);
          if (!updatePathsResult.ok) {
            errorMessage = updatePathsResult.error || $t('ERRORS.paths.add_path');
            notifyError(errorMessage);
            return;
          }

          // Если это был первый добавленный путь или текущий game_path пуст, устанавливаем его как game_path
          if (currentAllPaths.length === 0 || !currentSettings.game_path) {
            const updateGamePathResult = await updateGamePath(newPath);
            if (!updateGamePathResult.ok) {
              errorMessage = updateGamePathResult.error || $t('ERRORS.paths.select_game_path');
              notifyError(errorMessage);
              return;
            }
          }
          notifySuccess($t('SETTINGS.PATHS.path_added'));
        } else {
          notifyWarning($t('SETTINGS.PATHS.path_already_added'));
        }
      }
    } catch (error) {
      errorMessage = toErrorMessage(error, $t('ERRORS.paths.pick_folder'));
      notifyError(errorMessage);
    }
  }
</script>

{#if errorMessage}
  <div class="add-path-error">{errorMessage}</div>
{/if}

<!-- Плавающая кнопка для добавления пути -->
<button
  class="fab-add-button"
  on:click={handleAddPathClick}
  use:tt={{ content: $t('SETTINGS.PATHS.add_folder_tooltip') }}
>
  +
</button>

<style>
  /* Стили для плавающей кнопки */
  .fab-add-button {
    position: absolute; /* Относительно .general-settings */
    bottom: 20px;
    right: 20px;
    width: 50px;
    height: 50px;
    border-radius: 50%;
    background-color: var(--action-primary-bg, #3ba475); /* Цвет кнопки */
    color: var(--text-color-primary, #fff);
    font-size: 2em;
    border: none;
    box-shadow:
      0 6px 12px rgba(0, 0, 0, 0.4),
      0 0 0 3px rgba(0, 0, 0, 0.2); /* Более отчетливая тень */
    cursor: pointer;
    display: flex;
    justify-content: center;
    align-items: center;
    text-align: center;
    transition:
      background-color 0.3s ease,
      transform 0.3s ease,
      box-shadow 0.3s ease;
    z-index: 999; /* Чтобы была поверх контента */
  }

  .add-path-error {
    position: absolute;
    right: 20px;
    bottom: 78px;
    max-width: 340px;
    padding: 8px 10px;
    border-radius: 8px;
    border: 1px solid var(--status-error-border, rgba(220, 38, 38, 0.38));
    background: var(--status-error-bg, rgba(220, 38, 38, 0.16));
    color: var(--status-error-text, #fecaca);
    font-size: 12px;
    line-height: 1.35;
  }

  .fab-add-button:hover {
    background-color: var(--action-primary-bg-hover, #2e8b57);
    transform: scale(1.1);
    box-shadow:
      0 8px 16px rgba(0, 0, 0, 0.5),
      0 0 0 4px var(--status-success-border, rgba(36, 147, 79, 0.38)); /* Увеличиваем тень при наведении */
  }
</style>
