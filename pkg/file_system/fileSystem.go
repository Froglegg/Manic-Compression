package fileSystem

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/streaming"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
)

type FileSystem struct {
	ServiceClient *azblob.Client
	Files         []BlobInfo
	ContainerName string
}

type BlobInfo struct {
	Name string
	Size int64
}

func handleError(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

func CreateServiceClient(connectionString string) *azblob.Client {
	serviceClient, err := azblob.NewClientFromConnectionString(connectionString, nil)
	handleError(err)
	return serviceClient
}

func (fs *FileSystem) UploadFiles(directory string) []string {
	fmt.Println("UPLOADING FILES")
	items, _ := os.ReadDir(directory)
	inputFiles := []string{}
	for _, item := range items {
		fmt.Println("Uploading " + item.Name())

		filePath := directory + "/" + item.Name()
		f, _ := os.Open(filePath)
		defer f.Close()

		_, err := fs.ServiceClient.UploadFile(context.TODO(), fs.ContainerName, item.Name(), f, nil)
		handleError(err)
		inputFiles = append(inputFiles, item.Name())
	}
	return inputFiles
}

func (fs *FileSystem) UploadFile(r io.Reader, filename string) error {
	fmt.Println("Uploading " + filename)

	// create a temporary file to store the contents of the reader
	tmpFile, err := os.CreateTemp("", "upload-*")
	if err != nil {
		return err
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	// copy the contents of the reader to the temporary file
	_, err = io.Copy(tmpFile, r)
	if err != nil {
		return err
	}

	// seek to the beginning of the temporary file
	_, err = tmpFile.Seek(0, 0)
	if err != nil {
		return err
	}

	_, err = fs.ServiceClient.UploadFile(context.TODO(), fs.ContainerName, filename, tmpFile, nil)
	handleError(err)

	return nil
}

func (fs *FileSystem) DownloadFile(fileName string) {
	// Set up file to download the blob to
	destFile, err := os.Create(fileName)
	handleError(err)
	defer func(destFile *os.File) {
		err = destFile.Close()
		handleError(err)
	}(destFile)

	// Perform download
	_, err = fs.ServiceClient.DownloadFile(
		context.TODO(),
		fs.ContainerName,
		fileName,
		destFile,
		&azblob.DownloadFileOptions{},
	)

	// Assert download was successful
	handleError(err)
}

func (fs *FileSystem) DownloadFileToDst(fileName string, dstFileName string) {
	// Set up file to download the blob to
	destFile, err := os.Create(dstFileName)
	handleError(err)
	defer func(destFile *os.File) {
		err = destFile.Close()
		handleError(err)
	}(destFile)

	// Perform download
	_, err = fs.ServiceClient.DownloadFile(
		context.TODO(),
		fs.ContainerName,
		fileName,
		destFile,
		&azblob.DownloadFileOptions{},
	)

	// Assert download was successful
	handleError(err)
}

func (fs *FileSystem) DownloadHTTPFileStream(w http.ResponseWriter, fileName string) {
	// set expected headers
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	w.Header().Set("Content-Type", "application/octet-stream")

	// open download stream from blob
	response, err := fs.ServiceClient.DownloadStream(
		context.Background(),
		fs.ContainerName,
		fileName,
		&azblob.DownloadStreamOptions{},
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not download blob: %v", err), http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	// copy the blob stream over to the response writer
	if _, err = io.Copy(w, response.Body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (fs *FileSystem) ListBlobs() []BlobInfo {

	pager := fs.ServiceClient.NewListBlobsFlatPager(fs.ContainerName, &azblob.ListBlobsFlatOptions{
		// Include: container.ListBlobsInclude{Deleted: true, Versions: true},
	})

	blob_list := []BlobInfo{}

	for pager.More() {
		resp, err := pager.NextPage(context.TODO())
		handleError(err) // if err is not nil, break the loop.
		for _, _blob := range resp.Segment.BlobItems {
			element := &BlobInfo{
				Name: *_blob.Name,
				Size: *_blob.Properties.ContentLength,
			}
			blob_list = append(blob_list, *element)
		}
	}

	return blob_list
}

func (fs *FileSystem) DeleteBlob(blobName string) {
	_, err := fs.ServiceClient.DeleteBlob(context.TODO(), fs.ContainerName, blobName, nil)
	handleError(err)
}

func (fs *FileSystem) ClearContainer() {
	blob_list := fs.ListBlobs()
	for _, blob := range blob_list {
		fs.DeleteBlob(blob.Name)
	}
}

// Upload a blob (e.g., shards, intermediate files) to the file system
func (fs *FileSystem) UploadBlob(blobName string, blobData string) {
	// could also use bytes
	// blobContentReader := bytes.NewReader(blobData)
	_, err := fs.ServiceClient.UploadStream(
		context.TODO(),
		fs.ContainerName,
		blobName,
		strings.NewReader(blobData),
		&azblob.UploadStreamOptions{},
	)

	handleError(err)
}

// Download blob from the file system
// Uses the DownloadStream method to stream a blob's contents to a local file.
// Uses intelligent retries to download the blob.
func (fs *FileSystem) DownloadBlob(
	blobName string,
	rangeStart int64,
	rangeEnd int64,
	saveToFile bool,
) string {

	var downloadStreamOptions blob.DownloadStreamOptions

	if rangeStart >= 0 && rangeEnd >= 0 {
		downloadStreamOptions = azblob.DownloadStreamOptions{
			Range: azblob.HTTPRange{
				Offset: rangeStart, // specify the start of the range
				Count:  rangeEnd,   // specify the end of the range
			},
		}
	}

	// Download returns an intelligent retryable stream around a blob; it returns an io.ReadCloser.
	dr, err := fs.ServiceClient.DownloadStream(
		context.TODO(),
		fs.ContainerName,
		blobName,
		&downloadStreamOptions,
	)
	handleError(err)
	rs := dr.Body

	// NewResponseBodyProgress wraps the GetRetryStream with progress reporting; it returns an io.ReadCloser.
	stream := streaming.NewResponseProgress(
		rs,
		func(bytesTransferred int64) {
			// fmt.Printf("Downloaded %d of %d bytes.\n", bytesTransferred, contentLength)
		},
	)
	defer func(stream io.ReadCloser) {
		err := stream.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(stream) // The client must close the response body when finished with it

	if saveToFile {

		file, err := os.Create(blobName) // Create the file to hold the downloaded blob contents.
		handleError(err)
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				log.Fatal(err)
			}
		}(file)

		written, err := io.Copy(file, stream) // Write to the file by reading from the blob (with intelligent retries).
		handleError(err)
		fmt.Printf("Wrote %d bytes.\n", written)
	}

	buf := new(strings.Builder)
	_, err = io.Copy(buf, stream)
	handleError(err)

	return buf.String()
}
