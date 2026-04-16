package keymap

import (
	"testing"

	"github.com/programmersd21/kairo/internal/config"
)

func TestFromConfig_TaskSearch(t *testing.T) {
	t.Parallel()

	km := FromConfig(config.KeymapConfig{
		TaskSearch: "/",
	})
	keys := km.TaskSearch.Keys()
	if len(keys) != 1 || keys[0] != "/" {
		t.Fatalf("expected TaskSearch keys to be [/], got %v", keys)
	}
}
