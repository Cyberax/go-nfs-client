package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"github.com/aurorasolar/go-nfs-client/nfs4"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const NumScaleThreads = 40
const UploadSize = 1024*1024*5
const UploadTarget = 1024*1024*1024*5

//goland:noinspection GoNilness
func runTests(server, rootPath string) error {
	ctx := context.Background()

	hostname, _ := os.Hostname()
	client, err := nfs4.NewNfsClient(ctx, server, nfs4.AuthParams{
		MachineName: hostname,
	})
	if err != nil {
		return err
	}
	defer client.Close()

	if rootPath[len(rootPath)-1] != '/' {
		rootPath += "/"
	}

	// Remove anything from the old tests
	println("Removing old files")
	err = nfs4.RemoveRecursive(client, rootPath + "tests")
	checkErr(err)

	println("Making tests directory")
	err = client.MakePath(rootPath + "tests/directory/a")
	checkErr(err)

	// Path creation is idempotent
	err = client.MakePath(rootPath + "tests/directory/a")
	checkErr(err)

	println("Checking directory info")
	fi, err := client.GetFileInfo(rootPath + "tests/directory/a")
	checkErr(err)
	check(fi.IsDir && fi.Name == "a")
	check(approxNow(fi))

	println("Checking directory deletion")
	err = client.DeleteFile(rootPath + "tests/directory/a")
	checkErr(err)
	_, err = client.GetFileInfo(rootPath + "tests/directory/a")
	check(nfs4.IsNfsError(err, nfs4.ERROR_NOENT))

	// Deletion will fail with non-empty dir
	err = client.DeleteFile(rootPath + "tests")
	check(nfs4.IsNfsError(err, nfs4.ERROR_NOTEMPTY))

	// Check file ops
	testFileOps(client, rootPath + "tests/")

	testMassOps(client, rootPath + "tests/")

	println("Cleaning up")
	err = nfs4.RemoveRecursive(client, rootPath + "tests")
	checkErr(err)

	return nil
}

//goland:noinspection GoNilness
func testFileOps(cli nfs4.NfsInterface, path string) {
	data := make([]byte, 20*1024*1024)
	// Ganesha replaces all the data written by letters 'a'
	for i := range data {
		data[i] = 'a'
	}

	println("Checking file upload")
	written, err := cli.ReWriteFile(path+"file.bin", bytes.NewReader(data))
	checkErr(err)
	check(written == uint64(len(data)))

	println("Checking file download")
	buffer := bytes.NewBufferString("")
	read, err := cli.ReadFileAll(path + "file.bin", buffer)
	checkErr(err)
	check(read == uint64(len(data)))
	check(assert.ObjectsAreEqual(buffer.Bytes(), data))

	println("Checking file meta")
	info, err := cli.GetFileInfo(path + "file.bin")
	checkErr(err)
	check(info.Size == uint64(len(data)))
	check(!info.IsDir)
	check(info.Name == "file.bin")
	check(approxNow(info))

	println("Checking file deletion")
	err = cli.DeleteFile(path + "file.bin")
	checkErr(err)

	println("Verifying file deletion")
	_, err = cli.GetFileInfo(path + "file.bin")
	check(nfs4.IsNfsError(err, nfs4.ERROR_NOENT))
	_, err = cli.ReadFileAll(path + "file.bin", buffer)
	check(nfs4.IsNfsError(err, nfs4.ERROR_NOENT))
}

func testMassOps(cli nfs4.NfsInterface, path string) {
	println("Making the mass directory")
	err := cli.MakePath(path+"/mass")
	checkErr(err)

	println("Checking creation")
	st := time.Now()
	files := make(map[string]bool)
	for i := 0; i<2000; i++ {
		curFile := fmt.Sprintf("%s/mass/file-%d", path, i)
		_, err = cli.ReWriteFile(curFile, strings.NewReader("aaaaaaa"))
		files[fmt.Sprintf("file-%d", i)] = true
	}
	println("Time diff (ms): ", time.Now().Sub(st).Milliseconds())
	println("Getting the file list")

	lst, err := cli.GetFileList(path+"/mass")
	checkErr(err)
	for _, l := range lst {
		check(!l.IsDir)
		check(l.Size == 7)
		check(approxNow(l))
		delete(files, l.Name)
	}
	check(len(files) == 0)
}

func approxNow(fi nfs4.FileInfo) bool {
	ms := fi.Mtime.Sub(time.Now()).Milliseconds()
	if ms < 0 {
		ms = -ms
	}
	return ms < 600000
}

func checkErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func check(b bool) {
	if !b {
		panic("Check failed")
	}
}

func runScale(server, rootPath string) error {
	ctx := context.Background()

	hostname, _ := os.Hostname()
	var nfsClients []nfs4.NfsInterface

	defer func() {
		for _, c := range nfsClients {
			c.Close()
		}
	}()

	for i := 0; i<NumScaleThreads; i++ {
		client, err := nfs4.NewNfsClient(ctx, server, nfs4.AuthParams{
			MachineName: hostname,
		})
		if err != nil {
			return err
		}
		nfsClients = append(nfsClients, client)
	}

	if rootPath[len(rootPath)-1] != '/' {
		rootPath += "/"
	}

	// Remove anything from the old tests
	println("Removing old files")
	err := nfs4.RemoveRecursive(nfsClients[0], rootPath + "scale")
	checkErr(err)

	err = nfsClients[0].MakePath(rootPath + "scale")
	checkErr(err)

	defer func() {
		println("After test cleanup")
		_ = nfs4.RemoveRecursive(nfsClients[0], rootPath + "scale")
	}()

	data := make([]byte, UploadSize)
	_, _ = rand.Read(data)

	testWriteScaling(nfsClients, rootPath, data)
	testReadScaling(nfsClients, rootPath, data)

	return nil
}

func testWriteScaling(nfsClients []nfs4.NfsInterface, rootPath string, data []byte) {
	// Now run the scale test - upload files in multiple threads until we
	// reach the desired number of uploads
	infoMtx := sync.Mutex{}
	uploadedBytes := uint64(0)
	doneUploading := false
	start := time.Now()

	var count int32
	wait := sync.WaitGroup{}
	println("Running the upload test. Threads =", NumScaleThreads)
	for _, c := range nfsClients {
		wait.Add(1)
		go func(c nfs4.NfsInterface) {
			defer wait.Done()
			for ; !doneUploading; {
				nm := fmt.Sprintf(rootPath+"scale/test-%d", atomic.AddInt32(&count, 1))
				n, err := c.ReWriteFile(nm, bytes.NewReader(data))
				if err != nil {
					println("Error: ", err.Error())
					doneUploading = true
				}

				curUploaded := atomic.AddUint64(&uploadedBytes, n)
				if curUploaded > UploadTarget {
					infoMtx.Lock()
					if !doneUploading {
						doneUploading = true
						ms := time.Now().Sub(start).Milliseconds()
						rate := (float64(curUploaded) * 1000.0 / float64(ms)) / 1024 / 1024
						println("Uploaded bytes: ", curUploaded, ", time(ms): ", ms,
							" rate (MB/s): ", int64(rate))
					}
					infoMtx.Unlock()
				}
			}
		}(c)
	}
	wait.Wait()
}

func testReadScaling(nfsClients []nfs4.NfsInterface, rootPath string, data []byte) {
	// Now run the scale test - upload files in multiple threads until we
	// reach the desired number of uploads
	infoMtx := sync.Mutex{}
	downloadedBytes := uint64(0)
	done := false
	start := time.Now()

	var count int32
	wait := sync.WaitGroup{}
	println("Running the read test. Threads =", NumScaleThreads)
	for _, c := range nfsClients {
		wait.Add(1)
		go func(c nfs4.NfsInterface) {
			defer wait.Done()
			for ; !done; {
				nm := fmt.Sprintf(rootPath+"scale/test-%d", atomic.AddInt32(&count, 1))
				reader := bytes.NewBufferString("")
				n, err := c.ReadFileAll(nm, reader)
				if err != nil {
					println("Error: ", err.Error())
					done = true
				}
				resBytes := reader.Bytes()
				for i := 0; i < len(data); i++ {
					if data[i] != resBytes[i] {
						println("Data mismatch")
						done = true
						break
					}
				}

				curDownloaded := atomic.AddUint64(&downloadedBytes, n)
				if curDownloaded > UploadTarget {
					infoMtx.Lock()
					if !done {
						done = true
						ms := time.Now().Sub(start).Milliseconds()
						rate := (float64(curDownloaded) * 1000.0 / float64(ms)) / 1024 / 1024
						println("Downloaded bytes: ", curDownloaded, ", time(ms): ", ms,
							" rate (MB/s): ", int64(rate))
					}
					infoMtx.Unlock()
				}
			}
		}(c)
	}
	wait.Wait()
}

func main() {
	if len(os.Args) < 4 {
		_, _ = os.Stderr.WriteString("Usage: runtests [scale|test] <NFS-server> <root-path>\n")
		os.Exit(1)
	}
	defer func() {
		p := recover()
		if p != nil {
			_, _ = os.Stderr.WriteString(fmt.Sprintf("%v\n", p))
			os.Exit(2)
		}
	}()

	var err error

	if os.Args[1] == "scale" {
		println("Running scalability tests")
		err = runScale(os.Args[2], os.Args[3])
	} else if os.Args[1] == "test" {
		println("Running correctness tests")
		err = runTests(os.Args[2], os.Args[3])
	} else {
		println("Unknown test mode: ", os.Args[1])
		os.Exit(3)
	}

	if err != nil {
		_, _ = os.Stderr.WriteString(fmt.Sprintf("%v\n", err))
		os.Exit(2)
	}
}
