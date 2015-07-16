// Code generated by protoc-gen-go.
// source: domain.proto
// DO NOT EDIT!

package tao

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type DomainDetails struct {
	Name             *string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	PolicyKeysPath   *string `protobuf:"bytes,2,opt,name=policy_keys_path" json:"policy_keys_path,omitempty"`
	GuardType        *string `protobuf:"bytes,3,opt,name=guard_type" json:"guard_type,omitempty"`
	GuardNetwork     *string `protobuf:"bytes,4,opt,name=guard_network" json:"guard_network,omitempty"`
	GuardAddress     *string `protobuf:"bytes,5,opt,name=guard_address" json:"guard_address,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *DomainDetails) Reset()         { *m = DomainDetails{} }
func (m *DomainDetails) String() string { return proto.CompactTextString(m) }
func (*DomainDetails) ProtoMessage()    {}

func (m *DomainDetails) GetName() string {
	if m != nil && m.Name != nil {
		return *m.Name
	}
	return ""
}

func (m *DomainDetails) GetPolicyKeysPath() string {
	if m != nil && m.PolicyKeysPath != nil {
		return *m.PolicyKeysPath
	}
	return ""
}

func (m *DomainDetails) GetGuardType() string {
	if m != nil && m.GuardType != nil {
		return *m.GuardType
	}
	return ""
}

func (m *DomainDetails) GetGuardNetwork() string {
	if m != nil && m.GuardNetwork != nil {
		return *m.GuardNetwork
	}
	return ""
}

func (m *DomainDetails) GetGuardAddress() string {
	if m != nil && m.GuardAddress != nil {
		return *m.GuardAddress
	}
	return ""
}

type X509Details struct {
	CommonName         *string `protobuf:"bytes,1,opt,name=common_name" json:"common_name,omitempty"`
	Country            *string `protobuf:"bytes,2,opt,name=country" json:"country,omitempty"`
	State              *string `protobuf:"bytes,3,opt,name=state" json:"state,omitempty"`
	Organization       *string `protobuf:"bytes,4,opt,name=organization" json:"organization,omitempty"`
	OrganizationalUnit *string `protobuf:"bytes,5,opt,name=organizational_unit" json:"organizational_unit,omitempty"`
	SerialNumber       *int32  `protobuf:"varint,6,opt,name=serial_number" json:"serial_number,omitempty"`
	XXX_unrecognized   []byte  `json:"-"`
}

func (m *X509Details) Reset()         { *m = X509Details{} }
func (m *X509Details) String() string { return proto.CompactTextString(m) }
func (*X509Details) ProtoMessage()    {}

func (m *X509Details) GetCommonName() string {
	if m != nil && m.CommonName != nil {
		return *m.CommonName
	}
	return ""
}

func (m *X509Details) GetCountry() string {
	if m != nil && m.Country != nil {
		return *m.Country
	}
	return ""
}

func (m *X509Details) GetState() string {
	if m != nil && m.State != nil {
		return *m.State
	}
	return ""
}

func (m *X509Details) GetOrganization() string {
	if m != nil && m.Organization != nil {
		return *m.Organization
	}
	return ""
}

func (m *X509Details) GetOrganizationalUnit() string {
	if m != nil && m.OrganizationalUnit != nil {
		return *m.OrganizationalUnit
	}
	return ""
}

func (m *X509Details) GetSerialNumber() int32 {
	if m != nil && m.SerialNumber != nil {
		return *m.SerialNumber
	}
	return 0
}

type ACLGuardDetails struct {
	SignedAclsPath   *string `protobuf:"bytes,1,opt,name=signed_acls_path" json:"signed_acls_path,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *ACLGuardDetails) Reset()         { *m = ACLGuardDetails{} }
func (m *ACLGuardDetails) String() string { return proto.CompactTextString(m) }
func (*ACLGuardDetails) ProtoMessage()    {}

func (m *ACLGuardDetails) GetSignedAclsPath() string {
	if m != nil && m.SignedAclsPath != nil {
		return *m.SignedAclsPath
	}
	return ""
}

type DatalogGuardDetails struct {
	SignedRulesPath  *string `protobuf:"bytes,2,opt,name=signed_rules_path" json:"signed_rules_path,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *DatalogGuardDetails) Reset()         { *m = DatalogGuardDetails{} }
func (m *DatalogGuardDetails) String() string { return proto.CompactTextString(m) }
func (*DatalogGuardDetails) ProtoMessage()    {}

func (m *DatalogGuardDetails) GetSignedRulesPath() string {
	if m != nil && m.SignedRulesPath != nil {
		return *m.SignedRulesPath
	}
	return ""
}

type TPMDetails struct {
	TpmPath *string `protobuf:"bytes,1,opt,name=tpm_path" json:"tpm_path,omitempty"`
	AikPath *string `protobuf:"bytes,2,opt,name=aik_path" json:"aik_path,omitempty"`
	// A string representing the IDs of PCRs, like "17,18".
	Pcrs             *string `protobuf:"bytes,3,opt,name=pcrs" json:"pcrs,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *TPMDetails) Reset()         { *m = TPMDetails{} }
func (m *TPMDetails) String() string { return proto.CompactTextString(m) }
func (*TPMDetails) ProtoMessage()    {}

func (m *TPMDetails) GetTpmPath() string {
	if m != nil && m.TpmPath != nil {
		return *m.TpmPath
	}
	return ""
}

func (m *TPMDetails) GetAikPath() string {
	if m != nil && m.AikPath != nil {
		return *m.AikPath
	}
	return ""
}

func (m *TPMDetails) GetPcrs() string {
	if m != nil && m.Pcrs != nil {
		return *m.Pcrs
	}
	return ""
}

type DomainConfig struct {
	DomainInfo       *DomainDetails       `protobuf:"bytes,1,opt,name=domain_info" json:"domain_info,omitempty"`
	X509Info         *X509Details         `protobuf:"bytes,2,opt,name=x509_info" json:"x509_info,omitempty"`
	AclGuardInfo     *ACLGuardDetails     `protobuf:"bytes,3,opt,name=acl_guard_info" json:"acl_guard_info,omitempty"`
	DatalogGuardInfo *DatalogGuardDetails `protobuf:"bytes,4,opt,name=datalog_guard_info" json:"datalog_guard_info,omitempty"`
	TpmInfo          *TPMDetails          `protobuf:"bytes,5,opt,name=tpm_info" json:"tpm_info,omitempty"`
	XXX_unrecognized []byte               `json:"-"`
}

func (m *DomainConfig) Reset()         { *m = DomainConfig{} }
func (m *DomainConfig) String() string { return proto.CompactTextString(m) }
func (*DomainConfig) ProtoMessage()    {}

func (m *DomainConfig) GetDomainInfo() *DomainDetails {
	if m != nil {
		return m.DomainInfo
	}
	return nil
}

func (m *DomainConfig) GetX509Info() *X509Details {
	if m != nil {
		return m.X509Info
	}
	return nil
}

func (m *DomainConfig) GetAclGuardInfo() *ACLGuardDetails {
	if m != nil {
		return m.AclGuardInfo
	}
	return nil
}

func (m *DomainConfig) GetDatalogGuardInfo() *DatalogGuardDetails {
	if m != nil {
		return m.DatalogGuardInfo
	}
	return nil
}

func (m *DomainConfig) GetTpmInfo() *TPMDetails {
	if m != nil {
		return m.TpmInfo
	}
	return nil
}

type DomainTemplate struct {
	Config       *DomainConfig `protobuf:"bytes,1,opt,name=config" json:"config,omitempty"`
	DatalogRules []string      `protobuf:"bytes,2,rep,name=datalog_rules" json:"datalog_rules,omitempty"`
	AclRules     []string      `protobuf:"bytes,3,rep,name=acl_rules" json:"acl_rules,omitempty"`
	// The name of the host (used for policy statements)
	HostName          *string `protobuf:"bytes,4,opt,name=host_name" json:"host_name,omitempty"`
	HostPredicateName *string `protobuf:"bytes,5,opt,name=host_predicate_name" json:"host_predicate_name,omitempty"`
	// Program names (as paths to binaries)
	ProgramPaths         []string `protobuf:"bytes,6,rep,name=program_paths" json:"program_paths,omitempty"`
	ProgramPredicateName *string  `protobuf:"bytes,7,opt,name=program_predicate_name" json:"program_predicate_name,omitempty"`
	// Container names (as paths to images)
	ContainerPaths         []string `protobuf:"bytes,8,rep,name=container_paths" json:"container_paths,omitempty"`
	ContainerPredicateName *string  `protobuf:"bytes,9,opt,name=container_predicate_name" json:"container_predicate_name,omitempty"`
	// VM names (as paths to images)
	VmPaths         []string `protobuf:"bytes,10,rep,name=vm_paths" json:"vm_paths,omitempty"`
	VmPredicateName *string  `protobuf:"bytes,11,opt,name=vm_predicate_name" json:"vm_predicate_name,omitempty"`
	// LinuxHost names (as paths to images)
	LinuxHostPaths         []string `protobuf:"bytes,12,rep,name=linux_host_paths" json:"linux_host_paths,omitempty"`
	LinuxHostPredicateName *string  `protobuf:"bytes,13,opt,name=linux_host_predicate_name" json:"linux_host_predicate_name,omitempty"`
	// The name of the predicate to use for trusted guards.
	GuardPredicateName *string `protobuf:"bytes,14,opt,name=guard_predicate_name" json:"guard_predicate_name,omitempty"`
	// The name of the predicate to use for trusted TPMs.
	TpmPredicateName *string `protobuf:"bytes,15,opt,name=tpm_predicate_name" json:"tpm_predicate_name,omitempty"`
	// The name of the predicate to use for trusted OSs.
	OsPredicateName  *string `protobuf:"bytes,16,opt,name=os_predicate_name" json:"os_predicate_name,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *DomainTemplate) Reset()         { *m = DomainTemplate{} }
func (m *DomainTemplate) String() string { return proto.CompactTextString(m) }
func (*DomainTemplate) ProtoMessage()    {}

func (m *DomainTemplate) GetConfig() *DomainConfig {
	if m != nil {
		return m.Config
	}
	return nil
}

func (m *DomainTemplate) GetDatalogRules() []string {
	if m != nil {
		return m.DatalogRules
	}
	return nil
}

func (m *DomainTemplate) GetAclRules() []string {
	if m != nil {
		return m.AclRules
	}
	return nil
}

func (m *DomainTemplate) GetHostName() string {
	if m != nil && m.HostName != nil {
		return *m.HostName
	}
	return ""
}

func (m *DomainTemplate) GetHostPredicateName() string {
	if m != nil && m.HostPredicateName != nil {
		return *m.HostPredicateName
	}
	return ""
}

func (m *DomainTemplate) GetProgramPaths() []string {
	if m != nil {
		return m.ProgramPaths
	}
	return nil
}

func (m *DomainTemplate) GetProgramPredicateName() string {
	if m != nil && m.ProgramPredicateName != nil {
		return *m.ProgramPredicateName
	}
	return ""
}

func (m *DomainTemplate) GetContainerPaths() []string {
	if m != nil {
		return m.ContainerPaths
	}
	return nil
}

func (m *DomainTemplate) GetContainerPredicateName() string {
	if m != nil && m.ContainerPredicateName != nil {
		return *m.ContainerPredicateName
	}
	return ""
}

func (m *DomainTemplate) GetVmPaths() []string {
	if m != nil {
		return m.VmPaths
	}
	return nil
}

func (m *DomainTemplate) GetVmPredicateName() string {
	if m != nil && m.VmPredicateName != nil {
		return *m.VmPredicateName
	}
	return ""
}

func (m *DomainTemplate) GetLinuxHostPaths() []string {
	if m != nil {
		return m.LinuxHostPaths
	}
	return nil
}

func (m *DomainTemplate) GetLinuxHostPredicateName() string {
	if m != nil && m.LinuxHostPredicateName != nil {
		return *m.LinuxHostPredicateName
	}
	return ""
}

func (m *DomainTemplate) GetGuardPredicateName() string {
	if m != nil && m.GuardPredicateName != nil {
		return *m.GuardPredicateName
	}
	return ""
}

func (m *DomainTemplate) GetTpmPredicateName() string {
	if m != nil && m.TpmPredicateName != nil {
		return *m.TpmPredicateName
	}
	return ""
}

func (m *DomainTemplate) GetOsPredicateName() string {
	if m != nil && m.OsPredicateName != nil {
		return *m.OsPredicateName
	}
	return ""
}

func init() {
}
