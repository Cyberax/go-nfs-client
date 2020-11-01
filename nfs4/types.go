package nfs4

type Uint32_t = uint32
type Uint64_t = uint64
type Int64_t = int64

type XDR_Uint32_t = *XdrUint32
type XdrType_Uint32_t = XdrType_uint32
type XDR_Uint64_t = *XdrUint64
type XdrType_Uint64_t = XdrType_uint64
type XdrType_Int64_t = XdrType_int64
type XDR_Int64_t = *XdrInt64

const AUTH_NONE = 0
const AUTH_SYS = 1
const AUTH_SHORT = 2
const AUTH_DH = 3

//type XDR_Uint32_t = *xdr.XdrUint32
//type XdrType_Uint32_t = xdr.XdrType_uint32
//type XdrType_Uint64_t = xdr.XdrType_uint64
//type XDR_Uint64_t = *xdr.XdrUint64
//type XdrType_Int64_t = xdr.XdrType_int64
//type XDR_Int64_t = *xdr.XdrInt64

//The XdrType that unsigned hyper gets converted to for marshaling.
//type XdrUint64 uint64
//type XdrType_uint64 = *XdrUint64
//func (XdrUint64) XdrTypeName() string { return "uint64" }
//func (v XdrUint64) String() string { return fmt.Sprintf("%v", v.XdrValue()) }
//func (v *XdrUint64) Scan(ss fmt.ScanState, r rune) error {
//	_, err := fmt.Fscanf(ss, string([]rune{'%', r}), v.XdrPointer())
//	return err
//}
//func (v XdrUint64) GetU64() uint64 { return uint64(v) }
//func (v *XdrUint64) SetU64(nv uint64) { *v = XdrUint64(nv) }
//func (v *XdrUint64) XdrPointer() interface{} { return (*uint64)(v) }
//func (v XdrUint64) XdrValue() interface{} { return uint64(v) }
//func (v *XdrUint64) XdrMarshal(x XDR, name string) { x.Marshal(name, v) }
//func XDR_uint64(v *uint64) *XdrUint64 { return (*XdrUint64)(v) }
