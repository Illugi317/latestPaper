package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func main() {
	println("Deleting fetched directory (if it exists)")
	os.RemoveAll("fetched")

	println("Creating fetched directory")
	os.Mkdir("fetched", os.ModePerm)

	//check if any params have been put
	var check_version bool
	var version string
	var err error
	var url string

	check_v := flag.Bool("newest", false, "Get the newest version - cannot be used with -version")
	version_flag := flag.String("version", "", "Specify version - cannot be used with -newest")
	flag.Parse()

	if (*check_v == true) && (*version_flag == "") {
		check_version = true
	} else if (*check_v == true) && (*version_flag != "") {
		panic("Invalid arguments - Newest version and version flag are being used ")
	} else if (*check_v != true) && (*version_flag != "") {
		version = *version_flag
	}

	if check_version {
		version, err = getNewestVersion()
		if err != nil {
			panic(err)
		}
	}

	url, err = getURL(version)
	if err != nil {
		panic(err)
	}

	DownloadFile("fetched/paper.jar", url)

}

func getNewestVersion() (string, error) {
	fmt.Printf("Fetching latest version ")
	res, err := http.Get("https://papermc.io/api/v2/projects/paper/")
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		panic(errors.New("API status code is not OK - API contact failed"))
	}
	var data map[string]interface{}
	json.NewDecoder(res.Body).Decode(&data)
	builds_dstring := fmt.Sprintf("%v", data["versions"])
	builds_cstring := strings.Trim(builds_dstring, "[ ]")
	builds_list := strings.Fields(builds_cstring)
	version := builds_list[len(builds_list)-1] //find maxima!
	fmt.Printf(version + "\n")
	return version, nil
}

func getURL(version string) (string, error) {
	//find latest version
	fmt.Printf("Fetching latest build of version: ")
	res, err := http.Get("https://papermc.io/api/v2/projects/paper/versions/" + version)

	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", errors.New("StatusCode invalid check if api is down")
	}

	var data map[string]interface{}
	json.NewDecoder(res.Body).Decode(&data)

	builds_dstring := fmt.Sprintf("%v", data["builds"])
	builds_cstring := strings.Trim(builds_dstring, "[ ]")
	builds_list := strings.Fields(builds_cstring)
	build := builds_list[len(builds_list)-1]
	fmt.Printf(build + "\n")
	//https://papermc.io/api/v2/projects/paper/versions/1.16.5/builds/785/downloads/paper-1.16.5-785.jar
	url := "https://papermc.io/api/v2/projects/paper/versions/" + version + "/builds/" + build + "/downloads/" + "paper-" + version + "-" + build + ".jar"
	return url, nil
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string) error {
	println("Downloading latest paper jar")
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
