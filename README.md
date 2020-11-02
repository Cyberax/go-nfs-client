# Pure-Go NFSv4 client
This library is a pure-Go client for NFSv4 (NFSv3 is NOT supported at all). It's mostly 
designed and tested for AWS EFS, but will work fine with Linux's `nfs4-server` or userspace-based
Ganesha. It's fully synchronous, does nothing behind your back and fully supports 
`context.Context`-based cancellation and deadlines.

It's also a fairly minimal library with lots of limitations:
1. No support for locking.
2. No support for reconnection and session resumption.
3. Minimalistic API.
4. No support for ACLs or extended attributes.
5. No support for any authentication methods.

# Usage example
See `cmd/main/runtests.go` for the usage examples.

## Regenerating the XDR bindings
If you need to regenerate the XDR bindings, then there's some manual work involved.

First use:
```bash
go run github.com/xdrpp/goxdr/cmd/goxdr -B -enum-comments -p internal internal/nfs4.x > internal/nfs4.go
go run github.com/xdrpp/goxdr/cmd/goxdr -b -enum-comments -p internal internal/rpc.x > internal/rpc.go
```

After this, manually do the following:
1. Rename `_u` to `_U`.
2. Rename `xdrProc_NFSPROC4_COMPOUND` to `XdrProc_NFSPROC4_COMPOUND`
3. Rename `xdrProc_NFSPROC4_NULL` to `XdrProc_NFSPROC4_NULL`
4. Change all the methods like `XDR_Offset4(v *Offset4) XdrType_Offset4` to return pointers (not values).
