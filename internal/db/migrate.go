package db

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Migrate(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var upFiles []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".up.sql") {
			upFiles = append(upFiles, e.Name())
		}
	}
	sort.Strings(upFiles)

	for _, f := range upFiles {
		content, err := os.ReadFile(filepath.Join(dir, f))
		if err != nil {
			return fmt.Errorf("read %s: %w", f, err)
		}

		// Execute SQL with simple protocol (no prepared statements)
		sql := string(content)
		_, err = pool.Exec(ctx, sql, pgx.QueryExecModeSimpleProtocol)
		if err != nil {
			// #region agent log - migration failed
			logf, _ := os.OpenFile("/home/user/git/ai-for-developers-project-386/.cursor/debug-ff3c7c.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if logf != nil {
				enc := json.NewEncoder(logf)
				enc.Encode(map[string]interface{}{"sessionId":"ff3c7c","timestamp":time.Now().UnixMilli(),"location":"migrate.go:40","message":"migration failed","data":map[string]interface{}{"migration":f,"error":err.Error(),"hypothesisId":"C"}})
				logf.Close()
			}
			// #endregion
			return fmt.Errorf("exec %s: %w", f, err)
		}
		// #region agent log - migration applied
		logf, _ := os.OpenFile("/home/user/git/ai-for-developers-project-386/.cursor/debug-ff3c7c.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if logf != nil {
			enc := json.NewEncoder(logf)
			enc.Encode(map[string]interface{}{"sessionId":"ff3c7c","timestamp":time.Now().UnixMilli(),"location":"migrate.go:42","message":"migration applied","data":map[string]interface{}{"migration":f,"hypothesisId":"C"}})
			logf.Close()
		}
		// #endregion
		fmt.Printf("Applied migration: %s\n", f)
	}
	return nil
}
