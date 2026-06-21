package updater

import "errors"

var (
    ErrNoUpdateAvailable = errors.New("no update available")
    ErrInvalidVersion    = errors.New("invalid version")
    ErrUpdateFailed     = errors.New("update failed")
    ErrChecksumMismatch = errors.New("checksum mismatch")
)
