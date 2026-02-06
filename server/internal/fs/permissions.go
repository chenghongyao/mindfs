package fs

import "errors"

// Permissions defines basic access control for a managed directory.
type Permissions struct {
	ReadOnly bool
}

// RequireWriteAllowed returns an error if write operations are not permitted.
func RequireWriteAllowed(perms Permissions) error {
	if perms.ReadOnly {
		return errors.New("write operations are not allowed")
	}
	return nil
}
