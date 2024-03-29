// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.0
// source: ldap.proto

package api

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	structpb "google.golang.org/protobuf/types/known/structpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// https://datatracker.ietf.org/doc/html/rfc4511#section-4.1.11
type LDAPControl struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ControlType  string `protobuf:"bytes,1,opt,name=controlType,proto3" json:"controlType,omitempty"`   // LDAPOID,
	Criticality  bool   `protobuf:"varint,2,opt,name=criticality,proto3" json:"criticality,omitempty"`  // BOOLEAN DEFAULT FALSE,
	ControlValue string `protobuf:"bytes,3,opt,name=controlValue,proto3" json:"controlValue,omitempty"` // OCTET STRING OPTIONAL
}

func (x *LDAPControl) Reset() {
	*x = LDAPControl{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ldap_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LDAPControl) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LDAPControl) ProtoMessage() {}

func (x *LDAPControl) ProtoReflect() protoreflect.Message {
	mi := &file_ldap_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LDAPControl.ProtoReflect.Descriptor instead.
func (*LDAPControl) Descriptor() ([]byte, []int) {
	return file_ldap_proto_rawDescGZIP(), []int{0}
}

func (x *LDAPControl) GetControlType() string {
	if x != nil {
		return x.ControlType
	}
	return ""
}

func (x *LDAPControl) GetCriticality() bool {
	if x != nil {
		return x.Criticality
	}
	return false
}

func (x *LDAPControl) GetControlValue() string {
	if x != nil {
		return x.ControlValue
	}
	return ""
}

// https://datatracker.ietf.org/doc/html/rfc4511#section-4.5.1
type LDAPSearchRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// ----- connection -----
	// Optional. ID of the preconfigured LDAP catalog
	CatalogId int64 `protobuf:"varint,1,opt,name=catalog_id,json=catalogId,proto3" json:"catalog_id,omitempty"`
	// Optional. URL to establish connection to LDAP catalog
	Url string `protobuf:"bytes,5,opt,name=url,proto3" json:"url,omitempty"` // URL e.g.: [(ldap|ldapi|ldaps)://]host[:port]
	// // TLS configuration options
	// message TLSConfig {
	//     // TODO: (!)
	//     bytes cert = 1; // PEM: base64
	//     bytes key = 2; // PEM: base64
	//     bytes ca = 3; // PEM: base64
	// }
	// TLSConfig tls = 6;
	// ----- BIND: Authorization -----
	Bind     string `protobuf:"bytes,7,opt,name=bind,proto3" json:"bind,omitempty"`         // authorization method e.g.: SIMPLE, SAML, NTLM, etc.
	Username string `protobuf:"bytes,8,opt,name=username,proto3" json:"username,omitempty"` // bind_dn
	Password string `protobuf:"bytes,9,opt,name=password,proto3" json:"password,omitempty"` // password
	// ----- SearchRequest -----
	// baseObject [D]istinguished[N]ame
	BaseObject string `protobuf:"bytes,10,opt,name=baseObject,proto3" json:"baseObject,omitempty"`
	// baseObject              (0),
	// singleLevel             (1),
	// wholeSubtree            (2)
	Scope int32 `protobuf:"varint,11,opt,name=scope,proto3" json:"scope,omitempty"`
	// neverDerefAliases       (0),
	// derefInSearching        (1),
	// derefFindingBaseObj     (2),
	// derefAlways             (3)
	DerefAliases int32    `protobuf:"varint,12,opt,name=derefAliases,proto3" json:"derefAliases,omitempty"`
	SizeLimit    int64    `protobuf:"varint,13,opt,name=sizeLimit,proto3" json:"sizeLimit,omitempty"`  // INTEGER (0 ..  maxInt),
	TimeLimit    int64    `protobuf:"varint,14,opt,name=timeLimit,proto3" json:"timeLimit,omitempty"`  // INTEGER (0 ..  maxInt),
	TypesOnly    bool     `protobuf:"varint,15,opt,name=typesOnly,proto3" json:"typesOnly,omitempty"`  // BOOLEAN,
	Filter       string   `protobuf:"bytes,16,opt,name=filter,proto3" json:"filter,omitempty"`         // Filter,
	Attributes   []string `protobuf:"bytes,17,rep,name=attributes,proto3" json:"attributes,omitempty"` // AttributeSelection
	// ----- LDAPMessage -----
	// Controls NOT implemented yet !
	Controls []*LDAPControl `protobuf:"bytes,18,rep,name=controls,proto3" json:"controls,omitempty"`
}

func (x *LDAPSearchRequest) Reset() {
	*x = LDAPSearchRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ldap_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LDAPSearchRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LDAPSearchRequest) ProtoMessage() {}

func (x *LDAPSearchRequest) ProtoReflect() protoreflect.Message {
	mi := &file_ldap_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LDAPSearchRequest.ProtoReflect.Descriptor instead.
func (*LDAPSearchRequest) Descriptor() ([]byte, []int) {
	return file_ldap_proto_rawDescGZIP(), []int{1}
}

func (x *LDAPSearchRequest) GetCatalogId() int64 {
	if x != nil {
		return x.CatalogId
	}
	return 0
}

func (x *LDAPSearchRequest) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *LDAPSearchRequest) GetBind() string {
	if x != nil {
		return x.Bind
	}
	return ""
}

func (x *LDAPSearchRequest) GetUsername() string {
	if x != nil {
		return x.Username
	}
	return ""
}

func (x *LDAPSearchRequest) GetPassword() string {
	if x != nil {
		return x.Password
	}
	return ""
}

func (x *LDAPSearchRequest) GetBaseObject() string {
	if x != nil {
		return x.BaseObject
	}
	return ""
}

func (x *LDAPSearchRequest) GetScope() int32 {
	if x != nil {
		return x.Scope
	}
	return 0
}

func (x *LDAPSearchRequest) GetDerefAliases() int32 {
	if x != nil {
		return x.DerefAliases
	}
	return 0
}

func (x *LDAPSearchRequest) GetSizeLimit() int64 {
	if x != nil {
		return x.SizeLimit
	}
	return 0
}

func (x *LDAPSearchRequest) GetTimeLimit() int64 {
	if x != nil {
		return x.TimeLimit
	}
	return 0
}

func (x *LDAPSearchRequest) GetTypesOnly() bool {
	if x != nil {
		return x.TypesOnly
	}
	return false
}

func (x *LDAPSearchRequest) GetFilter() string {
	if x != nil {
		return x.Filter
	}
	return ""
}

func (x *LDAPSearchRequest) GetAttributes() []string {
	if x != nil {
		return x.Attributes
	}
	return nil
}

func (x *LDAPSearchRequest) GetControls() []*LDAPControl {
	if x != nil {
		return x.Controls
	}
	return nil
}

// https://datatracker.ietf.org/doc/html/rfc4511#section-4.5.2
type LDAPSearchResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// ----- SearchResult (Entry|Reference) -----
	// repeated LDAPSearchEntry entries = 1;
	Entries []*structpb.Struct `protobuf:"bytes,1,rep,name=entries,proto3" json:"entries,omitempty"`
	// ----- LDAPResult -----
	ResultCode        int32    `protobuf:"varint,2,opt,name=resultCode,proto3" json:"resultCode,omitempty"`
	MatchedDN         string   `protobuf:"bytes,3,opt,name=matchedDN,proto3" json:"matchedDN,omitempty"`                 // LDAPDN,
	DiagnosticMessage string   `protobuf:"bytes,4,opt,name=diagnosticMessage,proto3" json:"diagnosticMessage,omitempty"` // LDAPString,
	Referral          []string `protobuf:"bytes,5,rep,name=referral,proto3" json:"referral,omitempty"`                   // [3] Referral OPTIONAL
	// ----- LDAPMessage -----
	Controls []*LDAPControl `protobuf:"bytes,6,rep,name=controls,proto3" json:"controls,omitempty"`
}

func (x *LDAPSearchResponse) Reset() {
	*x = LDAPSearchResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ldap_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LDAPSearchResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LDAPSearchResponse) ProtoMessage() {}

func (x *LDAPSearchResponse) ProtoReflect() protoreflect.Message {
	mi := &file_ldap_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LDAPSearchResponse.ProtoReflect.Descriptor instead.
func (*LDAPSearchResponse) Descriptor() ([]byte, []int) {
	return file_ldap_proto_rawDescGZIP(), []int{2}
}

func (x *LDAPSearchResponse) GetEntries() []*structpb.Struct {
	if x != nil {
		return x.Entries
	}
	return nil
}

func (x *LDAPSearchResponse) GetResultCode() int32 {
	if x != nil {
		return x.ResultCode
	}
	return 0
}

func (x *LDAPSearchResponse) GetMatchedDN() string {
	if x != nil {
		return x.MatchedDN
	}
	return ""
}

func (x *LDAPSearchResponse) GetDiagnosticMessage() string {
	if x != nil {
		return x.DiagnosticMessage
	}
	return ""
}

func (x *LDAPSearchResponse) GetReferral() []string {
	if x != nil {
		return x.Referral
	}
	return nil
}

func (x *LDAPSearchResponse) GetControls() []*LDAPControl {
	if x != nil {
		return x.Controls
	}
	return nil
}

var File_ldap_proto protoreflect.FileDescriptor

var file_ldap_proto_rawDesc = []byte{
	0x0a, 0x0a, 0x6c, 0x64, 0x61, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x03, 0x61, 0x70,
	0x69, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2f, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0x75, 0x0a, 0x0b, 0x4c, 0x44, 0x41, 0x50, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x12, 0x20,
	0x0a, 0x0b, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x54, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0b, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x54, 0x79, 0x70, 0x65,
	0x12, 0x20, 0x0a, 0x0b, 0x63, 0x72, 0x69, 0x74, 0x69, 0x63, 0x61, 0x6c, 0x69, 0x74, 0x79, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0b, 0x63, 0x72, 0x69, 0x74, 0x69, 0x63, 0x61, 0x6c, 0x69,
	0x74, 0x79, 0x12, 0x22, 0x0a, 0x0c, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x56, 0x61, 0x6c,
	0x75, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f,
	0x6c, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x22, 0xaa, 0x03, 0x0a, 0x11, 0x4c, 0x44, 0x41, 0x50, 0x53,
	0x65, 0x61, 0x72, 0x63, 0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1d, 0x0a, 0x0a,
	0x63, 0x61, 0x74, 0x61, 0x6c, 0x6f, 0x67, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x09, 0x63, 0x61, 0x74, 0x61, 0x6c, 0x6f, 0x67, 0x49, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x75,
	0x72, 0x6c, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x75, 0x72, 0x6c, 0x12, 0x12, 0x0a,
	0x04, 0x62, 0x69, 0x6e, 0x64, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x62, 0x69, 0x6e,
	0x64, 0x12, 0x1a, 0x0a, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x08, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1a, 0x0a,
	0x08, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x08, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x12, 0x1e, 0x0a, 0x0a, 0x62, 0x61, 0x73,
	0x65, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x62,
	0x61, 0x73, 0x65, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x73, 0x63, 0x6f,
	0x70, 0x65, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x73, 0x63, 0x6f, 0x70, 0x65, 0x12,
	0x22, 0x0a, 0x0c, 0x64, 0x65, 0x72, 0x65, 0x66, 0x41, 0x6c, 0x69, 0x61, 0x73, 0x65, 0x73, 0x18,
	0x0c, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0c, 0x64, 0x65, 0x72, 0x65, 0x66, 0x41, 0x6c, 0x69, 0x61,
	0x73, 0x65, 0x73, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x69, 0x7a, 0x65, 0x4c, 0x69, 0x6d, 0x69, 0x74,
	0x18, 0x0d, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x73, 0x69, 0x7a, 0x65, 0x4c, 0x69, 0x6d, 0x69,
	0x74, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x4c, 0x69, 0x6d, 0x69, 0x74, 0x18, 0x0e,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x4c, 0x69, 0x6d, 0x69, 0x74, 0x12,
	0x1c, 0x0a, 0x09, 0x74, 0x79, 0x70, 0x65, 0x73, 0x4f, 0x6e, 0x6c, 0x79, 0x18, 0x0f, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x09, 0x74, 0x79, 0x70, 0x65, 0x73, 0x4f, 0x6e, 0x6c, 0x79, 0x12, 0x16, 0x0a,
	0x06, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x18, 0x10, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x66,
	0x69, 0x6c, 0x74, 0x65, 0x72, 0x12, 0x1e, 0x0a, 0x0a, 0x61, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75,
	0x74, 0x65, 0x73, 0x18, 0x11, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0a, 0x61, 0x74, 0x74, 0x72, 0x69,
	0x62, 0x75, 0x74, 0x65, 0x73, 0x12, 0x2c, 0x0a, 0x08, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c,
	0x73, 0x18, 0x12, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x4c, 0x44,
	0x41, 0x50, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x52, 0x08, 0x63, 0x6f, 0x6e, 0x74, 0x72,
	0x6f, 0x6c, 0x73, 0x22, 0xfd, 0x01, 0x0a, 0x12, 0x4c, 0x44, 0x41, 0x50, 0x53, 0x65, 0x61, 0x72,
	0x63, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x31, 0x0a, 0x07, 0x65, 0x6e,
	0x74, 0x72, 0x69, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x74,
	0x72, 0x75, 0x63, 0x74, 0x52, 0x07, 0x65, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x12, 0x1e, 0x0a,
	0x0a, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x43, 0x6f, 0x64, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x0a, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x1c, 0x0a,
	0x09, 0x6d, 0x61, 0x74, 0x63, 0x68, 0x65, 0x64, 0x44, 0x4e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x09, 0x6d, 0x61, 0x74, 0x63, 0x68, 0x65, 0x64, 0x44, 0x4e, 0x12, 0x2c, 0x0a, 0x11, 0x64,
	0x69, 0x61, 0x67, 0x6e, 0x6f, 0x73, 0x74, 0x69, 0x63, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x11, 0x64, 0x69, 0x61, 0x67, 0x6e, 0x6f, 0x73, 0x74,
	0x69, 0x63, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x72, 0x65, 0x66,
	0x65, 0x72, 0x72, 0x61, 0x6c, 0x18, 0x05, 0x20, 0x03, 0x28, 0x09, 0x52, 0x08, 0x72, 0x65, 0x66,
	0x65, 0x72, 0x72, 0x61, 0x6c, 0x12, 0x2c, 0x0a, 0x08, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c,
	0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x4c, 0x44,
	0x41, 0x50, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x52, 0x08, 0x63, 0x6f, 0x6e, 0x74, 0x72,
	0x6f, 0x6c, 0x73, 0x32, 0x47, 0x0a, 0x04, 0x4c, 0x44, 0x41, 0x50, 0x12, 0x3f, 0x0a, 0x0a, 0x4c,
	0x44, 0x41, 0x50, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x12, 0x16, 0x2e, 0x61, 0x70, 0x69, 0x2e,
	0x4c, 0x44, 0x41, 0x50, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x17, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x4c, 0x44, 0x41, 0x50, 0x53, 0x65, 0x61, 0x72,
	0x63, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x14, 0x5a, 0x12,
	0x77, 0x65, 0x62, 0x69, 0x74, 0x65, 0x6c, 0x2e, 0x67, 0x6f, 0x2f, 0x61, 0x70, 0x69, 0x3b, 0x61,
	0x70, 0x69, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_ldap_proto_rawDescOnce sync.Once
	file_ldap_proto_rawDescData = file_ldap_proto_rawDesc
)

func file_ldap_proto_rawDescGZIP() []byte {
	file_ldap_proto_rawDescOnce.Do(func() {
		file_ldap_proto_rawDescData = protoimpl.X.CompressGZIP(file_ldap_proto_rawDescData)
	})
	return file_ldap_proto_rawDescData
}

var file_ldap_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_ldap_proto_goTypes = []interface{}{
	(*LDAPControl)(nil),        // 0: api.LDAPControl
	(*LDAPSearchRequest)(nil),  // 1: api.LDAPSearchRequest
	(*LDAPSearchResponse)(nil), // 2: api.LDAPSearchResponse
	(*structpb.Struct)(nil),    // 3: google.protobuf.Struct
}
var file_ldap_proto_depIdxs = []int32{
	0, // 0: api.LDAPSearchRequest.controls:type_name -> api.LDAPControl
	3, // 1: api.LDAPSearchResponse.entries:type_name -> google.protobuf.Struct
	0, // 2: api.LDAPSearchResponse.controls:type_name -> api.LDAPControl
	1, // 3: api.LDAP.LDAPSearch:input_type -> api.LDAPSearchRequest
	2, // 4: api.LDAP.LDAPSearch:output_type -> api.LDAPSearchResponse
	4, // [4:5] is the sub-list for method output_type
	3, // [3:4] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_ldap_proto_init() }
func file_ldap_proto_init() {
	if File_ldap_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_ldap_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LDAPControl); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ldap_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LDAPSearchRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ldap_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LDAPSearchResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_ldap_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_ldap_proto_goTypes,
		DependencyIndexes: file_ldap_proto_depIdxs,
		MessageInfos:      file_ldap_proto_msgTypes,
	}.Build()
	File_ldap_proto = out.File
	file_ldap_proto_rawDesc = nil
	file_ldap_proto_goTypes = nil
	file_ldap_proto_depIdxs = nil
}
