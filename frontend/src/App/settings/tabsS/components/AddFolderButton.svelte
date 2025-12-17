<script>
  import { Dialogs } from '@wailsio/runtime';
  import { appSettings, updateAllPaths, updateGamePath } from '../../../lib/store/appSettings';
  import { get } from 'svelte/store';
  import { tt } from '../../../lib/tooltip';
  import { t } from 'svelte-i18n';

  // Функция для обработки нажатия на кнопку "Добавить путь"
  async function handleAddPathClick() {
    console.log('Открытие системного диалога для выбора папки.');
    try {
      const selectedDirectory = await Dialogs.OpenFile({
        Title: $t('select_game_folder_title'), // Заголовок диалога
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
          await updateAllPaths(updatedPaths); // Добавляем новый путь в all_paths
          console.log('Новый путь добавлен:', newPath);

          // Если это был первый добавленный путь или текущий game_path пуст, устанавливаем его как game_path
          if (currentAllPaths.length === 0 || !currentSettings.game_path) {
            await updateGamePath(newPath);
          }
        } else {
          console.warn('Путь уже существует:', newPath);
          // TODO: Возможно, показать уведомление пользователю
        }
      } else {
        console.log('Выбор директории отменен.');
      }
    } catch (error) {
      console.error('Ошибка при открытии диалога или выборе папки:', error);
      // TODO: Возможно, показать уведомление об ошибке
    }
  }
</script>

<!-- Плавающая кнопка для добавления пути -->
<button
  class="fab-add-button"
  on:click={handleAddPathClick}
  use:tt={{ content: $t('add_folder_tooltip') }}
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
    background-color: #3ba475; /* Цвет кнопки */
    color: white;
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

  .fab-add-button:hover {
    background-color: #2e8b57;
    transform: scale(1.1);
    box-shadow:
      0 8px 16px rgba(0, 0, 0, 0.5),
      0 0 0 4px rgba(255, 255, 255, 0.3); /* Увеличиваем тень при наведении */
  }
</style>
