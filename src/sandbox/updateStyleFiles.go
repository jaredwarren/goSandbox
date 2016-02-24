package main

import (
	"encoding/json"
	"filesystem"
	"fmt"
	"io/ioutil"
	"os/exec"
	"regexp"
	"ssh"
	"strings"
)

const (
	CF_ROOT     = "C:/data/CourseFiles/"
	STYLE_ROOT  = "C:/data/CourseFiles/ExtPlayer/Styles/default/"
	GLOBAL_ROOT = "C:/data/CourseFiles/ExtPlayer/Global/"

	SRC_ROOT          = "C:/data/www/sandbox/ExtPlayer/"
	PROD_SRC_ROOT     = "C:/data/www/sandbox/ExtPlayer/build/production/Player/"
	CoursePlayer_ROOT = "C:/data/www/sandbox/CoursePlayer/"

	CF_ORIGIN = "test"
)

var gitStatusClean = regexp.MustCompile(`nothing to commit, working directory clean`)

var commentReg = regexp.MustCompile(`(?m)((?:\/\*(?:[^*]|(?:\*+[^*\/]))*\*+\/)|(?:\/\/.*))`)

func main() {
	// commit src root
	if out, _ := execGit(SRC_ROOT, []string{"status"}); !gitStatusClean.Match(out) {
		execGit(SRC_ROOT, []string{"add", "."})
		execGit(SRC_ROOT, []string{"commit", "-a", "-m", `Build All`})
	}
	// build src
	execSencha(SRC_ROOT, []string{"app", "build"})

	builds := getBuilds()

	// update build files
	updateStyleFiles(builds)

	// update courePlayer
	updateCoursePlayer(builds)

	fmt.Println("Done")
}

func getBuilds() (builds []string) {
	builds = make([]string, 0)

	file, e := ioutil.ReadFile(`C:\data\www\sandbox\ExtPlayer\app.json`)
	if e != nil {
		panic(e)
	}
	// remove comments because the are causing errors
	file = commentReg.ReplaceAll(file, []byte(``))
	// parse file
	appJson := map[string]interface{}{}
	if err := json.Unmarshal(file, &appJson); err != nil {
		panic(err)
	}
	// get build keys
	for key, _ := range appJson["builds"].(map[string]interface{}) {
		builds = append(builds, key)
	}
	return
}

func execGit(root string, args []string) (out []byte, err error) {
	fmt.Println("> git", strings.Join(args, " "))
	return execCmd("C:/Program Files (x86)/Git/bin/git.exe", root, args)
}

var senchaError = regexp.MustCompile(`\[ERR\]`)

func execSencha(root string, args []string) (out []byte, err error) {
	fmt.Println("> sencha", strings.Join(args, " "))
	out, err = execCmd("C:/Users/jaredwarren/bin/Sencha/Cmd/6.0.1.76/sencha.exe", root, args)
	if senchaError.Match(out) {
		panic("Sencha Error!")
	}
	return
}

func execCmd(app string, root string, args []string) (out []byte, err error) {
	cmd := exec.Command(app, args...)
	if root != "" {
		cmd.Dir = root
	}
	out, err = cmd.Output()
	if err != nil {
		fmt.Println("OUT:", string(out))
		panic(err)
	}
	return
}

// LA style files
var styleFiles = []string{
	"lang",
	"resources",
	"cache.appcache",
	"index.html",
	"indexMobile.html",
	"resources",
	"resources",
}

func updateStyleFiles(builds []string) {
	// US and GB are the same for now
	filesystem.CopyItem(SRC_ROOT+"lang/en-US.js", SRC_ROOT+"lang/en-GB.js")
	filesystem.CopyDir(SRC_ROOT+"lang", PROD_SRC_ROOT+"lang")

	// Update Style Files
	for _, build := range builds {
		filesystem.CopyDir(PROD_SRC_ROOT+build, STYLE_ROOT+build)
		filesystem.CopyFile(PROD_SRC_ROOT+build+".json", STYLE_ROOT+build+".json")
	}
	for _, build := range styleFiles {
		filesystem.CopyItem(PROD_SRC_ROOT+build, STYLE_ROOT+build)
	}

	// Rename index to start make SCORM happy
	filesystem.CopyFile(PROD_SRC_ROOT+"index.html", STYLE_ROOT+"start.html")

	// check status
	if out, _ := execGit(CF_ROOT, []string{"status"}); gitStatusClean.Match(out) {
		panic("Clean!")
	}

	// add, commit and push changes
	execGit(CF_ROOT, []string{"add", "."})
	execGit(CF_ROOT, []string{"commit", "-a", "-m", `Build All`})
	execGit(CF_ROOT, []string{"push", CF_ORIGIN, "master"})

	// ssh
	ssh.ExecSSH("reset_cf" + CF_ORIGIN)
}

// CoursePlayer
var coursePlayerFiles = []string{
	"app",
	"build",
	"classic",
	"ext",
	"lang",
	"modern",
	"overrides",
	"packages",
	"resources",
	"sass",

	".gitignore",
	"app.js",
	"app.json",
	"bootstrap.css",
	"bootstrap.js",
	"build.xml",
	"consolePolyfill.js",
	"index.html",
	"loading.html",
	"mathjax-config.js",
	"workspace.json",
}

func updateCoursePlayer(builds []string) {
	for _, build := range coursePlayerFiles {
		filesystem.CopyItem(SRC_ROOT+build, CoursePlayer_ROOT+build)
	}
	for _, build := range builds {
		filesystem.CopyFile(SRC_ROOT+build+".json", CoursePlayer_ROOT+build+".json")
	}

	// TODO: eventually I should commit and push these changes
}
