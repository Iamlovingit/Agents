package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"

	appconfig "github.com/iamlovingit/clawmanager-openclaw-image/internal/config"
)

// Global openclaw package path (see Dockerfile: npm install -g openclaw).
var openclawGlobalNodeModules = "/usr/local/lib/node_modules/openclaw"

// ensureDingtalkOpenclawSymlink runs after defaults sync and applyOwnership. It
// makes dingtalk-connector resolve the "openclaw" package the same as the
// image-wide install: node_modules/openclaw -> /usr/local/lib/node_modules/openclaw
func ensureDingtalkOpenclawSymlink(cfg appconfig.Config) error {
	if cfg.OpenClawExtensionsDir == "" {
		return nil
	}
	connectorDir := filepath.Join(cfg.OpenClawExtensionsDir, "dingtalk-connector")
	if !pathExists(connectorDir) {
		return nil
	}
	linkPath := filepath.Join(connectorDir, "node_modules", "openclaw")
	parent := filepath.Dir(linkPath)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", parent, err)
	}

	if err := ensureSymlink(openclawGlobalNodeModules, linkPath); err != nil {
		return err
	}
	if os.Geteuid() == 0 && cfg.DropUserName != "" {
		uid, gid, err := lookupDropUser(cfg.DropUserName)
		if err != nil {
			return err
		}
		if err := os.Lchown(linkPath, uid, gid); err != nil {
			return fmt.Errorf("lchown %s: %w", linkPath, err)
		}
	}
	return nil
}

func ensureSymlink(target, linkPath string) error {
	if fi, err := os.Lstat(linkPath); err == nil {
		if fi.Mode()&os.ModeSymlink != 0 {
			cur, rerr := os.Readlink(linkPath)
			if rerr != nil {
				return fmt.Errorf("readlink %s: %w", linkPath, rerr)
			}
			if cur == target {
				return nil
			}
		}
		if err := os.RemoveAll(linkPath); err != nil {
			return fmt.Errorf("remove %s: %w", linkPath, err)
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.Symlink(target, linkPath); err != nil {
		return fmt.Errorf("symlink %s -> %s: %w", linkPath, target, err)
	}
	return nil
}
