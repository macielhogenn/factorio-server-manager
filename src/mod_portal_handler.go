package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func FactorioModPortalListModsHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var resp interface{}

	defer func() {
		WriteResponse(w, resp)
	}()

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")

	var statusCode int
	resp, err, statusCode = modPortalList()

	if err != nil {
		resp = fmt.Sprintf("Error in listing mods from mod portal: %s", err)
		log.Println(resp)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(statusCode)
}

// FactorioModPortalModInfoHandler returns JSON response with the mod details
func FactorioModPortalModInfoHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var resp interface{}

	defer func() {
		WriteResponse(w, resp)
	}()

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")

	vars := mux.Vars(r)
	modId := vars["mod"]

	var statusCode int
	resp, err, statusCode = modPortalModDetails(modId)

	if err != nil {
		resp = fmt.Sprintf("Error in getting mod details from mod portal: %s", err)
		log.Println(resp)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(statusCode)
}

func FactorioModPortalInstallHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var resp interface{}

	defer func() {
		WriteResponse(w, resp)
	}()

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")

	body, err := ReadRequestBody(w, r, &resp)
	if err != nil {
		return
	}

	// Get Data out of the request
	var data struct {
		DownloadURL string `json:"link"`
		Filename    string `json:"filename"`
		ModName     string `json:"modName"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		resp = fmt.Sprintf("Error reading data from request {%s}: %s", r.RequestURI, err)
		log.Println(resp)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	mods, err := CreateNewMods(w, &resp)
	if err != nil {
		return
	}

	err = mods.downloadMod(data.DownloadURL, data.Filename, data.ModName)
	if err != nil {
		resp = fmt.Sprintf("Error downloading a mod: %s", err)
		log.Println(resp)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp = mods.listInstalledMods()
}

func FactorioModPortalLoginHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var resp interface{}

	defer func() {
		WriteResponse(w, resp)
	}()

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")

	body, err := ReadRequestBody(w, r, &resp)
	if err != nil {
		return
	}

	var data struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		resp = fmt.Sprintf("Error reading data from request {%s}: %s", r.RequestURI, err)
		log.Println(resp)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	loginStatus, err, statusCode := factorioLogin(data.Username, data.Password)
	if err != nil {
		resp = fmt.Sprintf("Error trying to login into Factorio: %s", err)
		log.Println(resp)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if loginStatus == "" {
		resp = true
	}

	w.WriteHeader(statusCode)
}

func FactorioModPortalLoginStatusHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var resp interface{}

	defer func() {
		WriteResponse(w, resp)
	}()

	var credentials FactorioCredentials
	resp, err = credentials.load()

	if err != nil {
		resp = fmt.Sprintf("Error getting the factorio credentials: %s", err)
		log.Println(resp)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func FactorioModPortalLogoutHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var resp interface{}

	defer func() {
		WriteResponse(w, resp)
	}()

	var credentials FactorioCredentials
	err = credentials.del()

	if err != nil {
		resp = fmt.Sprintf("Error on logging out of factorio: %s", err)
		log.Println(resp)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp = false
}

func ModPortalInstallMultipleHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var resp interface{}

	defer func() {
		WriteResponse(w, resp)
	}()

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")

	body, err := ReadRequestBody(w, r, &resp)
	if err != nil {
		return
	}

	var data []struct {
		Name    string  `json:"name"`
		Version Version `json:"version"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		resp = fmt.Sprintf("Error reading data from request {%s}: %s", r.RequestURI, err)
		log.Println(resp)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	mods, err := CreateNewMods(w, &resp)
	if err != nil {
		return
	}

	for _, datum := range data {
		details, err, statusCode := modPortalModDetails(datum.Name)
		if err != nil || statusCode != http.StatusOK {
			resp = fmt.Sprintf("Error in getting mod details from mod portal: %s", err)
			log.Println(resp)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		//find correct mod-version
		for _, release := range details.Releases {
			if release.Version.Equals(datum.Version) {
				err := mods.downloadMod(release.DownloadURL, release.FileName, details.Name)
				if err != nil {
					resp = fmt.Sprintf("Error downloading mod {%s}, error: %s", details.Name, err)
					log.Println(resp)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				break
			}
		}
	}

	resp = mods.listInstalledMods()
}
