//@ts-nocheck
import tippy from "tippy.js";
import "tippy.js/dist/tippy.css";
// import "tippy.js/themes/light-border.css";
import "tippy.js/themes/material.css";
import "tippy.js/animations/scale.css";

export const defaultOptions = {
  placement: "auto",
  animation: "scale",
  arrow: true,
  delay: [200, 100],
  allowHTML: true,
  // theme: "light-border",
  theme: "material",
  inertia: true,
  popperOptions: {
    modifiers: [
      {
        name: "preventOverflow",
        options: {
          padding: 8,
        },
      },
    ],
  },
};

export function tt(node, params) {
  let instance;

  // Вспомогательная функция для создания или обновления tippy
  const initialize = (currentParams) => {
    // Проверяем, есть ли контент. Если нет, ничего не делаем или уничтожаем существующий экземпляр.
    if (currentParams && currentParams.content) {
      const options = { ...defaultOptions, ...currentParams };
      if (instance) {
        // Если экземпляр уже есть, просто обновляем его свойства
        instance.setProps(options);
      } else {
        // Если экземпляра нет, создаем его
        instance = tippy(node, options);
      }
    } else {
      // Если контента нет, а экземпляр существует, уничтожаем его
      if (instance) {
        instance.destroy();
        instance = null;
      }
    }
  };

  // Первый вызов при монтировании компонента
  initialize(params);

  return {
    update(newParams) {
      // Вызываем при каждом обновлении параметров
      initialize(newParams);
    },
    destroy() {
      // Вызываем при уничтожении компонента
      if (instance) {
        instance.destroy();
        instance = null;
      }
    },
  };
}

export function buildTooltipContent(items = []) {
  // items = [{ key: "tooltip.key1" }, { image: "/icons/foo.png" }, ...]
  return items
    .map((item) => {
      if (item.key) {
        return `<div style="margin:4px 0;">${item.key}</div>`;
      }
      if (item.image) {
        return `<div style="margin:4px 0; text-align:center;">
                  <img src="${item.image}" style="max-width:200px;"/>
                </div>`;
      }
      return "";
    })
    .join("");
}

if (typeof window !== "undefined") {
  const style = document.createElement("style");
  style.textContent = `
    .tippy-box {
      user-select: text !important;
      z-index: 100000000 !important;
    }
  `;
  document.head.appendChild(style);
}
