package main

import (
	"bufio"
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
	os.Chdir("fetched")
	os.Mkdir("plugins", os.ModePerm)
	os.Chdir("..")

	//check if any params have been put
	//var check_version bool
	var download_plugins bool
	var plugin_file string
	var version string
	//var err error
	var url string

	check_v := flag.Bool("newest", false, "Get the newest version - cannot be used with -version")
	version_flag := flag.String("version", "", "Specify version - cannot be used with -newest")
	plugin_flag := flag.String("plugin", "", "file that is a list of spigot plugins to be downloaded | check the docs online for more details on how to structure the .txt file :)")
	flag.Parse()

	if *plugin_flag != "" {
		if _, err := os.Stat(*plugin_flag); err == nil {
			// file exists
			download_plugins = true
			plugin_file = *plugin_flag
		} else if os.IsNotExist(err) {
			// file does not exists
			println("File does not exist ?")
			panic(err)
		} else {
			// file may or may not exist throw panic instead
			panic(err)
		}
	} else {
		download_plugins = false
	}

	if (*check_v == true) && (*version_flag != "") {
		panic("Invalid arguments - Newest version and version flag are being used ")
	} else if (*check_v != true) && (*version_flag != "") {
		version = *version_flag
	}
	url = getCorrectUrl(version)
	// correct url always. good
	fmt.Println(url)
	DownloadFile("fetched/paper.jar", url) //download the paper.jar and place it in the fetched folder
	if download_plugins == true && plugin_file != "" {
		DownloadPlugins(plugin_file)
	}
	//if download_plugins {
	//	DownloadPlugins(plugin_file)
	//}

	//download plugins
}

func getCorrectUrl(version string) string {
	/*
		Gets the correct url to download the paper.jar with the correct version and newest build for that version
		might to be able to use structs to accept the json data but for now i just mushed the two functions into one :)
	*/
	var input_version = version
	if input_version == "" {
		//get version and then continue.
		res, err := http.Get("https://papermc.io/api/v2/projects/paper/")
		if err != nil {
			panic(err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			panic(errors.New("API status code is not OK - API contact failed"))
		}
		var data map[string]interface{}                                   //Wizard stuff
		json.NewDecoder(res.Body).Decode(&data)                           //^^
		version_dirty_string := fmt.Sprintf("%v", data["versions"])       //No clue what's happening here
		version_clean_string := strings.Trim(version_dirty_string, "[ ]") // remove the [] from the string so i can use the strings.fields to get a slice/array
		version_list := strings.Fields(version_clean_string)
		latest_version := version_list[len(version_list)-1] // Get the latest version / get the last thing in that slice
		input_version = latest_version
		fmt.Println("Latest version is: " + input_version)
	}
	res, err := http.Get("https://papermc.io/api/v2/projects/paper/versions/" + input_version)

	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		panic("StatusCode not OK - Check if api down")
	}

	var data map[string]interface{}
	json.NewDecoder(res.Body).Decode(&data)

	builds_dstring := fmt.Sprintf("%v", data["builds"])
	builds_cstring := strings.Trim(builds_dstring, "[ ]")
	builds_list := strings.Fields(builds_cstring)
	build := builds_list[len(builds_list)-1]
	fmt.Println("Latest build of version " + input_version + " is :" + build)
	//https://papermc.io/api/v2/projects/paper/versions/1.16.5/builds/785/downloads/paper-1.16.5-785.jar
	url := "https://papermc.io/api/v2/projects/paper/versions/" + input_version + "/builds/" + build + "/downloads/" + "paper-" + input_version + "-" + build + ".jar"
	return url
}

type id_name struct {
	id   string
	name string
}

func DownloadPlugins(filepath string) {

	file, _ := os.Open(filepath)
	scanner := bufio.NewScanner(file)
	id_name_array := []id_name{}
	for scanner.Scan() {
		var id string
		var name string
		line := scanner.Text()
		line_slice := strings.Split(line, " ")
		if len(line_slice) > 1 {
			id = line_slice[0]
			name = line_slice[1]
			if strings.HasSuffix(name, ".jar") == false {
				name = name + ".jar"
			}
		} else {
			id = line_slice[0]
			name = line_slice[0] + ".jar"
		}
		id_name_array = append(id_name_array, id_name{id, name})
		fmt.Println(id_name_array)
	}
	fmt.Println("done")
	for _, id_name := range id_name_array { //download loop
		fmt.Println(id_name.id + " " + id_name.name)
		// https://api.spiget.org/v2/resources/31822/download
		download_url := "http://api.spiget.org/v2/resources/" + id_name.id + "/download"
		fmt.Println(download_url)
		DownloadFile("fetched/plugins/"+id_name.name, download_url)
	}

}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string) error {
	// Get the data
	client := &http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	request.Header.Set("User-Agent", "XGET/0.7") //wtf
	resp, err := client.Do(request)
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
