package main

import (
	"bytes"
	"code.google.com/p/goauth2/oauth"
	"encoding/base64"
	"flag"
	"fmt"
	"github.com/google/go-github/github"
	"log"
	"net/http"
	"os"
	"syscall"
)

const REVISION = "HEAD"

type Config struct {
	Host         string
	Port         string
	Token        string
	Organization string
	Repository   string
}

var config Config

func Daemonize(nochdir, noclose int) int {
	var ret uintptr
	var err syscall.Errno

	ret, _, err = syscall.Syscall(syscall.SYS_FORK, 0, 0, 0)
	if err != 0 {
		log.Fatal("failed to fall fork")
		return -1
	}
	switch ret {
	case 0:
		break
	default:
		os.Exit(0)
	}
	pid, err2 := syscall.Setsid()
	if err2 != nil {
		log.Fatal("failed to call setsid %+v", err2)
	}
	if pid == -1 {
		return -1
	}

	if nochdir == 0 {
		os.Chdir("/")
	}

	syscall.Umask(0)
	if noclose == 0 {
		f, e := os.OpenFile("/dev/null", os.O_RDWR, 0)
		if e == nil {
			fd := int(f.Fd())
			syscall.Dup2(fd, int(os.Stdin.Fd()))
			syscall.Dup2(fd, int(os.Stdout.Fd()))
			syscall.Dup2(fd, int(os.Stderr.Fd()))
		}
	}
	return 0
}

func root(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
}

func handler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	status := r.FormValue("status")
	version := r.FormValue("version")
	php_test_data := r.FormValue("php_test_data")

	if php_test_data != "" {
		if len(php_test_data) > 800000 {
			fmt.Fprint(w, "can't handle input that large.")
			log.Println("can't handle input that large.")
			return
		}
		switch status {
		case "failed":
			status = "failed"
			break
		case "success":
			status = "success"
		default:
			status = "unknown"
		}
		if version == "" {
			version = "unknown"
		}

		decode, _ := base64.StdEncoding.DecodeString(php_test_data)
		body := string(decode)

		title := "[run-test.php report]"
		buffer := bytes.NewBuffer(nil)
		buffer.WriteString(fmt.Sprintf("Status: %s\n", status))
		buffer.WriteString(fmt.Sprintf("PHP Version: %s\n", version))
		buffer.WriteString(fmt.Sprintf("Body: \n```\n%s\n```\n", body))
		s := buffer.String()

		t := &oauth.Transport{
			Token: &oauth.Token{AccessToken: config.Token}}
		client := github.NewClient(t.Client())
		req := &github.IssueRequest{
			Title: &title,
			Body:  &s,
		}

		client.Issues.Create(config.Organization, config.Repository, req)
	}

	fmt.Fprint(w, REVISION)
}

func initializeConfig() {
	config.Host = os.Getenv("HOST")
	config.Port = os.Getenv("PORT")
	config.Token = os.Getenv("TOKEN")
	config.Organization = os.Getenv("ORGANIZATION")
	config.Repository = os.Getenv("REPOSITORY")

	if config.Host == "" {
		config.Host = "127.0.0.1"
	}
	if config.Port == "" {
		config.Port = "9999"
	}
	if config.Organization == "" {
		config.Organization = "php-git-bot"
	}
	if config.Repository == "" {
		config.Repository = "test"
	}
	if config.Token == "" {
		log.Fatal("TOKEN env is required.")
	}
}

func main() {
	foreGround := flag.Bool("foreground", false, "run as foreground")
	flag.Parse()
	initializeConfig()

	if !*foreGround {
		err := Daemonize(0, 0)
		if err != 0 {
			log.Fatal("fork failed")
		}
	}

	http.HandleFunc("/", root)
	http.HandleFunc("/buildtest/process", handler)

	err := http.ListenAndServe(fmt.Sprintf("%s:%s", config.Host, config.Port), nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
