package filesystem

import (
	"os"
	"path"
	"errors"
	"strconv"

	"github.com/dapr/components-contrib/bindings"
	"github.com/dapr/dapr/pkg/logger"
)

// FileSystem allows writing to the local file system
type FileSystem struct {
	metadata fileSystemMetadata
	logger   logger.Logger
}

// The metadata holds path properties
type fileSystemMetadata struct {
	FolderName    string `json:"folderName"`
	FileName      string `json:"fileName"`
}

// NewFileSystem returns a new FileSystem bindings instance
func NewFileSystem(logger logger.Logger) *FileSystem {
	return &FileSystem{logger: logger}
}

// Helper to parse metadata
func (fs *FileSystem) parseMetadata(meta bindings.Metadata) (fileSystemMetadata, error) {
	fsMeta := fileSystemMetadata{}

	// Optional properties, these can also be set on a per request basis
	fsMeta.FolderName = meta.Properties["folderName"]
	fsMeta.FileName = meta.Properties["fileName"]

	return fsMeta, nil
}

func (fs *FileSystem) Init(metadata bindings.Metadata) error {

	fsMeta := fileSystemMetadata{}

	// Optional properties, these can also be set on a per request basis
	fsMeta.FolderName = metadata.Properties["folderName"]
	fsMeta.FileName = metadata.Properties["fileName"]

	fs.metadata = fsMeta

	return nil
}

func (fs *FileSystem) Operations() []bindings.OperationKind {
	return []bindings.OperationKind{bindings.CreateOperation}
}

func (fs *FileSystem) Invoke(req *bindings.InvokeRequest) (*bindings.InvokeResponse, error) {
	
	// We allow two possible sources of the properties we need,
	// the component metadata or request metadata.
	// Request takes priority if present.

	folderNameValue := fs.metadata.FolderName
	if folderNameValue == "" {
		folderNameFromRequest, ok := req.Metadata["folderName"]
		if !ok || folderNameFromRequest == "" {
			return nil, errors.New("filesystem missing \"folderName\" field")
		}
		folderNameValue = folderNameFromRequest
	}

	fileNameValue := fs.metadata.FileName
	if fileNameValue == "" {
		fileNameFromRequest, ok := req.Metadata["fileName"]
		if !ok || fileNameFromRequest == "" {
			return nil, errors.New("filesystem missing \"fileName\" field")
		}
		fileNameValue = fileNameFromRequest
	}

	// Ensure folder exists
	os.MkdirAll(folderNameValue, 0700)

	// Get content to write from request
	content, _ := strconv.Unquote(string(req.Data))

	// Create file
	fullPath := path.Join(folderNameValue, fileNameValue)
	f, err := os.Create(fullPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Write content to file
	f.WriteString(content)
	
	fs.logger.Info("written file with FileSystem: " + fullPath)
	fs.logger.Info(content)

	return nil, nil
}