package storageintel

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Entry struct {
	Path        string    `json:"path"`
	Name        string    `json:"name"`
	SizeBytes   int64     `json:"size_bytes"`
	Kind        string    `json:"kind"`
	Heat        string    `json:"heat"`
	Description string    `json:"description,omitempty"`
	ModifiedAt  time.Time `json:"modified_at,omitempty"`
	Children    []Entry   `json:"children,omitempty"`
}

type Summary struct {
	Root       string    `json:"root"`
	Generated  time.Time `json:"generated"`
	TotalBytes int64     `json:"total_bytes"`
	Entries    []Entry   `json:"entries"`
	Detected   []Entry   `json:"detected"`
}

func Analyze(root string, depth int, limit int) (Summary, error) {
	if root == "" {
		root = string(filepath.Separator)
	}
	if depth <= 0 {
		depth = 2
	}
	if filepath.Clean(root) == string(filepath.Separator) && depth < 4 {
		depth = 4
	}
	if limit <= 0 {
		limit = 24
	}
	if filepath.Clean(root) == string(filepath.Separator) && limit < 80 {
		limit = 80
	}
	entries, total := scan(root, depth)
	sort.Slice(entries, func(i, j int) bool { return entries[i].SizeBytes > entries[j].SizeBytes })
	if len(entries) > limit {
		entries = entries[:limit]
	}
	return Summary{
		Root:       root,
		Generated:  time.Now(),
		TotalBytes: total,
		Entries:    entries,
		Detected:   detect(root, limit),
	}, nil
}

func scan(root string, depth int) ([]Entry, int64) {
	dirents, err := os.ReadDir(root)
	if err != nil {
		return nil, 0
	}
	entries := []Entry{}
	var total int64
	for _, dirent := range dirents {
		if strings.HasPrefix(dirent.Name(), ".") {
			continue
		}
		path := filepath.Join(root, dirent.Name())
		if shouldSkip(path) {
			continue
		}
		size, mod := sizeOf(path, depth-1)
		if size == 0 {
			continue
		}
		total += size
		entries = append(entries, Entry{
			Path:        path,
			Name:        dirent.Name(),
			SizeBytes:   size,
			Kind:        classify(path),
			Heat:        heat(size),
			Description: Describe(path),
			ModifiedAt:  mod,
			Children:    childrenOf(path, depth-1, 14),
		})
	}
	return entries, total
}

func childrenOf(path string, depth int, limit int) []Entry {
	info, err := os.Lstat(path)
	if err != nil || !info.IsDir() || depth <= 0 {
		return nil
	}
	dirents, err := os.ReadDir(path)
	if err != nil {
		return nil
	}
	children := []Entry{}
	for _, dirent := range dirents {
		if strings.HasPrefix(dirent.Name(), ".") {
			continue
		}
		childPath := filepath.Join(path, dirent.Name())
		if shouldSkip(childPath) {
			continue
		}
		size, mod := sizeOf(childPath, depth-1)
		if size == 0 {
			continue
		}
		children = append(children, Entry{
			Path:        childPath,
			Name:        dirent.Name(),
			SizeBytes:   size,
			Kind:        classify(childPath),
			Heat:        heat(size),
			Description: Describe(childPath),
			ModifiedAt:  mod,
		})
	}
	sort.Slice(children, func(i, j int) bool { return children[i].SizeBytes > children[j].SizeBytes })
	if len(children) > limit {
		return children[:limit]
	}
	return children
}

func sizeOf(path string, depth int) (int64, time.Time) {
	info, err := os.Lstat(path)
	if err != nil {
		return 0, time.Time{}
	}
	if !info.IsDir() || depth <= 0 {
		return info.Size(), info.ModTime()
	}
	var total int64
	mod := info.ModTime()
	dirents, err := os.ReadDir(path)
	if err != nil {
		return info.Size(), mod
	}
	for _, dirent := range dirents {
		child := filepath.Join(path, dirent.Name())
		if shouldSkip(child) {
			continue
		}
		childSize, childMod := sizeOf(child, depth-1)
		total += childSize
		if childMod.After(mod) {
			mod = childMod
		}
	}
	return total, mod
}

func detect(root string, limit int) []Entry {
	patterns := artifactPatterns(root)
	found := []Entry{}
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		for _, path := range matches {
			if shouldSkip(path) {
				continue
			}
			if _, err := os.Stat(path); err != nil {
				continue
			}
			size, mod := sizeOf(path, 2)
			found = append(found, Entry{Path: path, Name: filepath.Base(path), SizeBytes: size, Kind: classify(path), Heat: heat(size), Description: Describe(path), ModifiedAt: mod})
		}
	}
	sort.Slice(found, func(i, j int) bool { return found[i].SizeBytes > found[j].SizeBytes })
	if len(found) > limit {
		return found[:limit]
	}
	return found
}

func Describe(path string) string {
	clean := filepath.Clean(path)
	lower := strings.ToLower(clean)
	switch {
	case clean == string(filepath.Separator):
		return "Filesystem root. Top-level operating system, user, application, and runtime directories live here."
	case clean == "/Users" || clean == "/home":
		return "User home directories. Personal files, profiles, application data, caches, and workspaces usually accumulate here."
	case strings.HasPrefix(clean, "/Users/") || strings.HasPrefix(clean, "/home/"):
		return "User-owned workspace. Review downloads, media, caches, documents, and application support data before deleting anything."
	case clean == "/System":
		return "macOS system files. Usually managed by the operating system and not a cleanup target."
	case clean == "/Library":
		return "Shared application support, logs, fonts, launch agents, caches, and machine-wide resources."
	case clean == "/Applications":
		return "Installed applications. Large apps can be removed through normal OS uninstall workflows."
	case clean == "/private" || clean == "/var":
		return "Runtime data, logs, temporary files, package state, and service storage. Inspect carefully before cleanup."
	case clean == "/usr" || clean == "/bin" || clean == "/sbin":
		return "System command and library area. Treat as operating-system managed."
	case clean == "/opt":
		return "Optional third-party software and local tool installations often live here."
	case clean == "/tmp" || clean == "/var/tmp" || strings.HasSuffix(lower, "/tmp"):
		return "Temporary files. Usually safe to inspect, but active applications may still be using recent files."
	case strings.Contains(lower, "downloads"):
		return "Downloaded files. Often a high-value cleanup area for installers, archives, exports, and media."
	case strings.Contains(lower, "cache"):
		return "Cache data created to speed up applications. Usually regenerates, but deleting it may temporarily slow apps."
	case strings.Contains(lower, "node_modules"):
		return "JavaScript package dependencies. Can be recreated with the project package manager."
	case strings.Contains(lower, "deriveddata"):
		return "Xcode build and index artifacts. Often safe to clear when Xcode is closed."
	case strings.Contains(lower, "docker") || strings.Contains(lower, "container"):
		return "Container images, volumes, and runtime data. Clean through Docker or container tooling to avoid removing active volumes."
	case strings.Contains(lower, "log"):
		return "Application and system logs. Useful for diagnostics; old logs can grow over time."
	default:
		return "Directory sampled by NexPerf storage intelligence. Size is estimated within the configured scan depth."
	}
}

func artifactPatterns(root string) []string {
	cleanRoot := filepath.Clean(root)
	local := func(parts ...string) string {
		all := append([]string{cleanRoot}, parts...)
		return filepath.Join(all...)
	}
	if cleanRoot != string(filepath.Separator) {
		return []string{
			local("node_modules"),
			local("Library", "Developer", "Xcode", "DerivedData"),
			local("Library", "Caches"),
			local(".npm"),
			local(".cache"),
			local(".docker"),
			local("Downloads"),
			local("tmp"),
		}
	}
	return []string{
		filepath.Join(string(filepath.Separator), "Users", "*", "Downloads"),
		filepath.Join(string(filepath.Separator), "Users", "*", "Library", "Caches"),
		filepath.Join(string(filepath.Separator), "Users", "*", "Library", "Developer", "Xcode", "DerivedData"),
		filepath.Join(string(filepath.Separator), "Users", "*", ".npm"),
		filepath.Join(string(filepath.Separator), "Users", "*", ".cache"),
		filepath.Join(string(filepath.Separator), "Users", "*", ".docker"),
		filepath.Join(string(filepath.Separator), "home", "*", "Downloads"),
		filepath.Join(string(filepath.Separator), "home", "*", ".cache"),
		filepath.Join(string(filepath.Separator), "home", "*", ".docker"),
		filepath.Join(string(filepath.Separator), "var", "log"),
		filepath.Join(string(filepath.Separator), "var", "tmp"),
		filepath.Join(string(filepath.Separator), "private", "tmp"),
		filepath.Join(string(filepath.Separator), "private", "var", "tmp"),
		filepath.Join(string(filepath.Separator), "private", "var", "folders"),
		filepath.Join(string(filepath.Separator), "private", "var", "log"),
		filepath.Join(string(filepath.Separator), "tmp"),
		filepath.Join(string(filepath.Separator), "opt"),
	}
}

func shouldSkip(path string) bool {
	clean := filepath.Clean(path)
	switch clean {
	case "/dev", "/proc", "/sys", "/run", "/private/var/run":
		return true
	default:
		return false
	}
}

func classify(path string) string {
	lower := strings.ToLower(path)
	switch {
	case strings.Contains(lower, "node_modules"):
		return "dependencies"
	case strings.Contains(lower, "docker"):
		return "containers"
	case strings.Contains(lower, "deriveddata") || strings.Contains(lower, "build"):
		return "build artifacts"
	case strings.Contains(lower, "cache"):
		return "cache"
	case strings.Contains(lower, "downloads"):
		return "downloads"
	case strings.Contains(lower, "sqlite") || strings.HasSuffix(lower, ".db"):
		return "database"
	default:
		return "directory"
	}
}

func heat(size int64) string {
	gb := float64(size) / 1024 / 1024 / 1024
	switch {
	case gb >= 10:
		return "hot"
	case gb >= 2:
		return "warm"
	default:
		return "cool"
	}
}
