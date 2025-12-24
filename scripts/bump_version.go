package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run scripts/bump_version.go <version>")
		os.Exit(1)
	}
	newVersion := os.Args[1]
	
	// Windows often requires a 4-part version string (e.g., 1.0.0.0)
	winVersion := newVersion
	if strings.Count(winVersion, ".") == 2 {
		winVersion += ".0"
	}

	fmt.Printf("Bumping version to %s (Windows: %s)\n", newVersion, winVersion)

	// 1. manifest.json
	if err := updateManifest("manifest.json", newVersion); err != nil {
		fmt.Printf("Error updating manifest.json: %v\n", err)
		os.Exit(1)
	}

	// 2. build/config.yml
	if err := updateConfig("build/config.yml", newVersion); err != nil {
		fmt.Printf("Error updating build/config.yml: %v\n", err)
		os.Exit(1)
	}

	// 3. build/windows/info.json
	if err := updateWindowsInfo("build/windows/info.json", winVersion); err != nil {
		fmt.Printf("Error updating build/windows/info.json: %v\n", err)
		os.Exit(1)
	}

	// 4. build/windows/wails.exe.manifest
	if err := updateWindowsManifest("build/windows/wails.exe.manifest", winVersion); err != nil {
		fmt.Printf("Error updating build/windows/wails.exe.manifest: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully updated all version files!")
}

func updateManifest(path string, version string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// We use regex to preserve structure/comments if any, though JSON doesn't support comments usually.
	// But standard JSON unmarshal/marshal might reorder keys.
	// Let's try to parse as map to update specific field, but regex is safer for minimal diffs.
	
	// Pattern: "app_version": { "version": "..."
	// This is a bit complex with regex because of newlines.
	// Let's use a simple approach: Read as map, update, write back with indentation.
	// The user's manifest is simple enough.

	var data map[string]interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		return err
	}

	appVer, ok := data["app_version"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("app_version not found in manifest.json")
	}
	appVer["version"] = version

	newContent, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, newContent, 0644)
}

func updateConfig(path string, version string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// YAML: version: "..."
	re := regexp.MustCompile(`version: ".*"`)
	newContent := re.ReplaceAllString(string(content), fmt.Sprintf(`version: "%s"`, version))

	// Also update the top-level version: '3' which is the file format version, we should NOT touch that.
	// The app version is under 'info:'.
	// The regex above might match the top level version if it was quoted, but it is version: '3'.
	// The info version is version: "0.9.6".
	// To be safe, let's target '  version: "..."' (indented).
	
	reIndented := regexp.MustCompile(`  version: ".*"`)
	newContent = reIndented.ReplaceAllString(string(content), fmt.Sprintf(`  version: "%s"`, version))

	return os.WriteFile(path, []byte(newContent), 0644)
}

func updateWindowsInfo(path string, version string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// JSON: "ProductVersion": "..."
	strContent := string(content)

	reProdVer := regexp.MustCompile(`"ProductVersion": ".*"`)
	strContent = reProdVer.ReplaceAllString(strContent, fmt.Sprintf(`"ProductVersion": "%s"`, version))

	return os.WriteFile(path, []byte(strContent), 0644)
}

func updateWindowsManifest(path string, version string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// XML: version="..." inside assemblyIdentity
	// <assemblyIdentity ... version="0.9.6.0" ... />
	// There are two assemblyIdentity tags. One for the app, one for Common-Controls.
	// The app one has name="com.phphacker.lodconfigeditor".
	
	strContent := string(content)
	
	// We want to replace version in the line containing our app name.
	lines := strings.Split(strContent, "\n")
	for i, line := range lines {
		if strings.Contains(line, "assemblyIdentity") && strings.Contains(line, "com.phphacker.lodconfigeditor") {
			re := regexp.MustCompile(`version="[^"]*"`)
			lines[i] = re.ReplaceAllString(line, fmt.Sprintf(`version="%s"`, version))
			break
		}
	}

	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
}
