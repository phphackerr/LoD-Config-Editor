package paths_scanner

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	// Этот импорт все еще нужен для `LoadSettings` и `SaveSettings` в других функциях, но не в CheckAndFindPaths.
	"golang.org/x/sys/windows/registry" // Для доступа к реестру Windows
)

// Scanner - это структура, которая будет привязана к фронтенду Wails.
// Она содержит методы для поиска путей к файлам игры.
type Scanner struct{}

// NewScanner создает новый экземпляр Scanner.
func NewScanner() *Scanner {
	return &Scanner{}
}

// isTargetFile проверяет, является ли файл целевым (config.lod.ini или war3.exe).
func isTargetFile(filename string) bool {
	lowerFilename := strings.ToLower(filename)
	return lowerFilename == "config.lod.ini" || lowerFilename == "war3.exe"
}

// findFilesInFolder ищет целевые файлы в указанной папке с заданной глубиной.
// Возвращает путь к папке, содержащей целевой файл, и булево значение, указывающее, найден ли он.
func findFilesInFolder(root string, excludedFolders []string, maxDepth int) (string, bool) {
	root = filepath.Clean(root) // Очищаем корневой путь
	rootInfo, err := os.Stat(root)
	if err != nil || !rootInfo.IsDir() {
		return "", false
	}

	foundPath := ""
	found := false

	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// Логируем ошибку, но продолжаем обход
			log.Printf("Ошибка доступа к пути %s: %v", path, err)
			return nil
		}

		// Вычисляем текущую глубину относительно корневого каталога
		depthFromRoot := 0
		if path != root {
			// Количество разделителей пути в части после корня
			depthFromRoot = strings.Count(path[len(root):], string(os.PathSeparator))
			if d.IsDir() {
				// Если это директория и она не является самим корнем, увеличиваем глубину на 1
				// Пример: C:\ -> depth 0; C:\Games -> depth 1; C:\Games\Warcraft -> depth 2
				if len(path[len(root):]) > 0 && path[len(root)] == os.PathSeparator {
					depthFromRoot = strings.Count(path[len(root):], string(os.PathSeparator))
				} else { // Handle direct subfolders without a leading separator in path[len(root):]
					depthFromRoot = 1
				}
			}
		}

		if d.IsDir() {
			if depthFromRoot > maxDepth {
				return filepath.SkipDir // Пропускаем папки, находящиеся глубже maxDepth
			}

			folderName := d.Name()
			for _, excluded := range excludedFolders {
				if strings.EqualFold(folderName, excluded) {
					return filepath.SkipDir // Пропускаем исключенные папки
				}
			}
		} else if d.Type().IsRegular() && isTargetFile(d.Name()) {
			foundPath = filepath.Dir(path)
			found = true
			return filepath.SkipAll // Найден целевой файл, останавливаем обход
		}
		return nil
	})

	if err != nil {
		log.Printf("Ошибка при обходе директории %s: %v", root, err)
	}

	return foundPath, found
}

// getLogicalDrives возвращает список логических дисков в Windows.
func getLogicalDrives() []string {
	var drives []string
	for char := 'A'; char <= 'Z'; char++ {
		drive := fmt.Sprintf("%c:\\", char)
		if _, err := os.Stat(drive); err == nil {
			drives = append(drives, drive)
		}
	}
	return drives
}

// findPathInRegistry ищет путь установки Warcraft III в реестре Windows.
func findPathInRegistry() (string, bool) {
	regPath := `SOFTWARE\Blizzard Entertainment\Warcraft III`

	// HKEY_LOCAL_MACHINE
	if k, err := registry.OpenKey(registry.LOCAL_MACHINE, regPath, registry.READ); err == nil {
		if installPath, _, err := k.GetStringValue("InstallPath"); err == nil {
			pathConfig := filepath.Join(installPath, "config.lod.ini")
			pathExe := filepath.Join(installPath, "war3.exe")

			// Проверяем наличие config.lod.ini или war3.exe в найденном пути
			_, errConfig := os.Stat(pathConfig)
			_, errExe := os.Stat(pathExe)

			if errConfig == nil || errExe == nil { // Исправленное условие
				k.Close()
				return installPath, true
			}
		}
		k.Close()
	}

	// HKEY_CURRENT_USER
	if k, err := registry.OpenKey(registry.CURRENT_USER, regPath, registry.READ); err == nil {
		if installPath, _, err := k.GetStringValue("InstallPath"); err == nil {
			pathConfig := filepath.Join(installPath, "config.lod.ini")
			pathExe := filepath.Join(installPath, "war3.exe")

			// Проверяем наличие config.lod.ini или war3.exe в найденном пути
			_, errConfig := os.Stat(pathConfig)
			_, errExe := os.Stat(pathExe)

			if errConfig == nil || errExe == nil {
				k.Close()
				return installPath, true
			}
		}
		k.Close()
	}

	return "", false
}

// FindConfigOrExeParallel параллельно ищет пути к файлам config.lod.ini или war3.exe на всех логических дисках.
// Эта функция привязана к фронтенду Wails.
func (s *Scanner) FindConfigOrExeParallel() []string {
	excludedFolders := []string{"Windows", "Users", "ProgramData", "System Volume Information"}
	drives := getLogicalDrives()

	var wg sync.WaitGroup
	results := make(chan string, len(drives)) // Буферизованный канал для результатов

	for _, drive := range drives {
		wg.Add(1)
		go func(drive string) {
			defer wg.Done()
			if path, found := findFilesInFolder(drive, excludedFolders, 5); found {
				results <- path
			}
		}(drive)
	}

	wg.Wait()      // Ждем завершения всех горутин
	close(results) // Закрываем канал после завершения всех записей

	uniquePaths := make(map[string]struct{})
	var foundPaths []string
	for path := range results {
		if _, exists := uniquePaths[path]; !exists {
			uniquePaths[path] = struct{}{}
			foundPaths = append(foundPaths, path)
		}
	}
	return foundPaths
}

// CheckAndFindPaths - основная функция для поиска путей.
// Эта функция будет привязана к фронтенду Wails.
func (s *Scanner) CheckAndFindPaths() ([]string, error) {
	log.Println("=== Начало CheckAndFindPaths ===")
	log.Println("Выполняем поиск путей")

	// Находим путь в реестре
	registryPath, foundInRegistry := findPathInRegistry()
	log.Printf("Путь из реестра: \"%s\", найдено: %t\n", registryPath, foundInRegistry)

	// Параллельно ищем пути на дисках
	foundFolders := s.FindConfigOrExeParallel()
	log.Printf("Найденные папки: %+v\n", foundFolders)

	// Объединяем все найденные пути
	combinedPathsMap := make(map[string]struct{})
	if foundInRegistry {
		combinedPathsMap[registryPath] = struct{}{}
	}
	for _, p := range foundFolders {
		combinedPathsMap[p] = struct{}{}
	}

	var resultPaths []string
	for p := range combinedPathsMap {
		resultPaths = append(resultPaths, p)
	}

	log.Printf("Всего найдено уникальных путей: %+v\n", resultPaths)
	log.Println("=== Конец CheckAndFindPaths ===")

	return resultPaths, nil
}

// contains - вспомогательная функция для проверки, содержит ли слайс строку
// func contains(slice []string, item string) bool {
// 	for _, s := range slice {
// 		if s == item {
// 			return true
// 		}
// 	}
// 	return false
// }
