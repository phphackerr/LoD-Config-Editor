<script>
  //@ts-nocheck
  import { onMount, onDestroy, tick } from 'svelte';
  import { searchQuery, searchableItems, activeTab } from '../lib/store/search';
  import { t } from 'svelte-i18n';

  let filteredItems = [];
  let showResults = false;
  let wrapperElement;
  let inputElement;
  let resultElements = [];

  function handleGlobalKeyDown(event) {
    // Проверяем, нажаты ли Ctrl и F
    if (event.ctrlKey && event.key === 'f') {
      // Предотвращаем стандартное действие браузера (поиск по странице)
      event.preventDefault();
      // Устанавливаем фокус на наше поле ввода
      inputElement?.focus();
    }
  }

  onMount(() => {
    window.addEventListener('keydown', handleGlobalKeyDown);
  });

  onDestroy(() => {
    window.removeEventListener('keydown', handleGlobalKeyDown);
  });

  // Реактивно фильтруем элементы
  $: {
    if ($searchQuery && $searchQuery.length > 0) {
      filteredItems = $searchableItems.filter((item) =>
        item.label.toLowerCase().includes($searchQuery.toLowerCase())
      );
      showResults = filteredItems.length > 0;
    } else {
      filteredItems = [];
      showResults = false;
    }
    // Сбрасываем массив элементов при изменении фильтра
    resultElements = [];
  }

  async function handleItemClick(item) {
    // 1. Переключаем вкладку, если это необходимо
    if (item.tabId && $activeTab !== item.tabId) {
      console.log(`Search.svelte: Switching tab from '${$activeTab}' to '${item.tabId}'`);
      activeTab.set(item.tabId);
      await tick();
    } else {
      console.log('Search.svelte: No tab switch needed.');
    }

    const elementRect = item.element?.getBoundingClientRect();
    const isVisible =
      elementRect.top >= 0 &&
      elementRect.bottom <= (window.innerHeight || document.documentElement.clientHeight);

    if (!isVisible) {
      item.element?.scrollIntoView({
        behavior: 'smooth',
        block: 'center'
      });
    }

    item.element?.classList.add('highlight');
    setTimeout(() => {
      item.element?.classList.remove('highlight');
    }, 2000);

    $searchQuery = '';
    inputElement?.focus();
  }

  function handleFocusIn() {
    if (filteredItems.length > 0) {
      showResults = true;
    }
  }

  function handleFocusOut(event) {
    if (!wrapperElement.contains(event.relatedTarget)) {
      showResults = false;
    }
  }

  // --- ИЗМЕНЕНИЕ 2: Централизованный обработчик нажатий клавиш ---
  function handleKeyDown(event) {
    if (!showResults) return;

    const { key } = event;
    const activeElement = document.activeElement;

    if (key === 'ArrowDown') {
      event.preventDefault(); // Предотвращаем прокрутку страницы

      if (activeElement === inputElement) {
        resultElements[0]?.focus(); // С поля ввода на первый результат
      } else {
        const currentIndex = resultElements.indexOf(activeElement);
        if (currentIndex > -1 && currentIndex < resultElements.length - 1) {
          resultElements[currentIndex + 1]?.focus(); // На следующий результат
        }
      }
    } else if (key === 'ArrowUp') {
      event.preventDefault(); // Предотвращаем прокрутку страницы
      const currentIndex = resultElements.indexOf(activeElement);
      if (currentIndex > 0) {
        resultElements[currentIndex - 1]?.focus(); // На предыдущий результат
      } else if (currentIndex === 0) {
        inputElement?.focus(); // С первого результата обратно на поле ввода
      }
    } else if (key === 'Escape') {
      $searchQuery = ''; // Очищаем и закрываем по нажатию Escape
      inputElement?.focus();
    }
  }

  function handleItemKeyPress(event, item) {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      handleItemClick(item);
    }
  }
</script>

<div
  class="search-wrapper"
  bind:this={wrapperElement}
  on:focusin={handleFocusIn}
  on:focusout={handleFocusOut}
  on:keydown={handleKeyDown}
  role="presentation"
>
  <div class="container-input">
    <label class="search-label">
      <input
        type="text"
        name="text"
        class="input"
        required=""
        placeholder={$t('TITLE.search')}
        bind:value={$searchQuery}
        bind:this={inputElement}
        autocomplete="off"
        on:keydown={(e) => {
          if (e.key === 'Escape') {
            e.preventDefault();
            if ($searchQuery) {
              $searchQuery = '';
            } else {
              inputElement?.blur();
            }
          }
        }}
      />
      <kbd class="slash-icon">Ctrl + F</kbd>
      <svg
        class="search-icon"
        xmlns="http://www.w3.org/2000/svg"
        version="1.1"
        xmlns:xlink="http://www.w3.org/1999/xlink"
        width="12"
        height="12"
        viewBox="0 0 56.966 56.966"
        style="enable-background:new 0 0 512 512"
        xml:space="preserve"
      >
        <path
          d="M55.146 51.887 41.588 37.786A22.926 22.926 0 0 0 46.984 23c0-12.682-10.318-23-23-23s-23 10.318-23 23 10.318 23 23 23c4.761 0 9.298-1.436 13.177-4.162l13.661 14.208c.571.593 1.339.92 2.162.92.779 0 1.518-.297 2.079-.837a3.004 3.004 0 0 0 .083-4.242zM23.984 6c9.374 0 17 7.626 17 17s-7.626 17-17 17-17-7.626-17-17 7.626-17 17-17z"
          fill="currentColor"
          data-original="#000000"
        ></path>
      </svg>
    </label>
  </div>

  {#if showResults}
    <ul class="search-results" role="listbox">
      {#each filteredItems as item, i}
        <li
          role="button"
          tabindex="0"
          bind:this={resultElements[i]}
          on:mousedown={() => handleItemClick(item)}
          on:keydown={(e) => handleItemKeyPress(e, item)}
        >
          {item.label}
        </li>
      {/each}
    </ul>
  {/if}
</div>

<style>
  :global(.highlight) {
    position: relative; /* важно, чтобы ::after позиционировался от элемента */
  }

  :global(.highlight::after) {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    border: 2px solid #ffeb3b;
    border-radius: 4px;
    box-shadow: inset 0 0 8px #ffeb3b;
    pointer-events: none; /* чтобы клики проходили сквозь подсветку */
    animation: highlight-fade 2s ease-out forwards;
  }

  @keyframes highlight-fade {
    0% {
      opacity: 1;
    }
    20% {
      opacity: 0;
    }
    40% {
      opacity: 1;
    }
    60% {
      opacity: 0;
    }
    80% {
      opacity: 1;
    }
    100% {
      opacity: 0;
    }
  }

  .search-wrapper {
    position: relative;
    pointer-events: auto;
  }

  .container-input {
    margin: 0 auto;
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .search-label {
    --wails-draggable: no-drag;
    display: flex;
    height: 25px;
    align-items: center;
    box-sizing: border-box;
    position: relative;
    border: 1px solid var(--titlebar-search-border-color);
    border-radius: 12px;
    overflow: hidden;
    background: var(--titlebar-search-bg-color);
    padding: 7px;
    cursor: text;
  }

  .search-label:hover {
    border-color: var(--titlebar-search-border-color-hover);
  }

  .search-label:focus-within {
    background: var(--titlebar-search-bg-color-focus);
    border-color: var(--titlebar-search-border-color-focus);
  }

  .search-label input {
    outline: none;
    width: 100%;
    border: none;
    background: none;
    color: var(--titlebar-search-input-text-color);
  }

  .search-label input:focus + .slash-icon {
    display: none;
  }

  .search-label input:valid ~ .search-icon {
    display: block;
  }

  .search-label input:valid {
    width: calc(100% - 22px);
    transform: translateX(20px);
  }

  .search-label svg,
  .slash-icon {
    position: absolute;
    color: var(--titlebar-search-hotkey-text-color);
  }

  .search-icon {
    display: none;
    width: 12px;
    height: auto;
  }

  .slash-icon {
    right: 7px;
    border: 1px solid var(--titlebar-search-hotkey-border-color);
    background: var(--titlebar-search-hotkey-bg-color);
    display: flex; /* Добавляем flexbox */
    align-items: center; /* Выравнивание по вертикали по центру */
    justify-content: center; /* Выравнивание по горизонтали по центру */
    border-radius: 3px;
    box-shadow: var(--titlebar-search-hotkey-box-shadow);
    cursor: text;
    font-size: 12px;
    width: fit-content;
    height: 17px;
    padding: 0 3px;
  }

  .slash-icon:active {
    box-shadow: var(--titlebar-search-hotkey-box-shadow-active);
    text-shadow: var(--titlebar-search-hotkey-text-shadow-active);
    color: var(--titlebar-search-hotkey-text-color-active);
  }

  .search-results {
    position: absolute; /* Главное изменение: вырываем список из потока */
    top: 100%; /* Располагаем его сразу под родительским элементом */
    left: 0;
    right: 0;
    margin-top: 5px; /* Небольшой отступ сверху */
    padding: 5px 0;
    list-style: none;
    background: var(--titlebar-search-bg-color-focus); /* Задаем фон */
    border: 1px solid var(--titlebar-search-border-color-focus);
    border-radius: 8px;
    max-height: 300px;
    overflow-y: auto;
    z-index: 200; /* Гарантируем, что список будет поверх других элементов */
  }

  .search-results li {
    padding: 8px 12px;
    cursor: pointer;
    color: var(--titlebar-search-input-text-color);
    font-size: 14px;
  }

  .search-results li:hover,
  .search-results li:focus {
    background-color: rgba(255, 215, 0, 0.2);
    outline: none;
  }
</style>
