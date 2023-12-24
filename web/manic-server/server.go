package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	audioTypes "manic-compression/pkg/audio_types"
	fileSystem "manic-compression/pkg/file_system"
	serviceBus "manic-compression/pkg/service_bus"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	uuid "github.com/google/uuid"
)

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// queue constants
const (
	taskQueue        = "audiotasks"
	taskResultsQueue = "audiotaskresults"
)

var (
	connectionString = os.Getenv("AZURE_STORAGE_CONNECTION_STRING")
	inputContainer   = getEnvOrDefault("INPUT_CONTAINER_NAME", "audio-input")
	outputContainer  = getEnvOrDefault("OUTPUT_CONTAINER_NAME", "audio-output")
)

type App struct {
	Router           *chi.Mux
	InputFileSystem  *fileSystem.FileSystem
	OutputFileSystem *fileSystem.FileSystem
	ServiceBus       *serviceBus.ServiceBus
}

// start request specifies all the files to be processed and the audio functions to be applied to each file
type StartRequest struct {
	InputFiles            []string `json:"inputFiles"`
	ClientID              string   `json:"clientID"`
	AudioFunctionPipeline []string `json:"audioFunctionPipeline"`
}

type TaskStatusRequest struct {
	Tasks map[string]audioTypes.AudioTask `json:"tasks"`
}

func main() {

	serviceClient := fileSystem.CreateServiceClient(connectionString)

	app := &App{
		Router: chi.NewRouter(),
		InputFileSystem: &fileSystem.FileSystem{
			ContainerName: inputContainer,
			ServiceClient: serviceClient,
		},
		OutputFileSystem: &fileSystem.FileSystem{
			ContainerName: outputContainer,
			ServiceClient: serviceClient,
		},
		ServiceBus: serviceBus.NewServiceBus(),
	}

	// Initialize CORS middleware with desired options
	corsMiddleware := cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	})

	// Use the CORS middleware
	app.Router.Use(corsMiddleware)

	// get and set available files
	app.initializeFiles()

	// initialize app routes
	app.InitializeRoutes()

	// mount the app router under /api (all calls must go through /api)
	apiRouter := chi.NewRouter()
	apiRouter.Mount("/api", app.Router)

	// start API server
	log.Println("Starting server on port :8080...")
	http.Handle("/", apiRouter)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func (app *App) initializeFiles() {
	app.InputFileSystem.Files = app.InputFileSystem.ListBlobs()
	app.OutputFileSystem.Files = app.OutputFileSystem.ListBlobs()
}

func (app *App) InitializeRoutes() {

	app.Router.Route("/hello", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello from Manic Compression Server!"))
		})
	})

	app.Router.Route("/start", func(r chi.Router) {
		r.Post("/", app.ManicCompressionHandler())
	})

	app.Router.Route("/activeTasks", func(r chi.Router) {
		r.Get("/", app.GetActiveTasksHandler())
	})

	app.Router.Route("/completedTasks", func(r chi.Router) {
		r.Get("/", app.GetCompletedTasksHandler())
	})

	app.Router.Route("/clearActiveTasks", func(r chi.Router) {
		r.Post("/", app.ClearActiveTasks())
	})

	app.Router.Route("/clearCompletedTasks", func(r chi.Router) {
		r.Post("/", app.ClearCompletedTasks())
	})

	app.Router.Route("/functions", func(r chi.Router) {
		r.Get("/", GetAudioFunctions())
	})

	app.Router.Route("/input", func(r chi.Router) {
		r.Get("/", ListFilesHandler(app.InputFileSystem))
		r.Get("/{name}", DownloadFileHandler(app.InputFileSystem))
		r.Post("/", UploadFileHandler(app.InputFileSystem))
		r.Delete("/{name}", DeleteFileHandler(app.InputFileSystem))
		r.Delete("/", ClearContainerHandler(app.InputFileSystem))
	})

	app.Router.Route("/output", func(r chi.Router) {
		r.Get("/", ListFilesHandler(app.OutputFileSystem))
		r.Get("/{name}", DownloadFileHandler(app.OutputFileSystem))
		r.Post("/", UploadFileHandler(app.OutputFileSystem))
		r.Delete("/{name}", DeleteFileHandler(app.OutputFileSystem))
		r.Delete("/", ClearContainerHandler(app.OutputFileSystem))
	})
}

func (app *App) ManicCompressionHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling start request")

		var req StartRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, fmt.Sprintf("could not decode request body: %v", err), http.StatusBadRequest)
			return
		}

		messages := []serviceBus.Msg{}
		tasks := []audioTypes.AudioTask{}

		for _, inputFile := range req.InputFiles {
			taskID := uuid.New().String()
			task := audioTypes.AudioTask{
				ClientID:              req.ClientID,
				TaskID:                taskID,
				Status:                serviceBus.TaskInProgress,
				InputFile:             inputFile,
				AudioFunctionPipeline: req.AudioFunctionPipeline,
			}
			msg := serviceBus.Msg{
				Type:    "processAudio",
				Content: task.Serialize(),
			}
			messages = append(messages, msg)
			tasks = append(tasks, task)
		}

		app.ServiceBus.SendMessageBatch(messages, taskQueue)

		// return task IDs to client so they can poll for results
		json.NewEncoder(w).Encode(map[string][]audioTypes.AudioTask{"tasks": tasks})
	}
}

func (app *App) GetActiveTasksHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling active tasks request")
		activeTasks, err := app.ServiceBus.PeekQueue(taskQueue)
		if err != nil {
			http.Error(w, fmt.Sprintf("could not get active tasks: %v", err), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(activeTasks)
	}
}

func (app *App) GetCompletedTasksHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling completed tasks request")
		completedTasks, err := app.ServiceBus.PeekQueue(taskResultsQueue)
		if err != nil {
			http.Error(w, fmt.Sprintf("could not get active tasks: %v", err), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(completedTasks)
	}
}

func (app *App) ClearActiveTasks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling clear active tasks request")
		app.ServiceBus.ClearQueue(taskQueue)
		msg := fmt.Sprintf("%s cleared successfully", taskQueue)
		json.NewEncoder(w).Encode(msg)
	}
}

func (app *App) ClearCompletedTasks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling clear task results request")
		app.ServiceBus.ClearQueue(taskResultsQueue)
		msg := fmt.Sprintf("%s cleared successfully", taskResultsQueue)
		json.NewEncoder(w).Encode(msg)
	}
}

func GetAudioFunctions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]string{
			audioTypes.AudioFunctionApplyEffect1,
			audioTypes.AudioFunctionApplyEffect2,
			audioTypes.AudioFunctionWAVToMp3,
		})
	}
}

func ListFilesHandler(fs *fileSystem.FileSystem) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling get files request")
		files := fs.ListBlobs()
		fs.Files = files
		json.NewEncoder(w).Encode(files)
	}
}

func DownloadFileHandler(fs *fileSystem.FileSystem) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file := chi.URLParam(r, "name")
		log.Printf("Handling download file request for file %s", file)
		fs.DownloadHTTPFileStream(w, file)
	}
}

func UploadFileHandler(fs *fileSystem.FileSystem) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log.Println("Handling file upload request")
		// use multipart form data to upload files
		err := r.ParseMultipartForm(40 << 20) // 40 MB max file size allowed
		if err != nil {
			http.Error(w, fmt.Sprintf("could not parse multipart form: %v", err), http.StatusBadRequest)
			return
		}

		// gt files from form data
		files := r.MultipartForm.File["files"]
		if files == nil {
			http.Error(w, "no files provided", http.StatusBadRequest)
			return
		}

		filesUploaded := []string{}

		for _, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				http.Error(w, fmt.Sprintf("could not open file: %v", err), http.StatusInternalServerError)
				return
			}
			defer file.Close()

			fs.UploadFile(file, fileHeader.Filename)
			filesUploaded = append(filesUploaded, fileHeader.Filename)
		}
		json.NewEncoder(w).Encode(filesUploaded)
	}
}

// DeleteBlobHandler handles the DELETE requests to delete blobs.
func DeleteFileHandler(fs *fileSystem.FileSystem) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file := chi.URLParam(r, "name")
		fs.DeleteBlob(file)
		msg := fmt.Sprintf("%s deleted successfully", file)
		json.NewEncoder(w).Encode(msg)
	}
}

func ClearContainerHandler(fs *fileSystem.FileSystem) http.HandlerFunc {
	fmt.Printf("Clearing container: %s", fs.ContainerName)
	return func(w http.ResponseWriter, r *http.Request) {
		fs.ClearContainer()
		msg := fmt.Sprintf("%s cleared successfully", fs.ContainerName)
		json.NewEncoder(w).Encode(msg)
	}
}
