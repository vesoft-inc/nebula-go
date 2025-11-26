// Copyright 2025 vesoft inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// 
//     http://www.apache.org/licenses/LICENSE-2.0
// 
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"

	"github.com/vesoft-inc/nebula-go/v5/internal/generated_code/v5.0.0/proto/common"
)

var PROTOCOL_VERSION []byte
var META_PROTOCOL_VERSION []byte

func init() {
	common.File_common_proto.Options().ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		switch fd.FullName() {
		case protoreflect.FullName(common.E_ProtocolVersion.Name):
			f := common.File_common_proto.Options().ProtoReflect().Get(fd)
			i := f.Interface()
			if _, ok := i.([]byte); ok {
				PROTOCOL_VERSION = i.([]byte)
			}
		case protoreflect.FullName(common.E_MetaProtocolVersion.Name):
			f := common.File_common_proto.Options().ProtoReflect().Get(fd)
			i := f.Interface()
			if _, ok := i.([]byte); ok {
				META_PROTOCOL_VERSION = i.([]byte)
			}
		}
		return true
	})
}
