package main

import (
	"strings"

	"regexp"

	"log"

	"os/exec"

	"os"

	"github.com/blang/semver"
	"github.com/mmcdole/gofeed"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
)

type info struct {
	lastChange map[string]string
	lastTags   map[string]semver.Version
}

const githubRepoOwner = "Fankserver"
const githubRepoName = "docker-prosody"

var i *info

func main() {
	i = &info{
		lastChange: map[string]string{},
		lastTags:   map[string]semver.Version{},
	}
	check("0.9")
	check("0.10")
	c := cron.New()
	c.AddFunc("@every 5m", func() { check("0.9") })
	c.AddFunc("@every 5m", func() { check("0.10") })
	c.Run()
}

func check(version string) {
	skipChange := false
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL("https://hg.prosody.im/" + version + "/rss-log")
	for _, item := range feed.Items {
		c, err := parseContent(item)
		if err != nil {
			return
		}

		if strings.Contains(c.Files, ".hgtags") {
			log.Println(version, "Tag found:", item.Title)
			re, err := regexp.Compile(`(` + version + `\.\d+)`)
			if err != nil {
				logrus.Error(err)
				continue
			}
			result := re.FindStringSubmatch(item.Title)
			if len(result) != 2 {
				logrus.Error("Did not found tag regex")
				continue
			}
			ver, err := semver.Make(result[1])
			if err != nil {
				logrus.Error(err)
				continue
			}

			if _, ok := i.lastTags[version]; !ok {
				log.Println("Initial version setup")
				v, _ := semver.Make("0.0.0")
				i.lastTags[version] = v
			}

			if !ver.GT(i.lastTags[version]) {
				continue
			}

			versionString := ver.String()
			err = createGitTag(item, versionString)
			if err != nil {
				logrus.Error(err)
			}

			i.lastTags[version] = ver
		} else {
			if skipChange {
				continue
			}

			if _, ok := i.lastChange[version]; !ok {
				log.Println("Initial change setup")
				i.lastChange[version] = item.GUID
			}

			if i.lastChange[version] != item.GUID {
				log.Println("Change")

				cmd := exec.Command(`curl`, `-H`, `Content-Type: application/json`, `--data`, `{"docker_tag": "`+version+`-dev"}`, `-X`, `POST`, `https://registry.hub.docker.com/u/fankserver/prosody/trigger/`+os.Getenv("DOCKER_KEY")+`/`)
				stdoutStderr, err := cmd.CombinedOutput()
				log.Println(string(stdoutStderr))
				if err != nil {
					log.Fatal(err)
				}
				i.lastChange[version] = item.GUID
			}
			skipChange = true
		}
	}
}
