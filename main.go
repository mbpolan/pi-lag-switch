package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"

	rice "github.com/GeertJohan/go.rice"
)

var speedLimits map[string]string = map[string]string{
	"saucy":     "96Kbit",
	"jittery":   "119Kbit",
	"meh":       "200Kbit",
	"unlimited": "400Mbit",
}

var limitSpeeds map[string]string
var networkInterface string

type ErrorModel struct {
	Error string `json:"error"`
}

type LagModel struct {
	Speed string `json:"speed"`
}

func httpError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)

	body := ErrorModel{
		Error: message,
	}

	json.NewEncoder(w).Encode(body)
}

func getSpeed() (string, error) {
	args := []string{"class", "show", "dev", networkInterface}

	cmd := exec.Command("tc", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error: %s (%d)", output, err)
		return "", err
	}

	parts := strings.Split(string(output), " ")
	return parts[7], nil
}

func applySpeed(speed string) error {
	args := []string{"tc", "class", "replace", "dev", networkInterface, "root", "classid", "1:10", "htb", "rate", speed}

	cmd := exec.Command("sudo", args...)
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	return nil
}

func handleGetLag(w http.ResponseWriter, r *http.Request) {
	speed, err := getSpeed()
	if err != nil {
		httpError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	speedValue := ""
	value, ok := limitSpeeds[strings.ToLower(speed)]
	if ok {
		speedValue = value
	}

	model := LagModel{
		Speed: speedValue,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model)
}

func handleUpdateLag(w http.ResponseWriter, r *http.Request) {
	var model LagModel

	err := json.NewDecoder(r.Body).Decode(&model)
	if err != nil {
		httpError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	speed, ok := speedLimits[strings.ToLower(model.Speed)]
	if !ok {
		httpError(w, "Unrecognized speed category", http.StatusBadRequest)
		return
	}

	err = applySpeed(speed)
	if err != nil {
		httpError(w, "Failed to set new speed", http.StatusBadRequest)
		return
	}

	handleGetLag(w, r)
}

func handleLag(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handleGetLag(w, r)
	case "POST":
		handleUpdateLag(w, r)
	default:
		httpError(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func main() {
	hostPtr := flag.String("host", "", "the host to listen on")
	portPtr := flag.Int("port", 8080, "the port to listen on")
	interfacePtr := flag.String("interface", "wlan0", "the network interface to apply rules to")
	flag.Parse()

	networkInterface = *interfacePtr

	limitSpeeds = make(map[string]string, len(speedLimits))
	for k, v := range speedLimits {
		limitSpeeds[strings.ToLower(v)] = k
	}

	http.Handle("/", http.FileServer(rice.MustFindBox("website").HTTPBox()))
	http.HandleFunc("/api/lag", handleLag)

	addr := fmt.Sprintf("%s:%d", *hostPtr, *portPtr)
	log.Printf("Starting server on %s", addr)

	log.Fatal(http.ListenAndServe(addr, nil))
}
