// Package e2e provides an E2E test for flaap.
package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	gocmd "github.com/go-cmd/cmd"

	"github.com/go-logr/logr"
	"github.com/jlewi/flaap/go/pkg/networking"
	"github.com/jlewi/flaap/go/pkg/subprocess"
	"github.com/jlewi/p22h/backend/pkg/logging"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// rootDir returns the assumed root directory of the repository
func rootDir() string {
	wd, err := os.Getwd()
	if err != nil {
		wd = "/tmp"
	}

	root, _ := filepath.Abs(path.Join(wd, ".."))
	return root
}

func newRootCmd() *cobra.Command {
	var level string
	var jsonLog bool

	runner := &Runner{
		jsonLog: jsonLog,
		level:   level,
	}

	rootCmd := &cobra.Command{
		Short: "Run the E2E test",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			newLogger, err := logging.InitLogger(level, !jsonLog)
			if err != nil {
				panic(err)
			}
			log = *newLogger
		},
		Run: func(cmd *cobra.Command, args []string) {
			if err := runner.Run(); err != nil {
				fmt.Printf("run failed with error: %+v", err)
				os.Exit(1)
			}
		},
	}

	defaultLogsDir := path.Join(os.TempDir(), "flaapE2ELogs")
	root := rootDir()
	defaultServer := filepath.Join(root, ".build", "server")
	rootCmd.PersistentFlags().StringVarP(&runner.taskStore, "taskstore", "", defaultServer, "Path to the task store binary")
	rootCmd.PersistentFlags().StringVarP(&level, "level", "", "info", "The logging level.")
	rootCmd.PersistentFlags().BoolVarP(&jsonLog, "json-logs", "", true, "Enable json logging.")
	rootCmd.PersistentFlags().StringVarP(&runner.logsDir, "logs-dir", "", defaultLogsDir, "Directory where logs should be written.")
	rootCmd.Flags().IntVarP(&runner.port, "port", "p", 0, "Port to serve on. Value of 0 means pick an available port")
	return rootCmd
}

type Runner struct {
	logsDir   string
	h         *subprocess.ExecHelper
	taskStore string
	jsonLog   bool
	level     string
	port      int

	logFiles map[string]*os.File

	// Keep track of all the commands so we can close them.
	cmds map[string]*gocmd.Cmd
}

func (r *Runner) Run() error {
	numWorkers := 2
	defer r.closeFiles()
	r.h = &subprocess.ExecHelper{
		Log: log,
	}

	r.logFiles = map[string]*os.File{}
	r.cmds = make(map[string]*gocmd.Cmd)

	if _, err := os.Stat(r.logsDir); os.IsNotExist(err) {
		log.Info("Creating logs directory", "dir", r.logsDir)
		if err := os.MkdirAll(r.logsDir, 0777); err != nil {
			return errors.Wrapf(err, "Failed to create logs directory: %v", r.logsDir)
		}
	}

	if err := r.createCoordinatorLogs(); err != nil {
		return err
	}

	if err := r.createTaskstoreLogs(); err != nil {
		return err
	}

	if err := r.createWorkerLogs(numWorkers); err != nil {
		return err
	}

	if r.port == 0 {
		var err error
		r.port, err = networking.GetFreePort()

		if err != nil {
			return errors.Wrapf(err, "Failed to get free port to run the server on")
		}
	}

	log.Info("Selected port", "port", r.port)

	// Run the taskstore asynchronously because it runs forever
	taskStoreChan := r.startTaskstore()

	// Run the worker asynchronously because it will run forever
	workerChan := make([]<-chan gocmd.Status, 0, numWorkers)
	for i := 0; i < numWorkers; i = i + 1 {
		workerChan = append(workerChan, r.startWorker(i))
	}

	coordChan := r.startCoordinator()

	if err := r.waitForCoordinator(coordChan, taskStoreChan, workerChan); err != nil {
		log.Error(err, "test didn't run successfully")
	}

	log.Info("Cleaning up the e2e tests.")
	// Stop all the subprocesses
	for name, c := range r.cmds {
		// Null op if already stopped
		log.Info("Stopping subprocess", "process", name)
		if err := c.Stop(); err != nil {
			log.Error(err, "Error stopping subprocess", "process", name)
		}
	}
	return nil
}

func (r *Runner) closeFiles() {
	for _, f := range r.logFiles {
		log.Info("Closing log file", "file", f.Name())
		if err := f.Close(); err != nil {
			log.Error(err, "Error closing file", "file", f.Name())
		}
	}
}

// waitForCoordinator waits for the coordinator to finish.
func (r *Runner) waitForCoordinator(coordChan <-chan gocmd.Status, taskStoreChan <-chan gocmd.Status, workerChan []<-chan gocmd.Status) error {
	// We need to wait for the coordinator to complete. If the workers or taskstore exit first that's a problem.
	log.Info("Waiting for coordinator to finish")
	for {
		// Check if the coordinator is done
		select {
		case status := <-coordChan:
			if status.Exit != 0 {
				return errors.Errorf("Coordinator exited with non-zero exit code")
			}
			return nil
		default:
			// Do Nothing. The default clause prevents the channel from blocking
		}

		select {
		case status := <-taskStoreChan:
			return errors.Errorf("taskstore exited with code %v before coordinator finished", status.Exit)
		default:
			// Do Nothing. The default clause prevents the channel from blocking
		}

		// loop over worker processes
		for i, c := range workerChan {
			select {
			case status := <-c:
				return errors.Errorf("worker %v exited with code %v before coordinator finished", i, status.Exit)
			default:
				// Do Nothing. The default clause prevents the channel from blocking
			}
		}
	}
}

func (r *Runner) createTaskstoreLogs() error {
	var err error
	var f *os.File
	taskPath := path.Join(r.logsDir, "taskstore.logs")
	f, err = os.Create(taskPath)
	if err != nil {
		return errors.Wrapf(err, "Failed to create file %v", taskPath)
	}
	r.logFiles["taskstore"] = f
	return nil
}

func (r *Runner) createCoordinatorLogs() error {
	var err error
	var f *os.File
	path := path.Join(r.logsDir, "coordinator.logs")
	f, err = os.Create(path)
	if err != nil {
		return errors.Wrapf(err, "Failed to create file %v", path)
	}
	r.logFiles["coordinator"] = f
	return nil
}

func (r *Runner) createWorkerLogs(numWorkers int) error {
	var err error
	var f *os.File
	for i := 0; i < numWorkers; i += 1 {
		path := path.Join(r.logsDir, workerKey(i)+".logs")
		f, err = os.Create(path)
		if err != nil {
			return errors.Wrapf(err, "Failed to create file %v", path)
		}
		r.logFiles[workerKey(i)] = f
	}
	return nil
}

func workerKey(index int) string {
	return fmt.Sprintf("worker-%v", index)
}

// startTaskstore starts the taskstore
func (r *Runner) startTaskstore() <-chan gocmd.Status {
	f := r.logFiles["taskstore"]
	log.Info("Starting taskstore server", "logsFile", f.Name(), "port", r.port)
	opts := gocmd.Options{
		Streaming: true,
	}
	cmd := gocmd.NewCmdOptions(opts, r.taskStore, fmt.Sprintf("--port=%v", r.port), fmt.Sprintf("--json-logs=%v", r.jsonLog), "--level="+r.level)
	r.cmds["taskstore"] = cmd
	return r.h.RunStreaming(cmd, f)
}

// startWorker starts the TFF worker that will claim tasks.
func (r *Runner) startWorker(index int) <-chan gocmd.Status {
	f := r.logFiles[workerKey(index)]
	log.Info("Starting TFF worker", "index", index, "logsFile", f.Name())

	opts := gocmd.Options{
		Streaming: true,
	}
	cmd := gocmd.NewCmdOptions(opts, "python3", "-m", "flaap.tff.task_handler", "run", fmt.Sprintf("--taskstore=localhost:%v", r.port))
	cmd.Env = []string{"PYTHONPATH=" + os.Getenv("PYTHONPATH")}
	r.cmds[fmt.Sprintf("worker-%v", index)] = cmd
	return r.h.RunStreaming(cmd, f)
}

// startCoordinator starts the coordinator
func (r *Runner) startCoordinator() <-chan gocmd.Status {
	f := r.logFiles["coordinator"]
	opts := gocmd.Options{
		Streaming: true,
	}
	cmd := gocmd.NewCmdOptions(opts, "python3", "-m", "flaap.testing.fed_average", "run", fmt.Sprintf("--port=localhost:%v", r.port))
	cmd.Env = []string{"PYTHONPATH=" + os.Getenv("PYTHONPATH")}
	r.cmds["coordinator"] = cmd
	log.Info("Starting coordinator program", "logsFile", f.Name())
	return r.h.RunStreaming(cmd, f)
}

func main() {
	rootCmd := newRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Command failed with error: %+v", err)
		os.Exit(1)
	}
}

var (
	log logr.Logger
)
