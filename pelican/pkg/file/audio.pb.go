// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        v5.28.3
// source: pelican/pkg/file/audio.proto

package file

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// AudioPacket represents a single chunk of audio data extracted from RTSP
type AudioPacket struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Raw audio data bytes
	AudioData []byte `protobuf:"bytes,1,opt,name=audio_data,json=audioData,proto3" json:"audio_data,omitempty"`
	// Timestamp of the audio packet (unix timestamp in milliseconds)
	Timestamp int64 `protobuf:"varint,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	// Source identifier for the RTSP stream
	StreamId string `protobuf:"bytes,3,opt,name=stream_id,json=streamId,proto3" json:"stream_id,omitempty"`
	// Audio format information
	Format *AudioFormat `protobuf:"bytes,4,opt,name=format,proto3" json:"format,omitempty"`
}

func (x *AudioPacket) Reset() {
	*x = AudioPacket{}
	mi := &file_pelican_pkg_file_audio_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AudioPacket) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AudioPacket) ProtoMessage() {}

func (x *AudioPacket) ProtoReflect() protoreflect.Message {
	mi := &file_pelican_pkg_file_audio_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AudioPacket.ProtoReflect.Descriptor instead.
func (*AudioPacket) Descriptor() ([]byte, []int) {
	return file_pelican_pkg_file_audio_proto_rawDescGZIP(), []int{0}
}

func (x *AudioPacket) GetAudioData() []byte {
	if x != nil {
		return x.AudioData
	}
	return nil
}

func (x *AudioPacket) GetTimestamp() int64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

func (x *AudioPacket) GetStreamId() string {
	if x != nil {
		return x.StreamId
	}
	return ""
}

func (x *AudioPacket) GetFormat() *AudioFormat {
	if x != nil {
		return x.Format
	}
	return nil
}

// AudioFormat contains the necessary metadata for processing the audio
type AudioFormat struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Sample rate in Hz (e.g., 16000, 44100)
	SampleRate int32 `protobuf:"varint,1,opt,name=sample_rate,json=sampleRate,proto3" json:"sample_rate,omitempty"`
	// Number of audio channels (1 for mono, 2 for stereo)
	Channels int32 `protobuf:"varint,2,opt,name=channels,proto3" json:"channels,omitempty"`
	// Audio encoding format (e.g., "pcm", "aac")
	Encoding string `protobuf:"bytes,3,opt,name=encoding,proto3" json:"encoding,omitempty"`
}

func (x *AudioFormat) Reset() {
	*x = AudioFormat{}
	mi := &file_pelican_pkg_file_audio_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AudioFormat) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AudioFormat) ProtoMessage() {}

func (x *AudioFormat) ProtoReflect() protoreflect.Message {
	mi := &file_pelican_pkg_file_audio_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AudioFormat.ProtoReflect.Descriptor instead.
func (*AudioFormat) Descriptor() ([]byte, []int) {
	return file_pelican_pkg_file_audio_proto_rawDescGZIP(), []int{1}
}

func (x *AudioFormat) GetSampleRate() int32 {
	if x != nil {
		return x.SampleRate
	}
	return 0
}

func (x *AudioFormat) GetChannels() int32 {
	if x != nil {
		return x.Channels
	}
	return 0
}

func (x *AudioFormat) GetEncoding() string {
	if x != nil {
		return x.Encoding
	}
	return ""
}

// Response from the downstream service
type StreamResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Status of the stream processing
	Success bool `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	// Error message if processing failed
	Error string `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *StreamResponse) Reset() {
	*x = StreamResponse{}
	mi := &file_pelican_pkg_file_audio_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StreamResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StreamResponse) ProtoMessage() {}

func (x *StreamResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pelican_pkg_file_audio_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StreamResponse.ProtoReflect.Descriptor instead.
func (*StreamResponse) Descriptor() ([]byte, []int) {
	return file_pelican_pkg_file_audio_proto_rawDescGZIP(), []int{2}
}

func (x *StreamResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *StreamResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

var File_pelican_pkg_file_audio_proto protoreflect.FileDescriptor

var file_pelican_pkg_file_audio_proto_rawDesc = []byte{
	0x0a, 0x1c, 0x70, 0x65, 0x6c, 0x69, 0x63, 0x61, 0x6e, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x66, 0x69,
	0x6c, 0x65, 0x2f, 0x61, 0x75, 0x64, 0x69, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x15,
	0x6d, 0x75, 0x73, 0x65, 0x2e, 0x70, 0x65, 0x6c, 0x69, 0x63, 0x61, 0x6e, 0x2e, 0x70, 0x6b, 0x67,
	0x2e, 0x66, 0x69, 0x6c, 0x65, 0x22, 0xa3, 0x01, 0x0a, 0x0b, 0x41, 0x75, 0x64, 0x69, 0x6f, 0x50,
	0x61, 0x63, 0x6b, 0x65, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x61, 0x75, 0x64, 0x69, 0x6f, 0x5f, 0x64,
	0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x61, 0x75, 0x64, 0x69, 0x6f,
	0x44, 0x61, 0x74, 0x61, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d,
	0x70, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61,
	0x6d, 0x70, 0x12, 0x1b, 0x0a, 0x09, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x5f, 0x69, 0x64, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x49, 0x64, 0x12,
	0x3a, 0x0a, 0x06, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x22, 0x2e, 0x6d, 0x75, 0x73, 0x65, 0x2e, 0x70, 0x65, 0x6c, 0x69, 0x63, 0x61, 0x6e, 0x2e, 0x70,
	0x6b, 0x67, 0x2e, 0x66, 0x69, 0x6c, 0x65, 0x2e, 0x41, 0x75, 0x64, 0x69, 0x6f, 0x46, 0x6f, 0x72,
	0x6d, 0x61, 0x74, 0x52, 0x06, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x22, 0x66, 0x0a, 0x0b, 0x41,
	0x75, 0x64, 0x69, 0x6f, 0x46, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x12, 0x1f, 0x0a, 0x0b, 0x73, 0x61,
	0x6d, 0x70, 0x6c, 0x65, 0x5f, 0x72, 0x61, 0x74, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x0a, 0x73, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x61, 0x74, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x63,
	0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x63,
	0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x65, 0x6e, 0x63, 0x6f, 0x64,
	0x69, 0x6e, 0x67, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x65, 0x6e, 0x63, 0x6f, 0x64,
	0x69, 0x6e, 0x67, 0x22, 0x40, 0x0a, 0x0e, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x12,
	0x14, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x65, 0x72, 0x72, 0x6f, 0x72, 0x32, 0x72, 0x0a, 0x12, 0x41, 0x75, 0x64, 0x69, 0x6f, 0x53, 0x74,
	0x72, 0x65, 0x61, 0x6d, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x5c, 0x0a, 0x0b, 0x53,
	0x74, 0x72, 0x65, 0x61, 0x6d, 0x41, 0x75, 0x64, 0x69, 0x6f, 0x12, 0x22, 0x2e, 0x6d, 0x75, 0x73,
	0x65, 0x2e, 0x70, 0x65, 0x6c, 0x69, 0x63, 0x61, 0x6e, 0x2e, 0x70, 0x6b, 0x67, 0x2e, 0x66, 0x69,
	0x6c, 0x65, 0x2e, 0x41, 0x75, 0x64, 0x69, 0x6f, 0x50, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x1a, 0x25,
	0x2e, 0x6d, 0x75, 0x73, 0x65, 0x2e, 0x70, 0x65, 0x6c, 0x69, 0x63, 0x61, 0x6e, 0x2e, 0x70, 0x6b,
	0x67, 0x2e, 0x66, 0x69, 0x6c, 0x65, 0x2e, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x28, 0x01, 0x42, 0x17, 0x5a, 0x15, 0x6d, 0x75, 0x73,
	0x65, 0x2f, 0x70, 0x65, 0x6c, 0x69, 0x63, 0x61, 0x6e, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x66, 0x69,
	0x6c, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pelican_pkg_file_audio_proto_rawDescOnce sync.Once
	file_pelican_pkg_file_audio_proto_rawDescData = file_pelican_pkg_file_audio_proto_rawDesc
)

func file_pelican_pkg_file_audio_proto_rawDescGZIP() []byte {
	file_pelican_pkg_file_audio_proto_rawDescOnce.Do(func() {
		file_pelican_pkg_file_audio_proto_rawDescData = protoimpl.X.CompressGZIP(file_pelican_pkg_file_audio_proto_rawDescData)
	})
	return file_pelican_pkg_file_audio_proto_rawDescData
}

var file_pelican_pkg_file_audio_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_pelican_pkg_file_audio_proto_goTypes = []any{
	(*AudioPacket)(nil),    // 0: muse.pelican.pkg.file.AudioPacket
	(*AudioFormat)(nil),    // 1: muse.pelican.pkg.file.AudioFormat
	(*StreamResponse)(nil), // 2: muse.pelican.pkg.file.StreamResponse
}
var file_pelican_pkg_file_audio_proto_depIdxs = []int32{
	1, // 0: muse.pelican.pkg.file.AudioPacket.format:type_name -> muse.pelican.pkg.file.AudioFormat
	0, // 1: muse.pelican.pkg.file.AudioStreamService.StreamAudio:input_type -> muse.pelican.pkg.file.AudioPacket
	2, // 2: muse.pelican.pkg.file.AudioStreamService.StreamAudio:output_type -> muse.pelican.pkg.file.StreamResponse
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_pelican_pkg_file_audio_proto_init() }
func file_pelican_pkg_file_audio_proto_init() {
	if File_pelican_pkg_file_audio_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_pelican_pkg_file_audio_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_pelican_pkg_file_audio_proto_goTypes,
		DependencyIndexes: file_pelican_pkg_file_audio_proto_depIdxs,
		MessageInfos:      file_pelican_pkg_file_audio_proto_msgTypes,
	}.Build()
	File_pelican_pkg_file_audio_proto = out.File
	file_pelican_pkg_file_audio_proto_rawDesc = nil
	file_pelican_pkg_file_audio_proto_goTypes = nil
	file_pelican_pkg_file_audio_proto_depIdxs = nil
}
