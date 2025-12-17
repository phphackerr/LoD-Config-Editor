# LoD Config Editor (LCE)

**LoD Config Editor** is a modern, feature-rich configuration tool for **Warcraft III: Legends of Dota (LoD)**. It allows players to easily modify game settings, manage hotkeys, download maps, and customize their gaming experience through a sleek, user-friendly interface.

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Wails](https://img.shields.io/badge/built%20with-Wails-red.svg)
![Svelte](https://img.shields.io/badge/frontend-Svelte-orange.svg)

## ✨ Features

- **Visual Config Editor**: Modify `config.lod.ini` without touching text files. Toggle checkboxes, adjust sliders, and pick colors with ease.
- **Map Downloader**: Automatically checks for the latest LoD maps from EpicWar and downloads them directly to your game folder.
- **Hotkey Manager**: Bind inventory items, skills, and game commands to any key. Supports complex modifiers (Alt, Ctrl, Shift).
- **Theme Support**: Customize the look of the editor with built-in themes or create your own.
- **Multilingual**: Fully localized interface (English, Russian, and more).
- **Game Path Scanner**: Automatically detects your Warcraft III installation.

## 🚀 Installation

1.  Download the latest release from the [Releases](https://github.com/phphacker/lce/releases) page.
2.  Run `LCE_Setup.exe` (or the portable executable).
3.  On first launch, the application will attempt to locate your Warcraft III folder. If not found, you can specify it manually in the settings.

## 🛠️ Development

This project is built using [Wails3](https://v3alpha.wails.io/) (Go + Svelte).

### Prerequisites

- **Go** (1.21+)
- **Node.js** (18+)
- **NPM**

### Setup

1.  Clone the repository:

    ```bash
    git clone https://github.com/phphacker/lce.git
    cd lce
    ```

2.  Install frontend dependencies:

    ```bash
    cd frontend
    npm install
    ```

3.  Run in development mode:
    ```bash
    wails3 dev
    ```

### Building

To build the application for production:

```bash
wails3 build
```

The output binary will be located in the `bin` directory.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1.  Fork the project
2.  Create your feature branch (`git checkout -b feature/AmazingFeature`)
3.  Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4.  Push to the branch (`git push origin feature/AmazingFeature`)
5.  Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

Portions of the code (StormLib wrapper) are Copyright (c) Ladislav Zezula.
