package cli

import (
	"os"
	"path/filepath"
	"sort"
)

func applyAdapterSync(state adapterSyncState) error {
	if err := os.MkdirAll(state.managedRoot, 0o755); err != nil {
		return err
	}
	if err := copyManagedFiles(state.sourceRoot, state.managedRoot, state.managedFiles); err != nil {
		return err
	}
	if err := removeStaleManagedFiles(state.managedRoot, state.managedFiles); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(state.manifestPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(state.manifestPath, state.manifest, 0o644); err != nil {
		return err
	}
	return nil
}

func copyManagedFiles(sourceRoot, managedRoot string, sourceFiles []string) error {
	for _, rel := range sourceFiles {
		srcPath := filepath.Join(sourceRoot, rel)
		dstPath := filepath.Join(managedRoot, rel)
		data, err := os.ReadFile(srcPath)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(dstPath, data, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func removeStaleManagedFiles(managedRoot string, sourceFiles []string) error {
	stale, err := collectStaleManagedFiles(managedRoot, sourceFiles)
	if err != nil {
		return err
	}
	for _, file := range stale {
		if err := os.Remove(file.absPath); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return pruneEmptyDirs(managedRoot)
}

func pruneEmptyDirs(root string) error {
	if !isDirectory(root) {
		return nil
	}
	dirs, err := collectDirectoryTree(root)
	if err != nil {
		return err
	}
	for _, path := range dirs {
		if path == root {
			continue
		}
		entries, err := os.ReadDir(path)
		if err != nil {
			continue
		}
		if len(entries) == 0 {
			_ = os.Remove(path)
		}
	}
	return nil
}

func collectDirectoryTree(root string) ([]string, error) {
	dirs := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			dirs = append(dirs, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(dirs, func(i, j int) bool { return len(dirs[i]) > len(dirs[j]) })
	return dirs, nil
}
