package fileexts

import (
	"fmt"
	"os"
	"strings"

	"github.com/bodrovis/lokalise-actions-common/v2/normalizers"
	"github.com/bodrovis/lokalise-actions-common/v2/parsers"
)

// ResolveFromEnv returns normalized file extensions from FILE_EXT or,
// if FILE_EXT is not provided, falls back to FILE_FORMAT.
func ResolveFromEnv(fileExtEnv, fileFormatEnv string) ([]string, error) {
	extensions := parsers.ParseStringArrayEnv(fileExtEnv)
	if len(extensions) == 0 {
		if inferred := strings.TrimSpace(os.Getenv(fileFormatEnv)); inferred != "" {
			extensions = []string{inferred}
		}
	}

	if len(extensions) == 0 {
		return nil, fmt.Errorf(
			"cannot infer file extension: make sure %s or %s environment variable is set",
			fileExtEnv,
			fileFormatEnv,
		)
	}

	return normalizers.NormalizeFileExtensions(extensions)
}
