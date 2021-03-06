// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// KlusterSpec kluster spec
// swagger:model KlusterSpec
type KlusterSpec struct {

	// advertise address
	AdvertiseAddress string `json:"advertiseAddress,omitempty"`

	// CIDR Range for Pods in the cluster. Can not be updated.
	// Pattern: ^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])(\/([0-9]|[1-2][0-9]|3[0-2]))$
	ClusterCIDR string `json:"clusterCIDR,omitempty"`

	// dns address
	DNSAddress string `json:"dnsAddress,omitempty"`

	// dns domain
	DNSDomain string `json:"dnsDomain,omitempty"`

	// name
	// Read Only: true
	Name string `json:"name,omitempty"`

	// node pools
	NodePools []NodePool `json:"nodePools"`

	// openstack
	Openstack OpenstackSpec `json:"openstack,omitempty"`

	// CIDR Range for Services in the cluster. Can not be updated.
	// Pattern: ^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])(\/([0-9]|[1-2][0-9]|3[0-2]))$
	ServiceCIDR string `json:"serviceCIDR,omitempty"`
}

// Validate validates this kluster spec
func (m *KlusterSpec) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateClusterCIDR(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateNodePools(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateOpenstack(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateServiceCIDR(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *KlusterSpec) validateClusterCIDR(formats strfmt.Registry) error {

	if swag.IsZero(m.ClusterCIDR) { // not required
		return nil
	}

	if err := validate.Pattern("clusterCIDR", "body", string(m.ClusterCIDR), `^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])(\/([0-9]|[1-2][0-9]|3[0-2]))$`); err != nil {
		return err
	}

	return nil
}

func (m *KlusterSpec) validateNodePools(formats strfmt.Registry) error {

	if swag.IsZero(m.NodePools) { // not required
		return nil
	}

	return nil
}

func (m *KlusterSpec) validateOpenstack(formats strfmt.Registry) error {

	if swag.IsZero(m.Openstack) { // not required
		return nil
	}

	if err := m.Openstack.Validate(formats); err != nil {
		if ve, ok := err.(*errors.Validation); ok {
			return ve.ValidateName("openstack")
		}
		return err
	}

	return nil
}

func (m *KlusterSpec) validateServiceCIDR(formats strfmt.Registry) error {

	if swag.IsZero(m.ServiceCIDR) { // not required
		return nil
	}

	if err := validate.Pattern("serviceCIDR", "body", string(m.ServiceCIDR), `^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])(\/([0-9]|[1-2][0-9]|3[0-2]))$`); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *KlusterSpec) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *KlusterSpec) UnmarshalBinary(b []byte) error {
	var res KlusterSpec
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
