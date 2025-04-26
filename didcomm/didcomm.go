// nolint:revive
package didcomm

/*
#cgo LDFLAGS: -L./lib -ldidcomm_uniffi

#include "didcomm.h"
*/
import "C"

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"
)

type RustBuffer = C.RustBuffer

type RustBufferI interface {
	AsReader() *bytes.Reader
	Free()
	ToGoBytes() []byte
	Data() unsafe.Pointer
	Len() int
	Capacity() int
}

func RustBufferFromExternal(b RustBufferI) RustBuffer {
	return RustBuffer{
		capacity: C.int(b.Capacity()),
		len:      C.int(b.Len()),
		data:     (*C.uchar)(b.Data()),
	}
}

func (cb RustBuffer) Capacity() int {
	return int(cb.capacity)
}

func (cb RustBuffer) Len() int {
	return int(cb.len)
}

func (cb RustBuffer) Data() unsafe.Pointer {
	return unsafe.Pointer(cb.data)
}

func (cb RustBuffer) AsReader() *bytes.Reader {
	b := unsafe.Slice((*byte)(cb.data), C.int(cb.len))
	return bytes.NewReader(b)
}

func (cb RustBuffer) Free() {
	rustCall(func(status *C.RustCallStatus) bool {
		C.ffi_didcomm_uniffi_rustbuffer_free(cb, status)
		return false
	})
}

func (cb RustBuffer) ToGoBytes() []byte {
	return C.GoBytes(unsafe.Pointer(cb.data), C.int(cb.len))
}

func stringToRustBuffer(str string) RustBuffer {
	return bytesToRustBuffer([]byte(str))
}

func bytesToRustBuffer(b []byte) RustBuffer {
	if len(b) == 0 {
		return RustBuffer{}
	}
	// We can pass the pointer along here, as it is pinned
	// for the duration of this call
	foreign := C.ForeignBytes{
		len:  C.int(len(b)),
		data: (*C.uchar)(unsafe.Pointer(&b[0])),
	}

	return rustCall(func(status *C.RustCallStatus) RustBuffer {
		return C.ffi_didcomm_uniffi_rustbuffer_from_bytes(foreign, status)
	})
}

type BufLifter[GoType any] interface {
	Lift(value RustBufferI) GoType
}

type BufLowerer[GoType any] interface {
	Lower(value GoType) RustBuffer
}

type FfiConverter[GoType any, FfiType any] interface {
	Lift(value FfiType) GoType
	Lower(value GoType) FfiType
}

type BufReader[GoType any] interface {
	Read(reader io.Reader) GoType
}

type BufWriter[GoType any] interface {
	Write(writer io.Writer, value GoType)
}

type FfiRustBufConverter[GoType any, FfiType any] interface {
	FfiConverter[GoType, FfiType]
	BufReader[GoType]
}

func LowerIntoRustBuffer[GoType any](bufWriter BufWriter[GoType], value GoType) RustBuffer {
	// This might be not the most efficient way but it does not require knowing allocation size
	// beforehand
	var buffer bytes.Buffer
	bufWriter.Write(&buffer, value)

	bytes, err := io.ReadAll(&buffer)
	if err != nil {
		panic(fmt.Errorf("reading written data: %w", err))
	}
	return bytesToRustBuffer(bytes)
}

func LiftFromRustBuffer[GoType any](bufReader BufReader[GoType], rbuf RustBufferI) GoType {
	defer rbuf.Free()
	reader := rbuf.AsReader()
	item := bufReader.Read(reader)
	if reader.Len() > 0 {
		// TODO: Remove this
		leftover, _ := io.ReadAll(reader)
		panic(fmt.Errorf("Junk remaining in buffer after lifting: %s", string(leftover)))
	}
	return item
}

func rustCallWithError[U any](converter BufLifter[error], callback func(*C.RustCallStatus) U) (U, error) {
	var status C.RustCallStatus
	returnValue := callback(&status)
	err := checkCallStatus(converter, status)

	return returnValue, err
}

func checkCallStatus(converter BufLifter[error], status C.RustCallStatus) error {
	switch status.code {
	case 0:
		return nil
	case 1:
		return converter.Lift(status.errorBuf)
	case 2:
		// when the rust code sees a panic, it tries to construct a rustbuffer
		// with the message.  but if that code panics, then it just sends back
		// an empty buffer.
		if status.errorBuf.len > 0 {
			panic(fmt.Errorf("%s", FfiConverterStringINSTANCE.Lift(status.errorBuf)))
		} else {
			panic(fmt.Errorf("Rust panicked while handling Rust panic"))
		}
	default:
		return fmt.Errorf("unknown status code: %d", status.code)
	}
}

func checkCallStatusUnknown(status C.RustCallStatus) error {
	switch status.code {
	case 0:
		return nil
	case 1:
		panic(fmt.Errorf("function not returning an error returned an error"))
	case 2:
		// when the rust code sees a panic, it tries to construct a rustbuffer
		// with the message.  but if that code panics, then it just sends back
		// an empty buffer.
		if status.errorBuf.len > 0 {
			panic(fmt.Errorf("%s", FfiConverterStringINSTANCE.Lift(status.errorBuf)))
		} else {
			panic(fmt.Errorf("Rust panicked while handling Rust panic"))
		}
	default:
		return fmt.Errorf("unknown status code: %d", status.code)
	}
}

func rustCall[U any](callback func(*C.RustCallStatus) U) U {
	returnValue, err := rustCallWithError(nil, callback)
	if err != nil {
		panic(err)
	}
	return returnValue
}

func writeInt8(writer io.Writer, value int8) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint8(writer io.Writer, value uint8) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeInt16(writer io.Writer, value int16) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint16(writer io.Writer, value uint16) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeInt32(writer io.Writer, value int32) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint32(writer io.Writer, value uint32) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeInt64(writer io.Writer, value int64) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint64(writer io.Writer, value uint64) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeFloat32(writer io.Writer, value float32) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeFloat64(writer io.Writer, value float64) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func readInt8(reader io.Reader) int8 {
	var result int8
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint8(reader io.Reader) uint8 {
	var result uint8
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readInt16(reader io.Reader) int16 {
	var result int16
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint16(reader io.Reader) uint16 {
	var result uint16
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readInt32(reader io.Reader) int32 {
	var result int32
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint32(reader io.Reader) uint32 {
	var result uint32
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readInt64(reader io.Reader) int64 {
	var result int64
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint64(reader io.Reader) uint64 {
	var result uint64
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readFloat32(reader io.Reader) float32 {
	var result float32
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readFloat64(reader io.Reader) float64 {
	var result float64
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func init() {

	(&FfiConverterCallbackInterfaceDIDResolver{}).register()
	(&FfiConverterCallbackInterfaceOnFromPriorPackResult{}).register()
	(&FfiConverterCallbackInterfaceOnFromPriorUnpackResult{}).register()
	(&FfiConverterCallbackInterfaceOnPackEncryptedResult{}).register()
	(&FfiConverterCallbackInterfaceOnPackPlaintextResult{}).register()
	(&FfiConverterCallbackInterfaceOnPackSignedResult{}).register()
	(&FfiConverterCallbackInterfaceOnUnpackResult{}).register()
	(&FfiConverterCallbackInterfaceOnWrapInForwardResult{}).register()
	(&FfiConverterCallbackInterfaceSecretsResolver{}).register()
	uniffiCheckChecksums()
}

func uniffiCheckChecksums() {
	// Get the bindings contract version from our ComponentInterface
	bindingsContractVersion := 24
	// Get the scaffolding contract version by calling the into the dylib
	scaffoldingContractVersion := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint32_t {
		return C.ffi_didcomm_uniffi_uniffi_contract_version(uniffiStatus)
	})
	if bindingsContractVersion != int(scaffoldingContractVersion) {
		// If this happens try cleaning and rebuilding your project
		panic("didcomm: UniFFI contract version mismatch")
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_didcomm_uniffi_checksum_method_didcomm_pack_encrypted(uniffiStatus)
		})
		if checksum != 39375 {
			// If this happens try cleaning and rebuilding your project
			panic("didcomm: uniffi_didcomm_uniffi_checksum_method_didcomm_pack_encrypted: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_didcomm_uniffi_checksum_method_didcomm_pack_from_prior(uniffiStatus)
		})
		if checksum != 15651 {
			// If this happens try cleaning and rebuilding your project
			panic("didcomm: uniffi_didcomm_uniffi_checksum_method_didcomm_pack_from_prior: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_didcomm_uniffi_checksum_method_didcomm_pack_plaintext(uniffiStatus)
		})
		if checksum != 52391 {
			// If this happens try cleaning and rebuilding your project
			panic("didcomm: uniffi_didcomm_uniffi_checksum_method_didcomm_pack_plaintext: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_didcomm_uniffi_checksum_method_didcomm_pack_signed(uniffiStatus)
		})
		if checksum != 55567 {
			// If this happens try cleaning and rebuilding your project
			panic("didcomm: uniffi_didcomm_uniffi_checksum_method_didcomm_pack_signed: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_didcomm_uniffi_checksum_method_didcomm_unpack(uniffiStatus)
		})
		if checksum != 6374 {
			// If this happens try cleaning and rebuilding your project
			panic("didcomm: uniffi_didcomm_uniffi_checksum_method_didcomm_unpack: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_didcomm_uniffi_checksum_method_didcomm_unpack_from_prior(uniffiStatus)
		})
		if checksum != 7340 {
			// If this happens try cleaning and rebuilding your project
			panic("didcomm: uniffi_didcomm_uniffi_checksum_method_didcomm_unpack_from_prior: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_didcomm_uniffi_checksum_method_didcomm_wrap_in_forward(uniffiStatus)
		})
		if checksum != 26257 {
			// If this happens try cleaning and rebuilding your project
			panic("didcomm: uniffi_didcomm_uniffi_checksum_method_didcomm_wrap_in_forward: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_didcomm_uniffi_checksum_method_exampledidresolver_resolve(uniffiStatus)
		})
		if checksum != 58414 {
			// If this happens try cleaning and rebuilding your project
			panic("didcomm: uniffi_didcomm_uniffi_checksum_method_exampledidresolver_resolve: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_didcomm_uniffi_checksum_method_examplesecretsresolver_find_secrets(uniffiStatus)
		})
		if checksum != 14951 {
			// If this happens try cleaning and rebuilding your project
			panic("didcomm: uniffi_didcomm_uniffi_checksum_method_examplesecretsresolver_find_secrets: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_didcomm_uniffi_checksum_method_examplesecretsresolver_get_secret(uniffiStatus)
		})
		if checksum != 22685 {
			// If this happens try cleaning and rebuilding your project
			panic("didcomm: uniffi_didcomm_uniffi_checksum_method_examplesecretsresolver_get_secret: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_didcomm_uniffi_checksum_method_ondidresolverresult_error(uniffiStatus)
		})
		if checksum != 28651 {
			// If this happens try cleaning and rebuilding your project
			panic("didcomm: uniffi_didcomm_uniffi_checksum_method_ondidresolverresult_error: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_didcomm_uniffi_checksum_method_ondidresolverresult_success(uniffiStatus)
		})
		if checksum != 37370 {
			// If this happens try cleaning and rebuilding your project
			panic("didcomm: uniffi_didcomm_uniffi_checksum_method_ondidresolverresult_success: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_didcomm_uniffi_checksum_method_onfindsecretsresult_error(uniffiStatus)
		})
		if checksum != 38944 {
			// If this happens try cleaning and rebuilding your project
			panic("didcomm: uniffi_didcomm_uniffi_checksum_method_onfindsecretsresult_error: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_didcomm_uniffi_checksum_method_onfindsecretsresult_success(uniffiStatus)
		})
		if checksum != 55827 {
			// If this happens try cleaning and rebuilding your project
			panic("didcomm: uniffi_didcomm_uniffi_checksum_method_onfindsecretsresult_success: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_didcomm_uniffi_checksum_method_ongetsecretresult_error(uniffiStatus)
		})
		if checksum != 38171 {
			// If this happens try cleaning and rebuilding your project
			panic("didcomm: uniffi_didcomm_uniffi_checksum_method_ongetsecretresult_error: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_didcomm_uniffi_checksum_method_ongetsecretresult_success(uniffiStatus)
		})
		if checksum != 19322 {
			// If this happens try cleaning and rebuilding your project
			panic("didcomm: uniffi_didcomm_uniffi_checksum_method_ongetsecretresult_success: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_didcomm_uniffi_checksum_constructor_didcomm_new(uniffiStatus)
		})
		if checksum != 58818 {
			// If this happens try cleaning and rebuilding your project
			panic("didcomm: uniffi_didcomm_uniffi_checksum_constructor_didcomm_new: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_didcomm_uniffi_checksum_constructor_exampledidresolver_new(uniffiStatus)
		})
		if checksum != 48936 {
			// If this happens try cleaning and rebuilding your project
			panic("didcomm: uniffi_didcomm_uniffi_checksum_constructor_exampledidresolver_new: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_didcomm_uniffi_checksum_constructor_examplesecretsresolver_new(uniffiStatus)
		})
		if checksum != 39891 {
			// If this happens try cleaning and rebuilding your project
			panic("didcomm: uniffi_didcomm_uniffi_checksum_constructor_examplesecretsresolver_new: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_didcomm_uniffi_checksum_method_didresolver_resolve(uniffiStatus)
		})
		if checksum != 25278 {
			// If this happens try cleaning and rebuilding your project
			panic("didcomm: uniffi_didcomm_uniffi_checksum_method_didresolver_resolve: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_didcomm_uniffi_checksum_method_onfrompriorpackresult_success(uniffiStatus)
		})
		if checksum != 4630 {
			// If this happens try cleaning and rebuilding your project
			panic("didcomm: uniffi_didcomm_uniffi_checksum_method_onfrompriorpackresult_success: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_didcomm_uniffi_checksum_method_onfrompriorpackresult_error(uniffiStatus)
		})
		if checksum != 30267 {
			// If this happens try cleaning and rebuilding your project
			panic("didcomm: uniffi_didcomm_uniffi_checksum_method_onfrompriorpackresult_error: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_didcomm_uniffi_checksum_method_onfrompriorunpackresult_success(uniffiStatus)
		})
		if checksum != 14120 {
			// If this happens try cleaning and rebuilding your project
			panic("didcomm: uniffi_didcomm_uniffi_checksum_method_onfrompriorunpackresult_success: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_didcomm_uniffi_checksum_method_onfrompriorunpackresult_error(uniffiStatus)
		})
		if checksum != 34539 {
			// If this happens try cleaning and rebuilding your project
			panic("didcomm: uniffi_didcomm_uniffi_checksum_method_onfrompriorunpackresult_error: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_didcomm_uniffi_checksum_method_onpackencryptedresult_success(uniffiStatus)
		})
		if checksum != 42036 {
			// If this happens try cleaning and rebuilding your project
			panic("didcomm: uniffi_didcomm_uniffi_checksum_method_onpackencryptedresult_success: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_didcomm_uniffi_checksum_method_onpackencryptedresult_error(uniffiStatus)
		})
		if checksum != 56424 {
			// If this happens try cleaning and rebuilding your project
			panic("didcomm: uniffi_didcomm_uniffi_checksum_method_onpackencryptedresult_error: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_didcomm_uniffi_checksum_method_onpackplaintextresult_success(uniffiStatus)
		})
		if checksum != 10777 {
			// If this happens try cleaning and rebuilding your project
			panic("didcomm: uniffi_didcomm_uniffi_checksum_method_onpackplaintextresult_success: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_didcomm_uniffi_checksum_method_onpackplaintextresult_error(uniffiStatus)
		})
		if checksum != 19574 {
			// If this happens try cleaning and rebuilding your project
			panic("didcomm: uniffi_didcomm_uniffi_checksum_method_onpackplaintextresult_error: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_didcomm_uniffi_checksum_method_onpacksignedresult_success(uniffiStatus)
		})
		if checksum != 25146 {
			// If this happens try cleaning and rebuilding your project
			panic("didcomm: uniffi_didcomm_uniffi_checksum_method_onpacksignedresult_success: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_didcomm_uniffi_checksum_method_onpacksignedresult_error(uniffiStatus)
		})
		if checksum != 60782 {
			// If this happens try cleaning and rebuilding your project
			panic("didcomm: uniffi_didcomm_uniffi_checksum_method_onpacksignedresult_error: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_didcomm_uniffi_checksum_method_onunpackresult_success(uniffiStatus)
		})
		if checksum != 59804 {
			// If this happens try cleaning and rebuilding your project
			panic("didcomm: uniffi_didcomm_uniffi_checksum_method_onunpackresult_success: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_didcomm_uniffi_checksum_method_onunpackresult_error(uniffiStatus)
		})
		if checksum != 519 {
			// If this happens try cleaning and rebuilding your project
			panic("didcomm: uniffi_didcomm_uniffi_checksum_method_onunpackresult_error: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_didcomm_uniffi_checksum_method_onwrapinforwardresult_success(uniffiStatus)
		})
		if checksum != 25003 {
			// If this happens try cleaning and rebuilding your project
			panic("didcomm: uniffi_didcomm_uniffi_checksum_method_onwrapinforwardresult_success: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_didcomm_uniffi_checksum_method_onwrapinforwardresult_error(uniffiStatus)
		})
		if checksum != 17982 {
			// If this happens try cleaning and rebuilding your project
			panic("didcomm: uniffi_didcomm_uniffi_checksum_method_onwrapinforwardresult_error: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_didcomm_uniffi_checksum_method_secretsresolver_get_secret(uniffiStatus)
		})
		if checksum != 48141 {
			// If this happens try cleaning and rebuilding your project
			panic("didcomm: uniffi_didcomm_uniffi_checksum_method_secretsresolver_get_secret: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_didcomm_uniffi_checksum_method_secretsresolver_find_secrets(uniffiStatus)
		})
		if checksum != 27454 {
			// If this happens try cleaning and rebuilding your project
			panic("didcomm: uniffi_didcomm_uniffi_checksum_method_secretsresolver_find_secrets: UniFFI API checksum mismatch")
		}
	}
}

type FfiConverterUint64 struct{}

var FfiConverterUint64INSTANCE = FfiConverterUint64{}

func (FfiConverterUint64) Lower(value uint64) C.uint64_t {
	return C.uint64_t(value)
}

func (FfiConverterUint64) Write(writer io.Writer, value uint64) {
	writeUint64(writer, value)
}

func (FfiConverterUint64) Lift(value C.uint64_t) uint64 {
	return uint64(value)
}

func (FfiConverterUint64) Read(reader io.Reader) uint64 {
	return readUint64(reader)
}

type FfiDestroyerUint64 struct{}

func (FfiDestroyerUint64) Destroy(_ uint64) {}

type FfiConverterBool struct{}

var FfiConverterBoolINSTANCE = FfiConverterBool{}

func (FfiConverterBool) Lower(value bool) C.int8_t {
	if value {
		return C.int8_t(1)
	}
	return C.int8_t(0)
}

func (FfiConverterBool) Write(writer io.Writer, value bool) {
	if value {
		writeInt8(writer, 1)
	} else {
		writeInt8(writer, 0)
	}
}

func (FfiConverterBool) Lift(value C.int8_t) bool {
	return value != 0
}

func (FfiConverterBool) Read(reader io.Reader) bool {
	return readInt8(reader) != 0
}

type FfiDestroyerBool struct{}

func (FfiDestroyerBool) Destroy(_ bool) {}

type FfiConverterString struct{}

var FfiConverterStringINSTANCE = FfiConverterString{}

func (FfiConverterString) Lift(rb RustBufferI) string {
	defer rb.Free()
	reader := rb.AsReader()
	b, err := io.ReadAll(reader)
	if err != nil {
		panic(fmt.Errorf("reading reader: %w", err))
	}
	return string(b)
}

func (FfiConverterString) Read(reader io.Reader) string {
	length := readInt32(reader)
	buffer := make([]byte, length)
	read_length, err := reader.Read(buffer)
	if err != nil {
		panic(err)
	}
	if read_length != int(length) {
		panic(fmt.Errorf("bad read length when reading string, expected %d, read %d", length, read_length))
	}
	return string(buffer)
}

func (FfiConverterString) Lower(value string) RustBuffer {
	return stringToRustBuffer(value)
}

func (FfiConverterString) Write(writer io.Writer, value string) {
	if len(value) > math.MaxInt32 {
		panic("String is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	write_length, err := io.WriteString(writer, value)
	if err != nil {
		panic(err)
	}
	if write_length != len(value) {
		panic(fmt.Errorf("bad write length when writing string, expected %d, written %d", len(value), write_length))
	}
}

type FfiDestroyerString struct{}

func (FfiDestroyerString) Destroy(_ string) {}

// Below is an implementation of synchronization requirements outlined in the link.
// https://github.com/mozilla/uniffi-rs/blob/0dc031132d9493ca812c3af6e7dd60ad2ea95bf0/uniffi_bindgen/src/bindings/kotlin/templates/ObjectRuntime.kt#L31

type FfiObject struct {
	pointer      unsafe.Pointer
	callCounter  atomic.Int64
	freeFunction func(unsafe.Pointer, *C.RustCallStatus)
	destroyed    atomic.Bool
}

func newFfiObject(pointer unsafe.Pointer, freeFunction func(unsafe.Pointer, *C.RustCallStatus)) FfiObject {
	return FfiObject{
		pointer:      pointer,
		freeFunction: freeFunction,
	}
}

func (ffiObject *FfiObject) incrementPointer(debugName string) unsafe.Pointer {
	for {
		counter := ffiObject.callCounter.Load()
		if counter <= -1 {
			panic(fmt.Errorf("%v object has already been destroyed", debugName))
		}
		if counter == math.MaxInt64 {
			panic(fmt.Errorf("%v object call counter would overflow", debugName))
		}
		if ffiObject.callCounter.CompareAndSwap(counter, counter+1) {
			break
		}
	}

	return ffiObject.pointer
}

func (ffiObject *FfiObject) decrementPointer() {
	if ffiObject.callCounter.Add(-1) == -1 {
		ffiObject.freeRustArcPtr()
	}
}

func (ffiObject *FfiObject) destroy() {
	if ffiObject.destroyed.CompareAndSwap(false, true) {
		if ffiObject.callCounter.Add(-1) == -1 {
			ffiObject.freeRustArcPtr()
		}
	}
}

func (ffiObject *FfiObject) freeRustArcPtr() {
	rustCall(func(status *C.RustCallStatus) int32 {
		ffiObject.freeFunction(ffiObject.pointer, status)
		return 0
	})
}

type DidComm struct {
	ffiObject FfiObject
}

func NewDidComm(didResolver DidResolver, secretResolver SecretsResolver) *DidComm {
	return FfiConverterDIDCommINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_didcomm_uniffi_fn_constructor_didcomm_new(FfiConverterCallbackInterfaceDIDResolverINSTANCE.Lower(didResolver), FfiConverterCallbackInterfaceSecretsResolverINSTANCE.Lower(secretResolver), _uniffiStatus)
	}))
}

func (_self *DidComm) PackEncrypted(msg Message, to string, from *string, signBy *string, options PackEncryptedOptions, cb OnPackEncryptedResult) ErrorCode {
	_pointer := _self.ffiObject.incrementPointer("*DidComm")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTypeErrorCodeINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_didcomm_uniffi_fn_method_didcomm_pack_encrypted(
			_pointer, FfiConverterTypeMessageINSTANCE.Lower(msg), FfiConverterStringINSTANCE.Lower(to), FfiConverterOptionalStringINSTANCE.Lower(from), FfiConverterOptionalStringINSTANCE.Lower(signBy), FfiConverterTypePackEncryptedOptionsINSTANCE.Lower(options), FfiConverterCallbackInterfaceOnPackEncryptedResultINSTANCE.Lower(cb), _uniffiStatus)
	}))
}

func (_self *DidComm) PackFromPrior(msg FromPrior, issuerKid *string, cb OnFromPriorPackResult) ErrorCode {
	_pointer := _self.ffiObject.incrementPointer("*DidComm")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTypeErrorCodeINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_didcomm_uniffi_fn_method_didcomm_pack_from_prior(
			_pointer, FfiConverterTypeFromPriorINSTANCE.Lower(msg), FfiConverterOptionalStringINSTANCE.Lower(issuerKid), FfiConverterCallbackInterfaceOnFromPriorPackResultINSTANCE.Lower(cb), _uniffiStatus)
	}))
}

func (_self *DidComm) PackPlaintext(msg Message, cb OnPackPlaintextResult) ErrorCode {
	_pointer := _self.ffiObject.incrementPointer("*DidComm")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTypeErrorCodeINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_didcomm_uniffi_fn_method_didcomm_pack_plaintext(
			_pointer, FfiConverterTypeMessageINSTANCE.Lower(msg), FfiConverterCallbackInterfaceOnPackPlaintextResultINSTANCE.Lower(cb), _uniffiStatus)
	}))
}

func (_self *DidComm) PackSigned(msg Message, signBy string, cb OnPackSignedResult) ErrorCode {
	_pointer := _self.ffiObject.incrementPointer("*DidComm")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTypeErrorCodeINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_didcomm_uniffi_fn_method_didcomm_pack_signed(
			_pointer, FfiConverterTypeMessageINSTANCE.Lower(msg), FfiConverterStringINSTANCE.Lower(signBy), FfiConverterCallbackInterfaceOnPackSignedResultINSTANCE.Lower(cb), _uniffiStatus)
	}))
}

func (_self *DidComm) Unpack(msg string, options UnpackOptions, cb OnUnpackResult) ErrorCode {
	_pointer := _self.ffiObject.incrementPointer("*DidComm")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTypeErrorCodeINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_didcomm_uniffi_fn_method_didcomm_unpack(
			_pointer, FfiConverterStringINSTANCE.Lower(msg), FfiConverterTypeUnpackOptionsINSTANCE.Lower(options), FfiConverterCallbackInterfaceOnUnpackResultINSTANCE.Lower(cb), _uniffiStatus)
	}))
}

func (_self *DidComm) UnpackFromPrior(fromPriorJwt string, cb OnFromPriorUnpackResult) ErrorCode {
	_pointer := _self.ffiObject.incrementPointer("*DidComm")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTypeErrorCodeINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_didcomm_uniffi_fn_method_didcomm_unpack_from_prior(
			_pointer, FfiConverterStringINSTANCE.Lower(fromPriorJwt), FfiConverterCallbackInterfaceOnFromPriorUnpackResultINSTANCE.Lower(cb), _uniffiStatus)
	}))
}

func (_self *DidComm) WrapInForward(msg string, headers map[string]JsonValue, to string, routingKeys []string, encAlgAnon AnonCryptAlg, cb OnWrapInForwardResult) ErrorCode {
	_pointer := _self.ffiObject.incrementPointer("*DidComm")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTypeErrorCodeINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_didcomm_uniffi_fn_method_didcomm_wrap_in_forward(
			_pointer, FfiConverterStringINSTANCE.Lower(msg), FfiConverterMapStringTypeJsonValueINSTANCE.Lower(headers), FfiConverterStringINSTANCE.Lower(to), FfiConverterSequenceStringINSTANCE.Lower(routingKeys), FfiConverterTypeAnonCryptAlgINSTANCE.Lower(encAlgAnon), FfiConverterCallbackInterfaceOnWrapInForwardResultINSTANCE.Lower(cb), _uniffiStatus)
	}))
}

func (object *DidComm) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterDIDComm struct{}

var FfiConverterDIDCommINSTANCE = FfiConverterDIDComm{}

func (c FfiConverterDIDComm) Lift(pointer unsafe.Pointer) *DidComm {
	result := &DidComm{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_didcomm_uniffi_fn_free_didcomm(pointer, status)
			}),
	}
	runtime.SetFinalizer(result, (*DidComm).Destroy)
	return result
}

func (c FfiConverterDIDComm) Read(reader io.Reader) *DidComm {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterDIDComm) Lower(value *DidComm) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*DidComm")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterDIDComm) Write(writer io.Writer, value *DidComm) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerDidComm struct{}

func (_ FfiDestroyerDidComm) Destroy(value *DidComm) {
	value.Destroy()
}

type ExampleDidResolver struct {
	ffiObject FfiObject
}

func NewExampleDidResolver(knownDids []DidDoc) *ExampleDidResolver {
	return FfiConverterExampleDIDResolverINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_didcomm_uniffi_fn_constructor_exampledidresolver_new(FfiConverterSequenceTypeDIDDocINSTANCE.Lower(knownDids), _uniffiStatus)
	}))
}

func (_self *ExampleDidResolver) Resolve(did string, cb *OnDidResolverResult) ErrorCode {
	_pointer := _self.ffiObject.incrementPointer("*ExampleDidResolver")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTypeErrorCodeINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_didcomm_uniffi_fn_method_exampledidresolver_resolve(
			_pointer, FfiConverterStringINSTANCE.Lower(did), FfiConverterOnDIDResolverResultINSTANCE.Lower(cb), _uniffiStatus)
	}))
}

func (object *ExampleDidResolver) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterExampleDIDResolver struct{}

var FfiConverterExampleDIDResolverINSTANCE = FfiConverterExampleDIDResolver{}

func (c FfiConverterExampleDIDResolver) Lift(pointer unsafe.Pointer) *ExampleDidResolver {
	result := &ExampleDidResolver{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_didcomm_uniffi_fn_free_exampledidresolver(pointer, status)
			}),
	}
	runtime.SetFinalizer(result, (*ExampleDidResolver).Destroy)
	return result
}

func (c FfiConverterExampleDIDResolver) Read(reader io.Reader) *ExampleDidResolver {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterExampleDIDResolver) Lower(value *ExampleDidResolver) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*ExampleDidResolver")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterExampleDIDResolver) Write(writer io.Writer, value *ExampleDidResolver) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerExampleDidResolver struct{}

func (_ FfiDestroyerExampleDidResolver) Destroy(value *ExampleDidResolver) {
	value.Destroy()
}

type ExampleSecretsResolver struct {
	ffiObject FfiObject
}

func NewExampleSecretsResolver(knownSecrets []Secret) *ExampleSecretsResolver {
	return FfiConverterExampleSecretsResolverINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_didcomm_uniffi_fn_constructor_examplesecretsresolver_new(FfiConverterSequenceTypeSecretINSTANCE.Lower(knownSecrets), _uniffiStatus)
	}))
}

func (_self *ExampleSecretsResolver) FindSecrets(secretIds []string, cb *OnFindSecretsResult) ErrorCode {
	_pointer := _self.ffiObject.incrementPointer("*ExampleSecretsResolver")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTypeErrorCodeINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_didcomm_uniffi_fn_method_examplesecretsresolver_find_secrets(
			_pointer, FfiConverterSequenceStringINSTANCE.Lower(secretIds), FfiConverterOnFindSecretsResultINSTANCE.Lower(cb), _uniffiStatus)
	}))
}

func (_self *ExampleSecretsResolver) GetSecret(secretId string, cb *OnGetSecretResult) ErrorCode {
	_pointer := _self.ffiObject.incrementPointer("*ExampleSecretsResolver")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTypeErrorCodeINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_didcomm_uniffi_fn_method_examplesecretsresolver_get_secret(
			_pointer, FfiConverterStringINSTANCE.Lower(secretId), FfiConverterOnGetSecretResultINSTANCE.Lower(cb), _uniffiStatus)
	}))
}

func (object *ExampleSecretsResolver) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterExampleSecretsResolver struct{}

var FfiConverterExampleSecretsResolverINSTANCE = FfiConverterExampleSecretsResolver{}

func (c FfiConverterExampleSecretsResolver) Lift(pointer unsafe.Pointer) *ExampleSecretsResolver {
	result := &ExampleSecretsResolver{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_didcomm_uniffi_fn_free_examplesecretsresolver(pointer, status)
			}),
	}
	runtime.SetFinalizer(result, (*ExampleSecretsResolver).Destroy)
	return result
}

func (c FfiConverterExampleSecretsResolver) Read(reader io.Reader) *ExampleSecretsResolver {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterExampleSecretsResolver) Lower(value *ExampleSecretsResolver) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*ExampleSecretsResolver")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterExampleSecretsResolver) Write(writer io.Writer, value *ExampleSecretsResolver) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerExampleSecretsResolver struct{}

func (_ FfiDestroyerExampleSecretsResolver) Destroy(value *ExampleSecretsResolver) {
	value.Destroy()
}

type OnDidResolverResult struct {
	ffiObject FfiObject
}

func (_self *OnDidResolverResult) Error(err *ErrorKind, msg string) error {
	_pointer := _self.ffiObject.incrementPointer("*OnDidResolverResult")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeErrorKind{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_didcomm_uniffi_fn_method_ondidresolverresult_error(
			_pointer, FfiConverterTypeErrorKindINSTANCE.Lower(err), FfiConverterStringINSTANCE.Lower(msg), _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (_self *OnDidResolverResult) Success(result *DidDoc) error {
	_pointer := _self.ffiObject.incrementPointer("*OnDidResolverResult")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeErrorKind{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_didcomm_uniffi_fn_method_ondidresolverresult_success(
			_pointer, FfiConverterOptionalTypeDIDDocINSTANCE.Lower(result), _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (object *OnDidResolverResult) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterOnDIDResolverResult struct{}

var FfiConverterOnDIDResolverResultINSTANCE = FfiConverterOnDIDResolverResult{}

func (c FfiConverterOnDIDResolverResult) Lift(pointer unsafe.Pointer) *OnDidResolverResult {
	result := &OnDidResolverResult{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_didcomm_uniffi_fn_free_ondidresolverresult(pointer, status)
			}),
	}
	runtime.SetFinalizer(result, (*OnDidResolverResult).Destroy)
	return result
}

func (c FfiConverterOnDIDResolverResult) Read(reader io.Reader) *OnDidResolverResult {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterOnDIDResolverResult) Lower(value *OnDidResolverResult) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*OnDidResolverResult")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterOnDIDResolverResult) Write(writer io.Writer, value *OnDidResolverResult) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerOnDidResolverResult struct{}

func (_ FfiDestroyerOnDidResolverResult) Destroy(value *OnDidResolverResult) {
	value.Destroy()
}

type OnFindSecretsResult struct {
	ffiObject FfiObject
}

func (_self *OnFindSecretsResult) Error(err *ErrorKind, msg string) error {
	_pointer := _self.ffiObject.incrementPointer("*OnFindSecretsResult")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeErrorKind{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_didcomm_uniffi_fn_method_onfindsecretsresult_error(
			_pointer, FfiConverterTypeErrorKindINSTANCE.Lower(err), FfiConverterStringINSTANCE.Lower(msg), _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (_self *OnFindSecretsResult) Success(result []string) error {
	_pointer := _self.ffiObject.incrementPointer("*OnFindSecretsResult")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeErrorKind{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_didcomm_uniffi_fn_method_onfindsecretsresult_success(
			_pointer, FfiConverterSequenceStringINSTANCE.Lower(result), _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (object *OnFindSecretsResult) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterOnFindSecretsResult struct{}

var FfiConverterOnFindSecretsResultINSTANCE = FfiConverterOnFindSecretsResult{}

func (c FfiConverterOnFindSecretsResult) Lift(pointer unsafe.Pointer) *OnFindSecretsResult {
	result := &OnFindSecretsResult{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_didcomm_uniffi_fn_free_onfindsecretsresult(pointer, status)
			}),
	}
	runtime.SetFinalizer(result, (*OnFindSecretsResult).Destroy)
	return result
}

func (c FfiConverterOnFindSecretsResult) Read(reader io.Reader) *OnFindSecretsResult {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterOnFindSecretsResult) Lower(value *OnFindSecretsResult) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*OnFindSecretsResult")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterOnFindSecretsResult) Write(writer io.Writer, value *OnFindSecretsResult) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerOnFindSecretsResult struct{}

func (_ FfiDestroyerOnFindSecretsResult) Destroy(value *OnFindSecretsResult) {
	value.Destroy()
}

type OnGetSecretResult struct {
	ffiObject FfiObject
}

func (_self *OnGetSecretResult) Error(err *ErrorKind, msg string) error {
	_pointer := _self.ffiObject.incrementPointer("*OnGetSecretResult")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeErrorKind{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_didcomm_uniffi_fn_method_ongetsecretresult_error(
			_pointer, FfiConverterTypeErrorKindINSTANCE.Lower(err), FfiConverterStringINSTANCE.Lower(msg), _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (_self *OnGetSecretResult) Success(result *Secret) error {
	_pointer := _self.ffiObject.incrementPointer("*OnGetSecretResult")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeErrorKind{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_didcomm_uniffi_fn_method_ongetsecretresult_success(
			_pointer, FfiConverterOptionalTypeSecretINSTANCE.Lower(result), _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (object *OnGetSecretResult) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterOnGetSecretResult struct{}

var FfiConverterOnGetSecretResultINSTANCE = FfiConverterOnGetSecretResult{}

func (c FfiConverterOnGetSecretResult) Lift(pointer unsafe.Pointer) *OnGetSecretResult {
	result := &OnGetSecretResult{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_didcomm_uniffi_fn_free_ongetsecretresult(pointer, status)
			}),
	}
	runtime.SetFinalizer(result, (*OnGetSecretResult).Destroy)
	return result
}

func (c FfiConverterOnGetSecretResult) Read(reader io.Reader) *OnGetSecretResult {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterOnGetSecretResult) Lower(value *OnGetSecretResult) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*OnGetSecretResult")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterOnGetSecretResult) Write(writer io.Writer, value *OnGetSecretResult) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerOnGetSecretResult struct{}

func (_ FfiDestroyerOnGetSecretResult) Destroy(value *OnGetSecretResult) {
	value.Destroy()
}

type Attachment struct {
	Data        AttachmentData
	Id          *string
	Description *string
	Filename    *string
	MediaType   *string
	Format      *string
	LastmodTime *uint64
	ByteCount   *uint64
}

func (r *Attachment) Destroy() {
	FfiDestroyerTypeAttachmentData{}.Destroy(r.Data)
	FfiDestroyerOptionalString{}.Destroy(r.Id)
	FfiDestroyerOptionalString{}.Destroy(r.Description)
	FfiDestroyerOptionalString{}.Destroy(r.Filename)
	FfiDestroyerOptionalString{}.Destroy(r.MediaType)
	FfiDestroyerOptionalString{}.Destroy(r.Format)
	FfiDestroyerOptionalUint64{}.Destroy(r.LastmodTime)
	FfiDestroyerOptionalUint64{}.Destroy(r.ByteCount)
}

type FfiConverterTypeAttachment struct{}

var FfiConverterTypeAttachmentINSTANCE = FfiConverterTypeAttachment{}

func (c FfiConverterTypeAttachment) Lift(rb RustBufferI) Attachment {
	return LiftFromRustBuffer[Attachment](c, rb)
}

func (c FfiConverterTypeAttachment) Read(reader io.Reader) Attachment {
	return Attachment{
		FfiConverterTypeAttachmentDataINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterOptionalUint64INSTANCE.Read(reader),
		FfiConverterOptionalUint64INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeAttachment) Lower(value Attachment) RustBuffer {
	return LowerIntoRustBuffer[Attachment](c, value)
}

func (c FfiConverterTypeAttachment) Write(writer io.Writer, value Attachment) {
	FfiConverterTypeAttachmentDataINSTANCE.Write(writer, value.Data)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.Id)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.Description)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.Filename)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.MediaType)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.Format)
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.LastmodTime)
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.ByteCount)
}

type FfiDestroyerTypeAttachment struct{}

func (_ FfiDestroyerTypeAttachment) Destroy(value Attachment) {
	value.Destroy()
}

type Base64AttachmentData struct {
	Base64 string
	Jws    *string
}

func (r *Base64AttachmentData) Destroy() {
	FfiDestroyerString{}.Destroy(r.Base64)
	FfiDestroyerOptionalString{}.Destroy(r.Jws)
}

type FfiConverterTypeBase64AttachmentData struct{}

var FfiConverterTypeBase64AttachmentDataINSTANCE = FfiConverterTypeBase64AttachmentData{}

func (c FfiConverterTypeBase64AttachmentData) Lift(rb RustBufferI) Base64AttachmentData {
	return LiftFromRustBuffer[Base64AttachmentData](c, rb)
}

func (c FfiConverterTypeBase64AttachmentData) Read(reader io.Reader) Base64AttachmentData {
	return Base64AttachmentData{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeBase64AttachmentData) Lower(value Base64AttachmentData) RustBuffer {
	return LowerIntoRustBuffer[Base64AttachmentData](c, value)
}

func (c FfiConverterTypeBase64AttachmentData) Write(writer io.Writer, value Base64AttachmentData) {
	FfiConverterStringINSTANCE.Write(writer, value.Base64)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.Jws)
}

type FfiDestroyerTypeBase64AttachmentData struct{}

func (_ FfiDestroyerTypeBase64AttachmentData) Destroy(value Base64AttachmentData) {
	value.Destroy()
}

type DidCommMessagingService struct {
	Uri         string
	Accept      *[]string
	RoutingKeys []string
}

func (r *DidCommMessagingService) Destroy() {
	FfiDestroyerString{}.Destroy(r.Uri)
	FfiDestroyerOptionalSequenceString{}.Destroy(r.Accept)
	FfiDestroyerSequenceString{}.Destroy(r.RoutingKeys)
}

type FfiConverterTypeDIDCommMessagingService struct{}

var FfiConverterTypeDIDCommMessagingServiceINSTANCE = FfiConverterTypeDIDCommMessagingService{}

func (c FfiConverterTypeDIDCommMessagingService) Lift(rb RustBufferI) DidCommMessagingService {
	return LiftFromRustBuffer[DidCommMessagingService](c, rb)
}

func (c FfiConverterTypeDIDCommMessagingService) Read(reader io.Reader) DidCommMessagingService {
	return DidCommMessagingService{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterOptionalSequenceStringINSTANCE.Read(reader),
		FfiConverterSequenceStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeDIDCommMessagingService) Lower(value DidCommMessagingService) RustBuffer {
	return LowerIntoRustBuffer[DidCommMessagingService](c, value)
}

func (c FfiConverterTypeDIDCommMessagingService) Write(writer io.Writer, value DidCommMessagingService) {
	FfiConverterStringINSTANCE.Write(writer, value.Uri)
	FfiConverterOptionalSequenceStringINSTANCE.Write(writer, value.Accept)
	FfiConverterSequenceStringINSTANCE.Write(writer, value.RoutingKeys)
}

type FfiDestroyerTypeDidCommMessagingService struct{}

func (_ FfiDestroyerTypeDidCommMessagingService) Destroy(value DidCommMessagingService) {
	value.Destroy()
}

type DidDoc struct {
	Id                 string
	KeyAgreement       []string
	Authentication     []string
	VerificationMethod []VerificationMethod
	Service            []Service
}

func (r *DidDoc) Destroy() {
	FfiDestroyerString{}.Destroy(r.Id)
	FfiDestroyerSequenceString{}.Destroy(r.KeyAgreement)
	FfiDestroyerSequenceString{}.Destroy(r.Authentication)
	FfiDestroyerSequenceTypeVerificationMethod{}.Destroy(r.VerificationMethod)
	FfiDestroyerSequenceTypeService{}.Destroy(r.Service)
}

type FfiConverterTypeDIDDoc struct{}

var FfiConverterTypeDIDDocINSTANCE = FfiConverterTypeDIDDoc{}

func (c FfiConverterTypeDIDDoc) Lift(rb RustBufferI) DidDoc {
	return LiftFromRustBuffer[DidDoc](c, rb)
}

func (c FfiConverterTypeDIDDoc) Read(reader io.Reader) DidDoc {
	return DidDoc{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterSequenceStringINSTANCE.Read(reader),
		FfiConverterSequenceStringINSTANCE.Read(reader),
		FfiConverterSequenceTypeVerificationMethodINSTANCE.Read(reader),
		FfiConverterSequenceTypeServiceINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeDIDDoc) Lower(value DidDoc) RustBuffer {
	return LowerIntoRustBuffer[DidDoc](c, value)
}

func (c FfiConverterTypeDIDDoc) Write(writer io.Writer, value DidDoc) {
	FfiConverterStringINSTANCE.Write(writer, value.Id)
	FfiConverterSequenceStringINSTANCE.Write(writer, value.KeyAgreement)
	FfiConverterSequenceStringINSTANCE.Write(writer, value.Authentication)
	FfiConverterSequenceTypeVerificationMethodINSTANCE.Write(writer, value.VerificationMethod)
	FfiConverterSequenceTypeServiceINSTANCE.Write(writer, value.Service)
}

type FfiDestroyerTypeDidDoc struct{}

func (_ FfiDestroyerTypeDidDoc) Destroy(value DidDoc) {
	value.Destroy()
}

type FromPrior struct {
	Iss string
	Sub string
	Aud *string
	Exp *uint64
	Nbf *uint64
	Iat *uint64
	Jti *string
}

func (r *FromPrior) Destroy() {
	FfiDestroyerString{}.Destroy(r.Iss)
	FfiDestroyerString{}.Destroy(r.Sub)
	FfiDestroyerOptionalString{}.Destroy(r.Aud)
	FfiDestroyerOptionalUint64{}.Destroy(r.Exp)
	FfiDestroyerOptionalUint64{}.Destroy(r.Nbf)
	FfiDestroyerOptionalUint64{}.Destroy(r.Iat)
	FfiDestroyerOptionalString{}.Destroy(r.Jti)
}

type FfiConverterTypeFromPrior struct{}

var FfiConverterTypeFromPriorINSTANCE = FfiConverterTypeFromPrior{}

func (c FfiConverterTypeFromPrior) Lift(rb RustBufferI) FromPrior {
	return LiftFromRustBuffer[FromPrior](c, rb)
}

func (c FfiConverterTypeFromPrior) Read(reader io.Reader) FromPrior {
	return FromPrior{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterOptionalUint64INSTANCE.Read(reader),
		FfiConverterOptionalUint64INSTANCE.Read(reader),
		FfiConverterOptionalUint64INSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeFromPrior) Lower(value FromPrior) RustBuffer {
	return LowerIntoRustBuffer[FromPrior](c, value)
}

func (c FfiConverterTypeFromPrior) Write(writer io.Writer, value FromPrior) {
	FfiConverterStringINSTANCE.Write(writer, value.Iss)
	FfiConverterStringINSTANCE.Write(writer, value.Sub)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.Aud)
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.Exp)
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.Nbf)
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.Iat)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.Jti)
}

type FfiDestroyerTypeFromPrior struct{}

func (_ FfiDestroyerTypeFromPrior) Destroy(value FromPrior) {
	value.Destroy()
}

type JsonAttachmentData struct {
	Json JsonValue
	Jws  *string
}

func (r *JsonAttachmentData) Destroy() {
	FfiDestroyerTypeJsonValue{}.Destroy(r.Json)
	FfiDestroyerOptionalString{}.Destroy(r.Jws)
}

type FfiConverterTypeJsonAttachmentData struct{}

var FfiConverterTypeJsonAttachmentDataINSTANCE = FfiConverterTypeJsonAttachmentData{}

func (c FfiConverterTypeJsonAttachmentData) Lift(rb RustBufferI) JsonAttachmentData {
	return LiftFromRustBuffer[JsonAttachmentData](c, rb)
}

func (c FfiConverterTypeJsonAttachmentData) Read(reader io.Reader) JsonAttachmentData {
	return JsonAttachmentData{
		FfiConverterTypeJsonValueINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeJsonAttachmentData) Lower(value JsonAttachmentData) RustBuffer {
	return LowerIntoRustBuffer[JsonAttachmentData](c, value)
}

func (c FfiConverterTypeJsonAttachmentData) Write(writer io.Writer, value JsonAttachmentData) {
	FfiConverterTypeJsonValueINSTANCE.Write(writer, value.Json)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.Jws)
}

type FfiDestroyerTypeJsonAttachmentData struct{}

func (_ FfiDestroyerTypeJsonAttachmentData) Destroy(value JsonAttachmentData) {
	value.Destroy()
}

type LinksAttachmentData struct {
	Links []string
	Hash  string
	Jws   *string
}

func (r *LinksAttachmentData) Destroy() {
	FfiDestroyerSequenceString{}.Destroy(r.Links)
	FfiDestroyerString{}.Destroy(r.Hash)
	FfiDestroyerOptionalString{}.Destroy(r.Jws)
}

type FfiConverterTypeLinksAttachmentData struct{}

var FfiConverterTypeLinksAttachmentDataINSTANCE = FfiConverterTypeLinksAttachmentData{}

func (c FfiConverterTypeLinksAttachmentData) Lift(rb RustBufferI) LinksAttachmentData {
	return LiftFromRustBuffer[LinksAttachmentData](c, rb)
}

func (c FfiConverterTypeLinksAttachmentData) Read(reader io.Reader) LinksAttachmentData {
	return LinksAttachmentData{
		FfiConverterSequenceStringINSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeLinksAttachmentData) Lower(value LinksAttachmentData) RustBuffer {
	return LowerIntoRustBuffer[LinksAttachmentData](c, value)
}

func (c FfiConverterTypeLinksAttachmentData) Write(writer io.Writer, value LinksAttachmentData) {
	FfiConverterSequenceStringINSTANCE.Write(writer, value.Links)
	FfiConverterStringINSTANCE.Write(writer, value.Hash)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.Jws)
}

type FfiDestroyerTypeLinksAttachmentData struct{}

func (_ FfiDestroyerTypeLinksAttachmentData) Destroy(value LinksAttachmentData) {
	value.Destroy()
}

type Message struct {
	Id           string
	Typ          string
	Type         string
	Body         JsonValue
	From         *string
	To           *[]string
	Thid         *string
	Pthid        *string
	ExtraHeaders map[string]JsonValue
	CreatedTime  *uint64
	ExpiresTime  *uint64
	FromPrior    *string
	Attachments  *[]Attachment
}

func (r *Message) Destroy() {
	FfiDestroyerString{}.Destroy(r.Id)
	FfiDestroyerString{}.Destroy(r.Typ)
	FfiDestroyerString{}.Destroy(r.Type)
	FfiDestroyerTypeJsonValue{}.Destroy(r.Body)
	FfiDestroyerOptionalString{}.Destroy(r.From)
	FfiDestroyerOptionalSequenceString{}.Destroy(r.To)
	FfiDestroyerOptionalString{}.Destroy(r.Thid)
	FfiDestroyerOptionalString{}.Destroy(r.Pthid)
	FfiDestroyerMapStringTypeJsonValue{}.Destroy(r.ExtraHeaders)
	FfiDestroyerOptionalUint64{}.Destroy(r.CreatedTime)
	FfiDestroyerOptionalUint64{}.Destroy(r.ExpiresTime)
	FfiDestroyerOptionalString{}.Destroy(r.FromPrior)
	FfiDestroyerOptionalSequenceTypeAttachment{}.Destroy(r.Attachments)
}

type FfiConverterTypeMessage struct{}

var FfiConverterTypeMessageINSTANCE = FfiConverterTypeMessage{}

func (c FfiConverterTypeMessage) Lift(rb RustBufferI) Message {
	return LiftFromRustBuffer[Message](c, rb)
}

func (c FfiConverterTypeMessage) Read(reader io.Reader) Message {
	return Message{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterTypeJsonValueINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterOptionalSequenceStringINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterMapStringTypeJsonValueINSTANCE.Read(reader),
		FfiConverterOptionalUint64INSTANCE.Read(reader),
		FfiConverterOptionalUint64INSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterOptionalSequenceTypeAttachmentINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeMessage) Lower(value Message) RustBuffer {
	return LowerIntoRustBuffer[Message](c, value)
}

func (c FfiConverterTypeMessage) Write(writer io.Writer, value Message) {
	FfiConverterStringINSTANCE.Write(writer, value.Id)
	FfiConverterStringINSTANCE.Write(writer, value.Typ)
	FfiConverterStringINSTANCE.Write(writer, value.Type)
	FfiConverterTypeJsonValueINSTANCE.Write(writer, value.Body)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.From)
	FfiConverterOptionalSequenceStringINSTANCE.Write(writer, value.To)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.Thid)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.Pthid)
	FfiConverterMapStringTypeJsonValueINSTANCE.Write(writer, value.ExtraHeaders)
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.CreatedTime)
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.ExpiresTime)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.FromPrior)
	FfiConverterOptionalSequenceTypeAttachmentINSTANCE.Write(writer, value.Attachments)
}

type FfiDestroyerTypeMessage struct{}

func (_ FfiDestroyerTypeMessage) Destroy(value Message) {
	value.Destroy()
}

type MessagingServiceMetadata struct {
	Id              string
	ServiceEndpoint string
}

func (r *MessagingServiceMetadata) Destroy() {
	FfiDestroyerString{}.Destroy(r.Id)
	FfiDestroyerString{}.Destroy(r.ServiceEndpoint)
}

type FfiConverterTypeMessagingServiceMetadata struct{}

var FfiConverterTypeMessagingServiceMetadataINSTANCE = FfiConverterTypeMessagingServiceMetadata{}

func (c FfiConverterTypeMessagingServiceMetadata) Lift(rb RustBufferI) MessagingServiceMetadata {
	return LiftFromRustBuffer[MessagingServiceMetadata](c, rb)
}

func (c FfiConverterTypeMessagingServiceMetadata) Read(reader io.Reader) MessagingServiceMetadata {
	return MessagingServiceMetadata{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeMessagingServiceMetadata) Lower(value MessagingServiceMetadata) RustBuffer {
	return LowerIntoRustBuffer[MessagingServiceMetadata](c, value)
}

func (c FfiConverterTypeMessagingServiceMetadata) Write(writer io.Writer, value MessagingServiceMetadata) {
	FfiConverterStringINSTANCE.Write(writer, value.Id)
	FfiConverterStringINSTANCE.Write(writer, value.ServiceEndpoint)
}

type FfiDestroyerTypeMessagingServiceMetadata struct{}

func (_ FfiDestroyerTypeMessagingServiceMetadata) Destroy(value MessagingServiceMetadata) {
	value.Destroy()
}

type PackEncryptedMetadata struct {
	MessagingService *MessagingServiceMetadata
	FromKid          *string
	SignByKid        *string
	ToKids           []string
}

func (r *PackEncryptedMetadata) Destroy() {
	FfiDestroyerOptionalTypeMessagingServiceMetadata{}.Destroy(r.MessagingService)
	FfiDestroyerOptionalString{}.Destroy(r.FromKid)
	FfiDestroyerOptionalString{}.Destroy(r.SignByKid)
	FfiDestroyerSequenceString{}.Destroy(r.ToKids)
}

type FfiConverterTypePackEncryptedMetadata struct{}

var FfiConverterTypePackEncryptedMetadataINSTANCE = FfiConverterTypePackEncryptedMetadata{}

func (c FfiConverterTypePackEncryptedMetadata) Lift(rb RustBufferI) PackEncryptedMetadata {
	return LiftFromRustBuffer[PackEncryptedMetadata](c, rb)
}

func (c FfiConverterTypePackEncryptedMetadata) Read(reader io.Reader) PackEncryptedMetadata {
	return PackEncryptedMetadata{
		FfiConverterOptionalTypeMessagingServiceMetadataINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterSequenceStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypePackEncryptedMetadata) Lower(value PackEncryptedMetadata) RustBuffer {
	return LowerIntoRustBuffer[PackEncryptedMetadata](c, value)
}

func (c FfiConverterTypePackEncryptedMetadata) Write(writer io.Writer, value PackEncryptedMetadata) {
	FfiConverterOptionalTypeMessagingServiceMetadataINSTANCE.Write(writer, value.MessagingService)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.FromKid)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.SignByKid)
	FfiConverterSequenceStringINSTANCE.Write(writer, value.ToKids)
}

type FfiDestroyerTypePackEncryptedMetadata struct{}

func (_ FfiDestroyerTypePackEncryptedMetadata) Destroy(value PackEncryptedMetadata) {
	value.Destroy()
}

type PackEncryptedOptions struct {
	ProtectSender    bool
	Forward          bool
	ForwardHeaders   *map[string]JsonValue
	MessagingService *string
	EncAlgAuth       AuthCryptAlg
	EncAlgAnon       AnonCryptAlg
}

func (r *PackEncryptedOptions) Destroy() {
	FfiDestroyerBool{}.Destroy(r.ProtectSender)
	FfiDestroyerBool{}.Destroy(r.Forward)
	FfiDestroyerOptionalMapStringTypeJsonValue{}.Destroy(r.ForwardHeaders)
	FfiDestroyerOptionalString{}.Destroy(r.MessagingService)
	FfiDestroyerTypeAuthCryptAlg{}.Destroy(r.EncAlgAuth)
	FfiDestroyerTypeAnonCryptAlg{}.Destroy(r.EncAlgAnon)
}

type FfiConverterTypePackEncryptedOptions struct{}

var FfiConverterTypePackEncryptedOptionsINSTANCE = FfiConverterTypePackEncryptedOptions{}

func (c FfiConverterTypePackEncryptedOptions) Lift(rb RustBufferI) PackEncryptedOptions {
	return LiftFromRustBuffer[PackEncryptedOptions](c, rb)
}

func (c FfiConverterTypePackEncryptedOptions) Read(reader io.Reader) PackEncryptedOptions {
	return PackEncryptedOptions{
		FfiConverterBoolINSTANCE.Read(reader),
		FfiConverterBoolINSTANCE.Read(reader),
		FfiConverterOptionalMapStringTypeJsonValueINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterTypeAuthCryptAlgINSTANCE.Read(reader),
		FfiConverterTypeAnonCryptAlgINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypePackEncryptedOptions) Lower(value PackEncryptedOptions) RustBuffer {
	return LowerIntoRustBuffer[PackEncryptedOptions](c, value)
}

func (c FfiConverterTypePackEncryptedOptions) Write(writer io.Writer, value PackEncryptedOptions) {
	FfiConverterBoolINSTANCE.Write(writer, value.ProtectSender)
	FfiConverterBoolINSTANCE.Write(writer, value.Forward)
	FfiConverterOptionalMapStringTypeJsonValueINSTANCE.Write(writer, value.ForwardHeaders)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.MessagingService)
	FfiConverterTypeAuthCryptAlgINSTANCE.Write(writer, value.EncAlgAuth)
	FfiConverterTypeAnonCryptAlgINSTANCE.Write(writer, value.EncAlgAnon)
}

type FfiDestroyerTypePackEncryptedOptions struct{}

func (_ FfiDestroyerTypePackEncryptedOptions) Destroy(value PackEncryptedOptions) {
	value.Destroy()
}

type PackSignedMetadata struct {
	SignByKid string
}

func (r *PackSignedMetadata) Destroy() {
	FfiDestroyerString{}.Destroy(r.SignByKid)
}

type FfiConverterTypePackSignedMetadata struct{}

var FfiConverterTypePackSignedMetadataINSTANCE = FfiConverterTypePackSignedMetadata{}

func (c FfiConverterTypePackSignedMetadata) Lift(rb RustBufferI) PackSignedMetadata {
	return LiftFromRustBuffer[PackSignedMetadata](c, rb)
}

func (c FfiConverterTypePackSignedMetadata) Read(reader io.Reader) PackSignedMetadata {
	return PackSignedMetadata{
		FfiConverterStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypePackSignedMetadata) Lower(value PackSignedMetadata) RustBuffer {
	return LowerIntoRustBuffer[PackSignedMetadata](c, value)
}

func (c FfiConverterTypePackSignedMetadata) Write(writer io.Writer, value PackSignedMetadata) {
	FfiConverterStringINSTANCE.Write(writer, value.SignByKid)
}

type FfiDestroyerTypePackSignedMetadata struct{}

func (_ FfiDestroyerTypePackSignedMetadata) Destroy(value PackSignedMetadata) {
	value.Destroy()
}

type Secret struct {
	Id             string
	Type           SecretType
	SecretMaterial SecretMaterial
}

func (r *Secret) Destroy() {
	FfiDestroyerString{}.Destroy(r.Id)
	FfiDestroyerTypeSecretType{}.Destroy(r.Type)
	FfiDestroyerTypeSecretMaterial{}.Destroy(r.SecretMaterial)
}

type FfiConverterTypeSecret struct{}

var FfiConverterTypeSecretINSTANCE = FfiConverterTypeSecret{}

func (c FfiConverterTypeSecret) Lift(rb RustBufferI) Secret {
	return LiftFromRustBuffer[Secret](c, rb)
}

func (c FfiConverterTypeSecret) Read(reader io.Reader) Secret {
	return Secret{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterTypeSecretTypeINSTANCE.Read(reader),
		FfiConverterTypeSecretMaterialINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeSecret) Lower(value Secret) RustBuffer {
	return LowerIntoRustBuffer[Secret](c, value)
}

func (c FfiConverterTypeSecret) Write(writer io.Writer, value Secret) {
	FfiConverterStringINSTANCE.Write(writer, value.Id)
	FfiConverterTypeSecretTypeINSTANCE.Write(writer, value.Type)
	FfiConverterTypeSecretMaterialINSTANCE.Write(writer, value.SecretMaterial)
}

type FfiDestroyerTypeSecret struct{}

func (_ FfiDestroyerTypeSecret) Destroy(value Secret) {
	value.Destroy()
}

type Service struct {
	Id              string
	ServiceEndpoint ServiceKind
}

func (r *Service) Destroy() {
	FfiDestroyerString{}.Destroy(r.Id)
	FfiDestroyerTypeServiceKind{}.Destroy(r.ServiceEndpoint)
}

type FfiConverterTypeService struct{}

var FfiConverterTypeServiceINSTANCE = FfiConverterTypeService{}

func (c FfiConverterTypeService) Lift(rb RustBufferI) Service {
	return LiftFromRustBuffer[Service](c, rb)
}

func (c FfiConverterTypeService) Read(reader io.Reader) Service {
	return Service{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterTypeServiceKindINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeService) Lower(value Service) RustBuffer {
	return LowerIntoRustBuffer[Service](c, value)
}

func (c FfiConverterTypeService) Write(writer io.Writer, value Service) {
	FfiConverterStringINSTANCE.Write(writer, value.Id)
	FfiConverterTypeServiceKindINSTANCE.Write(writer, value.ServiceEndpoint)
}

type FfiDestroyerTypeService struct{}

func (_ FfiDestroyerTypeService) Destroy(value Service) {
	value.Destroy()
}

type UnpackMetadata struct {
	Encrypted          bool
	Authenticated      bool
	NonRepudiation     bool
	AnonymousSender    bool
	ReWrappedInForward bool
	EncryptedFromKid   *string
	EncryptedToKids    *[]string
	SignFrom           *string
	FromPriorIssuerKid *string
	EncAlgAuth         *AuthCryptAlg
	EncAlgAnon         *AnonCryptAlg
	SignAlg            *SignAlg
	SignedMessage      *string
	FromPrior          *FromPrior
}

func (r *UnpackMetadata) Destroy() {
	FfiDestroyerBool{}.Destroy(r.Encrypted)
	FfiDestroyerBool{}.Destroy(r.Authenticated)
	FfiDestroyerBool{}.Destroy(r.NonRepudiation)
	FfiDestroyerBool{}.Destroy(r.AnonymousSender)
	FfiDestroyerBool{}.Destroy(r.ReWrappedInForward)
	FfiDestroyerOptionalString{}.Destroy(r.EncryptedFromKid)
	FfiDestroyerOptionalSequenceString{}.Destroy(r.EncryptedToKids)
	FfiDestroyerOptionalString{}.Destroy(r.SignFrom)
	FfiDestroyerOptionalString{}.Destroy(r.FromPriorIssuerKid)
	FfiDestroyerOptionalTypeAuthCryptAlg{}.Destroy(r.EncAlgAuth)
	FfiDestroyerOptionalTypeAnonCryptAlg{}.Destroy(r.EncAlgAnon)
	FfiDestroyerOptionalTypeSignAlg{}.Destroy(r.SignAlg)
	FfiDestroyerOptionalString{}.Destroy(r.SignedMessage)
	FfiDestroyerOptionalTypeFromPrior{}.Destroy(r.FromPrior)
}

type FfiConverterTypeUnpackMetadata struct{}

var FfiConverterTypeUnpackMetadataINSTANCE = FfiConverterTypeUnpackMetadata{}

func (c FfiConverterTypeUnpackMetadata) Lift(rb RustBufferI) UnpackMetadata {
	return LiftFromRustBuffer[UnpackMetadata](c, rb)
}

func (c FfiConverterTypeUnpackMetadata) Read(reader io.Reader) UnpackMetadata {
	return UnpackMetadata{
		FfiConverterBoolINSTANCE.Read(reader),
		FfiConverterBoolINSTANCE.Read(reader),
		FfiConverterBoolINSTANCE.Read(reader),
		FfiConverterBoolINSTANCE.Read(reader),
		FfiConverterBoolINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterOptionalSequenceStringINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterOptionalTypeAuthCryptAlgINSTANCE.Read(reader),
		FfiConverterOptionalTypeAnonCryptAlgINSTANCE.Read(reader),
		FfiConverterOptionalTypeSignAlgINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterOptionalTypeFromPriorINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeUnpackMetadata) Lower(value UnpackMetadata) RustBuffer {
	return LowerIntoRustBuffer[UnpackMetadata](c, value)
}

func (c FfiConverterTypeUnpackMetadata) Write(writer io.Writer, value UnpackMetadata) {
	FfiConverterBoolINSTANCE.Write(writer, value.Encrypted)
	FfiConverterBoolINSTANCE.Write(writer, value.Authenticated)
	FfiConverterBoolINSTANCE.Write(writer, value.NonRepudiation)
	FfiConverterBoolINSTANCE.Write(writer, value.AnonymousSender)
	FfiConverterBoolINSTANCE.Write(writer, value.ReWrappedInForward)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.EncryptedFromKid)
	FfiConverterOptionalSequenceStringINSTANCE.Write(writer, value.EncryptedToKids)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.SignFrom)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.FromPriorIssuerKid)
	FfiConverterOptionalTypeAuthCryptAlgINSTANCE.Write(writer, value.EncAlgAuth)
	FfiConverterOptionalTypeAnonCryptAlgINSTANCE.Write(writer, value.EncAlgAnon)
	FfiConverterOptionalTypeSignAlgINSTANCE.Write(writer, value.SignAlg)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.SignedMessage)
	FfiConverterOptionalTypeFromPriorINSTANCE.Write(writer, value.FromPrior)
}

type FfiDestroyerTypeUnpackMetadata struct{}

func (_ FfiDestroyerTypeUnpackMetadata) Destroy(value UnpackMetadata) {
	value.Destroy()
}

type UnpackOptions struct {
	ExpectDecryptByAllKeys  bool
	UnwrapReWrappingForward bool
}

func (r *UnpackOptions) Destroy() {
	FfiDestroyerBool{}.Destroy(r.ExpectDecryptByAllKeys)
	FfiDestroyerBool{}.Destroy(r.UnwrapReWrappingForward)
}

type FfiConverterTypeUnpackOptions struct{}

var FfiConverterTypeUnpackOptionsINSTANCE = FfiConverterTypeUnpackOptions{}

func (c FfiConverterTypeUnpackOptions) Lift(rb RustBufferI) UnpackOptions {
	return LiftFromRustBuffer[UnpackOptions](c, rb)
}

func (c FfiConverterTypeUnpackOptions) Read(reader io.Reader) UnpackOptions {
	return UnpackOptions{
		FfiConverterBoolINSTANCE.Read(reader),
		FfiConverterBoolINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeUnpackOptions) Lower(value UnpackOptions) RustBuffer {
	return LowerIntoRustBuffer[UnpackOptions](c, value)
}

func (c FfiConverterTypeUnpackOptions) Write(writer io.Writer, value UnpackOptions) {
	FfiConverterBoolINSTANCE.Write(writer, value.ExpectDecryptByAllKeys)
	FfiConverterBoolINSTANCE.Write(writer, value.UnwrapReWrappingForward)
}

type FfiDestroyerTypeUnpackOptions struct{}

func (_ FfiDestroyerTypeUnpackOptions) Destroy(value UnpackOptions) {
	value.Destroy()
}

type VerificationMethod struct {
	Id                   string
	Type                 VerificationMethodType
	Controller           string
	VerificationMaterial VerificationMaterial
}

func (r *VerificationMethod) Destroy() {
	FfiDestroyerString{}.Destroy(r.Id)
	FfiDestroyerTypeVerificationMethodType{}.Destroy(r.Type)
	FfiDestroyerString{}.Destroy(r.Controller)
	FfiDestroyerTypeVerificationMaterial{}.Destroy(r.VerificationMaterial)
}

type FfiConverterTypeVerificationMethod struct{}

var FfiConverterTypeVerificationMethodINSTANCE = FfiConverterTypeVerificationMethod{}

func (c FfiConverterTypeVerificationMethod) Lift(rb RustBufferI) VerificationMethod {
	return LiftFromRustBuffer[VerificationMethod](c, rb)
}

func (c FfiConverterTypeVerificationMethod) Read(reader io.Reader) VerificationMethod {
	return VerificationMethod{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterTypeVerificationMethodTypeINSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterTypeVerificationMaterialINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeVerificationMethod) Lower(value VerificationMethod) RustBuffer {
	return LowerIntoRustBuffer[VerificationMethod](c, value)
}

func (c FfiConverterTypeVerificationMethod) Write(writer io.Writer, value VerificationMethod) {
	FfiConverterStringINSTANCE.Write(writer, value.Id)
	FfiConverterTypeVerificationMethodTypeINSTANCE.Write(writer, value.Type)
	FfiConverterStringINSTANCE.Write(writer, value.Controller)
	FfiConverterTypeVerificationMaterialINSTANCE.Write(writer, value.VerificationMaterial)
}

type FfiDestroyerTypeVerificationMethod struct{}

func (_ FfiDestroyerTypeVerificationMethod) Destroy(value VerificationMethod) {
	value.Destroy()
}

type AnonCryptAlg uint

const (
	AnonCryptAlgA256cbcHs512EcdhEsA256kw AnonCryptAlg = 1
	AnonCryptAlgXc20pEcdhEsA256kw        AnonCryptAlg = 2
	AnonCryptAlgA256gcmEcdhEsA256kw      AnonCryptAlg = 3
)

type FfiConverterTypeAnonCryptAlg struct{}

var FfiConverterTypeAnonCryptAlgINSTANCE = FfiConverterTypeAnonCryptAlg{}

func (c FfiConverterTypeAnonCryptAlg) Lift(rb RustBufferI) AnonCryptAlg {
	return LiftFromRustBuffer[AnonCryptAlg](c, rb)
}

func (c FfiConverterTypeAnonCryptAlg) Lower(value AnonCryptAlg) RustBuffer {
	return LowerIntoRustBuffer[AnonCryptAlg](c, value)
}
func (FfiConverterTypeAnonCryptAlg) Read(reader io.Reader) AnonCryptAlg {
	id := readInt32(reader)
	return AnonCryptAlg(id)
}

func (FfiConverterTypeAnonCryptAlg) Write(writer io.Writer, value AnonCryptAlg) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeAnonCryptAlg struct{}

func (_ FfiDestroyerTypeAnonCryptAlg) Destroy(value AnonCryptAlg) {
}

type AttachmentData interface {
	Destroy()
}
type AttachmentDataBase64 struct {
	Value Base64AttachmentData
}

func (e AttachmentDataBase64) Destroy() {
	FfiDestroyerTypeBase64AttachmentData{}.Destroy(e.Value)
}

type AttachmentDataJson struct {
	Value JsonAttachmentData
}

func (e AttachmentDataJson) Destroy() {
	FfiDestroyerTypeJsonAttachmentData{}.Destroy(e.Value)
}

type AttachmentDataLinks struct {
	Value LinksAttachmentData
}

func (e AttachmentDataLinks) Destroy() {
	FfiDestroyerTypeLinksAttachmentData{}.Destroy(e.Value)
}

type FfiConverterTypeAttachmentData struct{}

var FfiConverterTypeAttachmentDataINSTANCE = FfiConverterTypeAttachmentData{}

func (c FfiConverterTypeAttachmentData) Lift(rb RustBufferI) AttachmentData {
	return LiftFromRustBuffer[AttachmentData](c, rb)
}

func (c FfiConverterTypeAttachmentData) Lower(value AttachmentData) RustBuffer {
	return LowerIntoRustBuffer[AttachmentData](c, value)
}
func (FfiConverterTypeAttachmentData) Read(reader io.Reader) AttachmentData {
	id := readInt32(reader)
	switch id {
	case 1:
		return AttachmentDataBase64{
			FfiConverterTypeBase64AttachmentDataINSTANCE.Read(reader),
		}
	case 2:
		return AttachmentDataJson{
			FfiConverterTypeJsonAttachmentDataINSTANCE.Read(reader),
		}
	case 3:
		return AttachmentDataLinks{
			FfiConverterTypeLinksAttachmentDataINSTANCE.Read(reader),
		}
	default:
		panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeAttachmentData.Read()", id))
	}
}

func (FfiConverterTypeAttachmentData) Write(writer io.Writer, value AttachmentData) {
	switch variant_value := value.(type) {
	case AttachmentDataBase64:
		writeInt32(writer, 1)
		FfiConverterTypeBase64AttachmentDataINSTANCE.Write(writer, variant_value.Value)
	case AttachmentDataJson:
		writeInt32(writer, 2)
		FfiConverterTypeJsonAttachmentDataINSTANCE.Write(writer, variant_value.Value)
	case AttachmentDataLinks:
		writeInt32(writer, 3)
		FfiConverterTypeLinksAttachmentDataINSTANCE.Write(writer, variant_value.Value)
	default:
		_ = variant_value
		panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeAttachmentData.Write", value))
	}
}

type FfiDestroyerTypeAttachmentData struct{}

func (_ FfiDestroyerTypeAttachmentData) Destroy(value AttachmentData) {
	value.Destroy()
}

type AuthCryptAlg uint

const (
	AuthCryptAlgA256cbcHs512Ecdh1puA256kw AuthCryptAlg = 1
)

type FfiConverterTypeAuthCryptAlg struct{}

var FfiConverterTypeAuthCryptAlgINSTANCE = FfiConverterTypeAuthCryptAlg{}

func (c FfiConverterTypeAuthCryptAlg) Lift(rb RustBufferI) AuthCryptAlg {
	return LiftFromRustBuffer[AuthCryptAlg](c, rb)
}

func (c FfiConverterTypeAuthCryptAlg) Lower(value AuthCryptAlg) RustBuffer {
	return LowerIntoRustBuffer[AuthCryptAlg](c, value)
}
func (FfiConverterTypeAuthCryptAlg) Read(reader io.Reader) AuthCryptAlg {
	id := readInt32(reader)
	return AuthCryptAlg(id)
}

func (FfiConverterTypeAuthCryptAlg) Write(writer io.Writer, value AuthCryptAlg) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeAuthCryptAlg struct{}

func (_ FfiDestroyerTypeAuthCryptAlg) Destroy(value AuthCryptAlg) {
}

type ErrorCode uint

const (
	ErrorCodeSuccess ErrorCode = 1
	ErrorCodeError   ErrorCode = 2
)

type FfiConverterTypeErrorCode struct{}

var FfiConverterTypeErrorCodeINSTANCE = FfiConverterTypeErrorCode{}

func (c FfiConverterTypeErrorCode) Lift(rb RustBufferI) ErrorCode {
	return LiftFromRustBuffer[ErrorCode](c, rb)
}

func (c FfiConverterTypeErrorCode) Lower(value ErrorCode) RustBuffer {
	return LowerIntoRustBuffer[ErrorCode](c, value)
}
func (FfiConverterTypeErrorCode) Read(reader io.Reader) ErrorCode {
	id := readInt32(reader)
	return ErrorCode(id)
}

func (FfiConverterTypeErrorCode) Write(writer io.Writer, value ErrorCode) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeErrorCode struct{}

func (_ FfiDestroyerTypeErrorCode) Destroy(value ErrorCode) {
}

type ErrorKind struct {
	err error
}

func (err ErrorKind) Error() string {
	return fmt.Sprintf("ErrorKind: %s", err.err.Error())
}

func (err ErrorKind) Unwrap() error {
	return err.err
}

// Err* are used for checking error type with `errors.Is`
var ErrErrorKindDidNotResolved = fmt.Errorf("ErrorKindDidNotResolved")
var ErrErrorKindDidUrlNotFound = fmt.Errorf("ErrorKindDidUrlNotFound")
var ErrErrorKindSecretNotFound = fmt.Errorf("ErrorKindSecretNotFound")
var ErrErrorKindMalformed = fmt.Errorf("ErrorKindMalformed")
var ErrErrorKindIoError = fmt.Errorf("ErrorKindIoError")
var ErrErrorKindInvalidState = fmt.Errorf("ErrorKindInvalidState")
var ErrErrorKindNoCompatibleCrypto = fmt.Errorf("ErrorKindNoCompatibleCrypto")
var ErrErrorKindUnsupported = fmt.Errorf("ErrorKindUnsupported")
var ErrErrorKindIllegalArgument = fmt.Errorf("ErrorKindIllegalArgument")

// Variant structs
type ErrorKindDidNotResolved struct {
	message string
}

func NewErrorKindDidNotResolved() *ErrorKind {
	return &ErrorKind{
		err: &ErrorKindDidNotResolved{},
	}
}

func (err ErrorKindDidNotResolved) Error() string {
	return fmt.Sprintf("DidNotResolved: %s", err.message)
}

func (self ErrorKindDidNotResolved) Is(target error) bool {
	return target == ErrErrorKindDidNotResolved
}

type ErrorKindDidUrlNotFound struct {
	message string
}

func NewErrorKindDidUrlNotFound() *ErrorKind {
	return &ErrorKind{
		err: &ErrorKindDidUrlNotFound{},
	}
}

func (err ErrorKindDidUrlNotFound) Error() string {
	return fmt.Sprintf("DidUrlNotFound: %s", err.message)
}

func (self ErrorKindDidUrlNotFound) Is(target error) bool {
	return target == ErrErrorKindDidUrlNotFound
}

type ErrorKindSecretNotFound struct {
	message string
}

func NewErrorKindSecretNotFound() *ErrorKind {
	return &ErrorKind{
		err: &ErrorKindSecretNotFound{},
	}
}

func (err ErrorKindSecretNotFound) Error() string {
	return fmt.Sprintf("SecretNotFound: %s", err.message)
}

func (self ErrorKindSecretNotFound) Is(target error) bool {
	return target == ErrErrorKindSecretNotFound
}

type ErrorKindMalformed struct {
	message string
}

func NewErrorKindMalformed() *ErrorKind {
	return &ErrorKind{
		err: &ErrorKindMalformed{},
	}
}

func (err ErrorKindMalformed) Error() string {
	return fmt.Sprintf("Malformed: %s", err.message)
}

func (self ErrorKindMalformed) Is(target error) bool {
	return target == ErrErrorKindMalformed
}

type ErrorKindIoError struct {
	message string
}

func NewErrorKindIoError() *ErrorKind {
	return &ErrorKind{
		err: &ErrorKindIoError{},
	}
}

func (err ErrorKindIoError) Error() string {
	return fmt.Sprintf("IoError: %s", err.message)
}

func (self ErrorKindIoError) Is(target error) bool {
	return target == ErrErrorKindIoError
}

type ErrorKindInvalidState struct {
	message string
}

func NewErrorKindInvalidState() *ErrorKind {
	return &ErrorKind{
		err: &ErrorKindInvalidState{},
	}
}

func (err ErrorKindInvalidState) Error() string {
	return fmt.Sprintf("InvalidState: %s", err.message)
}

func (self ErrorKindInvalidState) Is(target error) bool {
	return target == ErrErrorKindInvalidState
}

type ErrorKindNoCompatibleCrypto struct {
	message string
}

func NewErrorKindNoCompatibleCrypto() *ErrorKind {
	return &ErrorKind{
		err: &ErrorKindNoCompatibleCrypto{},
	}
}

func (err ErrorKindNoCompatibleCrypto) Error() string {
	return fmt.Sprintf("NoCompatibleCrypto: %s", err.message)
}

func (self ErrorKindNoCompatibleCrypto) Is(target error) bool {
	return target == ErrErrorKindNoCompatibleCrypto
}

type ErrorKindUnsupported struct {
	message string
}

func NewErrorKindUnsupported() *ErrorKind {
	return &ErrorKind{
		err: &ErrorKindUnsupported{},
	}
}

func (err ErrorKindUnsupported) Error() string {
	return fmt.Sprintf("Unsupported: %s", err.message)
}

func (self ErrorKindUnsupported) Is(target error) bool {
	return target == ErrErrorKindUnsupported
}

type ErrorKindIllegalArgument struct {
	message string
}

func NewErrorKindIllegalArgument() *ErrorKind {
	return &ErrorKind{
		err: &ErrorKindIllegalArgument{},
	}
}

func (err ErrorKindIllegalArgument) Error() string {
	return fmt.Sprintf("IllegalArgument: %s", err.message)
}

func (self ErrorKindIllegalArgument) Is(target error) bool {
	return target == ErrErrorKindIllegalArgument
}

type FfiConverterTypeErrorKind struct{}

var FfiConverterTypeErrorKindINSTANCE = FfiConverterTypeErrorKind{}

func (c FfiConverterTypeErrorKind) Lift(eb RustBufferI) error {
	return LiftFromRustBuffer[error](c, eb)
}

func (c FfiConverterTypeErrorKind) Lower(value *ErrorKind) RustBuffer {
	return LowerIntoRustBuffer[*ErrorKind](c, value)
}

func (c FfiConverterTypeErrorKind) Read(reader io.Reader) error {
	errorID := readUint32(reader)

	message := FfiConverterStringINSTANCE.Read(reader)
	switch errorID {
	case 1:
		return &ErrorKind{&ErrorKindDidNotResolved{message}}
	case 2:
		return &ErrorKind{&ErrorKindDidUrlNotFound{message}}
	case 3:
		return &ErrorKind{&ErrorKindSecretNotFound{message}}
	case 4:
		return &ErrorKind{&ErrorKindMalformed{message}}
	case 5:
		return &ErrorKind{&ErrorKindIoError{message}}
	case 6:
		return &ErrorKind{&ErrorKindInvalidState{message}}
	case 7:
		return &ErrorKind{&ErrorKindNoCompatibleCrypto{message}}
	case 8:
		return &ErrorKind{&ErrorKindUnsupported{message}}
	case 9:
		return &ErrorKind{&ErrorKindIllegalArgument{message}}
	default:
		panic(fmt.Sprintf("Unknown error code %d in FfiConverterTypeErrorKind.Read()", errorID))
	}

}

func (c FfiConverterTypeErrorKind) Write(writer io.Writer, value *ErrorKind) {
	switch variantValue := value.err.(type) {
	case *ErrorKindDidNotResolved:
		writeInt32(writer, 1)
	case *ErrorKindDidUrlNotFound:
		writeInt32(writer, 2)
	case *ErrorKindSecretNotFound:
		writeInt32(writer, 3)
	case *ErrorKindMalformed:
		writeInt32(writer, 4)
	case *ErrorKindIoError:
		writeInt32(writer, 5)
	case *ErrorKindInvalidState:
		writeInt32(writer, 6)
	case *ErrorKindNoCompatibleCrypto:
		writeInt32(writer, 7)
	case *ErrorKindUnsupported:
		writeInt32(writer, 8)
	case *ErrorKindIllegalArgument:
		writeInt32(writer, 9)
	default:
		_ = variantValue
		panic(fmt.Sprintf("invalid error value `%v` in FfiConverterTypeErrorKind.Write", value))
	}
}

type SecretMaterial interface {
	Destroy()
}
type SecretMaterialJwk struct {
	PrivateKeyJwk JsonValue
}

func (e SecretMaterialJwk) Destroy() {
	FfiDestroyerTypeJsonValue{}.Destroy(e.PrivateKeyJwk)
}

type SecretMaterialMultibase struct {
	PrivateKeyMultibase string
}

func (e SecretMaterialMultibase) Destroy() {
	FfiDestroyerString{}.Destroy(e.PrivateKeyMultibase)
}

type SecretMaterialBase58 struct {
	PrivateKeyBase58 string
}

func (e SecretMaterialBase58) Destroy() {
	FfiDestroyerString{}.Destroy(e.PrivateKeyBase58)
}

type FfiConverterTypeSecretMaterial struct{}

var FfiConverterTypeSecretMaterialINSTANCE = FfiConverterTypeSecretMaterial{}

func (c FfiConverterTypeSecretMaterial) Lift(rb RustBufferI) SecretMaterial {
	return LiftFromRustBuffer[SecretMaterial](c, rb)
}

func (c FfiConverterTypeSecretMaterial) Lower(value SecretMaterial) RustBuffer {
	return LowerIntoRustBuffer[SecretMaterial](c, value)
}
func (FfiConverterTypeSecretMaterial) Read(reader io.Reader) SecretMaterial {
	id := readInt32(reader)
	switch id {
	case 1:
		return SecretMaterialJwk{
			FfiConverterTypeJsonValueINSTANCE.Read(reader),
		}
	case 2:
		return SecretMaterialMultibase{
			FfiConverterStringINSTANCE.Read(reader),
		}
	case 3:
		return SecretMaterialBase58{
			FfiConverterStringINSTANCE.Read(reader),
		}
	default:
		panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeSecretMaterial.Read()", id))
	}
}

func (FfiConverterTypeSecretMaterial) Write(writer io.Writer, value SecretMaterial) {
	switch variant_value := value.(type) {
	case SecretMaterialJwk:
		writeInt32(writer, 1)
		FfiConverterTypeJsonValueINSTANCE.Write(writer, variant_value.PrivateKeyJwk)
	case SecretMaterialMultibase:
		writeInt32(writer, 2)
		FfiConverterStringINSTANCE.Write(writer, variant_value.PrivateKeyMultibase)
	case SecretMaterialBase58:
		writeInt32(writer, 3)
		FfiConverterStringINSTANCE.Write(writer, variant_value.PrivateKeyBase58)
	default:
		_ = variant_value
		panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeSecretMaterial.Write", value))
	}
}

type FfiDestroyerTypeSecretMaterial struct{}

func (_ FfiDestroyerTypeSecretMaterial) Destroy(value SecretMaterial) {
	value.Destroy()
}

type SecretType uint

const (
	SecretTypeJsonWebKey2020                    SecretType = 1
	SecretTypeX25519KeyAgreementKey2019         SecretType = 2
	SecretTypeEd25519VerificationKey2018        SecretType = 3
	SecretTypeEcdsaSecp256k1VerificationKey2019 SecretType = 4
	SecretTypeX25519KeyAgreementKey2020         SecretType = 5
	SecretTypeEd25519VerificationKey2020        SecretType = 6
	SecretTypeOther                             SecretType = 7
)

type FfiConverterTypeSecretType struct{}

var FfiConverterTypeSecretTypeINSTANCE = FfiConverterTypeSecretType{}

func (c FfiConverterTypeSecretType) Lift(rb RustBufferI) SecretType {
	return LiftFromRustBuffer[SecretType](c, rb)
}

func (c FfiConverterTypeSecretType) Lower(value SecretType) RustBuffer {
	return LowerIntoRustBuffer[SecretType](c, value)
}
func (FfiConverterTypeSecretType) Read(reader io.Reader) SecretType {
	id := readInt32(reader)
	return SecretType(id)
}

func (FfiConverterTypeSecretType) Write(writer io.Writer, value SecretType) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeSecretType struct{}

func (_ FfiDestroyerTypeSecretType) Destroy(value SecretType) {
}

type ServiceKind interface {
	Destroy()
}
type ServiceKindDidCommMessaging struct {
	Value DidCommMessagingService
}

func (e ServiceKindDidCommMessaging) Destroy() {
	FfiDestroyerTypeDidCommMessagingService{}.Destroy(e.Value)
}

type ServiceKindOther struct {
	Value JsonValue
}

func (e ServiceKindOther) Destroy() {
	FfiDestroyerTypeJsonValue{}.Destroy(e.Value)
}

type FfiConverterTypeServiceKind struct{}

var FfiConverterTypeServiceKindINSTANCE = FfiConverterTypeServiceKind{}

func (c FfiConverterTypeServiceKind) Lift(rb RustBufferI) ServiceKind {
	return LiftFromRustBuffer[ServiceKind](c, rb)
}

func (c FfiConverterTypeServiceKind) Lower(value ServiceKind) RustBuffer {
	return LowerIntoRustBuffer[ServiceKind](c, value)
}
func (FfiConverterTypeServiceKind) Read(reader io.Reader) ServiceKind {
	id := readInt32(reader)
	switch id {
	case 1:
		return ServiceKindDidCommMessaging{
			FfiConverterTypeDIDCommMessagingServiceINSTANCE.Read(reader),
		}
	case 2:
		return ServiceKindOther{
			FfiConverterTypeJsonValueINSTANCE.Read(reader),
		}
	default:
		panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeServiceKind.Read()", id))
	}
}

func (FfiConverterTypeServiceKind) Write(writer io.Writer, value ServiceKind) {
	switch variant_value := value.(type) {
	case ServiceKindDidCommMessaging:
		writeInt32(writer, 1)
		FfiConverterTypeDIDCommMessagingServiceINSTANCE.Write(writer, variant_value.Value)
	case ServiceKindOther:
		writeInt32(writer, 2)
		FfiConverterTypeJsonValueINSTANCE.Write(writer, variant_value.Value)
	default:
		_ = variant_value
		panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeServiceKind.Write", value))
	}
}

type FfiDestroyerTypeServiceKind struct{}

func (_ FfiDestroyerTypeServiceKind) Destroy(value ServiceKind) {
	value.Destroy()
}

type SignAlg uint

const (
	SignAlgEdDsa  SignAlg = 1
	SignAlgEs256  SignAlg = 2
	SignAlgEs256k SignAlg = 3
)

type FfiConverterTypeSignAlg struct{}

var FfiConverterTypeSignAlgINSTANCE = FfiConverterTypeSignAlg{}

func (c FfiConverterTypeSignAlg) Lift(rb RustBufferI) SignAlg {
	return LiftFromRustBuffer[SignAlg](c, rb)
}

func (c FfiConverterTypeSignAlg) Lower(value SignAlg) RustBuffer {
	return LowerIntoRustBuffer[SignAlg](c, value)
}
func (FfiConverterTypeSignAlg) Read(reader io.Reader) SignAlg {
	id := readInt32(reader)
	return SignAlg(id)
}

func (FfiConverterTypeSignAlg) Write(writer io.Writer, value SignAlg) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeSignAlg struct{}

func (_ FfiDestroyerTypeSignAlg) Destroy(value SignAlg) {
}

type VerificationMaterial interface {
	Destroy()
}
type VerificationMaterialJwk struct {
	PublicKeyJwk JsonValue
}

func (e VerificationMaterialJwk) Destroy() {
	FfiDestroyerTypeJsonValue{}.Destroy(e.PublicKeyJwk)
}

type VerificationMaterialMultibase struct {
	PublicKeyMultibase string
}

func (e VerificationMaterialMultibase) Destroy() {
	FfiDestroyerString{}.Destroy(e.PublicKeyMultibase)
}

type VerificationMaterialBase58 struct {
	PublicKeyBase58 string
}

func (e VerificationMaterialBase58) Destroy() {
	FfiDestroyerString{}.Destroy(e.PublicKeyBase58)
}

type FfiConverterTypeVerificationMaterial struct{}

var FfiConverterTypeVerificationMaterialINSTANCE = FfiConverterTypeVerificationMaterial{}

func (c FfiConverterTypeVerificationMaterial) Lift(rb RustBufferI) VerificationMaterial {
	return LiftFromRustBuffer[VerificationMaterial](c, rb)
}

func (c FfiConverterTypeVerificationMaterial) Lower(value VerificationMaterial) RustBuffer {
	return LowerIntoRustBuffer[VerificationMaterial](c, value)
}
func (FfiConverterTypeVerificationMaterial) Read(reader io.Reader) VerificationMaterial {
	id := readInt32(reader)
	switch id {
	case 1:
		return VerificationMaterialJwk{
			FfiConverterTypeJsonValueINSTANCE.Read(reader),
		}
	case 2:
		return VerificationMaterialMultibase{
			FfiConverterStringINSTANCE.Read(reader),
		}
	case 3:
		return VerificationMaterialBase58{
			FfiConverterStringINSTANCE.Read(reader),
		}
	default:
		panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeVerificationMaterial.Read()", id))
	}
}

func (FfiConverterTypeVerificationMaterial) Write(writer io.Writer, value VerificationMaterial) {
	switch variant_value := value.(type) {
	case VerificationMaterialJwk:
		writeInt32(writer, 1)
		FfiConverterTypeJsonValueINSTANCE.Write(writer, variant_value.PublicKeyJwk)
	case VerificationMaterialMultibase:
		writeInt32(writer, 2)
		FfiConverterStringINSTANCE.Write(writer, variant_value.PublicKeyMultibase)
	case VerificationMaterialBase58:
		writeInt32(writer, 3)
		FfiConverterStringINSTANCE.Write(writer, variant_value.PublicKeyBase58)
	default:
		_ = variant_value
		panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeVerificationMaterial.Write", value))
	}
}

type FfiDestroyerTypeVerificationMaterial struct{}

func (_ FfiDestroyerTypeVerificationMaterial) Destroy(value VerificationMaterial) {
	value.Destroy()
}

type VerificationMethodType uint

const (
	VerificationMethodTypeJsonWebKey2020                    VerificationMethodType = 1
	VerificationMethodTypeX25519KeyAgreementKey2019         VerificationMethodType = 2
	VerificationMethodTypeEd25519VerificationKey2018        VerificationMethodType = 3
	VerificationMethodTypeEcdsaSecp256k1VerificationKey2019 VerificationMethodType = 4
	VerificationMethodTypeX25519KeyAgreementKey2020         VerificationMethodType = 5
	VerificationMethodTypeEd25519VerificationKey2020        VerificationMethodType = 6
	VerificationMethodTypeOther                             VerificationMethodType = 7
)

type FfiConverterTypeVerificationMethodType struct{}

var FfiConverterTypeVerificationMethodTypeINSTANCE = FfiConverterTypeVerificationMethodType{}

func (c FfiConverterTypeVerificationMethodType) Lift(rb RustBufferI) VerificationMethodType {
	return LiftFromRustBuffer[VerificationMethodType](c, rb)
}

func (c FfiConverterTypeVerificationMethodType) Lower(value VerificationMethodType) RustBuffer {
	return LowerIntoRustBuffer[VerificationMethodType](c, value)
}
func (FfiConverterTypeVerificationMethodType) Read(reader io.Reader) VerificationMethodType {
	id := readInt32(reader)
	return VerificationMethodType(id)
}

func (FfiConverterTypeVerificationMethodType) Write(writer io.Writer, value VerificationMethodType) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeVerificationMethodType struct{}

func (_ FfiDestroyerTypeVerificationMethodType) Destroy(value VerificationMethodType) {
}

type uniffiCallbackResult C.int32_t

const (
	uniffiIdxCallbackFree               uniffiCallbackResult = 0
	uniffiCallbackResultSuccess         uniffiCallbackResult = 0
	uniffiCallbackResultError           uniffiCallbackResult = 1
	uniffiCallbackUnexpectedResultError uniffiCallbackResult = 2
	uniffiCallbackCancelled             uniffiCallbackResult = 3
)

type concurrentHandleMap[T any] struct {
	leftMap       map[uint64]*T
	rightMap      map[*T]uint64
	currentHandle uint64
	lock          sync.RWMutex
}

func newConcurrentHandleMap[T any]() *concurrentHandleMap[T] {
	return &concurrentHandleMap[T]{
		leftMap:  map[uint64]*T{},
		rightMap: map[*T]uint64{},
	}
}

func (cm *concurrentHandleMap[T]) insert(obj *T) uint64 {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	if existingHandle, ok := cm.rightMap[obj]; ok {
		return existingHandle
	}
	cm.currentHandle = cm.currentHandle + 1
	cm.leftMap[cm.currentHandle] = obj
	cm.rightMap[obj] = cm.currentHandle
	return cm.currentHandle
}

func (cm *concurrentHandleMap[T]) remove(handle uint64) bool {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	if val, ok := cm.leftMap[handle]; ok {
		delete(cm.leftMap, handle)
		delete(cm.rightMap, val)
	}
	return false
}

func (cm *concurrentHandleMap[T]) tryGet(handle uint64) (*T, bool) {
	cm.lock.RLock()
	defer cm.lock.RUnlock()

	val, ok := cm.leftMap[handle]
	return val, ok
}

type FfiConverterCallbackInterface[CallbackInterface any] struct {
	handleMap *concurrentHandleMap[CallbackInterface]
}

func (c *FfiConverterCallbackInterface[CallbackInterface]) drop(handle uint64) RustBuffer {
	c.handleMap.remove(handle)
	return RustBuffer{}
}

func (c *FfiConverterCallbackInterface[CallbackInterface]) Lift(handle uint64) CallbackInterface {
	val, ok := c.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}
	return *val
}

func (c *FfiConverterCallbackInterface[CallbackInterface]) Read(reader io.Reader) CallbackInterface {
	return c.Lift(readUint64(reader))
}

func (c *FfiConverterCallbackInterface[CallbackInterface]) Lower(value CallbackInterface) C.uint64_t {
	return C.uint64_t(c.handleMap.insert(&value))
}

func (c *FfiConverterCallbackInterface[CallbackInterface]) Write(writer io.Writer, value CallbackInterface) {
	writeUint64(writer, uint64(c.Lower(value)))
}

// Declaration and FfiConverters for DidResolver Callback Interface
type DidResolver interface {
	Resolve(did string, cb *OnDidResolverResult) ErrorCode
}

// foreignCallbackCallbackInterfaceDIDResolver cannot be callable be a compiled function at a same time
type foreignCallbackCallbackInterfaceDIDResolver struct{}

//export didcomm_uniffi_cgo_DIDResolver
func didcomm_uniffi_cgo_DIDResolver(handle C.uint64_t, method C.int32_t, argsPtr *C.uint8_t, argsLen C.int32_t, outBuf *C.RustBuffer) C.int32_t {
	cb := FfiConverterCallbackInterfaceDIDResolverINSTANCE.Lift(uint64(handle))
	switch method {
	case 0:
		// 0 means Rust is done with the callback, and the callback
		// can be dropped by the foreign language.
		*outBuf = FfiConverterCallbackInterfaceDIDResolverINSTANCE.drop(uint64(handle))
		// See docs of ForeignCallback in `uniffi/src/ffi/foreigncallbacks.rs`
		return C.int32_t(uniffiIdxCallbackFree)

	case 1:
		var result uniffiCallbackResult
		args := unsafe.Slice((*byte)(argsPtr), argsLen)
		result = foreignCallbackCallbackInterfaceDIDResolver{}.InvokeResolve(cb, args, outBuf)
		return C.int32_t(result)

	default:
		// This should never happen, because an out of bounds method index won't
		// ever be used. Once we can catch errors, we should return an InternalException.
		// https://github.com/mozilla/uniffi-rs/issues/351
		return C.int32_t(uniffiCallbackUnexpectedResultError)
	}
}

func (foreignCallbackCallbackInterfaceDIDResolver) InvokeResolve(callback DidResolver, args []byte, outBuf *C.RustBuffer) uniffiCallbackResult {
	reader := bytes.NewReader(args)
	result := callback.Resolve(FfiConverterStringINSTANCE.Read(reader), FfiConverterOnDIDResolverResultINSTANCE.Read(reader))

	*outBuf = LowerIntoRustBuffer[ErrorCode](FfiConverterTypeErrorCodeINSTANCE, result)
	return uniffiCallbackResultSuccess
}

type FfiConverterCallbackInterfaceDIDResolver struct {
	FfiConverterCallbackInterface[DidResolver]
}

var FfiConverterCallbackInterfaceDIDResolverINSTANCE = &FfiConverterCallbackInterfaceDIDResolver{
	FfiConverterCallbackInterface: FfiConverterCallbackInterface[DidResolver]{
		handleMap: newConcurrentHandleMap[DidResolver](),
	},
}

// This is a static function because only 1 instance is supported for registering
func (c *FfiConverterCallbackInterfaceDIDResolver) register() {
	rustCall(func(status *C.RustCallStatus) int32 {
		C.uniffi_didcomm_uniffi_fn_init_callback_didresolver(C.ForeignCallback(C.didcomm_uniffi_cgo_DIDResolver), status)
		return 0
	})
}

type FfiDestroyerCallbackInterfaceDidResolver struct{}

func (FfiDestroyerCallbackInterfaceDidResolver) Destroy(value DidResolver) {
}

// Declaration and FfiConverters for OnFromPriorPackResult Callback Interface
type OnFromPriorPackResult interface {
	Success(frompriorjwt string, kid string)
	Error(err *ErrorKind, msg string)
}

// foreignCallbackCallbackInterfaceOnFromPriorPackResult cannot be callable be a compiled function at a same time
type foreignCallbackCallbackInterfaceOnFromPriorPackResult struct{}

//export didcomm_uniffi_cgo_OnFromPriorPackResult
func didcomm_uniffi_cgo_OnFromPriorPackResult(handle C.uint64_t, method C.int32_t, argsPtr *C.uint8_t, argsLen C.int32_t, outBuf *C.RustBuffer) C.int32_t {
	cb := FfiConverterCallbackInterfaceOnFromPriorPackResultINSTANCE.Lift(uint64(handle))
	switch method {
	case 0:
		// 0 means Rust is done with the callback, and the callback
		// can be dropped by the foreign language.
		*outBuf = FfiConverterCallbackInterfaceOnFromPriorPackResultINSTANCE.drop(uint64(handle))
		// See docs of ForeignCallback in `uniffi/src/ffi/foreigncallbacks.rs`
		return C.int32_t(uniffiIdxCallbackFree)

	case 1:
		var result uniffiCallbackResult
		args := unsafe.Slice((*byte)(argsPtr), argsLen)
		result = foreignCallbackCallbackInterfaceOnFromPriorPackResult{}.InvokeSuccess(cb, args, outBuf)
		return C.int32_t(result)
	case 2:
		var result uniffiCallbackResult
		args := unsafe.Slice((*byte)(argsPtr), argsLen)
		result = foreignCallbackCallbackInterfaceOnFromPriorPackResult{}.InvokeError(cb, args, outBuf)
		return C.int32_t(result)

	default:
		// This should never happen, because an out of bounds method index won't
		// ever be used. Once we can catch errors, we should return an InternalException.
		// https://github.com/mozilla/uniffi-rs/issues/351
		return C.int32_t(uniffiCallbackUnexpectedResultError)
	}
}

func (foreignCallbackCallbackInterfaceOnFromPriorPackResult) InvokeSuccess(callback OnFromPriorPackResult, args []byte, outBuf *C.RustBuffer) uniffiCallbackResult {
	reader := bytes.NewReader(args)
	callback.Success(FfiConverterStringINSTANCE.Read(reader), FfiConverterStringINSTANCE.Read(reader))

	return uniffiCallbackResultSuccess
}
func (foreignCallbackCallbackInterfaceOnFromPriorPackResult) InvokeError(callback OnFromPriorPackResult, args []byte, outBuf *C.RustBuffer) uniffiCallbackResult {
	reader := bytes.NewReader(args)
	err := FfiConverterTypeErrorKindINSTANCE.Read(reader)

	// Type assertion to convert err to *ErrorKind
	if errKind, ok := err.(*ErrorKind); ok {
		callback.Error(errKind, FfiConverterStringINSTANCE.Read(reader))
	} else {
		// Handle the case where the error is not of type *ErrorKind
		// You may want to log an error or handle it appropriately.
	}

	return uniffiCallbackResultSuccess
}

type FfiConverterCallbackInterfaceOnFromPriorPackResult struct {
	FfiConverterCallbackInterface[OnFromPriorPackResult]
}

var FfiConverterCallbackInterfaceOnFromPriorPackResultINSTANCE = &FfiConverterCallbackInterfaceOnFromPriorPackResult{
	FfiConverterCallbackInterface: FfiConverterCallbackInterface[OnFromPriorPackResult]{
		handleMap: newConcurrentHandleMap[OnFromPriorPackResult](),
	},
}

// This is a static function because only 1 instance is supported for registering
func (c *FfiConverterCallbackInterfaceOnFromPriorPackResult) register() {
	rustCall(func(status *C.RustCallStatus) int32 {
		C.uniffi_didcomm_uniffi_fn_init_callback_onfrompriorpackresult(C.ForeignCallback(C.didcomm_uniffi_cgo_OnFromPriorPackResult), status)
		return 0
	})
}

type FfiDestroyerCallbackInterfaceOnFromPriorPackResult struct{}

func (FfiDestroyerCallbackInterfaceOnFromPriorPackResult) Destroy(value OnFromPriorPackResult) {
}

// Declaration and FfiConverters for OnFromPriorUnpackResult Callback Interface
type OnFromPriorUnpackResult interface {
	Success(fromprior FromPrior, kid string)
	Error(err *ErrorKind, msg string)
}

// foreignCallbackCallbackInterfaceOnFromPriorUnpackResult cannot be callable be a compiled function at a same time
type foreignCallbackCallbackInterfaceOnFromPriorUnpackResult struct{}

//export didcomm_uniffi_cgo_OnFromPriorUnpackResult
func didcomm_uniffi_cgo_OnFromPriorUnpackResult(handle C.uint64_t, method C.int32_t, argsPtr *C.uint8_t, argsLen C.int32_t, outBuf *C.RustBuffer) C.int32_t {
	cb := FfiConverterCallbackInterfaceOnFromPriorUnpackResultINSTANCE.Lift(uint64(handle))
	switch method {
	case 0:
		// 0 means Rust is done with the callback, and the callback
		// can be dropped by the foreign language.
		*outBuf = FfiConverterCallbackInterfaceOnFromPriorUnpackResultINSTANCE.drop(uint64(handle))
		// See docs of ForeignCallback in `uniffi/src/ffi/foreigncallbacks.rs`
		return C.int32_t(uniffiIdxCallbackFree)

	case 1:
		var result uniffiCallbackResult
		args := unsafe.Slice((*byte)(argsPtr), argsLen)
		result = foreignCallbackCallbackInterfaceOnFromPriorUnpackResult{}.InvokeSuccess(cb, args, outBuf)
		return C.int32_t(result)
	case 2:
		var result uniffiCallbackResult
		args := unsafe.Slice((*byte)(argsPtr), argsLen)
		result = foreignCallbackCallbackInterfaceOnFromPriorUnpackResult{}.InvokeError(cb, args, outBuf)
		return C.int32_t(result)

	default:
		// This should never happen, because an out of bounds method index won't
		// ever be used. Once we can catch errors, we should return an InternalException.
		// https://github.com/mozilla/uniffi-rs/issues/351
		return C.int32_t(uniffiCallbackUnexpectedResultError)
	}
}

func (foreignCallbackCallbackInterfaceOnFromPriorUnpackResult) InvokeSuccess(callback OnFromPriorUnpackResult, args []byte, outBuf *C.RustBuffer) uniffiCallbackResult {
	reader := bytes.NewReader(args)
	callback.Success(FfiConverterTypeFromPriorINSTANCE.Read(reader), FfiConverterStringINSTANCE.Read(reader))

	return uniffiCallbackResultSuccess
}
func (foreignCallbackCallbackInterfaceOnFromPriorUnpackResult) InvokeError(callback OnFromPriorUnpackResult, args []byte, outBuf *C.RustBuffer) uniffiCallbackResult {
	reader := bytes.NewReader(args)
	err := FfiConverterTypeErrorKindINSTANCE.Read(reader)

	// Type assertion to convert err to *ErrorKind
	if errKind, ok := err.(*ErrorKind); ok {
		callback.Error(errKind, FfiConverterStringINSTANCE.Read(reader))
	} else {
		// Handle the case where the error is not of type *ErrorKind
		// You may want to log an error or handle it appropriately.
	}

	return uniffiCallbackResultSuccess
}

type FfiConverterCallbackInterfaceOnFromPriorUnpackResult struct {
	FfiConverterCallbackInterface[OnFromPriorUnpackResult]
}

var FfiConverterCallbackInterfaceOnFromPriorUnpackResultINSTANCE = &FfiConverterCallbackInterfaceOnFromPriorUnpackResult{
	FfiConverterCallbackInterface: FfiConverterCallbackInterface[OnFromPriorUnpackResult]{
		handleMap: newConcurrentHandleMap[OnFromPriorUnpackResult](),
	},
}

// This is a static function because only 1 instance is supported for registering
func (c *FfiConverterCallbackInterfaceOnFromPriorUnpackResult) register() {
	rustCall(func(status *C.RustCallStatus) int32 {
		C.uniffi_didcomm_uniffi_fn_init_callback_onfrompriorunpackresult(C.ForeignCallback(C.didcomm_uniffi_cgo_OnFromPriorUnpackResult), status)
		return 0
	})
}

type FfiDestroyerCallbackInterfaceOnFromPriorUnpackResult struct{}

func (FfiDestroyerCallbackInterfaceOnFromPriorUnpackResult) Destroy(value OnFromPriorUnpackResult) {
}

// Declaration and FfiConverters for OnPackEncryptedResult Callback Interface
type OnPackEncryptedResult interface {
	Success(result string, metadata PackEncryptedMetadata)
	Error(err *ErrorKind, msg string)
}

// foreignCallbackCallbackInterfaceOnPackEncryptedResult cannot be callable be a compiled function at a same time
type foreignCallbackCallbackInterfaceOnPackEncryptedResult struct{}

//export didcomm_uniffi_cgo_OnPackEncryptedResult
func didcomm_uniffi_cgo_OnPackEncryptedResult(handle C.uint64_t, method C.int32_t, argsPtr *C.uint8_t, argsLen C.int32_t, outBuf *C.RustBuffer) C.int32_t {
	cb := FfiConverterCallbackInterfaceOnPackEncryptedResultINSTANCE.Lift(uint64(handle))
	switch method {
	case 0:
		// 0 means Rust is done with the callback, and the callback
		// can be dropped by the foreign language.
		*outBuf = FfiConverterCallbackInterfaceOnPackEncryptedResultINSTANCE.drop(uint64(handle))
		// See docs of ForeignCallback in `uniffi/src/ffi/foreigncallbacks.rs`
		return C.int32_t(uniffiIdxCallbackFree)

	case 1:
		var result uniffiCallbackResult
		args := unsafe.Slice((*byte)(argsPtr), argsLen)
		result = foreignCallbackCallbackInterfaceOnPackEncryptedResult{}.InvokeSuccess(cb, args, outBuf)
		return C.int32_t(result)
	case 2:
		var result uniffiCallbackResult
		args := unsafe.Slice((*byte)(argsPtr), argsLen)
		result = foreignCallbackCallbackInterfaceOnPackEncryptedResult{}.InvokeError(cb, args, outBuf)
		return C.int32_t(result)

	default:
		// This should never happen, because an out of bounds method index won't
		// ever be used. Once we can catch errors, we should return an InternalException.
		// https://github.com/mozilla/uniffi-rs/issues/351
		return C.int32_t(uniffiCallbackUnexpectedResultError)
	}
}

func (foreignCallbackCallbackInterfaceOnPackEncryptedResult) InvokeSuccess(callback OnPackEncryptedResult, args []byte, outBuf *C.RustBuffer) uniffiCallbackResult {
	reader := bytes.NewReader(args)
	callback.Success(FfiConverterStringINSTANCE.Read(reader), FfiConverterTypePackEncryptedMetadataINSTANCE.Read(reader))

	return uniffiCallbackResultSuccess
}
func (foreignCallbackCallbackInterfaceOnPackEncryptedResult) InvokeError(callback OnPackEncryptedResult, args []byte, outBuf *C.RustBuffer) uniffiCallbackResult {
	reader := bytes.NewReader(args)
	err := FfiConverterTypeErrorKindINSTANCE.Read(reader)

	// Type assertion to convert err to *ErrorKind
	if errKind, ok := err.(*ErrorKind); ok {
		callback.Error(errKind, FfiConverterStringINSTANCE.Read(reader))
	} else {
		// Handle the case where the error is not of type *ErrorKind
		// You may want to log an error or handle it appropriately.
	}

	return uniffiCallbackResultSuccess
}

type FfiConverterCallbackInterfaceOnPackEncryptedResult struct {
	FfiConverterCallbackInterface[OnPackEncryptedResult]
}

var FfiConverterCallbackInterfaceOnPackEncryptedResultINSTANCE = &FfiConverterCallbackInterfaceOnPackEncryptedResult{
	FfiConverterCallbackInterface: FfiConverterCallbackInterface[OnPackEncryptedResult]{
		handleMap: newConcurrentHandleMap[OnPackEncryptedResult](),
	},
}

// This is a static function because only 1 instance is supported for registering
func (c *FfiConverterCallbackInterfaceOnPackEncryptedResult) register() {
	rustCall(func(status *C.RustCallStatus) int32 {
		C.uniffi_didcomm_uniffi_fn_init_callback_onpackencryptedresult(C.ForeignCallback(C.didcomm_uniffi_cgo_OnPackEncryptedResult), status)
		return 0
	})
}

type FfiDestroyerCallbackInterfaceOnPackEncryptedResult struct{}

func (FfiDestroyerCallbackInterfaceOnPackEncryptedResult) Destroy(value OnPackEncryptedResult) {
}

// Declaration and FfiConverters for OnPackPlaintextResult Callback Interface
type OnPackPlaintextResult interface {
	Success(result string)
	Error(err *ErrorKind, msg string)
}

// foreignCallbackCallbackInterfaceOnPackPlaintextResult cannot be callable be a compiled function at a same time
type foreignCallbackCallbackInterfaceOnPackPlaintextResult struct{}

//export didcomm_uniffi_cgo_OnPackPlaintextResult
func didcomm_uniffi_cgo_OnPackPlaintextResult(handle C.uint64_t, method C.int32_t, argsPtr *C.uint8_t, argsLen C.int32_t, outBuf *C.RustBuffer) C.int32_t {
	cb := FfiConverterCallbackInterfaceOnPackPlaintextResultINSTANCE.Lift(uint64(handle))
	switch method {
	case 0:
		// 0 means Rust is done with the callback, and the callback
		// can be dropped by the foreign language.
		*outBuf = FfiConverterCallbackInterfaceOnPackPlaintextResultINSTANCE.drop(uint64(handle))
		// See docs of ForeignCallback in `uniffi/src/ffi/foreigncallbacks.rs`
		return C.int32_t(uniffiIdxCallbackFree)

	case 1:
		var result uniffiCallbackResult
		args := unsafe.Slice((*byte)(argsPtr), argsLen)
		result = foreignCallbackCallbackInterfaceOnPackPlaintextResult{}.InvokeSuccess(cb, args, outBuf)
		return C.int32_t(result)
	case 2:
		var result uniffiCallbackResult
		args := unsafe.Slice((*byte)(argsPtr), argsLen)
		result = foreignCallbackCallbackInterfaceOnPackPlaintextResult{}.InvokeError(cb, args, outBuf)
		return C.int32_t(result)

	default:
		// This should never happen, because an out of bounds method index won't
		// ever be used. Once we can catch errors, we should return an InternalException.
		// https://github.com/mozilla/uniffi-rs/issues/351
		return C.int32_t(uniffiCallbackUnexpectedResultError)
	}
}

func (foreignCallbackCallbackInterfaceOnPackPlaintextResult) InvokeSuccess(callback OnPackPlaintextResult, args []byte, outBuf *C.RustBuffer) uniffiCallbackResult {
	reader := bytes.NewReader(args)
	callback.Success(FfiConverterStringINSTANCE.Read(reader))

	return uniffiCallbackResultSuccess
}
func (foreignCallbackCallbackInterfaceOnPackPlaintextResult) InvokeError(callback OnPackPlaintextResult, args []byte, outBuf *C.RustBuffer) uniffiCallbackResult {
	reader := bytes.NewReader(args)
	err := FfiConverterTypeErrorKindINSTANCE.Read(reader)

	// Type assertion to convert err to *ErrorKind
	if errKind, ok := err.(*ErrorKind); ok {
		callback.Error(errKind, FfiConverterStringINSTANCE.Read(reader))
	} else {
		// Handle the case where the error is not of type *ErrorKind
		// You may want to log an error or handle it appropriately.
	}

	return uniffiCallbackResultSuccess
}

type FfiConverterCallbackInterfaceOnPackPlaintextResult struct {
	FfiConverterCallbackInterface[OnPackPlaintextResult]
}

var FfiConverterCallbackInterfaceOnPackPlaintextResultINSTANCE = &FfiConverterCallbackInterfaceOnPackPlaintextResult{
	FfiConverterCallbackInterface: FfiConverterCallbackInterface[OnPackPlaintextResult]{
		handleMap: newConcurrentHandleMap[OnPackPlaintextResult](),
	},
}

// This is a static function because only 1 instance is supported for registering
func (c *FfiConverterCallbackInterfaceOnPackPlaintextResult) register() {
	rustCall(func(status *C.RustCallStatus) int32 {
		C.uniffi_didcomm_uniffi_fn_init_callback_onpackplaintextresult(C.ForeignCallback(C.didcomm_uniffi_cgo_OnPackPlaintextResult), status)
		return 0
	})
}

type FfiDestroyerCallbackInterfaceOnPackPlaintextResult struct{}

func (FfiDestroyerCallbackInterfaceOnPackPlaintextResult) Destroy(value OnPackPlaintextResult) {
}

// Declaration and FfiConverters for OnPackSignedResult Callback Interface
type OnPackSignedResult interface {
	Success(result string, metadata PackSignedMetadata)
	Error(err *ErrorKind, msg string)
}

// foreignCallbackCallbackInterfaceOnPackSignedResult cannot be callable be a compiled function at a same time
type foreignCallbackCallbackInterfaceOnPackSignedResult struct{}

//export didcomm_uniffi_cgo_OnPackSignedResult
func didcomm_uniffi_cgo_OnPackSignedResult(handle C.uint64_t, method C.int32_t, argsPtr *C.uint8_t, argsLen C.int32_t, outBuf *C.RustBuffer) C.int32_t {
	cb := FfiConverterCallbackInterfaceOnPackSignedResultINSTANCE.Lift(uint64(handle))
	switch method {
	case 0:
		// 0 means Rust is done with the callback, and the callback
		// can be dropped by the foreign language.
		*outBuf = FfiConverterCallbackInterfaceOnPackSignedResultINSTANCE.drop(uint64(handle))
		// See docs of ForeignCallback in `uniffi/src/ffi/foreigncallbacks.rs`
		return C.int32_t(uniffiIdxCallbackFree)

	case 1:
		var result uniffiCallbackResult
		args := unsafe.Slice((*byte)(argsPtr), argsLen)
		result = foreignCallbackCallbackInterfaceOnPackSignedResult{}.InvokeSuccess(cb, args, outBuf)
		return C.int32_t(result)
	case 2:
		var result uniffiCallbackResult
		args := unsafe.Slice((*byte)(argsPtr), argsLen)
		result = foreignCallbackCallbackInterfaceOnPackSignedResult{}.InvokeError(cb, args, outBuf)
		return C.int32_t(result)

	default:
		// This should never happen, because an out of bounds method index won't
		// ever be used. Once we can catch errors, we should return an InternalException.
		// https://github.com/mozilla/uniffi-rs/issues/351
		return C.int32_t(uniffiCallbackUnexpectedResultError)
	}
}

func (foreignCallbackCallbackInterfaceOnPackSignedResult) InvokeSuccess(callback OnPackSignedResult, args []byte, outBuf *C.RustBuffer) uniffiCallbackResult {
	reader := bytes.NewReader(args)
	callback.Success(FfiConverterStringINSTANCE.Read(reader), FfiConverterTypePackSignedMetadataINSTANCE.Read(reader))

	return uniffiCallbackResultSuccess
}
func (foreignCallbackCallbackInterfaceOnPackSignedResult) InvokeError(callback OnPackSignedResult, args []byte, outBuf *C.RustBuffer) uniffiCallbackResult {
	reader := bytes.NewReader(args)
	err := FfiConverterTypeErrorKindINSTANCE.Read(reader)

	// Type assertion to convert err to *ErrorKind
	if errKind, ok := err.(*ErrorKind); ok {
		callback.Error(errKind, FfiConverterStringINSTANCE.Read(reader))
	} else {
		// Handle the case where the error is not of type *ErrorKind
		// You may want to log an error or handle it appropriately.
	}

	return uniffiCallbackResultSuccess
}

type FfiConverterCallbackInterfaceOnPackSignedResult struct {
	FfiConverterCallbackInterface[OnPackSignedResult]
}

var FfiConverterCallbackInterfaceOnPackSignedResultINSTANCE = &FfiConverterCallbackInterfaceOnPackSignedResult{
	FfiConverterCallbackInterface: FfiConverterCallbackInterface[OnPackSignedResult]{
		handleMap: newConcurrentHandleMap[OnPackSignedResult](),
	},
}

// This is a static function because only 1 instance is supported for registering
func (c *FfiConverterCallbackInterfaceOnPackSignedResult) register() {
	rustCall(func(status *C.RustCallStatus) int32 {
		C.uniffi_didcomm_uniffi_fn_init_callback_onpacksignedresult(C.ForeignCallback(C.didcomm_uniffi_cgo_OnPackSignedResult), status)
		return 0
	})
}

type FfiDestroyerCallbackInterfaceOnPackSignedResult struct{}

func (FfiDestroyerCallbackInterfaceOnPackSignedResult) Destroy(value OnPackSignedResult) {
}

// Declaration and FfiConverters for OnUnpackResult Callback Interface
type OnUnpackResult interface {
	Success(result Message, metadata UnpackMetadata)
	Error(err *ErrorKind, msg string)
}

// foreignCallbackCallbackInterfaceOnUnpackResult cannot be callable be a compiled function at a same time
type foreignCallbackCallbackInterfaceOnUnpackResult struct{}

//export didcomm_uniffi_cgo_OnUnpackResult
func didcomm_uniffi_cgo_OnUnpackResult(handle C.uint64_t, method C.int32_t, argsPtr *C.uint8_t, argsLen C.int32_t, outBuf *C.RustBuffer) C.int32_t {
	cb := FfiConverterCallbackInterfaceOnUnpackResultINSTANCE.Lift(uint64(handle))
	switch method {
	case 0:
		// 0 means Rust is done with the callback, and the callback
		// can be dropped by the foreign language.
		*outBuf = FfiConverterCallbackInterfaceOnUnpackResultINSTANCE.drop(uint64(handle))
		// See docs of ForeignCallback in `uniffi/src/ffi/foreigncallbacks.rs`
		return C.int32_t(uniffiIdxCallbackFree)

	case 1:
		var result uniffiCallbackResult
		args := unsafe.Slice((*byte)(argsPtr), argsLen)
		result = foreignCallbackCallbackInterfaceOnUnpackResult{}.InvokeSuccess(cb, args, outBuf)
		return C.int32_t(result)
	case 2:
		var result uniffiCallbackResult
		args := unsafe.Slice((*byte)(argsPtr), argsLen)
		result = foreignCallbackCallbackInterfaceOnUnpackResult{}.InvokeError(cb, args, outBuf)
		return C.int32_t(result)

	default:
		// This should never happen, because an out of bounds method index won't
		// ever be used. Once we can catch errors, we should return an InternalException.
		// https://github.com/mozilla/uniffi-rs/issues/351
		return C.int32_t(uniffiCallbackUnexpectedResultError)
	}
}

func (foreignCallbackCallbackInterfaceOnUnpackResult) InvokeSuccess(callback OnUnpackResult, args []byte, outBuf *C.RustBuffer) uniffiCallbackResult {
	reader := bytes.NewReader(args)
	callback.Success(FfiConverterTypeMessageINSTANCE.Read(reader), FfiConverterTypeUnpackMetadataINSTANCE.Read(reader))

	return uniffiCallbackResultSuccess
}
func (foreignCallbackCallbackInterfaceOnUnpackResult) InvokeError(callback OnUnpackResult, args []byte, outBuf *C.RustBuffer) uniffiCallbackResult {
	reader := bytes.NewReader(args)
	err := FfiConverterTypeErrorKindINSTANCE.Read(reader)

	// Type assertion to convert err to *ErrorKind
	if errKind, ok := err.(*ErrorKind); ok {
		callback.Error(errKind, FfiConverterStringINSTANCE.Read(reader))
	} else {
		// Handle the case where the error is not of type *ErrorKind
		// You may want to log an error or handle it appropriately.
	}

	return uniffiCallbackResultSuccess
}

type FfiConverterCallbackInterfaceOnUnpackResult struct {
	FfiConverterCallbackInterface[OnUnpackResult]
}

var FfiConverterCallbackInterfaceOnUnpackResultINSTANCE = &FfiConverterCallbackInterfaceOnUnpackResult{
	FfiConverterCallbackInterface: FfiConverterCallbackInterface[OnUnpackResult]{
		handleMap: newConcurrentHandleMap[OnUnpackResult](),
	},
}

// This is a static function because only 1 instance is supported for registering
func (c *FfiConverterCallbackInterfaceOnUnpackResult) register() {
	rustCall(func(status *C.RustCallStatus) int32 {
		C.uniffi_didcomm_uniffi_fn_init_callback_onunpackresult(C.ForeignCallback(C.didcomm_uniffi_cgo_OnUnpackResult), status)
		return 0
	})
}

type FfiDestroyerCallbackInterfaceOnUnpackResult struct{}

func (FfiDestroyerCallbackInterfaceOnUnpackResult) Destroy(value OnUnpackResult) {
}

// Declaration and FfiConverters for OnWrapInForwardResult Callback Interface
type OnWrapInForwardResult interface {
	Success(result string)
	Error(err *ErrorKind, msg string)
}

// foreignCallbackCallbackInterfaceOnWrapInForwardResult cannot be callable be a compiled function at a same time
type foreignCallbackCallbackInterfaceOnWrapInForwardResult struct{}

//export didcomm_uniffi_cgo_OnWrapInForwardResult
func didcomm_uniffi_cgo_OnWrapInForwardResult(handle C.uint64_t, method C.int32_t, argsPtr *C.uint8_t, argsLen C.int32_t, outBuf *C.RustBuffer) C.int32_t {
	cb := FfiConverterCallbackInterfaceOnWrapInForwardResultINSTANCE.Lift(uint64(handle))
	switch method {
	case 0:
		// 0 means Rust is done with the callback, and the callback
		// can be dropped by the foreign language.
		*outBuf = FfiConverterCallbackInterfaceOnWrapInForwardResultINSTANCE.drop(uint64(handle))
		// See docs of ForeignCallback in `uniffi/src/ffi/foreigncallbacks.rs`
		return C.int32_t(uniffiIdxCallbackFree)

	case 1:
		var result uniffiCallbackResult
		args := unsafe.Slice((*byte)(argsPtr), argsLen)
		result = foreignCallbackCallbackInterfaceOnWrapInForwardResult{}.InvokeSuccess(cb, args, outBuf)
		return C.int32_t(result)
	case 2:
		var result uniffiCallbackResult
		args := unsafe.Slice((*byte)(argsPtr), argsLen)
		result = foreignCallbackCallbackInterfaceOnWrapInForwardResult{}.InvokeError(cb, args, outBuf)
		return C.int32_t(result)

	default:
		// This should never happen, because an out of bounds method index won't
		// ever be used. Once we can catch errors, we should return an InternalException.
		// https://github.com/mozilla/uniffi-rs/issues/351
		return C.int32_t(uniffiCallbackUnexpectedResultError)
	}
}

func (foreignCallbackCallbackInterfaceOnWrapInForwardResult) InvokeSuccess(callback OnWrapInForwardResult, args []byte, outBuf *C.RustBuffer) uniffiCallbackResult {
	reader := bytes.NewReader(args)
	callback.Success(FfiConverterStringINSTANCE.Read(reader))

	return uniffiCallbackResultSuccess
}
func (foreignCallbackCallbackInterfaceOnWrapInForwardResult) InvokeError(callback OnWrapInForwardResult, args []byte, outBuf *C.RustBuffer) uniffiCallbackResult {
	reader := bytes.NewReader(args)
	err := FfiConverterTypeErrorKindINSTANCE.Read(reader)

	// Type assertion to convert err to *ErrorKind
	if errKind, ok := err.(*ErrorKind); ok {
		callback.Error(errKind, FfiConverterStringINSTANCE.Read(reader))
	} else {
		// Handle the case where the error is not of type *ErrorKind
		// You may want to log an error or handle it appropriately.
	}

	return uniffiCallbackResultSuccess
}

type FfiConverterCallbackInterfaceOnWrapInForwardResult struct {
	FfiConverterCallbackInterface[OnWrapInForwardResult]
}

var FfiConverterCallbackInterfaceOnWrapInForwardResultINSTANCE = &FfiConverterCallbackInterfaceOnWrapInForwardResult{
	FfiConverterCallbackInterface: FfiConverterCallbackInterface[OnWrapInForwardResult]{
		handleMap: newConcurrentHandleMap[OnWrapInForwardResult](),
	},
}

// This is a static function because only 1 instance is supported for registering
func (c *FfiConverterCallbackInterfaceOnWrapInForwardResult) register() {
	rustCall(func(status *C.RustCallStatus) int32 {
		C.uniffi_didcomm_uniffi_fn_init_callback_onwrapinforwardresult(C.ForeignCallback(C.didcomm_uniffi_cgo_OnWrapInForwardResult), status)
		return 0
	})
}

type FfiDestroyerCallbackInterfaceOnWrapInForwardResult struct{}

func (FfiDestroyerCallbackInterfaceOnWrapInForwardResult) Destroy(value OnWrapInForwardResult) {
}

// Declaration and FfiConverters for SecretsResolver Callback Interface
type SecretsResolver interface {
	GetSecret(secretid string, cb *OnGetSecretResult) ErrorCode
	FindSecrets(secretids []string, cb *OnFindSecretsResult) ErrorCode
}

// foreignCallbackCallbackInterfaceSecretsResolver cannot be callable be a compiled function at a same time
type foreignCallbackCallbackInterfaceSecretsResolver struct{}

//export didcomm_uniffi_cgo_SecretsResolver
func didcomm_uniffi_cgo_SecretsResolver(handle C.uint64_t, method C.int32_t, argsPtr *C.uint8_t, argsLen C.int32_t, outBuf *C.RustBuffer) C.int32_t {
	cb := FfiConverterCallbackInterfaceSecretsResolverINSTANCE.Lift(uint64(handle))
	switch method {
	case 0:
		// 0 means Rust is done with the callback, and the callback
		// can be dropped by the foreign language.
		*outBuf = FfiConverterCallbackInterfaceSecretsResolverINSTANCE.drop(uint64(handle))
		// See docs of ForeignCallback in `uniffi/src/ffi/foreigncallbacks.rs`
		return C.int32_t(uniffiIdxCallbackFree)

	case 1:
		var result uniffiCallbackResult
		args := unsafe.Slice((*byte)(argsPtr), argsLen)
		result = foreignCallbackCallbackInterfaceSecretsResolver{}.InvokeGetSecret(cb, args, outBuf)
		return C.int32_t(result)
	case 2:
		var result uniffiCallbackResult
		args := unsafe.Slice((*byte)(argsPtr), argsLen)
		result = foreignCallbackCallbackInterfaceSecretsResolver{}.InvokeFindSecrets(cb, args, outBuf)
		return C.int32_t(result)

	default:
		// This should never happen, because an out of bounds method index won't
		// ever be used. Once we can catch errors, we should return an InternalException.
		// https://github.com/mozilla/uniffi-rs/issues/351
		return C.int32_t(uniffiCallbackUnexpectedResultError)
	}
}

func (foreignCallbackCallbackInterfaceSecretsResolver) InvokeGetSecret(callback SecretsResolver, args []byte, outBuf *C.RustBuffer) uniffiCallbackResult {
	reader := bytes.NewReader(args)
	result := callback.GetSecret(FfiConverterStringINSTANCE.Read(reader), FfiConverterOnGetSecretResultINSTANCE.Read(reader))

	*outBuf = LowerIntoRustBuffer[ErrorCode](FfiConverterTypeErrorCodeINSTANCE, result)
	return uniffiCallbackResultSuccess
}
func (foreignCallbackCallbackInterfaceSecretsResolver) InvokeFindSecrets(callback SecretsResolver, args []byte, outBuf *C.RustBuffer) uniffiCallbackResult {
	reader := bytes.NewReader(args)
	result := callback.FindSecrets(FfiConverterSequenceStringINSTANCE.Read(reader), FfiConverterOnFindSecretsResultINSTANCE.Read(reader))

	*outBuf = LowerIntoRustBuffer[ErrorCode](FfiConverterTypeErrorCodeINSTANCE, result)
	return uniffiCallbackResultSuccess
}

type FfiConverterCallbackInterfaceSecretsResolver struct {
	FfiConverterCallbackInterface[SecretsResolver]
}

var FfiConverterCallbackInterfaceSecretsResolverINSTANCE = &FfiConverterCallbackInterfaceSecretsResolver{
	FfiConverterCallbackInterface: FfiConverterCallbackInterface[SecretsResolver]{
		handleMap: newConcurrentHandleMap[SecretsResolver](),
	},
}

// This is a static function because only 1 instance is supported for registering
func (c *FfiConverterCallbackInterfaceSecretsResolver) register() {
	rustCall(func(status *C.RustCallStatus) int32 {
		C.uniffi_didcomm_uniffi_fn_init_callback_secretsresolver(C.ForeignCallback(C.didcomm_uniffi_cgo_SecretsResolver), status)
		return 0
	})
}

type FfiDestroyerCallbackInterfaceSecretsResolver struct{}

func (FfiDestroyerCallbackInterfaceSecretsResolver) Destroy(value SecretsResolver) {
}

type FfiConverterOptionalUint64 struct{}

var FfiConverterOptionalUint64INSTANCE = FfiConverterOptionalUint64{}

func (c FfiConverterOptionalUint64) Lift(rb RustBufferI) *uint64 {
	return LiftFromRustBuffer[*uint64](c, rb)
}

func (_ FfiConverterOptionalUint64) Read(reader io.Reader) *uint64 {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterUint64INSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalUint64) Lower(value *uint64) RustBuffer {
	return LowerIntoRustBuffer[*uint64](c, value)
}

func (_ FfiConverterOptionalUint64) Write(writer io.Writer, value *uint64) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterUint64INSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalUint64 struct{}

func (_ FfiDestroyerOptionalUint64) Destroy(value *uint64) {
	if value != nil {
		FfiDestroyerUint64{}.Destroy(*value)
	}
}

type FfiConverterOptionalString struct{}

var FfiConverterOptionalStringINSTANCE = FfiConverterOptionalString{}

func (c FfiConverterOptionalString) Lift(rb RustBufferI) *string {
	return LiftFromRustBuffer[*string](c, rb)
}

func (_ FfiConverterOptionalString) Read(reader io.Reader) *string {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterStringINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalString) Lower(value *string) RustBuffer {
	return LowerIntoRustBuffer[*string](c, value)
}

func (_ FfiConverterOptionalString) Write(writer io.Writer, value *string) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterStringINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalString struct{}

func (_ FfiDestroyerOptionalString) Destroy(value *string) {
	if value != nil {
		FfiDestroyerString{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypeDIDDoc struct{}

var FfiConverterOptionalTypeDIDDocINSTANCE = FfiConverterOptionalTypeDIDDoc{}

func (c FfiConverterOptionalTypeDIDDoc) Lift(rb RustBufferI) *DidDoc {
	return LiftFromRustBuffer[*DidDoc](c, rb)
}

func (_ FfiConverterOptionalTypeDIDDoc) Read(reader io.Reader) *DidDoc {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeDIDDocINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeDIDDoc) Lower(value *DidDoc) RustBuffer {
	return LowerIntoRustBuffer[*DidDoc](c, value)
}

func (_ FfiConverterOptionalTypeDIDDoc) Write(writer io.Writer, value *DidDoc) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeDIDDocINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeDidDoc struct{}

func (_ FfiDestroyerOptionalTypeDidDoc) Destroy(value *DidDoc) {
	if value != nil {
		FfiDestroyerTypeDidDoc{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypeFromPrior struct{}

var FfiConverterOptionalTypeFromPriorINSTANCE = FfiConverterOptionalTypeFromPrior{}

func (c FfiConverterOptionalTypeFromPrior) Lift(rb RustBufferI) *FromPrior {
	return LiftFromRustBuffer[*FromPrior](c, rb)
}

func (_ FfiConverterOptionalTypeFromPrior) Read(reader io.Reader) *FromPrior {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeFromPriorINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeFromPrior) Lower(value *FromPrior) RustBuffer {
	return LowerIntoRustBuffer[*FromPrior](c, value)
}

func (_ FfiConverterOptionalTypeFromPrior) Write(writer io.Writer, value *FromPrior) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeFromPriorINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeFromPrior struct{}

func (_ FfiDestroyerOptionalTypeFromPrior) Destroy(value *FromPrior) {
	if value != nil {
		FfiDestroyerTypeFromPrior{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypeMessagingServiceMetadata struct{}

var FfiConverterOptionalTypeMessagingServiceMetadataINSTANCE = FfiConverterOptionalTypeMessagingServiceMetadata{}

func (c FfiConverterOptionalTypeMessagingServiceMetadata) Lift(rb RustBufferI) *MessagingServiceMetadata {
	return LiftFromRustBuffer[*MessagingServiceMetadata](c, rb)
}

func (_ FfiConverterOptionalTypeMessagingServiceMetadata) Read(reader io.Reader) *MessagingServiceMetadata {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeMessagingServiceMetadataINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeMessagingServiceMetadata) Lower(value *MessagingServiceMetadata) RustBuffer {
	return LowerIntoRustBuffer[*MessagingServiceMetadata](c, value)
}

func (_ FfiConverterOptionalTypeMessagingServiceMetadata) Write(writer io.Writer, value *MessagingServiceMetadata) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeMessagingServiceMetadataINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeMessagingServiceMetadata struct{}

func (_ FfiDestroyerOptionalTypeMessagingServiceMetadata) Destroy(value *MessagingServiceMetadata) {
	if value != nil {
		FfiDestroyerTypeMessagingServiceMetadata{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypeSecret struct{}

var FfiConverterOptionalTypeSecretINSTANCE = FfiConverterOptionalTypeSecret{}

func (c FfiConverterOptionalTypeSecret) Lift(rb RustBufferI) *Secret {
	return LiftFromRustBuffer[*Secret](c, rb)
}

func (_ FfiConverterOptionalTypeSecret) Read(reader io.Reader) *Secret {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeSecretINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeSecret) Lower(value *Secret) RustBuffer {
	return LowerIntoRustBuffer[*Secret](c, value)
}

func (_ FfiConverterOptionalTypeSecret) Write(writer io.Writer, value *Secret) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeSecretINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeSecret struct{}

func (_ FfiDestroyerOptionalTypeSecret) Destroy(value *Secret) {
	if value != nil {
		FfiDestroyerTypeSecret{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypeAnonCryptAlg struct{}

var FfiConverterOptionalTypeAnonCryptAlgINSTANCE = FfiConverterOptionalTypeAnonCryptAlg{}

func (c FfiConverterOptionalTypeAnonCryptAlg) Lift(rb RustBufferI) *AnonCryptAlg {
	return LiftFromRustBuffer[*AnonCryptAlg](c, rb)
}

func (_ FfiConverterOptionalTypeAnonCryptAlg) Read(reader io.Reader) *AnonCryptAlg {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeAnonCryptAlgINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeAnonCryptAlg) Lower(value *AnonCryptAlg) RustBuffer {
	return LowerIntoRustBuffer[*AnonCryptAlg](c, value)
}

func (_ FfiConverterOptionalTypeAnonCryptAlg) Write(writer io.Writer, value *AnonCryptAlg) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeAnonCryptAlgINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeAnonCryptAlg struct{}

func (_ FfiDestroyerOptionalTypeAnonCryptAlg) Destroy(value *AnonCryptAlg) {
	if value != nil {
		FfiDestroyerTypeAnonCryptAlg{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypeAuthCryptAlg struct{}

var FfiConverterOptionalTypeAuthCryptAlgINSTANCE = FfiConverterOptionalTypeAuthCryptAlg{}

func (c FfiConverterOptionalTypeAuthCryptAlg) Lift(rb RustBufferI) *AuthCryptAlg {
	return LiftFromRustBuffer[*AuthCryptAlg](c, rb)
}

func (_ FfiConverterOptionalTypeAuthCryptAlg) Read(reader io.Reader) *AuthCryptAlg {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeAuthCryptAlgINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeAuthCryptAlg) Lower(value *AuthCryptAlg) RustBuffer {
	return LowerIntoRustBuffer[*AuthCryptAlg](c, value)
}

func (_ FfiConverterOptionalTypeAuthCryptAlg) Write(writer io.Writer, value *AuthCryptAlg) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeAuthCryptAlgINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeAuthCryptAlg struct{}

func (_ FfiDestroyerOptionalTypeAuthCryptAlg) Destroy(value *AuthCryptAlg) {
	if value != nil {
		FfiDestroyerTypeAuthCryptAlg{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypeSignAlg struct{}

var FfiConverterOptionalTypeSignAlgINSTANCE = FfiConverterOptionalTypeSignAlg{}

func (c FfiConverterOptionalTypeSignAlg) Lift(rb RustBufferI) *SignAlg {
	return LiftFromRustBuffer[*SignAlg](c, rb)
}

func (_ FfiConverterOptionalTypeSignAlg) Read(reader io.Reader) *SignAlg {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeSignAlgINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeSignAlg) Lower(value *SignAlg) RustBuffer {
	return LowerIntoRustBuffer[*SignAlg](c, value)
}

func (_ FfiConverterOptionalTypeSignAlg) Write(writer io.Writer, value *SignAlg) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeSignAlgINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeSignAlg struct{}

func (_ FfiDestroyerOptionalTypeSignAlg) Destroy(value *SignAlg) {
	if value != nil {
		FfiDestroyerTypeSignAlg{}.Destroy(*value)
	}
}

type FfiConverterOptionalSequenceString struct{}

var FfiConverterOptionalSequenceStringINSTANCE = FfiConverterOptionalSequenceString{}

func (c FfiConverterOptionalSequenceString) Lift(rb RustBufferI) *[]string {
	return LiftFromRustBuffer[*[]string](c, rb)
}

func (_ FfiConverterOptionalSequenceString) Read(reader io.Reader) *[]string {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterSequenceStringINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalSequenceString) Lower(value *[]string) RustBuffer {
	return LowerIntoRustBuffer[*[]string](c, value)
}

func (_ FfiConverterOptionalSequenceString) Write(writer io.Writer, value *[]string) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterSequenceStringINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalSequenceString struct{}

func (_ FfiDestroyerOptionalSequenceString) Destroy(value *[]string) {
	if value != nil {
		FfiDestroyerSequenceString{}.Destroy(*value)
	}
}

type FfiConverterOptionalSequenceTypeAttachment struct{}

var FfiConverterOptionalSequenceTypeAttachmentINSTANCE = FfiConverterOptionalSequenceTypeAttachment{}

func (c FfiConverterOptionalSequenceTypeAttachment) Lift(rb RustBufferI) *[]Attachment {
	return LiftFromRustBuffer[*[]Attachment](c, rb)
}

func (_ FfiConverterOptionalSequenceTypeAttachment) Read(reader io.Reader) *[]Attachment {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterSequenceTypeAttachmentINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalSequenceTypeAttachment) Lower(value *[]Attachment) RustBuffer {
	return LowerIntoRustBuffer[*[]Attachment](c, value)
}

func (_ FfiConverterOptionalSequenceTypeAttachment) Write(writer io.Writer, value *[]Attachment) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterSequenceTypeAttachmentINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalSequenceTypeAttachment struct{}

func (_ FfiDestroyerOptionalSequenceTypeAttachment) Destroy(value *[]Attachment) {
	if value != nil {
		FfiDestroyerSequenceTypeAttachment{}.Destroy(*value)
	}
}

type FfiConverterOptionalMapStringTypeJsonValue struct{}

var FfiConverterOptionalMapStringTypeJsonValueINSTANCE = FfiConverterOptionalMapStringTypeJsonValue{}

func (c FfiConverterOptionalMapStringTypeJsonValue) Lift(rb RustBufferI) *map[string]JsonValue {
	return LiftFromRustBuffer[*map[string]JsonValue](c, rb)
}

func (_ FfiConverterOptionalMapStringTypeJsonValue) Read(reader io.Reader) *map[string]JsonValue {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterMapStringTypeJsonValueINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalMapStringTypeJsonValue) Lower(value *map[string]JsonValue) RustBuffer {
	return LowerIntoRustBuffer[*map[string]JsonValue](c, value)
}

func (_ FfiConverterOptionalMapStringTypeJsonValue) Write(writer io.Writer, value *map[string]JsonValue) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterMapStringTypeJsonValueINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalMapStringTypeJsonValue struct{}

func (_ FfiDestroyerOptionalMapStringTypeJsonValue) Destroy(value *map[string]JsonValue) {
	if value != nil {
		FfiDestroyerMapStringTypeJsonValue{}.Destroy(*value)
	}
}

type FfiConverterSequenceString struct{}

var FfiConverterSequenceStringINSTANCE = FfiConverterSequenceString{}

func (c FfiConverterSequenceString) Lift(rb RustBufferI) []string {
	return LiftFromRustBuffer[[]string](c, rb)
}

func (c FfiConverterSequenceString) Read(reader io.Reader) []string {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]string, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterStringINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceString) Lower(value []string) RustBuffer {
	return LowerIntoRustBuffer[[]string](c, value)
}

func (c FfiConverterSequenceString) Write(writer io.Writer, value []string) {
	if len(value) > math.MaxInt32 {
		panic("[]string is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterStringINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceString struct{}

func (FfiDestroyerSequenceString) Destroy(sequence []string) {
	for _, value := range sequence {
		FfiDestroyerString{}.Destroy(value)
	}
}

type FfiConverterSequenceTypeAttachment struct{}

var FfiConverterSequenceTypeAttachmentINSTANCE = FfiConverterSequenceTypeAttachment{}

func (c FfiConverterSequenceTypeAttachment) Lift(rb RustBufferI) []Attachment {
	return LiftFromRustBuffer[[]Attachment](c, rb)
}

func (c FfiConverterSequenceTypeAttachment) Read(reader io.Reader) []Attachment {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]Attachment, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeAttachmentINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeAttachment) Lower(value []Attachment) RustBuffer {
	return LowerIntoRustBuffer[[]Attachment](c, value)
}

func (c FfiConverterSequenceTypeAttachment) Write(writer io.Writer, value []Attachment) {
	if len(value) > math.MaxInt32 {
		panic("[]Attachment is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeAttachmentINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeAttachment struct{}

func (FfiDestroyerSequenceTypeAttachment) Destroy(sequence []Attachment) {
	for _, value := range sequence {
		FfiDestroyerTypeAttachment{}.Destroy(value)
	}
}

type FfiConverterSequenceTypeDIDDoc struct{}

var FfiConverterSequenceTypeDIDDocINSTANCE = FfiConverterSequenceTypeDIDDoc{}

func (c FfiConverterSequenceTypeDIDDoc) Lift(rb RustBufferI) []DidDoc {
	return LiftFromRustBuffer[[]DidDoc](c, rb)
}

func (c FfiConverterSequenceTypeDIDDoc) Read(reader io.Reader) []DidDoc {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]DidDoc, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeDIDDocINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeDIDDoc) Lower(value []DidDoc) RustBuffer {
	return LowerIntoRustBuffer[[]DidDoc](c, value)
}

func (c FfiConverterSequenceTypeDIDDoc) Write(writer io.Writer, value []DidDoc) {
	if len(value) > math.MaxInt32 {
		panic("[]DidDoc is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeDIDDocINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeDidDoc struct{}

func (FfiDestroyerSequenceTypeDidDoc) Destroy(sequence []DidDoc) {
	for _, value := range sequence {
		FfiDestroyerTypeDidDoc{}.Destroy(value)
	}
}

type FfiConverterSequenceTypeSecret struct{}

var FfiConverterSequenceTypeSecretINSTANCE = FfiConverterSequenceTypeSecret{}

func (c FfiConverterSequenceTypeSecret) Lift(rb RustBufferI) []Secret {
	return LiftFromRustBuffer[[]Secret](c, rb)
}

func (c FfiConverterSequenceTypeSecret) Read(reader io.Reader) []Secret {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]Secret, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeSecretINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeSecret) Lower(value []Secret) RustBuffer {
	return LowerIntoRustBuffer[[]Secret](c, value)
}

func (c FfiConverterSequenceTypeSecret) Write(writer io.Writer, value []Secret) {
	if len(value) > math.MaxInt32 {
		panic("[]Secret is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeSecretINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeSecret struct{}

func (FfiDestroyerSequenceTypeSecret) Destroy(sequence []Secret) {
	for _, value := range sequence {
		FfiDestroyerTypeSecret{}.Destroy(value)
	}
}

type FfiConverterSequenceTypeService struct{}

var FfiConverterSequenceTypeServiceINSTANCE = FfiConverterSequenceTypeService{}

func (c FfiConverterSequenceTypeService) Lift(rb RustBufferI) []Service {
	return LiftFromRustBuffer[[]Service](c, rb)
}

func (c FfiConverterSequenceTypeService) Read(reader io.Reader) []Service {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]Service, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeServiceINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeService) Lower(value []Service) RustBuffer {
	return LowerIntoRustBuffer[[]Service](c, value)
}

func (c FfiConverterSequenceTypeService) Write(writer io.Writer, value []Service) {
	if len(value) > math.MaxInt32 {
		panic("[]Service is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeServiceINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeService struct{}

func (FfiDestroyerSequenceTypeService) Destroy(sequence []Service) {
	for _, value := range sequence {
		FfiDestroyerTypeService{}.Destroy(value)
	}
}

type FfiConverterSequenceTypeVerificationMethod struct{}

var FfiConverterSequenceTypeVerificationMethodINSTANCE = FfiConverterSequenceTypeVerificationMethod{}

func (c FfiConverterSequenceTypeVerificationMethod) Lift(rb RustBufferI) []VerificationMethod {
	return LiftFromRustBuffer[[]VerificationMethod](c, rb)
}

func (c FfiConverterSequenceTypeVerificationMethod) Read(reader io.Reader) []VerificationMethod {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]VerificationMethod, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeVerificationMethodINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeVerificationMethod) Lower(value []VerificationMethod) RustBuffer {
	return LowerIntoRustBuffer[[]VerificationMethod](c, value)
}

func (c FfiConverterSequenceTypeVerificationMethod) Write(writer io.Writer, value []VerificationMethod) {
	if len(value) > math.MaxInt32 {
		panic("[]VerificationMethod is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeVerificationMethodINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeVerificationMethod struct{}

func (FfiDestroyerSequenceTypeVerificationMethod) Destroy(sequence []VerificationMethod) {
	for _, value := range sequence {
		FfiDestroyerTypeVerificationMethod{}.Destroy(value)
	}
}

type FfiConverterMapStringTypeJsonValue struct{}

var FfiConverterMapStringTypeJsonValueINSTANCE = FfiConverterMapStringTypeJsonValue{}

func (c FfiConverterMapStringTypeJsonValue) Lift(rb RustBufferI) map[string]JsonValue {
	return LiftFromRustBuffer[map[string]JsonValue](c, rb)
}

func (_ FfiConverterMapStringTypeJsonValue) Read(reader io.Reader) map[string]JsonValue {
	result := make(map[string]JsonValue)
	length := readInt32(reader)
	for i := int32(0); i < length; i++ {
		key := FfiConverterStringINSTANCE.Read(reader)
		value := FfiConverterTypeJsonValueINSTANCE.Read(reader)
		result[key] = value
	}
	return result
}

func (c FfiConverterMapStringTypeJsonValue) Lower(value map[string]JsonValue) RustBuffer {
	return LowerIntoRustBuffer[map[string]JsonValue](c, value)
}

func (_ FfiConverterMapStringTypeJsonValue) Write(writer io.Writer, mapValue map[string]JsonValue) {
	if len(mapValue) > math.MaxInt32 {
		panic("map[string]JsonValue is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(mapValue)))
	for key, value := range mapValue {
		FfiConverterStringINSTANCE.Write(writer, key)
		FfiConverterTypeJsonValueINSTANCE.Write(writer, value)
	}
}

type FfiDestroyerMapStringTypeJsonValue struct{}

func (_ FfiDestroyerMapStringTypeJsonValue) Destroy(mapValue map[string]JsonValue) {
	for key, value := range mapValue {
		FfiDestroyerString{}.Destroy(key)
		FfiDestroyerTypeJsonValue{}.Destroy(value)
	}
}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type JsonValue = string
type FfiConverterTypeJsonValue = FfiConverterString
type FfiDestroyerTypeJsonValue = FfiDestroyerString

var FfiConverterTypeJsonValueINSTANCE = FfiConverterString{}
