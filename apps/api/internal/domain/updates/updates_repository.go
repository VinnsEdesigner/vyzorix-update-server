package updates

import (
	"context"
)

// Repository defines the interface for update data access.
type Repository interface {
	// Version operations
	CreateVersion(ctx context.Context, v *UpdateVersion) error
	GetVersionByID(ctx context.Context, id string) (*UpdateVersion, error)
	GetVersionByVersion(ctx context.Context, version string) (*UpdateVersion, error)
	GetLatestVersion(ctx context.Context) (*UpdateVersion, error)
	ListVersions(ctx context.Context, status string, limit, offset int) ([]*UpdateVersion, int, error)
	UpdateVersion(ctx context.Context, v *UpdateVersion) error
	UpdateLatestFlag(ctx context.Context, versionID string) error
	DeleteVersion(ctx context.Context, id string) error

	// Push operations
	CreatePush(ctx context.Context, p *UpdatePush) error
	GetPushByID(ctx context.Context, id string) (*UpdatePush, error)
	GetPushByIDWithVersion(ctx context.Context, id string) (*UpdatePush, *UpdateVersion, error)
	UpdatePushStatus(ctx context.Context, id string, status UpdateStatus) error
	CompletePush(ctx context.Context, id string) error
	CancelPush(ctx context.Context, id, cancelledBy string) error
	ListPushes(ctx context.Context, status string, limit, offset int) ([]*UpdatePush, int, error)

	// Push device operations
	CreatePushDevice(ctx context.Context, d *UpdatePushDevice) error
	GetPushDevices(ctx context.Context, pushID string) ([]*UpdatePushDevice, error)
	GetPushDeviceByPushAndDevice(ctx context.Context, pushID, deviceID string) (*UpdatePushDevice, error)
	UpdatePushDeviceStatus(ctx context.Context, id string, status DevicePushStatus, errorMsg string) error
	UpdatePushDeviceStatusByDispatch(ctx context.Context, dispatchID, deviceID string, status DevicePushStatus, errorMsg string) error
	CountPushDevicesByStatus(ctx context.Context, pushID string, status DevicePushStatus) (int, error)

	// Sync state operations
	GetSyncState(ctx context.Context) (*SyncState, error)
	UpdateSyncState(ctx context.Context, state *SyncState) error
	TryAcquireSyncLock(ctx context.Context) (bool, *SyncState, error)
}
