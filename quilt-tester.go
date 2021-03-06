package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"

	"github.com/quilt/quilt/api"
	"github.com/quilt/quilt/api/client/getter"
	apiUtil "github.com/quilt/quilt/api/util"
	"github.com/quilt/quilt/stitch"
	"github.com/quilt/quilt/util"
)

var (
	infrastructureBlueprint = "./config/infrastructure-runner.js"
)

// The global logger for this CI run.
var log logger

func main() {
	namespace := os.Getenv("TESTING_NAMESPACE")
	if namespace == "" {
		logrus.Error("Please set TESTING_NAMESPACE.")
		os.Exit(1)
	}

	var err error
	if log, err = newLogger(); err != nil {
		logrus.WithError(err).Error("Failed to create logger.")
		os.Exit(1)
	}

	tester, err := newTester(namespace)
	if err != nil {
		logrus.WithError(err).Error("Failed to create tester instance.")
		os.Exit(1)
	}

	if err := tester.run(); err != nil {
		logrus.WithError(err).Error("Test execution failed.")
		os.Exit(1)
	}
}

type tester struct {
	preserveFailed bool
	junitOut       string
	toRun          map[string]struct{}

	testSuites  []*testSuite
	initialized bool
	namespace   string
}

func newTester(namespace string) (tester, error) {
	t := tester{
		namespace: namespace,
		toRun:     map[string]struct{}{},
	}

	testRoot := flag.String("testRoot", "",
		"the root directory containing the integration tests")
	flag.BoolVar(&t.preserveFailed, "preserve-failed", false,
		"don't destroy machines on failed tests")
	flag.StringVar(&t.junitOut, "junitOut", "",
		"location to write junit report")
	flag.Var(stringSet(t.toRun), "run", "test to run")
	flag.Parse()

	if *testRoot == "" {
		return tester{}, errors.New("testRoot is required")
	}

	err := t.generateTestSuites(*testRoot)
	if err != nil {
		return tester{}, err
	}

	return t, nil
}

func (t *tester) generateTestSuites(testRoot string) error {
	l := log.testerLogger

	// First, we need to ls the testRoot, and find all of the folders. Then we can
	// generate a testSuite for each folder.
	testRootFiles, err := filepath.Glob(filepath.Join(testRoot, "*"))
	if err != nil {
		l.infoln("Could not access test suite folders")
		l.errorln(err.Error())
		return err
	}

	var testSuiteFolders []string
	for _, f := range testRootFiles {
		stat, err := os.Stat(f)
		if err != nil {
			l.infoln(fmt.Sprintf("Failed to stat potential test suite %s. "+
				"Ignoring.", f))
			l.errorln(err.Error())
			continue
		}

		if stat.IsDir() {
			testSuiteFolders = append(testSuiteFolders, f)
		}
	}

	sort.Sort(byPriorityPrefix(testSuiteFolders))
	for _, testSuiteFolder := range testSuiteFolders {
		files, err := ioutil.ReadDir(testSuiteFolder)
		if err != nil {
			l.infoln(fmt.Sprintf(
				"Error reading test suite %s", testSuiteFolder))
			l.errorln(err.Error())
			return err
		}

		var blueprint, test string
		for _, file := range files {
			path := filepath.Join(testSuiteFolder, file.Name())
			switch {
			case strings.HasSuffix(file.Name(), ".js"):
				blueprint, err = setupBlueprintSandbox(path, t.namespace)
				if err != nil {
					l.infoln(fmt.Sprintf(
						"Error updating namespace for %s.", blueprint))
					l.errorln(err.Error())
					return err
				}
			// If the file is executable by everyone, and is not a directory.
			case (file.Mode()&1 != 0) && !file.IsDir():
				test = path
			}
		}
		newSuite := testSuite{
			name:      filepath.Base(testSuiteFolder),
			blueprint: blueprint,
			test:      test,
		}
		t.testSuites = append(t.testSuites, &newSuite)
	}

	return nil
}

func (t tester) run() error {
	defer func() {
		if t.junitOut != "" {
			writeJUnitReport(t.testSuites, t.junitOut)
		}

		failed := false
		for _, suite := range t.testSuites {
			if suite.result == testFailed {
				failed = true
				break
			}
		}

		if failed && t.preserveFailed {
			return
		}

		cleanupMachines(t.namespace)
	}()

	if err := t.setup(); err != nil {
		log.testerLogger.errorln("Unable to setup the tests, bailing.")
		// All suites failed if we didn't run them.
		for _, suite := range t.testSuites {
			suite.result = testFailed
		}
		return err
	}

	return t.runTestSuites()
}

func (t *tester) setup() error {
	l := log.testerLogger

	l.infoln("Starting the Quilt daemon.")
	go runQuiltDaemon()

	// Do a preliminary quilt stop.
	l.infoln(fmt.Sprintf("Preliminary `quilt stop %s`", t.namespace))
	_, _, err := stop(t.namespace)
	if err != nil {
		l.infoln(fmt.Sprintf("Error stopping: %s", err.Error()))
		return err
	}

	// Setup infrastructure.
	l.infoln("Booting the machines the test suites will run on, and waiting " +
		"for them to connect back.")
	infrastructureBlueprint, err = setupBlueprintSandbox(infrastructureBlueprint, t.namespace)
	if err != nil {
		l.infoln(fmt.Sprintf("Error updating namespace for %s.",
			infrastructureBlueprint))
		l.errorln(err.Error())
		return err
	}
	contents, _ := fileContents(infrastructureBlueprint)
	l.infoln("Begin " + filepath.Base(infrastructureBlueprint))
	l.println(contents)
	l.infoln("End " + filepath.Base(infrastructureBlueprint))

	_, _, err = runBlueprintUntilConnected(infrastructureBlueprint)
	if err != nil {
		l.infoln("Failed to setup infrastructure")
		l.errorln(err.Error())
		return err
	}

	l.infoln("Booted Quilt")
	l.infoln("Machines")
	machines, _ := queryMachines()
	l.println(fmt.Sprintf("%v", machines))

	return nil
}

func (t tester) runTestSuites() error {
	var err error
	for _, suite := range t.testSuites {
		_, explicitRun := t.toRun[suite.name]
		if !explicitRun && len(t.toRun) != 0 {
			suite.result = testSkipped
			continue
		}

		if e := suite.run(); e != nil && err == nil {
			err = e
		}
	}
	return err
}

type testSuite struct {
	name      string
	blueprint string
	test      string

	output      string
	result      testResult
	timeElapsed time.Duration
}

func (ts *testSuite) run() error {
	testStart := time.Now()
	l := log.testerLogger

	defer func() {
		ts.timeElapsed = time.Since(testStart)
	}()
	defer func() {
		logsPath := filepath.Join(os.Getenv("WORKSPACE"), ts.name+"_debug_logs")
		cmd := exec.Command("quilt", "debug-logs", "-tar=false", "-o="+logsPath, "-all")
		stdout, stderr, err := execCmd(cmd, "DEBUG LOGS")
		if err != nil {
			l.errorln(fmt.Sprintf("Debug logs encountered an error:"+
				" %v\nstdout: %s\nstderr: %s", err, stdout, stderr))
		}
	}()

	l.infoln(fmt.Sprintf("Test Suite: %s", ts.name))
	l.infoln("Start " + ts.name + ".js")
	contents, _ := fileContents(ts.blueprint)
	l.println(contents)
	l.infoln("End " + ts.name + ".js")
	defer l.infoln(fmt.Sprintf("Finished Test Suite: %s", ts.name))

	runBlueprint(ts.blueprint)

	l.infoln("Waiting for containers to start up")
	if err := waitForContainers(ts.blueprint); err != nil {
		l.println(".. Containers never started: " + err.Error())
		ts.result = testFailed
		return err
	}

	// Wait a little bit longer for any container bootstrapping after boot.
	time.Sleep(90 * time.Second)

	var err error
	if ts.test != "" {
		l.infoln("Starting Test")
		l.println(".. " + filepath.Base(ts.test))

		ts.output, err = runTest(ts.test)
		if err == nil {
			l.println(".... Passed")
			ts.result = testPassed
		} else {
			l.println(".... Failed")
			ts.result = testFailed
		}
	}

	return err
}

func waitForContainers(blueprintPath string) error {
	stc, err := stitch.FromFile(blueprintPath)
	if err != nil {
		return err
	}

	localClient, err := getter.New().Client(api.DefaultSocket)
	if err != nil {
		return err
	}

	return util.WaitFor(func() bool {
		for _, exp := range stc.Containers {
			containerClient, err := getter.New().ContainerClient(localClient,
				exp.ID)
			if err != nil {
				return false
			}

			actual, err := apiUtil.GetContainer(containerClient, exp.ID)
			if err != nil || actual.Created.IsZero() {
				return false
			}
		}
		return true
	}, 15*time.Second, 10*time.Minute)
}

func runTest(testPath string) (string, error) {
	output, err := exec.Command(testPath).CombinedOutput()
	if err != nil || !strings.Contains(string(output), "PASSED") {
		_, testName := filepath.Split(testPath)
		err = fmt.Errorf("test failed: %s", testName)
	}
	return string(output), err
}
