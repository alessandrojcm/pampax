package discovery

import (
	"errors"
	"fmt"
	"os"
)

func classifyReadDirError(path string, err error) Warning {
	if errors.Is(err, os.ErrPermission) {
		return Warning{
			Code:    WarningPermissionDenied,
			Path:    path,
			Message: fmt.Sprintf("permission denied while reading directory: %v", err),
		}
	}

	return Warning{
		Code:    WarningReadDirFailed,
		Path:    path,
		Message: fmt.Sprintf("failed to read directory: %v", err),
	}
}

func classifyStatError(path string, err error) Warning {
	if errors.Is(err, os.ErrPermission) {
		return Warning{
			Code:    WarningPermissionDenied,
			Path:    path,
			Message: fmt.Sprintf("permission denied while reading path metadata: %v", err),
		}
	}

	if errors.Is(err, os.ErrNotExist) {
		return Warning{
			Code:    WarningBrokenSymlink,
			Path:    path,
			Message: fmt.Sprintf("broken symlink target: %v", err),
		}
	}

	return Warning{
		Code:    WarningStatFailed,
		Path:    path,
		Message: fmt.Sprintf("failed to read path metadata: %v", err),
	}
}
