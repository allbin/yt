package youtrack

import (
	"fmt"
	"net/url"
)

// bundlePaths maps a bundle $type to its admin API path segment.
var bundlePaths = map[string]string{
	"EnumBundle":    "enum",
	"OwnedBundle":   "ownedField",
	"VersionBundle": "version",
	"BuildBundle":   "build",
	"StateBundle":   "state",
}

// AddBundleValue adds a value to the bundle behind a custom field. A bundle
// can be shared by several projects; the value appears in all of them.
func (c *Client) AddBundleValue(f ProjectField, name string) error {
	seg, ok := bundlePaths[f.BundleType]
	if !ok {
		return fmt.Errorf("%s: cannot add values to a %s field", f.Name, f.Type)
	}
	path := "/api/admin/customFieldSettings/bundles/" + seg + "/" + url.PathEscape(f.BundleID) + "/values"
	body := namedValue{Name: name}
	if err := c.post(path, body); err != nil {
		return fmt.Errorf("add %q to %s: %w", name, f.Name, err)
	}
	return nil
}
