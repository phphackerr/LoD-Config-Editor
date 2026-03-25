<script>
  import { onMount, onDestroy, tick } from 'svelte';
  import { searchQuery, searchableItems, activeTab, setActiveTab } from '../lib/store/search';
  import { t } from 'svelte-i18n';
  import { tt } from '../lib/tooltip';

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
      setActiveTab(item.tabId);
      await tick();
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
    if (!wrapperElement?.contains(event.relatedTarget)) {
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
    <label
      class="search-label"
      use:tt={{ content: $t('TITLE.search_hotkey_tooltip'), placement: 'bottom' }}
    >
      <input
        type="text"
        name="text"
        class="input"
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
      <kbd
        class="slash-icon"
        use:tt={{ content: $t('TITLE.search_hotkey_tooltip'), placement: 'bottom' }}
      >
        {$t('TITLE.search_hotkey_label')}
      </kbd>
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
        <!-- svelte-ignore a11y_no_noninteractive_element_to_interactive_role -->
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
    position: relative;
  }

  :global(.highlight::after) {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    border: 2px solid var(--app-highlight-border-color, var(--accent-color, #ffeb3b));
    border-radius: 4px;
    box-shadow: inset 0 0 8px var(--app-highlight-glow-color, var(--accent-color, #ffeb3b));
    pointer-events: none;
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
    border: 1px solid transparent;
    border-radius: 12px;
    overflow: hidden;
    background: var(--titlebar-search-bg, var(--bg-color-medium, rgb(61, 61, 61)));
    padding: 7px;
    cursor: text;
  }

  .search-label:hover {
    border-color: var(--titlebar-search-border, var(--border-color, rgb(128, 128, 128)));
  }

  .search-label:focus-within {
    background: var(--titlebar-search-focus-bg, var(--bg-color, rgb(70, 70, 70)));
    border-color: var(--titlebar-search-focus-border, var(--border-color, rgb(128, 128, 128)));
  }

  .search-label input {
    outline: none;
    width: 100%;
    border: none;

    background: rgba(0, 0, 0, 0);
    color: var(--titlebar-search-text, var(--text-color-secondary, rgb(162, 162, 162)));
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

    color: var(--titlebar-search-icon, var(--text-color-muted, rgb(151, 148, 148)));
  }

  .search-icon {
    display: none;
    width: 12px;
    height: auto;
  }

  .slash-icon {
    right: 7px;
    border: 1px solid
      var(--titlebar-search-hotkey-border-color, var(--border-color, rgb(57, 56, 56)));
    background: var(
      --titlebar-search-hotkey-bg-gradient,
      linear-gradient(-225deg, var(--bg-color-dark, #343434), var(--bg-color-medium, #6d6d6d))
    );
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 3px;
    box-shadow: var(
      --titlebar-search-hotkey-box-shadow,
      inset 0 -2px 0 0 #3f3f3f,
      inset 0 0 1px 1px rgb(94, 93, 93),
      0 1px 2px 1px rgba(28, 28, 29, 0.4)
    );
    cursor: text;
    font-size: 12px;
    width: fit-content;
    height: 17px;
    padding: 0 3px;
  }

  .slash-icon:active {
    box-shadow: var(
      --titlebar-search-hotkey-box-shadow-active,
      inset 0 1px 0 0 #3f3f3f,
      inset 0 0 1px 1px rgb(94, 93, 93),
      0 1px 2px 0 rgba(28, 28, 29, 0.4)
    );
    text-shadow: 0 1px 0 var(--titlebar-search-hotkey-text-shadow-color, #7e7e7e);
    color: transparent;
  }

  .search-results {
    position: absolute;
    top: 100%;
    left: 0;
    right: 0;
    margin-top: 5px;
    padding: 5px 0;
    list-style: none;
    background: var(--titlebar-search-results-bg, var(--bg-color-medium, #464646));
    border: 1px solid var(--titlebar-search-results-border, var(--border-color, rgb(128, 128, 128)));
    border-radius: 8px;
    max-height: 300px;
    overflow-y: auto;
    z-index: 200;
  }

  .search-results li {
    padding: 8px 12px;
    cursor: pointer;
    color: var(--titlebar-search-results-text, var(--text-color-secondary, rgb(162, 162, 162)));
    font-size: 14px;
  }

  .search-results li:hover,
  .search-results li:focus {
    background-color: var(
      --titlebar-search-results-hover-bg,
      var(--status-warning-bg, rgba(255, 215, 0, 0.2))
    );
    outline: none;
  }
</style>
