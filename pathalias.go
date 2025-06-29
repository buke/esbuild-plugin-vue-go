// Copyright 2025 Brian Wang <wangbuke@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package vueplugin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"

	"github.com/evanw/esbuild/pkg/api"
)

// parseTsconfigPathAlias parses path aliases from tsconfig for module resolution.
// It supports both raw tsconfig JSON and file-based tsconfig.
// Returns a map of alias to real path.
func parseTsconfigPathAlias(buildOptions *api.BuildOptions) (map[string]string, error) {
	var tsconfigAbsDir string
	pathAlias := make(map[string]string)
	var tsconfig map[string]interface{}

	// Parse from raw tsconfig JSON if provided
	if buildOptions.TsconfigRaw != "" {
		err := json.Unmarshal([]byte(buildOptions.TsconfigRaw), &tsconfig)
		if err != nil {
			return pathAlias, err
		}
		if buildOptions.AbsWorkingDir != "" {
			tsconfigAbsDir = buildOptions.AbsWorkingDir
		} else {
			exePath, _ := os.Executable()
			tsconfigAbsDir, _ = filepath.Abs(filepath.Dir(exePath))
		}
		// Otherwise, parse from tsconfig file if provided
	} else if buildOptions.Tsconfig != "" {
		file, err := os.Open(buildOptions.Tsconfig)
		if err != nil {
			return pathAlias, err
		}
		defer file.Close()
		err = json.NewDecoder(file).Decode(&tsconfig)
		if err != nil {
			return pathAlias, err
		}
		tsconfigAbsDir, _ = filepath.Abs(filepath.Dir(buildOptions.Tsconfig))
	}

	// Extract path aliases from compilerOptions.paths
	if compilerOptions, ok := tsconfig["compilerOptions"].(map[string]interface{}); ok {
		if paths, ok := compilerOptions["paths"].(map[string]interface{}); ok {
			for key, value := range paths {
				if pathArray, ok := value.([]interface{}); ok && len(pathArray) > 0 {
					if pathStr, ok := pathArray[0].(string); ok {
						pathAlias[key] = filepath.Join(tsconfigAbsDir, pathStr)
					}
				}
			}
		}
	}

	return pathAlias, nil
}

// applyPathAlias applies path alias mapping to the given path.
// It supports wildcard '*' in the alias and replaces it accordingly.
func applyPathAlias(pathAlias map[string]string, path string) string {
	for alias, realPath := range pathAlias {
		aliasPattern := "^" + regexp.QuoteMeta(alias)
		// Support wildcard '*' in alias
		if alias[len(alias)-1] == '*' {
			aliasPattern = aliasPattern[:len(aliasPattern)-2] + "(.*)"
			realPath = realPath[:len(realPath)-1] + "$1"
		}
		re := regexp.MustCompile(aliasPattern)
		if re.MatchString(path) {
			return re.ReplaceAllString(path, realPath)
		}
	}
	return path
}
