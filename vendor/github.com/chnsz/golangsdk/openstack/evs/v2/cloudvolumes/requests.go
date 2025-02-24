package cloudvolumes

import (
	"strings"

	"github.com/chnsz/golangsdk"
	"github.com/chnsz/golangsdk/openstack/common/structs"
	"github.com/chnsz/golangsdk/pagination"
)

// CreateOptsBuilder allows extensions to add additional parameters to the
// Create request.
type CreateOptsBuilder interface {
	ToVolumeCreateMap() (map[string]interface{}, error)
}

// CreateOpts contains options for creating a Volume. This object is passed to
// the cloudvolumes.Create function.
type CreateOpts struct {
	Volume     VolumeOpts          `json:"volume" required:"true"`
	ChargeInfo *structs.ChargeInfo `json:"bssParam,omitempty"`
	Scheduler  *SchedulerOpts      `json:"OS-SCH-HNT:scheduler_hints,omitempty"`
	ServerID   string              `json:"server_id,omitempty"`
}

// VolumeOpts contains options for creating a Volume.
type VolumeOpts struct {
	// The availability zone
	AvailabilityZone string `json:"availability_zone" required:"true"`
	// The associated volume type
	VolumeType string `json:"volume_type" required:"true"`
	// The volume name
	Name string `json:"name,omitempty"`
	// The volume description
	Description string `json:"description,omitempty"`
	// The size of the volume, in GB
	Size int `json:"size,omitempty"`
	// The number to be created in a batch
	Count int `json:"count,omitempty"`
	// The backup_id
	BackupID string `json:"backup_id,omitempty"`
	// the ID of the existing volume snapshot
	SnapshotID string `json:"snapshot_id,omitempty"`
	// the ID of the image in IMS
	ImageID string `json:"imageRef,omitempty"`
	// Shared disk
	Multiattach bool `json:"multiattach,omitempty"`
	// One or more metadata key and value pairs to associate with the volume
	Metadata map[string]string `json:"metadata,omitempty"`
	// One or more tag key and value pairs to associate with the volume
	Tags map[string]string `json:"tags,omitempty"`
	// the enterprise project id
	EnterpriseProjectID string `json:"enterprise_project_id,omitempty"`
}

// SchedulerOpts contains the scheduler hints
type SchedulerOpts struct {
	StorageID string `json:"dedicated_storage_id,omitempty"`
}

// ToVolumeCreateMap assembles a request body based on the contents of a
// CreateOpts.
func (opts CreateOpts) ToVolumeCreateMap() (map[string]interface{}, error) {
	return golangsdk.BuildRequestBody(opts, "")
}

// Create will create a new Volume based on the values in CreateOpts.
func Create(client *golangsdk.ServiceClient, opts CreateOptsBuilder) (r JobResult) {
	b, err := opts.ToVolumeCreateMap()
	if err != nil {
		r.Err = err
		return
	}

	// the version of create API is v2.1
	newClient := *client
	baseURL := newClient.ResourceBaseURL()
	newClient.ResourceBase = strings.Replace(baseURL, "/v2/", "/v2.1/", 1)

	_, r.Err = newClient.Post(createURL(&newClient), b, &r.Body, nil)
	return
}

// ExtendOptsBuilder allows extensions to add additional parameters to the
// ExtendSize request.
type ExtendOptsBuilder interface {
	ToVolumeExtendMap() (map[string]interface{}, error)
}

// ExtendOpts contains options for extending the size of an existing Volume.
// This object is passed to the cloudvolumes.ExtendSize function.
type ExtendOpts struct {
	SizeOpts   ExtendSizeOpts    `json:"os-extend" required:"true"`
	ChargeInfo *ExtendChargeOpts `json:"bssParam,omitempty"`
}

// ExtendSizeOpts contains the new size of the volume, in GB.
type ExtendSizeOpts struct {
	NewSize int `json:"new_size" required:"true"`
}

// ExtendChargeOpts contains the charging parameters of the volume
type ExtendChargeOpts struct {
	IsAutoPay string `json:"is_auto_pay,omitempty"`
}

// ToVolumeExtendMap assembles a request body based on the contents of an
// ExtendOpts.
func (opts ExtendOpts) ToVolumeExtendMap() (map[string]interface{}, error) {
	return golangsdk.BuildRequestBody(opts, "")
}

// ExtendSize will extend the size of the volume based on the provided information.
// This operation does not return a response body.
func ExtendSize(client *golangsdk.ServiceClient, id string, opts ExtendOptsBuilder) (r JobResult) {
	b, err := opts.ToVolumeExtendMap()
	if err != nil {
		r.Err = err
		return
	}
	// the version of extend API is v2.1
	newClient := *client
	baseURL := newClient.ResourceBaseURL()
	newClient.ResourceBase = strings.Replace(baseURL, "/v2/", "/v2.1/", 1)

	_, r.Err = newClient.Post(actionURL(&newClient, id), b, nil, &golangsdk.RequestOpts{
		OkCodes: []int{202},
	})
	return
}

// UpdateOptsBuilder allows extensions to add additional parameters to the
// Update request.
type UpdateOptsBuilder interface {
	ToVolumeUpdateMap() (map[string]interface{}, error)
}

// UpdateOpts contain options for updating an existing Volume. This object is passed
// to the cloudvolumes.Update function.
type UpdateOpts struct {
	Name        string  `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// ToVolumeUpdateMap assembles a request body based on the contents of an
// UpdateOpts.
func (opts UpdateOpts) ToVolumeUpdateMap() (map[string]interface{}, error) {
	return golangsdk.BuildRequestBody(opts, "volume")
}

// Update will update the Volume with provided information. To extract the updated
// Volume from the response, call the Extract method on the UpdateResult.
func Update(client *golangsdk.ServiceClient, id string, opts UpdateOptsBuilder) (r UpdateResult) {
	b, err := opts.ToVolumeUpdateMap()
	if err != nil {
		r.Err = err
		return
	}
	_, r.Err = client.Put(resourceURL(client, id), b, &r.Body, &golangsdk.RequestOpts{
		OkCodes: []int{200},
	})
	return
}

// DeleteOptsBuilder is an interface by which can be able to build the query string
// of volume deletion.
type DeleteOptsBuilder interface {
	ToVolumeDeleteQuery() (string, error)
}

// DeleteOpts contain options for deleting an existing Volume. This object is passed
// to the cloudvolumes.Delete function.
type DeleteOpts struct {
	// Specifies to delete all snapshots associated with the EVS disk.
	Cascade bool `q:"cascade"`
}

// ToVolumeDeleteQuery assembles a request body based on the contents of an
// DeleteOpts.
func (opts DeleteOpts) ToVolumeDeleteQuery() (string, error) {
	q, err := golangsdk.BuildQueryString(opts)
	return q.String(), err
}

// Delete will delete the existing Volume with the provided ID
func Delete(client *golangsdk.ServiceClient, id string, opts DeleteOptsBuilder) (r DeleteResult) {
	url := resourceURL(client, id)
	if opts != nil {
		q, err := opts.ToVolumeDeleteQuery()
		if err != nil {
			r.Err = err
			return
		}
		url += q
	}
	_, r.Err = client.Delete(url, &golangsdk.RequestOpts{
		OkCodes: []int{200},
	})
	return
}

// Get retrieves the Volume with the provided ID. To extract the Volume object
// from the response, call the Extract method on the GetResult.
func Get(client *golangsdk.ServiceClient, id string) (r GetResult) {
	_, r.Err = client.Get(resourceURL(client, id), &r.Body, nil)
	return
}

// ListOptsBuilder allows extensions to add additional parameters to the List
// request.
type ListOptsBuilder interface {
	ToVolumeListQuery() (string, error)
}

// ListOpts holds options for listing Volumes. It is passed to the volumes.List
// function.
type ListOpts struct {
	// Name will filter by the specified volume name.
	Name string `q:"name"`

	// Status will filter by the specified status.
	Status string `q:"status"`

	// Metadata will filter results based on specified metadata.
	Metadata map[string]string `q:"metadata"`

	ID string `q:"id"`

	ServerID string `q:"server_id"`

	SortKey string `q:"sort_key"`
	SortDir string `q:"sort_dir"`

	// Requests a page size of items.
	Limit int `q:"limit"`

	// Used in conjunction with limit to return a slice of items.
	Offset int `q:"offset"`

	// The ID of the last-seen item.
	Marker string `q:"marker"`
}

// ToVolumeListQuery formats a ListOpts into a query string.
func (opts ListOpts) ToVolumeListQuery() (string, error) {
	q, err := golangsdk.BuildQueryString(opts)
	return q.String(), err
}

// List returns Volumes optionally limited by the conditions provided in ListOpts.
func List(client *golangsdk.ServiceClient, opts ListOptsBuilder) pagination.Pager {
	url := listURL(client)
	if opts != nil {
		query, err := opts.ToVolumeListQuery()
		if err != nil {
			return pagination.Pager{Err: err}
		}
		url += query
	}

	return pagination.NewPager(client, url, func(r pagination.PageResult) pagination.Page {
		return VolumePage{pagination.LinkedPageBase{PageResult: r}}
	})
}
