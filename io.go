package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/afero"
)

// appFs is an aero filesystem.  It is stored in a variable so that we can replace it
// with in-memory filesystems for unit tests.
var appFs = afero.NewOsFs()

type logger struct {
	cmdLogger    fileLogger
	testerLogger fileLogger
}

// Create a new logger that will log in the proper directory.
// Also initializes all necessary directories and files.
func newLogger() (logger, error) {
	cmdLoggerPath := filepath.Join(os.Getenv("WORKSPACE"), "commandOutputs.log")
	cmdLoggerFile, err := os.Create(cmdLoggerPath)
	if err != nil {
		return logger{}, err
	}

	return logger{
		testerLogger: fileLogger{os.Stdout},
		cmdLogger:    fileLogger{cmdLoggerFile},
	}, nil
}

type fileLogger struct {
	out io.Writer
}

func (l fileLogger) infoln(msg string) {
	timestamp := time.Now().Format("[15:04:05] ")
	l.println("\n" + timestamp + "=== " + msg + " ===")
}

func (l fileLogger) errorln(msg string) {
	l.println("\n=== Error Text ===\n" + msg + "\n")
}

func (l fileLogger) println(msg string) {
	fmt.Fprintln(l.out, msg)
}

func overwrite(file string, message string) error {
	a := afero.Afero{
		Fs: appFs,
	}
	return a.WriteFile(file, []byte(message), 0666)
}

func fileContents(file string) (string, error) {
	a := afero.Afero{
		Fs: appFs,
	}
	contents, err := a.ReadFile(file)
	if err != nil {
		return "", err
	}
	return string(contents), nil
}

// Update the given blueprint to have the given namespace, and its specified dependencies.
func setupBlueprintSandbox(blueprintFile, namespace string) (string, error) {
	outDir, err := ioutil.TempDir("", filepath.Base(blueprintFile))
	if err != nil {
		return "", err
	}

	if err := setupSandboxDependencies(blueprintFile, outDir); err != nil {
		return "", err
	}

	blueprintContents, err := fileContents(blueprintFile)
	if err != nil {
		return "", err
	}

	// Set the namespace of the global deployment to be `namespace`.
	updatedBlueprint := blueprintContents +
		fmt.Sprintf("; require('@quilt/quilt').getDeployment().namespace = %q;", namespace)

	newBlueprint := filepath.Join(outDir, filepath.Base(blueprintFile))
	return newBlueprint, overwrite(newBlueprint, updatedBlueprint)
}

func setupSandboxDependencies(blueprintFile, sandbox string) error {
	pkgFile := filepath.Join(filepath.Dir(blueprintFile), "package.json")
	if _, err := os.Stat(pkgFile); os.IsNotExist(err) {
		return nil
	}

	err := copyFile(pkgFile, filepath.Join(sandbox, "package.json"))
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	os.Chdir(sandbox)
	defer os.Chdir(cwd)
	_, _, err = npmInstall()
	return err
}

func copyFile(src, dst string) error {
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
