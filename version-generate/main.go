package main

import (
	"log"
	"os"
	"os/exec"
	"strings"
)

var (
	versionFilename = "version.go"
)

func main() {
	// versionData, err := ioutil.ReadFile("../version.go")
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// re := regexp.MustCompile(".*majorVersion = \"(.*)\".*")
	// majorVersion := re.FindAllStringSubmatch(string(versionData), -1)[0][1]
	// fmt.Println(majorVersion)
	// cmd := exec.Command("git", "rev-list", majorVersion+"..HEAD", "--count")
	cmd := exec.Command("git", "describe", "--tags")
	out, err := cmd.CombinedOutput()
	cmd.Wait()
	if err != nil {
		log.Fatalln(err)
	}

	versionFile, err := os.Create(versionFilename)
	if err != nil {
		log.Fatalln(err)
	}
	defer versionFile.Close()

	versionFile.Write([]byte("package main\n\nconst Version = \""))
	versionFile.Write([]byte(strings.TrimSpace(string(out)) + "\"\n"))

}
